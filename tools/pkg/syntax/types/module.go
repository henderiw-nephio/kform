package types

import (
	"context"
	"strings"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	kformpkgmetav1alpha1 "github.com/henderiw-nephio/kform/tools/apis/kform/pkg/meta/v1alpha1"
	"github.com/henderiw-nephio/kform/tools/pkg/dag"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
	"github.com/henderiw-nephio/kform/tools/pkg/util/sets"
)

type Module struct {
	nsn       cache.NSN
	Kind      ModuleKind
	recorder  diag.Recorder
	SourceDir string
	Version   string

	Backend *Backend

	ProviderRequirements cache.Cache[kformpkgmetav1alpha1.Provider]
	ProviderConfigs      cache.Cache[*ProviderConfig]
	//Providers            map[string][]string

	Inputs      cache.Cache[*Input]
	Locals      cache.Cache[*Local]
	Outputs     cache.Cache[*Output]
	Resources   cache.Cache[*Resource]
	ModuleCalls cache.Cache[*ModuleCall]
}

func NewModule(nsn cache.NSN, kind ModuleKind, recorder diag.Recorder) *Module {
	return &Module{
		nsn:                  nsn,
		Kind:                 kind,
		recorder:             recorder,
		ProviderRequirements: cache.New[kformpkgmetav1alpha1.Provider](),
		ProviderConfigs:      cache.New[*ProviderConfig](),

		Inputs:      cache.New[*Input](),
		Locals:      cache.New[*Local](),
		Outputs:     cache.New[*Output](),
		Resources:   cache.New[*Resource](),
		ModuleCalls: cache.New[*ModuleCall](),
	}
}

func (r *Module) GetProvidersFromResources(ctx context.Context) sets.Set[cache.NSN] {
	providers := sets.New[cache.NSN]()
	for _, resource := range r.Resources.List() {
		providers.Insert(cache.NSN{Name: resource.GetProvider()})
	}
	return providers
}

// interface signature
type DependencyBlock interface {
	GetDependencies() map[string]string
	GetModDependencies() map[string]string
	GetContext(string) string
	GetAttributes() *KformBlockAttributes
}

type modDeps map[string]string

func (r modDeps) add(d map[string]string) {
	for k, v := range d {
		r[k] = v
	}
}

func (r *Module) GetModuleDependencies(ctx context.Context) map[string]string {
	modDeps := modDeps{}
	for _, x := range r.Inputs.List() {
		modDeps.add(x.GetModDependencies())
	}
	for _, x := range r.Outputs.List() {
		modDeps.add(x.GetModDependencies())
	}
	for _, x := range r.Locals.List() {
		modDeps.add(x.GetModDependencies())
	}
	for _, x := range r.ModuleCalls.List() {
		modDeps.add(x.GetModDependencies())
	}
	for _, x := range r.Resources.List() {
		modDeps.add(x.GetModDependencies())
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
	for d, dctx := range v.GetDependencies() {
		switch strings.Split(d, ".")[0] {
		case string(BlockTypeInput):
			if _, err := r.Inputs.Get(cache.NSN{Name: d}); err != nil {
				r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(nsn.Name), "%s module: %s dependency resolution failed for %s, ctx: %s", r.Kind, r.nsn.Name, d, dctx))
			}
		case string(BlockTypeOutput):
			if _, err := r.Outputs.Get(cache.NSN{Name: d}); err != nil {
				r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(nsn.Name), "%s module: %s dependency resolution failed for %s, ctx: %s", r.Kind, r.nsn.Name, d, dctx))
			}
		case string(BlockTypeLocal):
			if _, err := r.Locals.Get(cache.NSN{Name: d}); err != nil {
				r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(nsn.Name), "%s module: %s dependency resolution failed for %s, ctx: %s", r.Kind, r.nsn.Name, d, dctx))
			}
		case string(BlockTypeModule):
			if _, err := r.ModuleCalls.Get(cache.NSN{Name: d}); err != nil {
				r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(nsn.Name), "%s module: %s dependency resolution failed for %s, ctx: %s", r.Kind, r.nsn.Name, d, dctx))
			}
		case "each":
			if v.GetAttributes().ForEach == nil {
				r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(nsn.Name), "%s module: %s dependency resolution failed each requires a for_each attribute dependency: %s, ctx: %s", r.Kind, r.nsn.Name, d, dctx))
			}
		case "count":
			if v.GetAttributes().Count == nil {
				r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(nsn.Name), "%s module: %s dependency resolution failed each requires a count attribute dependency: %s, ctx: %s", r.Kind, r.nsn.Name, d, dctx))
			}
		default:
			// resources - resource or data
			if _, err := r.Resources.Get(cache.NSN{Name: d}); err != nil {
				r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(nsn.Name), "%s module: %s dependency resolution failed for %s, ctx: %s", r.Kind, r.nsn.Name, d, dctx))
			}
		}
	}
}

