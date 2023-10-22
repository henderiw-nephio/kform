package markers

import (
	"bytes"
	"fmt"
	"reflect"
	sc "text/scanner"

	"github.com/henderiw-nephio/kform/tools/pkg/loader"
)

// Definition is a parsed definition of a marker.
type Definition struct {
	// Output is the deserialized Go type of the marker.
	Output reflect.Type
	// Name is the marker's name.
	Name string
	// Target indicates which kind of node this marker can be associated with.
	Target TargetType
	// Fields lists out the types of each field that this marker has, by
	// argument name as used in the marker (if the output type isn't a struct,
	// it'll have a single, blank field name).  This only lists exported fields,
	// (as per reflection rules).
	Fields map[string]Argument
	// FieldNames maps argument names (as used in the marker) to struct field name
	// in the output type.
	FieldNames map[string]string
	// Strict indicates that this definition should error out when parsing if
	// not all non-optional fields were seen.
	Strict bool
}

// TargetType describes which kind of node a given marker is associated with.
type TargetType int

const (
	// DescribesPackage indicates that a marker is associated with a package.
	DescribesPackage TargetType = iota
	// DescribesType indicates that a marker is associated with a type declaration.
	DescribesType
	// DescribesField indicates that a marker is associated with a struct field.
	DescribesField
)

// loadFields uses reflection to populate argument information from the Output type.
func (d *Definition) loadFields() error {
	if d.Fields == nil {
		d.Fields = make(map[string]Argument)
		d.FieldNames = make(map[string]string)
	}
	if d.Output.Kind() != reflect.Struct {
		// anonymous field type
		argType, err := ArgumentFromType(d.Output)
		if err != nil {
			return err
		}
		d.Fields[""] = argType
		d.FieldNames[""] = ""
		return nil
	}

	for i := 0; i < d.Output.NumField(); i++ {
		field := d.Output.Field(i)
		if field.PkgPath != "" {
			// as per the reflect package docs, pkgpath is empty for exported fields,
			// so non-empty package path means a private field, which we should skip
			continue
		}
		argName, optionalOpt := argumentInfo(field.Name, field.Tag)

		argType, err := ArgumentFromType(field.Type)
		if err != nil {
			return fmt.Errorf("unable to extract type information for field %q: %w", field.Name, err)
		}

		if argType.Type == RawType {
			return fmt.Errorf("RawArguments must be the direct type of a marker, and not a field")
		}

		argType.Optional = optionalOpt || argType.Optional

		d.Fields[argName] = argType
		d.FieldNames[argName] = field.Name
	}

	return nil
}

// MakeDefinition constructs a definition from a name, type, and the output type.
// All such definitions are strict by default.  If a struct is passed as the output
// type, its public fields will automatically be populated into Fields (and similar
// fields in Definition).  Other values will have a single, empty-string-named Fields
// entry.
func MakeDefinition(name string, target TargetType, output interface{}) (*Definition, error) {
	def := &Definition{
		Name:   name,
		Target: target,
		Output: reflect.TypeOf(output),
		Strict: true,
	}

	if err := def.loadFields(); err != nil {
		return nil, err
	}

	return def, nil
}

