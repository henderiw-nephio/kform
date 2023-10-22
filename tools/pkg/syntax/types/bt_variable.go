package types

import (
	"context"
	"fmt"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	blockv1alpha1 "github.com/henderiw-nephio/kform/tools/apis/kform/block/v1alpha1"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/exttypes"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/sctx"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cctx"
)

func newVariable(n string) Block {
	return &variable{
		bt{
			Level: 1,
			Name:  n,
		},
	}
}

type variable struct{ bt }

func (r *variable) AddData(ctx context.Context) {
	attrs := cctx.GetContextValue[map[string]any](ctx, sctx.CtxKeyAttributes)

	allDeps := []string{}
	for _, attr := range attrs {
		deps := r.getDependencies(ctx, attr)
		if len(deps) > 0 {
			allDeps = append(allDeps, deps...)
		}
	}

	execCfg := cctx.GetContextValue[exttypes.ExecConfig](ctx, sctx.CtxExecConfig)
	if execCfg == nil {
		r.recorder.Record(diag.DiagErrorf("cannot add variable without execConfig"))
	}

	name := fmt.Sprintf("var.%s", cctx.GetContextValue[string](ctx, sctx.CtxKeyVarName))

	execCfg.GetVars().AddVertex(ctx, name, blockv1alpha1.Variable{
		FileName:     cctx.GetContextValue[string](ctx, sctx.CtxKeyFileName),
		BlockType:    blockv1alpha1.BlockTypeVariable,
		Attributes:   attrs,
		Dependencies: allDeps,
	})
}
