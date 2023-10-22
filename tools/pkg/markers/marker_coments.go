package markers

import (
	"go/ast"
	"strings"
)

// markerComment is an AST comment that contains a marker.
// It may or may not be from a Godoc comment, which affects
// marker re-associated (from type-level to package-level)
type markerComment struct {
	*ast.Comment
	fromGodoc bool
}

// Text returns the text of the marker, stripped of the comment
// marker and leading spaces, as should be passed to Registry.Lookup
// and Registry.Parse.
func (c markerComment) Text() string {
	return strings.TrimSpace(c.Comment.Text[2:])
}

// isMarkerComment checks that the given comment is a single-line (`//`)
// comment and it's first non-space content is `+`.
func isMarkerComment(comment string) bool {
	if comment[0:2] != "//" {
		return false
	}
	stripped := strings.TrimSpace(comment[2:])
	if len(stripped) < 1 || stripped[0] != '+' {
		return false
	}
	return true
}
