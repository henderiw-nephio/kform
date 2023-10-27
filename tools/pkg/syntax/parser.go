package syntax

import (
	"context"
	"fmt"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	blockv1alpha1 "github.com/henderiw-nephio/kform/tools/apis/kform/block/v1alpha1"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/types"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cctx"
)

type Parser interface {
	Parse(ctx context.Context) (blockv1alpha1.ExecConfig, diag.Diagnostics)
}

func NewParser(ctx context.Context, kforms map[string]*blockv1alpha1.Kform) Parser {
	if kforms == nil {
		kforms = map[string]*blockv1alpha1.Kform{}
	}
	return &parser{
		controllerName: cctx.GetContextValue[string](ctx, "controllerName"),
		kforms:         kforms,
		//diagCh:         make(chan diag.Diagnostics),
		recorder: diag.NewRecorder(),
	}
}

type parser struct {
	controllerName string
	kforms         map[string]*blockv1alpha1.Kform

	recorder diag.Recorder
}

func (r *parser) Parse(ctx context.Context) (blockv1alpha1.ExecConfig, diag.Diagnostics) {
	m := &types.Module{}
	ctx = context.WithValue(ctx, types.CtxKeyModule, m)
	// validate syntax of each nested block
	// populate the resources
	// build dependencies
	// validate duplicate entries
	r.Validate(ctx)
	if r.recorder.Get().HasError() {
		fmt.Println("error", r.recorder.Get().Error())
		return nil, r.recorder.Get()
	}

	// resolve dependencies and providers
	r.Resolve(ctx)
	if r.recorder.Get().HasError() {
		fmt.Println("error", r.recorder.Get().Error())
		return nil, r.recorder.Get()
	}

	// connect vertices
	r.Connect(ctx)

	// Transitive reduction
	execCfg.GetVars().TransitiveReduction(ctx)

	return execCfg, r.recorder.Get()
}
