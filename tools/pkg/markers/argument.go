package markers

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
	sc "text/scanner"
	"unicode"
)

// RawArguments is a special type that can be used for a marker
// to receive *all* raw, underparsed argument data for a marker.
// You probably want to use `interface{}` to match any type instead.
// Use *only* for legacy markers that don't follow Definition's normal
// parsing logic.  It should *not* be used as a field in a marker struct.
type RawArguments []byte

// ArgumentType is the kind of a marker argument type.
// It's roughly analogous to a subset of reflect.Kind, with
// an extra "AnyType" to represent the empty interface.
type ArgumentType int

const (
	// Invalid represents a type that can't be parsed, and should never be used.
	InvalidType ArgumentType = iota
	// IntType is an int
	IntType
	// NumberType is a float64
	NumberType
	// StringType is a string
	StringType
	// BoolType is a bool
	BoolType
	// AnyType is the empty interface, and matches the rest of the content
	AnyType
	// SliceType is any slice constructed of the ArgumentTypes
	SliceType
	// MapType is any map constructed of string keys, and ArgumentType values.
	// Keys are strings, and it's common to see AnyType (non-uniform) values.
	MapType
	// RawType represents content that gets passed directly to the marker
	// without any parsing. It should *only* be used with anonymous markers.
	RawType
)

// Argument is the type of a marker argument.
type Argument struct {
	// Type is the type of this argument For non-scalar types (map and slice),
	// further information is specified in ItemType.
	Type ArgumentType
	// Optional indicates if this argument is optional.
	Optional bool
	// Pointer indicates if this argument was a pointer (this is really only
	// needed for deserialization, and should alway imply optional)
	Pointer bool

	// ItemType is the type of the slice item for slices, and the value type
	// for maps.
	ItemType *Argument
}

// Parse attempts to consume the argument from the given scanner (based on the given
// raw input as well for collecting ranges of content), and places the output value
// in the given reflect.Value.  Errors are reported via the given scanner.
func (a *Argument) Parse(scanner *sc.Scanner, raw string, out reflect.Value) {
	a.parse(scanner, raw, out, false)
}

// parse functions like Parse, except that it allows passing down whether or not we're
// already in a slice, to avoid duplicate legacy slice detection for AnyType
func (a *Argument) parse(scanner *sc.Scanner, raw string, out reflect.Value, inSlice bool) {
	// nolint:gocyclo
	if a.Type == InvalidType {
		scanner.Error(scanner, "cannot parse invalid type")
		return
	}
	if a.Pointer {
		out.Set(reflect.New(out.Type().Elem()))
		out = reflect.Indirect(out)
	}
	switch a.Type {
	case RawType:
		// raw consumes everything else
		castAndSet(out, reflect.ValueOf(raw[scanner.Pos().Offset:]))
		// consume everything else
		var tok rune
		for {
			tok = scanner.Scan()
			if tok == sc.EOF {
				break
			}
		}
	case NumberType:
		nextChar := scanner.Peek()
		isNegative := false
		if nextChar == '-' {
			isNegative = true
			scanner.Scan() // eat the '-'
		}

		tok := scanner.Scan()
		if tok != sc.Float && tok != sc.Int {
			scanner.Error(scanner, fmt.Sprintf("expected integer or float, got %q", scanner.TokenText()))
			return
		}

		text := scanner.TokenText()
		if isNegative {
			text = "-" + text
		}

		val, err := strconv.ParseFloat(text, 64)
		if err != nil {
			scanner.Error(scanner, fmt.Sprintf("unable to parse number: %v", err))
			return
		}

		castAndSet(out, reflect.ValueOf(val))
	case IntType:
		nextChar := scanner.Peek()
		isNegative := false
		if nextChar == '-' {
			isNegative = true
			scanner.Scan() // eat the '-'
		}
		if !expect(scanner, sc.Int, "integer") {
			return
		}
		// TODO(directxman12): respect the size when parsing
		text := scanner.TokenText()
		if isNegative {
			text = "-" + text
		}
		val, err := strconv.Atoi(text)
		if err != nil {
			scanner.Error(scanner, fmt.Sprintf("unable to parse integer: %v", err))
			return
		}
		castAndSet(out, reflect.ValueOf(val))
	case StringType:
		// strings are a bit weird -- the "easy" case is quoted strings (tokenized as strings),
		// the "hard" case (present for backwards compat) is a bare sequence of tokens that aren't
		// a comma.
		a.parseString(scanner, raw, out)
	case BoolType:
		if !expect(scanner, sc.Ident, "true or false") {
			return
		}
		switch scanner.TokenText() {
		case "true":
			castAndSet(out, reflect.ValueOf(true))
		case "false":
			castAndSet(out, reflect.ValueOf(false))
		default:
			scanner.Error(scanner, fmt.Sprintf("expected true or false, got %q", scanner.TokenText()))
			return
		}
	case AnyType:
		guessedType := guessType(scanner, raw, !inSlice)
		newOut := out

		// we need to be able to construct the right element types, below
		// in parse, so construct a concretely-typed value to use as "out"
		switch guessedType.Type {
		case SliceType:
			newType, err := makeSliceType(*guessedType.ItemType)
			if err != nil {
				scanner.Error(scanner, err.Error())
				return
			}
			newOut = reflect.Indirect(reflect.New(newType))
		case MapType:
			newType, err := makeMapType(*guessedType.ItemType)
			if err != nil {
				scanner.Error(scanner, err.Error())
				return
			}
			newOut = reflect.Indirect(reflect.New(newType))
		}
		if !newOut.CanSet() {
			panic("at the disco") // TODO(directxman12): this is left over from debugging -- it might need to be an error
		}
		guessedType.Parse(scanner, raw, newOut)
		castAndSet(out, newOut)
	case SliceType:
		// slices have two supported formats, like string:
		// - `{val, val, val}` (preferred)
		// - `val;val;val` (legacy)
		a.parseSlice(scanner, raw, out)
	case MapType:
		// maps are {string: val, string: val, string: val}
		a.parseMap(scanner, raw, out)
	}
}

