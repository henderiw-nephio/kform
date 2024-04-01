package types

import (
	"context"
	"fmt"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/tools/pkg/recorder"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cctx"
)

func newModule(ctx context.Context, n string) Block {
	return &module{
		config: config{
			level:     1,
			blockType: GetBlockType(n),
			expectedKeywords: map[BlockContextKey]bool{
				BlockContextKeyAttributes:  mandatory,
				BlockContextKeyInputParams: mandatory,
			},
			expectedAttributes: map[string]bool{
				string(MetaArgumentSource):    mandatory,
				string(MetaArgumentProviders): optional,
				string(MetaArgumentDependsOn): optional,
				string(MetaArgumentCount):     optional,
				string(MetaArgumentForEach):   optional,
			},
			recorder: cctx.GetContextValue[recorder.Recorder[diag.Diagnostic]](ctx, CtxKeyRecorder),
		},
	}
}

type module struct {
	config
}

func (r *module) UpdateModule(ctx context.Context) {
	r.initAndValidateBlockConfig(ctx)

	x := &ModuleCall{
		config: r.config,
		name:   cctx.GetContextValue[string](ctx, CtxKeyVarName),
	}
	x.getSource(ctx)
	x.getProviders(ctx)

	// update module
	m := cctx.GetContextValue[*Module](ctx, CtxKeyModule)
	if m == nil {
		r.recorder.Record(diag.DiagFromErrWithContext(GetContext(ctx), fmt.Errorf("cannot add backend without module")))
		return
	}

	if err := m.ModuleCalls.Add(ctx, cache.NSN{Name: x.GetBlockName()}, x); err != nil {
		r.recorder.Record(diag.DiagFromErrWithContext(
			GetContext(ctx),
			fmt.Errorf("duplicate resource with fileName: %s, name: %s, type: %s",
				x.GetFileName(),
				x.GetBlockName(),
				x.GetBlockType(),
			)))
	}
}

type ModuleCall struct {
	config

	name   string
	source string
	// providers key is the target provider, value is the source provider
	providers map[string]string
}

func (r *ModuleCall) GetBlockName() string {
	return fmt.Sprintf("%s.%s", r.blockType, r.name)
}

func (r *ModuleCall) getSource(_ context.Context) {
	if r.KformBlockContext.Attributes != nil && r.KformBlockContext.Attributes.Source != nil {
		r.source = *r.KformBlockContext.Attributes.Source
	}
}

func (r *ModuleCall) getProviders(_ context.Context) {
	if r.KformBlockContext.Attributes != nil && r.KformBlockContext.Attributes.Providers != nil {
		r.providers = r.KformBlockContext.Attributes.Providers
	} else {
		r.providers = map[string]string{}
	}
}

func (r *ModuleCall) GetProviders() map[string]string {
	return r.providers
}
