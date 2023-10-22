package loader

import (
	"fmt"
	"go/token"
)

// PositionedError represents some error with an associated position.
// The position is tied to some external token.FileSet.
type PositionedError struct {
	Pos token.Pos
	error
}

// Node is the intersection of go/ast.Node and go/types.Var.
type Node interface {
	Pos() token.Pos // position of first character belonging to the node
}

// ErrFromNode returns the given error, with additional information
// attaching it to the given AST node.  It will automatically map
// over error lists.
func ErrFromNode(err error, node Node) error {
	if asList, isList := err.(ErrList); isList {
		resList := make(ErrList, len(asList))
		for i, baseErr := range asList {
			resList[i] = ErrFromNode(baseErr, node)
		}
		return resList
	}
	return PositionedError{
		Pos:   node.Pos(),
		error: err,
	}
}

// MaybeErrList constructs an ErrList if the given list of
// errors has any errors, otherwise returning nil.
func MaybeErrList(errs []error) error {
	if len(errs) == 0 {
		return nil
	}
	return ErrList(errs)
}

// ErrList is a list of errors aggregated together into a single error.
type ErrList []error

func (l ErrList) Error() string {
	return fmt.Sprintf("%v", []error(l))
}
