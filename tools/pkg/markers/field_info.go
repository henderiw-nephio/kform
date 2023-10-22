package markers

import (
	"go/ast"
	"reflect"
)

// FieldInfo contains marker values and commonly used information for a struct field.
type FieldInfo struct {
	// Name is the name of the field (or "" for embedded fields)
	Name string
	// Doc is the Godoc of the field, pre-processed to remove markers and joine
	// single newlines together.
	Doc string
	// Tag struct tag associated with this field (or "" if non existed).
	Tag reflect.StructTag

	// Markers are all registered markers associated with this field.
	Markers MarkerValues

	// RawField is the raw, underlying field AST object that this field represents.
	RawField *ast.Field
}