// Parse uses the type information in this Definition to parse the given
// raw marker in the form `+a:b:c=arg,d=arg` into an output object of the
// type specified in the definition.
func (r *Definition) Parse(rawMarker string) (interface{}, error) {
	name, anonName, fields := splitMarker(rawMarker)
	//fmt.Println("name: ", name)
	//fmt.Println("anonName: ", anonName)
	//fmt.Println("fields: ", fields)

	out := reflect.Indirect(reflect.New(r.Output))

	// if we're a not a struct or have no arguments, treat the full `a:b:c` as the name,
	// otherwise, treat `c` as a field name, and `a:b` as the marker name.
	if !r.AnonymousField() && !r.Empty() && len(anonName) >= len(name)+1 {
		fields = anonName[len(name)+1:] + "=" + fields
	}

	var errs []error
	scanner := parserScanner(fields, func(scanner *sc.Scanner, msg string) {
		errs = append(errs, &ScannerError{Msg: msg, Pos: scanner.Position})
	})

	// TODO(directxman12): strict parsing where we error out if certain fields aren't optional
	seen := make(map[string]struct{}, len(r.Fields))
	if r.AnonymousField() && scanner.Peek() != sc.EOF {
		// might still be a struct that something fiddled with, so double check
		structFieldName := r.FieldNames[""]
		outTarget := out
		if structFieldName != "" {
			// it's a struct field mapped to an anonymous marker
			outTarget = out.FieldByName(structFieldName)
			if !outTarget.CanSet() {
				scanner.Error(scanner, fmt.Sprintf("cannot set field %q (might not exist)", structFieldName))
				return out.Interface(), loader.MaybeErrList(errs)
			}
		}

		// no need for trying to parse field names if we're not a struct
		field := r.Fields[""]
		field.Parse(scanner, fields, outTarget)
		seen[""] = struct{}{} // mark as seen for strict definitions
	} else if !r.Empty() && scanner.Peek() != sc.EOF {
		// if we expect *and* actually have arguments passed
		for {
			// parse the argument name
			if !expect(scanner, sc.Ident, "argument name") {
				break
			}
			argName := scanner.TokenText()
			if !expect(scanner, '=', "equals") {
				break
			}

			// make sure we know the field
			fieldName, known := r.FieldNames[argName]
			if !known {
				scanner.Error(scanner, fmt.Sprintf("unknown argument %q", argName))
				break
			}
			fieldType, known := r.Fields[argName]
			if !known {
				scanner.Error(scanner, fmt.Sprintf("unknown argument %q", argName))
				break
			}
			seen[argName] = struct{}{} // mark as seen for strict definitions

			// parse the field value
			fieldVal := out.FieldByName(fieldName)
			if !fieldVal.CanSet() {
				scanner.Error(scanner, fmt.Sprintf("cannot set field %q (might not exist)", fieldName))
				break
			}
			fieldType.Parse(scanner, fields, fieldVal)

			if len(errs) > 0 {
				break
			}

			if scanner.Peek() == sc.EOF {
				break
			}
			if !expect(scanner, ',', "comma") {
				break
			}
		}
	}

	if tok := scanner.Scan(); tok != sc.EOF {
		scanner.Error(scanner, fmt.Sprintf("extra arguments provided: %q", fields[scanner.Position.Offset:]))
	}

	if r.Strict {
		for argName, arg := range r.Fields {
			if _, wasSeen := seen[argName]; !wasSeen && !arg.Optional {
				scanner.Error(scanner, fmt.Sprintf("missing argument %q", argName))
			}
		}
	}

	return out.Interface(), loader.MaybeErrList(errs)
}

// AnonymousField indicates that the definition has one field,
// (actually the original object), and thus the field
// doesn't get named as part of the name.
func (r *Definition) AnonymousField() bool {
	if len(r.Fields) != 1 {
		return false
	}
	_, hasAnonField := r.Fields[""]
	return hasAnonField
}

// Empty indicates that this definition has no fields.
func (r *Definition) Empty() bool {
	return len(r.Fields) == 0
}

// parserScanner makes a new scanner appropriate for use in parsing definitions and arguments.
func parserScanner(raw string, err func(*sc.Scanner, string)) *sc.Scanner {
	scanner := &sc.Scanner{}
	scanner.Init(bytes.NewBufferString(raw))
	scanner.Mode = sc.ScanIdents | sc.ScanInts | sc.ScanFloats | sc.ScanStrings | sc.ScanRawStrings | sc.SkipComments
	scanner.Error = err

	return scanner
}

type ScannerError struct {
	Msg string
	Pos sc.Position
}

func (e *ScannerError) Error() string {
	return fmt.Sprintf("%s (at %s)", e.Msg, e.Pos)
}

// expect checks that the next token of the scanner is the given token, adding an error
// to the scanner if not.  It returns whether the token was as expected.
func expect(scanner *sc.Scanner, expected rune, errDesc string) bool {
	tok := scanner.Scan()
	if tok != expected {
		scanner.Error(scanner, fmt.Sprintf("expected %s, got %q", errDesc, scanner.TokenText()))
		return false
	}
	return true
}
