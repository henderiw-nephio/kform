package types

import (
	"context"
	"sort"
	"strings"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	kformpkgmetav1alpha1 "github.com/henderiw-nephio/kform/tools/apis/kform/pkg/meta/v1alpha1"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
)

type Module struct {
	nsn      cache.NSN
	recorder  diag.Recorder
	SourceDir string
	Version   string

	Backend *Backend

	ProviderRequirements cache.Cache[kformpkgmetav1alpha1.Provider]
	ProviderConfigs      cache.Cache[*Provider]

	Inputs      cache.Cache[*Input]
	Locals      cache.Cache[*Local]
	Outputs     cache.Cache[*Output]
	Resources   cache.Cache[*Resource]
	ModuleCalls cache.Cache[*ModuleCall]
}

func NewModule(nsn cache.NSN, recorder diag.Recorder) *Module {
	return &Module{
		nsn:                 nsn,
		recorder:             recorder,
		ProviderRequirements: cache.New[kformpkgmetav1alpha1.Provider](),
		ProviderConfigs:      cache.New[*Provider](),

		Inputs:      cache.New[*Input](),
		Locals:      cache.New[*Local](),
		Outputs:     cache.New[*Output](),
		Resources:   cache.New[*Resource](),
		ModuleCalls: cache.New[*ModuleCall](),
	}
}

// interface signature
type DependencyBlock interface {
	GetDependencies() []string
	GetModDependencies() []string
	GetContext(string) string
	GetAttributes() *KformBlockAttributes
}

func (r *Module) GetModuleDependencies(ctx context.Context) []string {
	modDeps := []string{}
	for _, x := range r.Inputs.List() {
		modDeps = append(modDeps, x.GetModDependencies()...)
	}
	for _, x := range r.Outputs.List() {
		modDeps = append(modDeps, x.GetModDependencies()...)
	}
	for _, x := range r.Locals.List() {
		modDeps = append(modDeps, x.GetModDependencies()...)
	}
	for _, x := range r.ModuleCalls.List() {
		modDeps = append(modDeps, x.GetModDependencies()...)
	}
	for _, x := range r.Resources.List() {
		modDeps = append(modDeps, x.GetModDependencies()...)
	}
	return modDeps
}

func (r *Module) ResolveDAGDependencies(ctx context.Context) {
	for nsn, x := range r.Inputs.List() {
		r.resolveDependencies(ctx, nsn, x)
	}
	for nsn, x := range r.Outputs.List() {
		r.resolveDependencies(ctx, nsn, x)
	}
	for nsn, x := range r.Locals.List() {
		r.resolveDependencies(ctx, nsn, x)
	}
	for nsn, x := range r.ModuleCalls.List() {
		r.resolveDependencies(ctx, nsn, x)
	}
	for nsn, x := range r.Resources.List() {
		r.resolveDependencies(ctx, nsn, x)
	}
}

func (r *Module) resolveDependencies(ctx context.Context, nsn cache.NSN, v DependencyBlock) {
	for _, d := range v.GetDependencies() {
		switch strings.Split(d, ".")[0] {
		case string(BlockTypeInput):
			if _, err := r.Inputs.Get(cache.NSN{Name: d}); err != nil {
				r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(nsn.Name), "module: %s dependency resolution failed for %s", r.nsn.Name, d))
			}
		case string(BlockTypeOutput):
			if _, err := r.Outputs.Get(cache.NSN{Name: d}); err != nil {
				r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(nsn.Name), "module: %s dependency resolution failed for %s", r.nsn.Name, d))
			}
		case string(BlockTypeLocal):
			if _, err := r.Locals.Get(cache.NSN{Name: d}); err != nil {
				r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(nsn.Name), "module: %s dependency resolution failed for %s", r.nsn.Name, d))
			}
		case string(BlockTypeModule):
			if _, err := r.ModuleCalls.Get(cache.NSN{Name: d}); err != nil {
				r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(nsn.Name), "module: %s dependency resolution failed for %s", r.nsn.Name, d))
			}
		case "each":
			if v.GetAttributes().ForEach == nil {
				r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(nsn.Name), "module: %s dependency resolution failed each requires a for_each attribute dependency: %s", r.nsn.Name, d))
			}
		case "count":
			if v.GetAttributes().Count == nil {
				r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(nsn.Name), "module: %s dependency resolution failed each requires a count attribute dependency: %s", r.nsn.Name, d))
			}
		default:
			// resources - resource or data
			if _, err := r.Resources.Get(cache.NSN{Name: d}); err != nil {
				r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(nsn.Name), "module: %s dependency resolution failed for %s", r.nsn.Name, d))
			}
		}
	}
}

func (r *Module) ResolveResource2ProviderConfig(ctx context.Context) {
	// TBD do we need a config for all the providers - right now we assume yes
	for nsn, v := range r.Resources.List() {
		provider := v.GetProvider()
		if _, err := r.ProviderConfigs.Get(cache.NSN{Name: provider}); err != nil {
			r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(nsn.Name), "module: %s provider resolution failed for %s", r.nsn.Name, provider))
		}
	}
}

func (r *Module) ResolveProviderConfig2ProviderRequirements(ctx context.Context) {
	for nsn, v := range r.ProviderConfigs.List() {
		// reteurn the raw provider name w/o alias
		provider := v.GetName()
		if _, err := r.ProviderRequirements.Get(cache.NSN{Name: provider}); err != nil {
			r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(nsn.Name), "module: %s provider requirements resolution failed for %s", r.nsn.Name, provider))
		}
	}
}

func (r *Module) ValidateUnReferencedProviderConfigs(ctx context.Context) {
	providerConfigs := r.ProviderConfigs.List()
	for _, v := range r.Resources.List() {
		delete(providerConfigs, cache.NSN{Name: v.GetProvider()})
		if len(providerConfigs) == 0 {
			return
		}
	}
	if len(providerConfigs) > 0 {
		unreferenceProviders := make([]string, 0, len(providerConfigs))
		for nsn, v := range providerConfigs {
			unreferenceProviders = append(unreferenceProviders, v.GetContext(nsn.Name))
		}
		sort.Strings(unreferenceProviders)
		r.recorder.Record(diag.DiagWarnf("module: %s provider configs are unreferenced: %v", r.nsn.Name, unreferenceProviders))
	}
}

func (r *Module) ValidateUnReferencedProviderRequirements(ctx context.Context) {
	providerRequirements := r.ProviderRequirements.List()
	for _, v := range r.ProviderConfigs.List() {
		delete(providerRequirements, cache.NSN{Name: v.GetName()})
		if len(providerRequirements) == 0 {
			return
		}
	}
	if len(providerRequirements) > 0 {
		unreferenceProviderRequirements := make([]string, 0, len(providerRequirements))
		for nsn := range providerRequirements {
			unreferenceProviderRequirements = append(unreferenceProviderRequirements, nsn.Name)
		}
		sort.Strings(unreferenceProviderRequirements)
		r.recorder.Record(diag.DiagWarnf("module %s provider requirements are unreferenced: %v", r.nsn.Name, unreferenceProviderRequirements))
	}

}
