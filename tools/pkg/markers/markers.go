package markers

import (
	"go/ast"
	"go/token"
	"strings"

	"github.com/henderiw-nephio/kform/tools/pkg/loader"
)

// PackageMarkers collects all the package-level marker values for the given package.
func PackageMarkers(col Collector, pkg *loader.Package) (MarkerValues, error) {
	markers, err := col.MarkersInPackage(pkg)
	if err != nil {
		return nil, err
	}
	res := make(MarkerValues)
	for _, file := range pkg.Syntax {
		fileMarkers := markers[file]
		for name, vals := range fileMarkers {
			res[name] = append(res[name], vals...)
		}
	}

	return res, nil
}

// TypeCallback is a callback called for each type declaration in a package.
type TypeCallback func(info *TypeInfo)

// EachType collects all markers, then calls the given callback for each type declaration in a package.
// Each individual spec is considered separate, so
//
//	type (
//	    Foo string
//	    Bar int
//	    Baz struct{}
//	)
//
// yields three calls to the callback.
func EachType(col Collector, pkg *loader.Package, cb TypeCallback) error {
	markers, err := col.MarkersInPackage(pkg)
	if err != nil {
		return err
	}

	pkg.EachType(func(file *ast.File, decl *ast.GenDecl, spec *ast.TypeSpec) {
		var fields []FieldInfo
		if structSpec, isStruct := spec.Type.(*ast.StructType); isStruct {
			for _, field := range structSpec.Fields.List {
				for _, name := range field.Names {
					fields = append(fields, FieldInfo{
						Name:     name.Name,
						Doc:      extractDoc(field, nil),
						Tag:      loader.ParseAstTag(field.Tag),
						Markers:  markers[field],
						RawField: field,
					})
				}
				if field.Names == nil {
					fields = append(fields, FieldInfo{
						Doc:      extractDoc(field, nil),
						Tag:      loader.ParseAstTag(field.Tag),
						Markers:  markers[field],
						RawField: field,
					})
				}
			}
		}

		cb(&TypeInfo{
			Name:    spec.Name.Name,
			Markers: markers[spec],
			Doc:     extractDoc(spec, decl),
			Fields:  fields,
			RawDecl: decl,
			RawSpec: spec,
			RawFile: file,
		})
	})

	return nil
}

// extractDoc extracts documentation from the given node, skipping markers
// in the godoc and falling back to the decl if necessary (for single-line decls).
func extractDoc(node ast.Node, decl *ast.GenDecl) string {
	var docs *ast.CommentGroup
	switch docced := node.(type) {
	case *ast.Field:
		docs = docced.Doc
	case *ast.File:
		docs = docced.Doc
	case *ast.GenDecl:
		docs = docced.Doc
	case *ast.TypeSpec:
		docs = docced.Doc
		// type Ident expr expressions get docs attached to the decl,
		// so check for that case (missing Lparen == single line type decl)
		if docs == nil && decl.Lparen == token.NoPos {
			docs = decl.Doc
		}
	}

	if docs == nil {
		return ""
	}

	// filter out markers
	var outGroup ast.CommentGroup
	outGroup.List = make([]*ast.Comment, 0, len(docs.List))
	for _, comment := range docs.List {
		if isMarkerComment(comment.Text) {
			continue
		}
		outGroup.List = append(outGroup.List, comment)
	}

	// split lines, and re-join together as a single
	// paragraph, respecting double-newlines as
	// paragraph markers.
	outLines := strings.Split(outGroup.Text(), "\n")
	if outLines[len(outLines)-1] == "" {
		// chop off the extraneous last part
		outLines = outLines[:len(outLines)-1]
	}

	for i, line := range outLines {
		// Trim any extranous whitespace,
		// for handling /*…*/-style comments,
		// which have whitespace preserved in go/ast:
		line = strings.TrimSpace(line)

		// Respect that double-newline means
		// actual newline:
		if line == "" {
			outLines[i] = "\n"
		} else {
			outLines[i] = line
		}
	}

	return strings.Join(outLines, " ")
}
