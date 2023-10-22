package syntax

import (
	"context"
	"fmt"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/tools/apis/kform/block/v1alpha1"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/exttypes"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/sctx"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cctx"
)

type Parser interface {
	Parse(ctx context.Context) (exttypes.ExecConfig, diag.Diagnostics)
}

func NewParser(ctx context.Context, kforms map[string]*v1alpha1.K8sForm) Parser {
	if kforms == nil {
		kforms = map[string]*v1alpha1.K8sForm{}
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
	kforms         map[string]*v1alpha1.K8sForm

	recorder diag.Recorder
}

func (r *parser) Parse(ctx context.Context) (exttypes.ExecConfig, diag.Diagnostics) {

	execCfg := exttypes.NewExecConfig()
	ctx = context.WithValue(ctx, sctx.CtxExecConfig, execCfg)
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
