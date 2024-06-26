package types

import (
	"context"
	"fmt"
	"strings"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/tools/pkg/recorder"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cctx"
)

func newResource(ctx context.Context, n string) Block {
	return &resource{
		config: config{
			level:     2,
			blockType: GetBlockType(n),
			expectedKeywords: map[BlockContextKey]bool{
				BlockContextKeyAttributes: mandatory,
				BlockContextKeyConfig:     mandatory,
			},
			expectedAttributes: map[string]bool{
				string(MetaArgumentSchema): mandatory,
				//string(MetaArgumentAlias):         optional,
				string(MetaArgumentProvider):      optional,
				string(MetaArgumentDependsOn):     optional,
				string(MetaArgumentCount):         optional,
				string(MetaArgumentForEach):       optional,
				string(MetaArgumentLifecycle):     optional,
				string(MetaArgumentPrecondition):  optional,
				string(MetaArgumentPostcondition): optional,
				string(MetaArgumentConnection):    optional,
				string(MetaArgumentProvisioner):   optional,
				string(MetaArgumentDescription):   optional,
				string(MetaArgumentSensitive):     optional,
				string(MetaArgumentValidation):    optional,
			},
			recorder: cctx.GetContextValue[recorder.Recorder[diag.Diagnostic]](ctx, CtxKeyRecorder),
		},
	}
}

type resource struct {
	config
}

func (r *resource) UpdateModule(ctx context.Context) {
	r.initAndValidateBlockConfig(ctx)

	x := &Resource{
		config:       r.config,
		resourceType: cctx.GetContextValue[string](ctx, CtxKeyVarType),
		resourceID:   cctx.GetContextValue[string](ctx, CtxKeyVarName),
	}

	if err := validateResourceSyntax(ResourceType, x.resourceType); err != nil {
		r.recorder.Record(diag.DiagFromErrWithContext(GetContext(ctx), err))
		return
	}
	x.getProvider(ctx)

	// update module
	m := cctx.GetContextValue[*Module](ctx, CtxKeyModule)
	if m == nil {
		r.recorder.Record(diag.DiagFromErrWithContext(GetContext(ctx), fmt.Errorf("cannot add backend without module")))
		return
	}

	if err := m.Resources.Add(ctx, cache.NSN{Name: x.GetBlockName()}, x); err != nil {
		r.recorder.Record(diag.DiagFromErrWithContext(
			GetContext(ctx),
			fmt.Errorf("duplicate resource with fileName: %s, name: %s, type: %s",
				x.GetFileName(),
				x.GetBlockName(),
				x.GetBlockType(),
			)))
	}
}

type Resource struct {
	config

	resourceType string
	resourceID   string
	provider     string
}

func (r *Resource) GetBlockName() string {
	return fmt.Sprintf("%s.%s", r.resourceType, r.resourceID)
}

func (r *Resource) getProvider(_ context.Context) {
	// the provider is the first element in the resourceType of the resource
	r.provider = strings.Split(r.resourceType, "_")[0]
	if r.KformBlockContext.Attributes != nil && r.KformBlockContext.Attributes.Provider != nil {
		r.provider = *r.KformBlockContext.Attributes.Provider
	}
}

func (r *Resource) GetProvider() string {
	return r.provider
}
