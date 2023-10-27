package types

import (
	"context"
	"strings"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	kformpkgmetav1alpha1 "github.com/henderiw-nephio/kform/tools/apis/kform/pkg/meta/v1alpha1"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
)

type Module struct {
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

func NewModule(recorder diag.Recorder) *Module {
	return &Module{
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
	GetContext(string) string
	GetAttributes() *KformBlockAttributes
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
				r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(nsn.Name), "dependency resolution failed for %s", d))
			}
		case string(BlockTypeOutput):
			if _, err := r.Outputs.Get(cache.NSN{Name: d}); err != nil {
				r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(nsn.Name), "dependency resolution failed for %s", d))
			}
		case string(BlockTypeLocal):
			if _, err := r.Locals.Get(cache.NSN{Name: d}); err != nil {
				r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(nsn.Name), "dependency resolution failed for %s", d))
			}
		case string(BlockTypeModule):
			if _, err := r.ModuleCalls.Get(cache.NSN{Name: d}); err != nil {
				r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(nsn.Name), "dependency resolution failed for %s", d))
			}
		case "each":
			if v.GetAttributes().ForEach == nil {
				r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(nsn.Name), "dependency resolution failed each requires a for_each attribute", d))
			}
		case "count":
			if v.GetAttributes().Count == nil {
				r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(nsn.Name), "dependency resolution failed each requires a count attribute", d))
			}
		default:
			// resources - resource or data
			if _, err := r.Resources.Get(cache.NSN{Name: d}); err != nil {
				r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(nsn.Name), "dependency resolution failed for %s", d))
			}
		}
	}
}

func (r *Module) ResolveResource2ProviderConfig(ctx context.Context) {
	// TBD do we need a config for all the providers - right now we assume yes
	for nsn, v := range r.Resources.List() {
		provider := v.GetProvider()
		if _, err := r.ProviderConfigs.Get(cache.NSN{Name: provider}); err != nil {
			r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(nsn.Name), "provider resolution failed for %s", provider))
		}
	}
}

func (r *Module) ResolveProviderConfig2ProviderRequirements(ctx context.Context) {
	for nsn, v := range r.ProviderConfigs.List() {
		// reteurn the raw provider name w/o alias
		provider := v.GetName()
		if _, err := r.ProviderRequirements.Get(cache.NSN{Name: provider}); err != nil {
			r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(nsn.Name), "provider requirements resolution failed for %s", provider))
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
	unreferenceProviders := make([]string, 0, len(providerConfigs))
	for nsn, v := range providerConfigs {
		unreferenceProviders = append(unreferenceProviders, v.GetContext(nsn.Name))
	}
	r.recorder.Record(diag.DiagWarnf("provider configs are unreferenced: %v", unreferenceProviders))
}

func (r *Module) ValidateUnReferencedProviderRequirements(ctx context.Context) {
	providerRequirements := r.ProviderRequirements.List()
	for _, v := range r.ProviderConfigs.List() {
		delete(providerRequirements, cache.NSN{Name: v.GetName()})
		if len(providerRequirements) == 0 {
			return
		}
	}
	//fmt.Println("unreferenceProviderRequirements", providerRequirements)
	unreferenceProviderRequirements := make([]string, 0, len(providerRequirements))
	for nsn := range providerRequirements {
		unreferenceProviderRequirements = append(unreferenceProviderRequirements, nsn.Name)
	}
	r.recorder.Record(diag.DiagWarnf("provider requirements are unreferenced: %v", unreferenceProviderRequirements))
}
