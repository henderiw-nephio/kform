package fns

import (
	"context"
	"fmt"

	"github.com/henderiw-nephio/kform/tools/pkg/exec/fn/render"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/vars"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
)

type renderer struct {
	vars cache.Cache[vars.Variable]
}

func (r *renderer) renderData(ctx context.Context, blockName string, x any, localVars map[string]any) error {
	renderer := render.Renderer{
		Vars:      r.vars,
		LocalVars: localVars,
	}
	d, err := renderer.Render(ctx, x)
	if err != nil {
		return fmt.Errorf("run output, render failed for blockName %s, err: %s", blockName, err.Error())
	}

	switch d := d.(type) {
	case []any:
		r.vars.Add(ctx, cache.NSN{Name: blockName}, vars.Variable{Data: map[string][]any{
			vars.DummyKey: d,
		}})
	default:
		r.vars.Add(ctx, cache.NSN{Name: blockName}, vars.Variable{Data: map[string][]any{
			vars.DummyKey: {d},
		}})
	}
	return nil
}
