package fns

import (
	"context"

	"github.com/henderiw-nephio/kform/tools/pkg/exec/fn"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/vars"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/vctx"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
	"github.com/henderiw/logger/log"
)

func NewLocalOrOutputFn(cfg *Config) fn.BlockInstanceRunner {
	return &localOrOutput{
		vars: cfg.Vars,
	}
}

type localOrOutput struct {
	vars cache.Cache[vars.Variable]
}

func (r *localOrOutput) Run(ctx context.Context, vCtx *vctx.VertexContext, localVars map[string]any) error {
	// NOTE: forEach or count expected and its respective values will be represented in localVars
	// ForEach: each.key/value
	// Count: count.index
	log := log.FromContext(ctx).With("vertexContext", vctx.GetContext(vCtx))
	log.Info("run instance")
	// if the BlockContext Value is defined we render the expected output
	// the syntax parser should validate this, meaning the value should always be defined
	if vCtx.BlockContext.Value != nil {
		renderer := &renderer{vars: r.vars}
		if err := renderer.renderData(ctx, vCtx.BlockName, vCtx.BlockContext.Value, localVars); err != nil {
			return err
		}
	}
	return nil
}
