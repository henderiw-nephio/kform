package parser

import (
	"context"

	"github.com/henderiw-nephio/kform/tools/pkg/syntax/types"
)

func (r *moduleparser) resolve(ctx context.Context, m *types.Module) {
	// creates a list
	m.ResolveDAGDependencies(ctx)

	// for each resource we should have a required provider in root/child modules
	//m.ResolveResource2ProviderRequirements(ctx)
	if m.Kind == types.ModuleKindRoot {
		// validate that for each provider in a resource there is a related provider config
		m.ResolveResource2ProviderConfig(ctx)
		// validate that for each provider confif we have defined the provider requirements
		//m.ResolveProviderConfig2ProviderRequirements(ctx)
	} else {
		// a child module must not have provider configs
		m.ValidateChildProviderConfigs(ctx)
	}

	// check if we have too many providers and required providers that are not referenced
	/*
	if m.Kind == types.ModuleKindRoot {
		m.ValidateUnReferencedProviderConfigs(ctx)
		m.ValidateUnReferencedProviderRequirements(ctx)
	}
	*/
}
