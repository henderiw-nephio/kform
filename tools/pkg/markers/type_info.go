package markers

import "go/ast"

// TypeInfo contains marker values and commonly used information for a type declaration.
type TypeInfo struct {
	// Name is the name of the type.
	Name string
	// Doc is the Godoc of the type, pre-processed to remove markers and joine
	// single newlines together.
	Doc string

	// Markers are all registered markers associated with the type.
	Markers MarkerValues

	// Fields are all the fields associated with the type, if it's a struct.
	// (if not, Fields will be nil).
	Fields []FieldInfo

	// RawDecl contains the raw GenDecl that the type was declared as part of.
	RawDecl *ast.GenDecl
	// RawSpec contains the raw Spec that declared this type.
	RawSpec *ast.TypeSpec
	// RawFile contains the file in which this type was declared.
	RawFile *ast.File
}
