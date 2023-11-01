package types

import (
	"context"
	"fmt"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/tools/pkg/recorder"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cctx"
)

func newInput(ctx context.Context, n string) Block {
	return &input{
		config: config{
			level:     1,
			blockType: GetBlockType(n),
			expectedKeywords: map[BlockContextKey]bool{
				BlockContextKeyAttributes: optional,
				BlockContextKeyDefault:    optional,
			},
			expectedAttributes: map[string]bool{
				string(MetaArgumentSchema):      mandatory,
				string(MetaArgumentDescription): optional,
				string(MetaArgumentSensitive):   optional,
			},
			recorder: cctx.GetContextValue[recorder.Recorder[diag.Diagnostic]](ctx, CtxKeyRecorder),
		},
	}
}

type input struct {
	config
}

func (r *input) UpdateModule(ctx context.Context) {
	r.initAndValidateBlockConfig(ctx)

	x := &Input{
		config: r.config,
		name:   cctx.GetContextValue[string](ctx, CtxKeyVarName),
	}

	if len(r.dependencies) > 0 {
		r.recorder.Record(diag.DiagFromErrWithContext(GetContext(ctx), fmt.Errorf("not expecting a dependency, got: %v", r.dependencies)))
	}

	// update module
	m := cctx.GetContextValue[*Module](ctx, CtxKeyModule)
	if m == nil {
		r.recorder.Record(diag.DiagFromErrWithContext(GetContext(ctx), fmt.Errorf("cannot add backend without module")))
		return
	}

	if err := m.Inputs.Add(ctx, cache.NSN{Name: x.GetBlockName()}, x); err != nil {
		r.recorder.Record(diag.DiagFromErrWithContext(
			GetContext(ctx),
			fmt.Errorf("duplicate resource with fileName: %s, name: %s, type: %s",
				x.GetFileName(),
				x.GetBlockName(),
				x.GetBlockType(),
			)))
	}
}

type Input struct {
	config

	name string
}

func (r *Input) GetBlockName() string {
	return fmt.Sprintf("%s.%s", r.GetBlockType(), r.name)
}
