package syntax

import (
	"context"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	blockv1alpha1 "github.com/henderiw-nephio/kform/tools/apis/kform/block/v1alpha1"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cctx"
)

func (r *parser) Resolve(ctx context.Context) {
	execCfg := cctx.GetContextValue[blockv1alpha1.ExecConfig](ctx, blockv1alpha1.CtxExecConfig)
	if execCfg == nil {
		r.recorder.Record(diag.DiagErrorf("cannot resolve without execConfig"))
	}
	for n, v := range execCfg.GetVars().GetVertices() {
		for _, d := range v.GetDependencies() {
			// validate if all dependencies can be resolved
			if _, err := execCfg.GetVars().GetVertex(d); err != nil {
				r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(n), "dependency resolution failed for %s", d))
			}
			// validate if the provider is defined
			if _, err := execCfg.GetProviders().GetVertex(v.GetProvider()); err != nil {
				r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(n), "provider resolution failed for %s", v.Provider))
			}
		}
	}
}
