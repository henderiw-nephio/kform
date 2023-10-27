package types

/*
import (
	"context"
	"fmt"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	blockv1alpha1 "github.com/henderiw-nephio/kform/tools/apis/kform/block/v1alpha1"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/exttypes"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cctx"
)

func newInput(n string) Block {
	return &input{
		bt{
			Level: 1,
			Name:  n,
		},
	}
}

type input struct{ bt }

func (r *input) AddData(ctx context.Context) {
	attrs := cctx.GetContextValue[map[string]any](ctx, blockv1alpha1.CtxKeyAttributes)

	allDeps := []string{}
	for _, attr := range attrs {
		deps := r.getDependencies(ctx, attr)
		if len(deps) > 0 {
			allDeps = append(allDeps, deps...)
		}
	}

	execCfg := cctx.GetContextValue[exttypes.ExecConfig](ctx, blockv1alpha1.CtxExecConfig)
	if execCfg == nil {
		r.recorder.Record(diag.DiagErrorf("cannot add variable without execConfig"))
	}

	name := fmt.Sprintf("var.%s", cctx.GetContextValue[string](ctx, blockv1alpha1.CtxKeyVarName))

	execCfg.GetVars().AddVertex(ctx, name, blockv1alpha1.Variable{
		FileName:     cctx.GetContextValue[string](ctx, blockv1alpha1.CtxKeyFileName),
		BlockType:    blockv1alpha1.BlockTypeInput,
		Attributes:   attrs,
		Dependencies: allDeps,
	})
}
*/