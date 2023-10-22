package openapi

import (
	"fmt"

	"github.com/henderiw-nephio/kform/tools/pkg/loader"
)

// TypeIdent represents some type in a Package.
type TypeIdent struct {
	Package *loader.Package
	Name    string
}

func (t TypeIdent) String() string {
	return fmt.Sprintf("%q.%s", t.Package.ID, t.Name)
}
