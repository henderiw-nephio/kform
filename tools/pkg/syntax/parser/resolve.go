package parser

import (
	"context"

	"github.com/henderiw-nephio/kform/tools/pkg/syntax/types"
)

func (r *parser) resolve(ctx context.Context, m *types.Module) {
	m.ResolveDAGDependencies(ctx)
	m.ResolveResource2ProviderConfig(ctx)
	m.ResolveProviderConfig2ProviderRequirements(ctx)

	// TODO check if we have too many providers and required providers that are not referenced
	m.ValidateUnReferencedProviderConfigs(ctx)
	m.ValidateUnReferencedProviderRequirements(ctx)

}
