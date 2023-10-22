package markers

import (
	"context"
	"fmt"

	"github.com/henderiw-nephio/kform/tools/pkg/markers"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cctx"
)

// AllDefinitions contains all marker definitions for this package.
var AllDefinitions []*definitionWithHelp

type definitionWithHelp struct {
	*markers.Definition
	Help *markers.DefinitionHelp
}

func (d *definitionWithHelp) WithHelp(help *markers.DefinitionHelp) *definitionWithHelp {
	d.Help = help
	return d
}

func (d *definitionWithHelp) Register(ctx context.Context) error {
	registry := cctx.GetContextValue[markers.Registry](ctx, "registry")
	if registry == nil {
		return fmt.Errorf("no registry provided")
	}

	if err := registry.Register(d.Definition); err != nil {
		return err
	}
	if d.Help != nil {
		registry.AddHelp(d.Definition, d.Help)
	}
	return nil
}

func must(def *markers.Definition, err error) *definitionWithHelp {
	return &definitionWithHelp{
		Definition: markers.Must(def, err),
	}
}

type hasHelp interface {
	Help() *markers.DefinitionHelp
}
