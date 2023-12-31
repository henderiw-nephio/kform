package types

import (
	"context"
	"fmt"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/tools/pkg/recorder"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cctx"
)

func newProvider(ctx context.Context, n string) Block {
	return &provider{
		config: config{
			level:     1,
			blockType: GetBlockType(n),
			expectedKeywords: map[BlockContextKey]bool{
				BlockContextKeyAttributes: mandatory,
				BlockContextKeyConfig:     mandatory,
			},
			expectedAttributes: map[string]bool{
				string(MetaArgumentSchema): mandatory,
			},
			recorder: cctx.GetContextValue[recorder.Recorder[diag.Diagnostic]](ctx, CtxKeyRecorder),
		},
	}
}

type provider struct {
	config
}

func (r *provider) UpdateModule(ctx context.Context) {
	r.initAndValidateBlockConfig(ctx)

	x := &ProviderConfig{
		config: r.config,
		name:   cctx.GetContextValue[string](ctx, CtxKeyVarName),
	}
	//x.GetAlias(ctx)

	if len(r.dependencies) > 0 {
		r.recorder.Record(diag.DiagFromErrWithContext(GetContext(ctx), fmt.Errorf("not expecting a dependency, got: %v", r.dependencies)))
	}

	// update module
	m := cctx.GetContextValue[*Module](ctx, CtxKeyModule)
	if m == nil {
		r.recorder.Record(diag.DiagFromErrWithContext(GetContext(ctx), fmt.Errorf("cannot add backend without module")))
		return
	}

	if err := m.ProviderConfigs.Add(ctx, cache.NSN{Name: x.GetBlockName()}, x); err != nil {
		r.recorder.Record(diag.DiagFromErrWithContext(
			GetContext(ctx),
			fmt.Errorf("duplicate resource with fileName: %s, name: %s, type: %s",
				x.GetFileName(),
				x.GetBlockName(),
				x.GetBlockType(),
			)))
	}
}

type ProviderConfig struct {
	config

	name string
}

func (r *ProviderConfig) GetBlockName() string {
	return r.name
}

func (r *ProviderConfig) GetName() string {
	return r.name
}
