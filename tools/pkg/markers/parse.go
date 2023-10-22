package markers

import "strings"

// splitMarker takes a marker in the form of `+a:b:c=arg,d=arg` and splits it
// into the name (`a:b`), the name if it's not a struct (`a:b:c`), and the parts
// that are definitely fields (`arg,d=arg`).
func splitMarker(raw string) (name string, anonymousName string, restFields string) {
	raw = raw[1:] // get rid of the leading '+'
	nameFieldParts := strings.SplitN(raw, "=", 2)
	if len(nameFieldParts) == 1 {
		return nameFieldParts[0], nameFieldParts[0], ""
	}
	anonymousName = nameFieldParts[0]
	name = anonymousName
	restFields = nameFieldParts[1]

	nameParts := strings.Split(name, ":")
	if len(nameParts) > 1 {
		name = strings.Join(nameParts[:len(nameParts)-1], ":")
	}
	return name, anonymousName, restFields
}

// MakeAnyTypeDefinition constructs a definition for an output struct with a
// field named `Value` of type `interface{}`. The argument to the marker will
// be parsed as AnyType and assigned to the field named `Value`.
func MakeAnyTypeDefinition(name string, target TargetType, output interface{}) (*Definition, error) {
	defn, err := MakeDefinition(name, target, output)
	if err != nil {
		return nil, err
	}
	defn.FieldNames = map[string]string{"": "Value"}
	defn.Fields = map[string]Argument{"": defn.Fields["value"]}
	return defn, nil
}
