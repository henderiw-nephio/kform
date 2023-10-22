package syntax

import (
	"context"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cctx"
)

func (r *parser) Resolve(ctx context.Context) {
	execCfg := cctx.GetContextValue[ExecConfig](ctx, CtxExecConfig)
	if execCfg == nil {
		r.recorder.Record(diag.DiagErrorf("cannot resolve without execConfig"))
	}
	for n, v := range execCfg.GetVars().GetVertices() {
		for _, d := range v.Dependencies {
			// validate if all dependencies can be resolved
			if _, err := execCfg.GetVars().GetVertex(d); err != nil {
				r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(n), "dependency resolution failed for %s", d))
			}
			// validate if the provider is defined
			if _, err := execCfg.GetProviders().GetVertex(v.Provider); err != nil {
				r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(n), "provider resolution failed for %s", v.Provider))
			}
		}
	}
}
