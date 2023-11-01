package types

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	kformpkgmetav1alpha1 "github.com/henderiw-nephio/kform/tools/apis/kform/pkg/meta/v1alpha1"
	"github.com/henderiw-nephio/kform/tools/pkg/dag"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio"
	"github.com/henderiw-nephio/kform/tools/pkg/recorder"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
	"github.com/henderiw-nephio/kform/tools/pkg/util/sets"
)

type Module struct {
	nsn       cache.NSN
	Kind      ModuleKind
	recorder  recorder.Recorder[diag.Diagnostic]
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

func NewModule(nsn cache.NSN, kind ModuleKind, recorder recorder.Recorder[diag.Diagnostic]) *Module {
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
				r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(nsn.Name), "%s module: %s dependency resolution failed for %s, ctx: %s, err: %s", r.Kind, r.nsn.Name, d, dctx, err.Error()))
			}
		case string(BlockTypeOutput):
			if _, err := r.Outputs.Get(cache.NSN{Name: d}); err != nil {
				r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(nsn.Name), "%s module: %s dependency resolution failed for %s, ctx: %s, err: %s", r.Kind, r.nsn.Name, d, dctx, err.Error()))
			}
		case string(BlockTypeLocal):
			if _, err := r.Locals.Get(cache.NSN{Name: d}); err != nil {
				r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(nsn.Name), "%s module: %s dependency resolution failed for %s, ctx: %s, err: %s", r.Kind, r.nsn.Name, d, dctx, err.Error()))
			}
		case string(BlockTypeModule):
			if _, err := r.ModuleCalls.Get(cache.NSN{Name: d}); err != nil {
				r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(nsn.Name), "%s module: %s dependency resolution failed for %s, ctx: %s, err: %s", r.Kind, r.nsn.Name, d, dctx, err.Error()))
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
				r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(nsn.Name), "%s module: %s dependency resolution failed for %s, ctx: %s, err: %s", r.Kind, r.nsn.Name, d, dctx, err.Error()))
			}
		}
	}
}

func (r *Module) ResolveResource2ProviderConfig(ctx context.Context) {
	for nsn, v := range r.Resources.List() {
		provider := v.GetProvider()
		if _, err := r.ProviderConfigs.Get(cache.NSN{Name: provider}); err != nil {
			r.recorder.Record(diag.DiagErrorfWithContext(v.GetContext(nsn.Name), "%s module: %s provider resolution resource2providerConfig failed for %s, err: %s", r.Kind, r.nsn.Name, provider, err.Error()))
		}
	}
}

func (r *Module) GenerateDAG(ctx context.Context) dag.DAG[*VertexContext] {
	// add the vertices with the right VertexContext to the dag
	d := r.generateDAG(ctx)
	// connect the dag based on the depdenencies
	for n, v := range d.GetVertices() {
		deps := v.GetBlockDependencies()
		fmt.Println("block dependencies", n, deps)
		for dep := range deps {
			d.Connect(ctx, dep, n)
		}
		if n != dag.Root {
			if len(deps) == 0 {
				d.Connect(ctx, dag.Root, n)
			}
		}

	}
	// optimize the dag by removing the transitive connection in the dag
	//d.TransitiveReduction(ctx)
	return d
}

func (r *Module) generateDAG(ctx context.Context) dag.DAG[*VertexContext] {
	d := dag.New[*VertexContext]()

	d.AddVertex(ctx, dag.Root, &VertexContext{
		FileName:     filepath.Join(r.SourceDir, pkgio.PkgFileMatch[0]),
		ModuleName:   r.nsn.Name,
		BlockType:    dag.Root,
		BlockContext: KformBlockContext{},
	})

	for nsn, x := range r.Inputs.List() {
		d.AddVertex(ctx, nsn.Name, &VertexContext{
			FileName:        x.fileName,
			ModuleName:      x.moduleName,
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
			ModuleName:      x.moduleName,
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
			ModuleName:      x.moduleName,
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
			ModuleName:      x.moduleName,
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
			ModuleName:      x.moduleName,
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
