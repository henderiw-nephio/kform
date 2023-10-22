package syntax

import (
	"context"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/exttypes"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/sctx"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cctx"
)

func (r *parser) Connect(ctx context.Context) {
	execCfg := cctx.GetContextValue[exttypes.ExecConfig](ctx, sctx.CtxExecConfig)
	if execCfg == nil {
		r.recorder.Record(diag.DiagErrorf("cannot connect without execConfig"))
	}
	for n, v := range execCfg.GetVars().GetVertices() {
		for _, d := range v.Dependencies {
			execCfg.GetVars().Connect(ctx, d, n)
		}
	}
}
