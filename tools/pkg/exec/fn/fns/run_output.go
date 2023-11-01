package fns

import (
	"context"

	"github.com/henderiw-nephio/kform/tools/pkg/exec/fn"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/vars"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/vctx"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
	"github.com/henderiw/logger/log"
)

func NewOutputFn(cfg *Config) fn.BlockInstanceRunner {
	return &output{
		vars: cfg.Vars,
	}
}

type output struct {
	vars cache.Cache[vars.Variable]
}

/*
	1. run for_each or count
	-> determines how many child modules will be instantiated
	-> assign a local variable each.key/value or count.key/value

	Per execution context (single or range (count/for_each))
	1. run the cell functions to generate the respected output

*/

func (r *output) Run(ctx context.Context, vCtx *vctx.VertexContext, localVars map[string]any) error {
	log := log.FromContext(ctx).With("vertexContext", vctx.GetContext(vCtx))
	log.Info("run instance")
	if vCtx.BlockContext.Value != nil {
		renderer := &renderer{vars: r.vars}
		if err := renderer.renderData(ctx, vCtx.BlockName, vCtx.BlockContext.Value, localVars); err != nil {
			return err
		}
	}

	return nil
}
