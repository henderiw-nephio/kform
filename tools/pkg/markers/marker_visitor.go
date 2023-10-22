package markers

import (
	"go/ast"
	"go/token"
)

// markerVisistor visits AST nodes, recording markers associated with each node.
type markerVisitor struct {
	allComments []*ast.CommentGroup
	commentInd  int

	declComments         []markerComment
	lastLineCommentGroup *ast.CommentGroup

	pkgMarkers  []markerComment
	nodeMarkers map[ast.Node][]markerComment
}

type markerSubVisitor struct {
	*markerVisitor
	node                ast.Node
	collectPackageLevel bool
}

// markersBetween grabs the markers between the given indicies in the list of all comments.
func (v *markerVisitor) markersBetween(fromGodoc bool, start, end int) []markerComment {
	if start < 0 || end < 0 {
		return nil
	}
	var res []markerComment
	for i := start; i < end; i++ {
		commentGroup := v.allComments[i]
		for _, comment := range commentGroup.List {
			if !isMarkerComment(comment.Text) {
				continue
			}
			res = append(res, markerComment{Comment: comment, fromGodoc: fromGodoc})
		}
	}
	return res
}

// Visit collects markers for each node in the AST, optionally
// collecting unassociated markers as package-level.
func (v markerSubVisitor) Visit(node ast.Node) ast.Visitor {
	if node == nil {
		// end of the node, so we might need to advance comments beyond the end
		// of the block if we don't want to collect package-level markers in
		// this block.

		if !v.collectPackageLevel {
			if v.commentInd < len(v.allComments) {
				lastCommentInd := v.commentInd
				nextGroup := v.allComments[lastCommentInd]
				for nextGroup.Pos() < v.node.End() {
					lastCommentInd++
					if lastCommentInd >= len(v.allComments) {
						// after the increment so our decrement below still makes sense
						break
					}
					nextGroup = v.allComments[lastCommentInd]
				}
				v.commentInd = lastCommentInd
			}
		}

		return nil
	}

	// skip comments on the same line as the previous node
	// making sure to double-check for the case where we've gone past the end of the comments
	// but still have to finish up typespec-gendecl association (see below).
	if v.lastLineCommentGroup != nil && v.commentInd < len(v.allComments) && v.lastLineCommentGroup.Pos() == v.allComments[v.commentInd].Pos() {
		v.commentInd++
	}

	// stop visiting if there are no more comments in the file
	// NB(directxman12): we can't just stop immediately, because we
	// still need to check if there are typespecs associated with gendecls.
	var markerCommentBlock []markerComment
	var docCommentBlock []markerComment
	lastCommentInd := v.commentInd
	if v.commentInd < len(v.allComments) {
		// figure out the first comment after the node in question...
		nextGroup := v.allComments[lastCommentInd]
		for nextGroup.Pos() < node.Pos() {
			lastCommentInd++
			if lastCommentInd >= len(v.allComments) {
				// after the increment so our decrement below still makes sense
				break
			}
			nextGroup = v.allComments[lastCommentInd]
		}
		lastCommentInd-- // ...then decrement to get the last comment before the node in question

		// figure out the godoc comment so we can deal with it separately
		var docGroup *ast.CommentGroup
		docGroup, v.lastLineCommentGroup = associatedCommentsFor(node)

		// find the last comment group that's not godoc
		markerCommentInd := lastCommentInd
		if docGroup != nil && v.allComments[markerCommentInd].Pos() == docGroup.Pos() {
			markerCommentInd--
		}

		// check if we have freestanding package markers,
		// and find the markers in our "closest non-godoc" comment block,
		// plus our godoc comment block
		if markerCommentInd >= v.commentInd {
			if v.collectPackageLevel {
				// assume anything between the comment ind and the marker ind (not including it)
				// are package-level
				v.pkgMarkers = append(v.pkgMarkers, v.markersBetween(false, v.commentInd, markerCommentInd)...)
			}
			markerCommentBlock = v.markersBetween(false, markerCommentInd, markerCommentInd+1)
			docCommentBlock = v.markersBetween(true, markerCommentInd+1, lastCommentInd+1)
		} else {
			docCommentBlock = v.markersBetween(true, markerCommentInd+1, lastCommentInd+1)
		}
	}

	resVisitor := markerSubVisitor{
		collectPackageLevel: false, // don't collect package level by default
		markerVisitor:       v.markerVisitor,
		node:                node,
	}

	// associate those markers with a node
	switch typedNode := node.(type) {
	case *ast.GenDecl:
		// save the comments associated with the gen-decl if it's a single-line type decl
		if typedNode.Lparen != token.NoPos || typedNode.Tok != token.TYPE {
			// not a single-line type spec, treat them as free comments
			v.pkgMarkers = append(v.pkgMarkers, markerCommentBlock...)
			break
		}
		// save these, we'll need them when we encounter the actual type spec
		v.declComments = append(v.declComments, markerCommentBlock...)
		v.declComments = append(v.declComments, docCommentBlock...)
	case *ast.TypeSpec:
		// add in comments attributed to the gen-decl, if any,
		// as well as comments associated with the actual type
		v.nodeMarkers[node] = append(v.nodeMarkers[node], v.declComments...)
		v.nodeMarkers[node] = append(v.nodeMarkers[node], markerCommentBlock...)
		v.nodeMarkers[node] = append(v.nodeMarkers[node], docCommentBlock...)

		v.declComments = nil
		v.collectPackageLevel = false // don't collect package-level inside type structs
	case *ast.Field:
		v.nodeMarkers[node] = append(v.nodeMarkers[node], markerCommentBlock...)
		v.nodeMarkers[node] = append(v.nodeMarkers[node], docCommentBlock...)
	case *ast.File:
		v.pkgMarkers = append(v.pkgMarkers, markerCommentBlock...)
		v.pkgMarkers = append(v.pkgMarkers, docCommentBlock...)

		// collect markers in root file scope
		resVisitor.collectPackageLevel = true
	default:
		// assume markers before anything else are package-level markers,
		// *but* don't include any markers in godoc
		if v.collectPackageLevel {
			v.pkgMarkers = append(v.pkgMarkers, markerCommentBlock...)
		}
	}

	// increment the comment ind so that we start at the right place for the next node
	v.commentInd = lastCommentInd + 1

	return resVisitor
}

// associatedCommentsFor returns the doc comment group (if relevant and present) and end-of-line comment
// (again if relevant and present) for the given AST node.
func associatedCommentsFor(node ast.Node) (docGroup *ast.CommentGroup, lastLineCommentGroup *ast.CommentGroup) {
	switch typedNode := node.(type) {
	case *ast.Field:
		docGroup = typedNode.Doc
		lastLineCommentGroup = typedNode.Comment
	case *ast.File:
		docGroup = typedNode.Doc
	case *ast.FuncDecl:
		docGroup = typedNode.Doc
	case *ast.GenDecl:
		docGroup = typedNode.Doc
	case *ast.ImportSpec:
		docGroup = typedNode.Doc
		lastLineCommentGroup = typedNode.Comment
	case *ast.TypeSpec:
		docGroup = typedNode.Doc
		lastLineCommentGroup = typedNode.Comment
	case *ast.ValueSpec:
		docGroup = typedNode.Doc
		lastLineCommentGroup = typedNode.Comment
	default:
		lastLineCommentGroup = nil
	}

	return docGroup, lastLineCommentGroup
}