// parseString parses either of the two accepted string forms (quoted, or bare tokens).
func (a *Argument) parseString(scanner *sc.Scanner, raw string, out reflect.Value) {
	// we need to temporarily disable the scanner's int/float parsing, since we want to
	// prevent number parsing errors.
	oldMode := scanner.Mode
	scanner.Mode = oldMode &^ sc.ScanInts &^ sc.ScanFloats
	defer func() {
		scanner.Mode = oldMode
	}()

	// strings are a bit weird -- the "easy" case is quoted strings (tokenized as strings),
	// the "hard" case (present for backwards compat) is a bare sequence of tokens that aren't
	// a comma.
	tok := scanner.Scan()
	if tok == sc.String || tok == sc.RawString {
		// the easy case
		val, err := strconv.Unquote(scanner.TokenText())
		if err != nil {
			scanner.Error(scanner, fmt.Sprintf("unable to parse string: %v", err))
			return
		}
		castAndSet(out, reflect.ValueOf(val))
		return
	}

	// the "hard" case -- bare tokens not including ',' (the argument
	// separator), ';' (the slice separator), ':' (the map separator), or '}'
	// (delimitted slice ender)
	startPos := scanner.Position.Offset
	for hint := peekNoSpace(scanner); hint != ',' && hint != ';' && hint != ':' && hint != '}' && hint != sc.EOF; hint = peekNoSpace(scanner) {
		// skip this token
		scanner.Scan()
	}
	endPos := scanner.Position.Offset + len(scanner.TokenText())
	castAndSet(out, reflect.ValueOf(raw[startPos:endPos]))
}

