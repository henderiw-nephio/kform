package syntax

import (
	"context"
	"fmt"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/syntax/apis/k8sform/v1alpha1"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cctx"
)

type Parser interface {
	Parse(ctx context.Context) (ExecConfig, diag.Diagnostics)
}

func NewParser(ctx context.Context, kforms []*v1alpha1.K8sFormCtx) Parser {
	if kforms == nil {
		kforms = []*v1alpha1.K8sFormCtx{}
	}
	return &parser{
		controllerName: cctx.GetContextValue[string](ctx, "controllerName"),
		kforms:         kforms,
		//diagCh:         make(chan diag.Diagnostics),
		recorder: diag.NewRecorder(),
		//providers: dag.New[kformtypes.Provider](),
		//vars:      dag.New[kformtypes.Variable](),
	}
}

type parser struct {
	controllerName string
	kforms         []*v1alpha1.K8sFormCtx

	//diagCh         chan diag.Diagnostics
	recorder diag.Recorder

	//providers dag.DAG[kformtypes.Provider]
	//vars      dag.DAG[kformtypes.Variable]
}

func (r *parser) Parse(ctx context.Context) (ExecConfig, diag.Diagnostics) {

	execCfg := NewExecConfig()
	ctx = context.WithValue(ctx, CtxExecConfig, execCfg)
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