/*
func (r *Module) ResolveResource2ProviderRequirements(ctx context.Context) {
	for nsn, v := range r.Resources.List() {
		// the raw provider reference (w/o alias tag) is the first segment of the provider name
		// <provider>_<alias>
		rawProvider := strings.Split(v.GetProvider(), "_")[0]
		if _, err := r.ProviderRequirements.Get(cache.NSN{Name: rawProvider}); err != nil {
			r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(nsn.Name), "%s module: %s raw provider resolution resource2providerReq failed for %s", r.Kind, r.nsn.Name, rawProvider))
		}
	}
}
*/

func (r *Module) ResolveResource2ProviderConfig(ctx context.Context) {
	for nsn, v := range r.Resources.List() {
		provider := v.GetProvider()
		if _, err := r.ProviderConfigs.Get(cache.NSN{Name: provider}); err != nil {
			r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(nsn.Name), "%s module: %s provider resolution resource2providerConfig failed for %s", r.Kind, r.nsn.Name, provider))
		}
	}
}

/*
func (r *Module) ResolveProviderConfig2ProviderRequirements(ctx context.Context) {
	for nsn, v := range r.ProviderConfigs.List() {
		// reteurn the raw provider name w/o alias
		provider := v.GetName()
		if _, err := r.ProviderRequirements.Get(cache.NSN{Name: provider}); err != nil {
			r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(nsn.Name), "%s module: %s provider requirements providerConfig2providerReq resolution failed for %s", r.Kind, r.nsn.Name, provider))
		}
	}
}
*/

/*
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
		r.recorder.Record(diag.DiagWarnf("%s module: %s provider configs are unreferenced: %v", r.Kind, r.nsn.Name, unreferenceProviders))
	}
}
*/

/*
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
		r.recorder.Record(diag.DiagWarnf("%s module %s provider requirements are unreferenced: %v", r.Kind, r.nsn.Name, unreferenceProviderRequirements))
	}
}
*/

func (r *Module) GenerateDAG(ctx context.Context) {
	// add the vertices with the right VertexContext to the dag
	dag := r.generateDAG(ctx)
	// connect the dag based on the depdenencies
	for n, v := range dag.GetVertices() {
		for d := range v.GetDependencies() {
			dag.Connect(ctx, d, n)
		}
	}
	// optimize the dag by removing the transitive connection in the dag
	dag.TransitiveReduction(ctx)
}

func (r *Module) generateDAG(ctx context.Context) dag.DAG[*VertexContext] {
	d := dag.New[*VertexContext]()
	for nsn, x := range r.Inputs.List() {
		d.AddVertex(ctx, nsn.Name, &VertexContext{
			FileName:        x.fileName,
			Module:          x.moduleName,
			BlockType:       x.blockType,
			GVK:             x.gvk,
			BlockContext:    x.KformBlockContext,
			Dependencies:    x.dependencies,
			ModDependencies: x.modDependencies,
		})
	}
	for nsn, x := range r.Outputs.List() {
		d.AddVertex(ctx, nsn.Name, &VertexContext{
			FileName:        x.fileName,
			Module:          x.moduleName,
			BlockType:       x.blockType,
			GVK:             x.gvk,
			BlockContext:    x.KformBlockContext,
			Dependencies:    x.dependencies,
			ModDependencies: x.modDependencies,
		})
	}
	for nsn, x := range r.Locals.List() {
		d.AddVertex(ctx, nsn.Name, &VertexContext{
			FileName:        x.fileName,
			Module:          x.moduleName,
			BlockType:       x.blockType,
			GVK:             x.gvk,
			BlockContext:    x.KformBlockContext,
			Dependencies:    x.dependencies,
			ModDependencies: x.modDependencies,
		})
	}
	for nsn, x := range r.ModuleCalls.List() {
		d.AddVertex(ctx, nsn.Name, &VertexContext{
			FileName:        x.fileName,
			Module:          x.moduleName,
			BlockType:       x.blockType,
			GVK:             x.gvk,
			BlockContext:    x.KformBlockContext,
			Dependencies:    x.dependencies,
			ModDependencies: x.modDependencies,
		})
	}
	for nsn, x := range r.Resources.List() {
		d.AddVertex(ctx, nsn.Name, &VertexContext{
			FileName:        x.fileName,
			Module:          x.moduleName,
			BlockType:       x.blockType,
			GVK:             x.gvk,
			BlockContext:    x.KformBlockContext,
			Dependencies:    x.dependencies,
			ModDependencies: x.modDependencies,
			Provider:        x.provider,
		})
	}
	return d
}

func (r *Module) ValidateChildProviderConfigs(ctx context.Context) {
	if r.Kind == ModuleKindChild {
		providerConfigs := []string{}
		for nsn := range r.ProviderConfigs.List() {
			providerConfigs = append(providerConfigs, nsn.Name)
		}
		if len(providerConfigs) > 0 {
			r.recorder.Record(diag.DiagErrorf("%s module: %s child modules cannot have provider configs, provider configs must come from the root module, providers: %v", r.Kind, r.nsn.Name, providerConfigs))
		}
	}
}

func (r *Module) GetProviderRequirements(ctx context.Context) map[cache.NSN]kformpkgmetav1alpha1.Provider {
	return r.ProviderRequirements.List()
}