// parseSlice parses either of the two slice forms (curly-brace-delimitted and semicolon-separated).
func (a *Argument) parseSlice(scanner *sc.Scanner, raw string, out reflect.Value) {
	// slices have two supported formats, like string:
	// - `{val, val, val}` (preferred)
	// - `val;val;val` (legacy)
	resSlice := reflect.Zero(out.Type())
	elem := reflect.Indirect(reflect.New(out.Type().Elem()))

	// preferred case
	if peekNoSpace(scanner) == '{' {
		// NB(directxman12): supporting delimitted slices in bare slices
		// would require an extra look-ahead here :-/

		scanner.Scan() // skip '{'
		for hint := peekNoSpace(scanner); hint != '}' && hint != sc.EOF; hint = peekNoSpace(scanner) {
			a.ItemType.parse(scanner, raw, elem, true /* parsing a slice */)
			resSlice = reflect.Append(resSlice, elem)
			tok := peekNoSpace(scanner)
			if tok == '}' {
				break
			}
			if !expect(scanner, ',', "comma") {
				return
			}
		}
		if !expect(scanner, '}', "close curly brace") {
			return
		}
		castAndSet(out, resSlice)
		return
	}

	// legacy case
	for hint := peekNoSpace(scanner); hint != ',' && hint != '}' && hint != sc.EOF; hint = peekNoSpace(scanner) {
		a.ItemType.parse(scanner, raw, elem, true /* parsing a slice */)
		resSlice = reflect.Append(resSlice, elem)
		tok := peekNoSpace(scanner)
		if tok == ',' || tok == '}' || tok == sc.EOF {
			break
		}
		scanner.Scan()
		if tok != ';' {
			scanner.Error(scanner, fmt.Sprintf("expected comma, got %q", scanner.TokenText()))
			return
		}
	}
	castAndSet(out, resSlice)
}

// parseMap parses a map of the form {string: val, string: val, string: val}
func (a *Argument) parseMap(scanner *sc.Scanner, raw string, out reflect.Value) {
	resMap := reflect.MakeMap(out.Type())
	elem := reflect.Indirect(reflect.New(out.Type().Elem()))
	key := reflect.Indirect(reflect.New(out.Type().Key()))

	if !expect(scanner, '{', "open curly brace") {
		return
	}

	for hint := peekNoSpace(scanner); hint != '}' && hint != sc.EOF; hint = peekNoSpace(scanner) {
		a.parseString(scanner, raw, key)
		if !expect(scanner, ':', "colon") {
			return
		}
		a.ItemType.parse(scanner, raw, elem, false /* not in a slice */)
		resMap.SetMapIndex(key, elem)

		if peekNoSpace(scanner) == '}' {
			break
		}
		if !expect(scanner, ',', "comma") {
			return
		}
	}

	if !expect(scanner, '}', "close curly brace") {
		return
	}

	castAndSet(out, resMap)
}

// castAndSet casts val to out's type if needed,
// then sets out to val.
func castAndSet(out, val reflect.Value) {
	outType := out.Type()
	if outType != val.Type() {
		val = val.Convert(outType)
	}
	out.Set(val)
}

// makeSliceType makes a reflect.Type for a slice of the given type.
// Useful for constructing the out value for when AnyType's guess returns a slice.
func makeSliceType(itemType Argument) (reflect.Type, error) {
	var itemReflectedType reflect.Type
	switch itemType.Type {
	case IntType:
		itemReflectedType = reflect.TypeOf(int(0))
	case NumberType:
		itemReflectedType = reflect.TypeOf(float64(0))
	case StringType:
		itemReflectedType = reflect.TypeOf("")
	case BoolType:
		itemReflectedType = reflect.TypeOf(false)
	case SliceType:
		subItemType, err := makeSliceType(*itemType.ItemType)
		if err != nil {
			return nil, err
		}
		itemReflectedType = subItemType
	case MapType:
		subItemType, err := makeMapType(*itemType.ItemType)
		if err != nil {
			return nil, err
		}
		itemReflectedType = subItemType
	// TODO(directxman12): support non-uniform slices?  (probably not)
	default:
		return nil, fmt.Errorf("invalid type when constructing guessed slice out: %v", itemType.Type)
	}

	if itemType.Pointer {
		itemReflectedType = reflect.PtrTo(itemReflectedType)
	}

	return reflect.SliceOf(itemReflectedType), nil
}

