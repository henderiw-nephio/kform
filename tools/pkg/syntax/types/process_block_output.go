package types

import (
	"context"
	"fmt"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/tools/pkg/recorder"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cctx"
)

func newOutput(ctx context.Context, n string) Block {
	return &output{
		config: config{
			level:     1,
			blockType: GetBlockType(n),
			expectedKeywords: map[BlockContextKey]bool{
				BlockContextKeyAttributes: mandatory,
				BlockContextKeyValue:      mandatory,
			},
			expectedAttributes: map[string]bool{
				string(MetaArgumentSchema):        mandatory,
				string(MetaArgumentDescription):   optional,
				string(MetaArgumentSensitive):     optional,
				string(MetaArgumentDependsOn):     optional,
				string(MetaArgumentCount):         optional,
				string(MetaArgumentForEach):       optional,
				string(MetaArgumentPrecondition):  optional,
				string(MetaArgumentPostcondition): optional,
				string(MetaArgumentValidation):    optional,
			},
			recorder: cctx.GetContextValue[recorder.Recorder[diag.Diagnostic]](ctx, CtxKeyRecorder),
		},
	}
}

type output struct {
	config
}

func (r *output) UpdateModule(ctx context.Context) {
	r.initAndValidateBlockConfig(ctx)

	x := &Output{
		config: r.config,
		name:   cctx.GetContextValue[string](ctx, CtxKeyVarName),
	}

	//fmt.Println("output processing: value", x.name, x.config.Value)

	// update module
	m := cctx.GetContextValue[*Module](ctx, CtxKeyModule)
	if m == nil {
		r.recorder.Record(diag.DiagFromErrWithContext(GetContext(ctx), fmt.Errorf("cannot add backend without module")))
		return
	}

	if err := m.Outputs.Add(ctx, cache.NSN{Name: x.GetBlockName()}, x); err != nil {
		r.recorder.Record(diag.DiagFromErrWithContext(
			GetContext(ctx),
			fmt.Errorf("duplicate resource with fileName: %s, name: %s, type: %s",
				x.GetFileName(),
				x.GetBlockName(),
				x.GetBlockType(),
			)))
	}
}

type Output struct {
	config
	name string
}

func (r *Output) GetBlockName() string {
	// TBD how we handle this
	return fmt.Sprintf("%s.%s", r.GetBlockType(), r.name)
}
