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

func NewResourceFn(cfg *Config) fn.BlockInstanceRunner {
	return &resource{
		vars: cfg.Vars,
	}
}

/*
	1. run for_each or count
	-> determines how many child modules will be instantiated
	-> assign a local variable each.key/value or count.key/value

	Per execution context (single or range (count/for_each))
	1. run the cell functions to generate the respected output

	2. run the provider

*/

type resource struct {
	vars cache.Cache[vars.Variable]
}

func (r *resource) Run(ctx context.Context, vCtx *types.VertexContext, localVars map[string]any) error {
	log := log.FromContext(ctx).With("vertexContext", vctx.GetContext(vCtx))
	log.Info("run instance")
	if vCtx.BlockContext.Config != nil {
		renderer := &renderer{vars: r.vars}
		if err := renderer.renderData(ctx, vCtx.BlockName, vCtx.BlockContext.Config, localVars); err != nil {
			return err
		}
	}

	// TODO run provider
	return nil
}