// makeMapType makes a reflect.Type for a map of the given item type.
// Useful for constructing the out value for when AnyType's guess returns a map.
func makeMapType(itemType Argument) (reflect.Type, error) {
	var itemReflectedType reflect.Type
	switch itemType.Type {
	case IntType:
		itemReflectedType = reflect.TypeOf(int(0))
	case NumberType:
		itemReflectedType = reflect.TypeOf(float64(0))
	case StringType:
		itemReflectedType = reflect.TypeOf("")
	case BoolType:
		itemReflectedType = reflect.TypeOf(false)
	case SliceType:
		subItemType, err := makeSliceType(*itemType.ItemType)
		if err != nil {
			return nil, err
		}
		itemReflectedType = subItemType
	// TODO(directxman12): support non-uniform slices?  (probably not)
	case MapType:
		subItemType, err := makeMapType(*itemType.ItemType)
		if err != nil {
			return nil, err
		}
		itemReflectedType = subItemType
	case AnyType:
		// NB(directxman12): maps explicitly allow non-uniform item types, unlike slices at the moment
		itemReflectedType = interfaceType
	default:
		return nil, fmt.Errorf("invalid type when constructing guessed slice out: %v", itemType.Type)
	}

	if itemType.Pointer {
		itemReflectedType = reflect.PtrTo(itemReflectedType)
	}

	return reflect.MapOf(reflect.TypeOf(""), itemReflectedType), nil
}

// guessType takes an educated guess about the type of the next field.  If allowSlice
// is false, it will not guess slices.  It's less efficient than parsing with actual
// type information, since we need to allocate to peek ahead full tokens, and the scanner
// only allows peeking ahead one character.
// Maps are *always* non-uniform (i.e. type the AnyType item type), since they're frequently
// used to represent things like defaults for an object in JSON.
func guessType(scanner *sc.Scanner, raw string, allowSlice bool) *Argument {
	if allowSlice {
		maybeItem := guessType(scanner, raw, false)

		subRaw := raw[scanner.Pos().Offset:]
		subScanner := parserScanner(subRaw, scanner.Error)

		var tok rune
		for {
			tok = subScanner.Scan()
			if tok == ',' || tok == sc.EOF || tok == ';' {
				break
			}
			// wait till we get something interesting
		}

		// semicolon means it's a legacy slice
		if tok == ';' {
			return &Argument{
				Type:     SliceType,
				ItemType: maybeItem,
			}
		}

		return maybeItem
	}

	// everything else needs a duplicate scanner to scan properly
	// (so we don't consume our scanner tokens until we actually
	// go to use this -- Go doesn't like scanners that can be rewound).
	subRaw := raw[scanner.Pos().Offset:]
	subScanner := parserScanner(subRaw, scanner.Error)

	// skip whitespace
	hint := peekNoSpace(subScanner)

	// first, try the easy case -- quoted strings strings
	switch hint {
	case '"', '\'', '`':
		return &Argument{Type: StringType}
	}

	// next, check for slices or maps
	if hint == '{' {
		subScanner.Scan()

		// TODO(directxman12): this can't guess at empty objects, but that's generally ok.
		// We'll cross that bridge when we get there.

		// look ahead till we can figure out if this is a map or a slice
		firstElemType := guessType(subScanner, subRaw, false)
		if firstElemType.Type == StringType {
			// might be a map or slice, parse the string and check for colon
			// (blech, basically arbitrary look-ahead due to raw strings).
			var keyVal string // just ignore this
			(&Argument{Type: StringType}).parseString(subScanner, raw, reflect.Indirect(reflect.ValueOf(&keyVal)))

			if subScanner.Scan() == ':' {
				// it's got a string followed by a colon -- it's a map
				return &Argument{
					Type:     MapType,
					ItemType: &Argument{Type: AnyType},
				}
			}
		}

		// definitely a slice -- maps have to have string keys and have a value followed by a colon
		return &Argument{
			Type:     SliceType,
			ItemType: firstElemType,
		}
	}

	// then, bools...
	probablyString := false
	if hint == 't' || hint == 'f' {
		// maybe a bool
		if nextTok := subScanner.Scan(); nextTok == sc.Ident {
			switch subScanner.TokenText() {
			case "true", "false":
				// definitely a bool
				return &Argument{Type: BoolType}
			}
			// probably a string
			probablyString = true
		} else {
			// we shouldn't ever get here
			scanner.Error(scanner, fmt.Sprintf("got a token (%q) that looked like an ident, but was not", scanner.TokenText()))
			return &Argument{Type: InvalidType}
		}
	}

	// then, integers...
	if !probablyString {
		nextTok := subScanner.Scan()
		if nextTok == '-' {
			nextTok = subScanner.Scan()
		}

		if nextTok == sc.Int {
			return &Argument{Type: IntType}
		}
		if nextTok == sc.Float {
			return &Argument{Type: NumberType}
		}
	}

	// otherwise assume bare strings
	return &Argument{Type: StringType}
}

