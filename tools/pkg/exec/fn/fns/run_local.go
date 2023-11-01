package fns

import (
	"context"

	"github.com/henderiw-nephio/kform/tools/pkg/exec/fn"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/vars"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/vctx"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
	"github.com/henderiw/logger/log"
)

func NewLocalFn(cfg *Config) fn.BlockInstanceRunner {
	return &local{
		vars: cfg.Vars,
	}
}

type local struct {
	vars cache.Cache[vars.Variable]
}

func (r *local) Run(ctx context.Context, vCtx *vctx.VertexContext, localVars map[string]any) error {
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
