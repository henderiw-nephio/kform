package types

import (
	"context"
	"fmt"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/tools/pkg/recorder"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cctx"
)

func newBackend(ctx context.Context, n string) Block {
	return &backend{
		config: config{
			level:     1,
			blockType: GetBlockType(n),
			expectedKeywords: map[BlockContextKey]bool{
				BlockContextKeyAttributes: optional,
				BlockContextKeyConfig:     optional, // specific config per backend
			},
			expectedAttributes: map[string]bool{
				string(MetaArgumentSource): mandatory, // mandatory
			},
			recorder: cctx.GetContextValue[recorder.Recorder[diag.Diagnostic]](ctx, CtxKeyRecorder),
		},
	}
}

type backend struct {
	config
}

func (r *backend) UpdateModule(ctx context.Context) {
	r.initAndValidateBlockConfig(ctx)

	x := &Backend{
		config: r.config,
		name:   cctx.GetContextValue[string](ctx, CtxKeyVarName),
	}
	fmt.Println(x.fileName)
	if len(r.dependencies) > 0 {
		r.recorder.Record(diag.DiagFromErrWithContext(GetContext(ctx), fmt.Errorf("not expecting a dependency, got: %v", r.dependencies)))
	}

	// update module
	m := cctx.GetContextValue[*Module](ctx, CtxKeyModule)
	if m == nil {
		r.recorder.Record(diag.DiagFromErrWithContext(GetContext(ctx), fmt.Errorf("cannot add backend without module")))
		return
	}

	// NOTE only expecting 1 backend
	if m.Backend != nil {
		r.recorder.Record(diag.DiagFromErrWithContext(
			GetContext(ctx),
			fmt.Errorf("expecting only 1 backend config duplicate %s with %s",
				fmt.Sprintf("filename: %s, blockName: %s, blockType: %s", m.Backend.GetFileName(), m.Backend.GetBlockName(), m.Backend.GetBlockType()),
				fmt.Sprintf("filename: %s, blockName: %s, blockType: %s", x.GetFileName(), x.GetBlockName(), x.GetBlockType()),
			)))
	}
	m.Backend = x
}

type Backend struct {
	config

	name string
}

func (r *Backend) GetBlockName() string {
	return fmt.Sprintf("%s.%s", r.GetBlockType(), r.name)
}
