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

func (r *input) Run(ctx context.Context, vCtx *vctx.VertexContext, localVars map[string]any) error {
	// NOTE: No forEach or count expected
	log := log.FromContext(ctx).With("vertexContext", vctx.GetContext(vCtx))
	log.Info("run instance")
	// check if the blockName (aka input variable) exists in the variable
	// if not copy the default parameters to the variable cache if default is defined 
	if _, err := r.vars.Get(cache.NSN{Name: vCtx.BlockName}); err != nil {
		if len(vCtx.BlockContext.Default) > 0 {
			r.vars.Upsert(ctx, cache.NSN{Name: vCtx.BlockName}, vars.Variable{Data: map[string][]any{
				vars.DummyKey: vCtx.BlockContext.Default,
			}})
		}
	}

	return nil
}
