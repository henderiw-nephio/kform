package openapi

import (
	"context"
	"fmt"

	"github.com/henderiw-nephio/kform/tools/pkg/genall"
	"github.com/henderiw-nephio/kform/tools/pkg/loader"
	"github.com/henderiw-nephio/kform/tools/pkg/openapi/markers"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cctx"
)

func NewGenerator() genall.Generator {
	return generator{}
}

type generator struct {
}

func (r generator) RegisterMarkers(ctx context.Context) error {
	return markers.Register(ctx)
}

func (r generator) Generate(ctx context.Context) error {
	p := NewParser(ctx)

	rootPkgs := cctx.GetContextValue[[]*loader.Package](ctx, "roots")
	if len(rootPkgs) == 0 {
		return fmt.Errorf("cannot generate openapi without root packages: %d", len(rootPkgs))
	}
	// Get the type and type-checkign information for the package
	for _, pkg := range rootPkgs {
		p.NeedPackage(pkg)
	}

	typeIdents := p.FindProviderAPIs()
	if len(typeIdents) == 0 {
		return nil
	}
	for _, typeIdent := range typeIdents {
		p.NeedOpenAPIFor(typeIdent)
	}

	p.PrintSchemata()

	return nil
}