// peekNoSpace is equivalent to scanner.Peek, except that it will consume intervening whitespace.
func peekNoSpace(scanner *sc.Scanner) rune {
	hint := scanner.Peek()
	for ; hint <= rune(' ') && ((1<<uint64(hint))&scanner.Whitespace) != 0; hint = scanner.Peek() {
		scanner.Next() // skip the whitespace
	}
	return hint
}

var (
	// interfaceType is a pre-computed reflect.Type representing the empty interface.
	interfaceType = reflect.TypeOf((*interface{})(nil)).Elem()
	rawArgsType   = reflect.TypeOf((*RawArguments)(nil)).Elem()
)

// lowerCamelCase converts PascalCase string to
// a camelCase string (by lowering the first rune).
func lowerCamelCase(in string) string {
	isFirst := true
	return strings.Map(func(inRune rune) rune {
		if isFirst {
			isFirst = false
			return unicode.ToLower(inRune)
		}
		return inRune
	}, in)
}

// ArgumentFromType constructs an Argument by examining the given
// raw reflect.Type.  It can construct arguments from the Go types
// corresponding to any of the types listed in ArgumentType.
func ArgumentFromType(rawType reflect.Type) (Argument, error) {
	if rawType == rawArgsType {
		return Argument{
			Type: RawType,
		}, nil
	}

	if rawType == interfaceType {
		return Argument{
			Type: AnyType,
		}, nil
	}

	arg := Argument{}
	if rawType.Kind() == reflect.Ptr {
		rawType = rawType.Elem()
		arg.Pointer = true
		arg.Optional = true
	}

	switch rawType.Kind() {
	case reflect.String:
		arg.Type = StringType
	case reflect.Int, reflect.Int32: // NB(directxman12): all ints in kubernetes are int32, so explicitly support that
		arg.Type = IntType
	case reflect.Float64:
		arg.Type = NumberType
	case reflect.Bool:
		arg.Type = BoolType
	case reflect.Slice:
		arg.Type = SliceType
		itemType, err := ArgumentFromType(rawType.Elem())
		if err != nil {
			return Argument{}, fmt.Errorf("bad slice item type: %w", err)
		}
		arg.ItemType = &itemType
	case reflect.Map:
		arg.Type = MapType
		if rawType.Key().Kind() != reflect.String {
			return Argument{}, fmt.Errorf("bad map key type: map keys must be strings")
		}
		itemType, err := ArgumentFromType(rawType.Elem())
		if err != nil {
			return Argument{}, fmt.Errorf("bad slice item type: %w", err)
		}
		arg.ItemType = &itemType
	default:
		return Argument{}, fmt.Errorf("type has unsupported kind %s", rawType.Kind())
	}

	return arg, nil
}

// argumentInfo returns information about an argument field as the marker parser's field loader
// would see it.  This can be useful if you have to interact with marker definition structs
// externally (e.g. at compile time).
func argumentInfo(fieldName string, tag reflect.StructTag) (argName string, optionalOpt bool) {
	argName = lowerCamelCase(fieldName)
	markerTag, tagSpecified := tag.Lookup("marker")
	markerTagParts := strings.Split(markerTag, ",")
	if tagSpecified && markerTagParts[0] != "" {
		// allow overriding to support legacy cases where we don't follow camelCase conventions
		argName = markerTagParts[0]
	}
	optionalOpt = false
	for _, tagOption := range markerTagParts[1:] {
		switch tagOption {
		case "optional":
			optionalOpt = true
		}
	}

	return argName, optionalOpt
}
