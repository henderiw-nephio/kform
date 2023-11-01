package fns

import (
	"context"

	"github.com/henderiw-nephio/kform/tools/pkg/exec/fn"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/vars"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/vctx"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
	"github.com/henderiw/logger/log"
)

// provide and input runner, which runs per input instance
func NewInputFn(cfg *Config) fn.BlockInstanceRunner {
	return &input{
		vars: cfg.Vars,
	}
}

type input struct {
	vars cache.Cache[vars.Variable]
}

/*
	NOTE: No for_each or count expected
	1. if input does not exist in variable copy the default to the vars
*/

func (r *input) Run(ctx context.Context, vCtx *vctx.VertexContext, localVars map[string]any) error {
	log := log.FromContext(ctx).With("vertexContext", vctx.GetContext(vCtx))
	log.Info("run instance")
	if _, err := r.vars.Get(cache.NSN{Name: vCtx.BlockName}); err != nil {
		if len(vCtx.BlockContext.Default) > 0 {
			r.vars.Upsert(ctx, cache.NSN{Name: vCtx.BlockName}, vars.Variable{Data: map[string][]any{
				vars.DummyKey: vCtx.BlockContext.Default,
			}})
		}
	}

	return nil
}
