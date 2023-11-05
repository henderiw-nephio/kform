package fns

import (
	"context"

	"github.com/henderiw-nephio/kform/tools/pkg/exec/fn"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/vars"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/vctx"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/types"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
	"github.com/henderiw/logger/log"
)

func NewLocalOrOutputFn(cfg *Config) fn.BlockInstanceRunner {
	return &localOrOutput{
		rootModuleName: cfg.RootModuleName,
		vars:           cfg.Vars,
	}
}

type localOrOutput struct {
	rootModuleName string
	vars           cache.Cache[vars.Variable]
}

func (r *localOrOutput) Run(ctx context.Context, vCtx *types.VertexContext, localVars map[string]any) error {
	// NOTE: forEach or count expected and its respective values will be represented in localVars
	// ForEach: each.key/value
	// Count: count.index
	log := log.FromContext(ctx).With("vertexContext", vctx.GetContext(r.rootModuleName, vCtx))
	log.Info("run instance")
	// if the BlockContext Value is defined we render the expected output
	// the syntax parser should validate this, meaning the value should always be defined
	if vCtx.BlockContext.Value != nil {
		renderer := &Renderer{Vars: r.vars}
		if err := renderer.RenderData(ctx, vCtx.BlockName, vCtx.BlockContext.Value, localVars); err != nil {
			return err
		}
	}
	return nil
}
