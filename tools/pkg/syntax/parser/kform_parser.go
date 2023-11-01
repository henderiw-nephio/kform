package parser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	kformpkgmetav1alpha1 "github.com/henderiw-nephio/kform/tools/apis/kform/pkg/meta/v1alpha1"
	"github.com/henderiw-nephio/kform/tools/pkg/dag"
	"github.com/henderiw-nephio/kform/tools/pkg/recorder"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/types"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cctx"
)

type KformParser interface {
	Parse(ctx context.Context)
	GetProviderRequirements(ctx context.Context) map[cache.NSN][]kformpkgmetav1alpha1.Provider
	GenerateDAG(ctx context.Context) map[cache.NSN]dag.DAG[*types.VertexContext]
}

func NewKformParser(ctx context.Context, path string) (KformParser, error) {
	recorder := cctx.GetContextValue[recorder.Recorder[diag.Diagnostic]](ctx, types.CtxKeyRecorder)
	if recorder == nil {
		return nil, fmt.Errorf("cannot parse without a recorder")
	}
	return &kformparser{
		rootModulePath: path,
		recorder:       recorder,
		modules:        cache.New[*types.Module](),
	}, nil
}

type kformparser struct {
	rootModulePath string
	rootModuleName cache.NSN
	recorder       recorder.Recorder[diag.Diagnostic]

	modules cache.Cache[*types.Module]
}

func (r *kformparser) Parse(ctx context.Context) {
	// we start by parsing the root module
	// if there are child modules they will be resolved concurrently
	r.rootModuleName = cache.NSN{Name: fmt.Sprintf("module.%s", filepath.Base(r.rootModulePath))}
	r.parseModule(ctx, r.rootModuleName, r.rootModulePath)
	if r.recorder.Get().HasError() {
		return
	}
	r.validateProviderConfigs(ctx)
	r.validateModuleCalls(ctx)
	r.validateUnreferencedProviderConfigs(ctx)
	r.validateUnreferencedProviderRequirements(ctx)
}

func (r *kformparser) parseModule(ctx context.Context, nsn cache.NSN, path string) {
	ctx = context.WithValue(ctx, types.CtxKeyModuleName, nsn)
	if r.rootModulePath == path {
		ctx = context.WithValue(ctx, types.CtxKeyModuleKind, types.ModuleKindRoot)
	} else {
		ctx = context.WithValue(ctx, types.CtxKeyModuleKind, types.ModuleKindChild)
	}
	p, err := NewModuleParser(ctx, path)
	if err != nil {
		r.recorder.Record(diag.DiagFromErr(err))
		return
	}

	m := p.Parse(ctx)
	if r.recorder.Get().HasError() {
		// if an error is found we stop processing
		return
	}
	r.modules.Add(ctx, nsn, m)

	// for each module that calls another module we need to continue
	// processing the new module
	var wg sync.WaitGroup
	for name, module := range m.ModuleCalls.List() {
		source := module.GetAttributes().GetSource()
		// TODO check local or remote module
		// The recursive modules always reference from the rootModule
		path := fmt.Sprintf("./%s", filepath.Join(".", r.rootModulePath, source))
		if _, err := os.Stat(path); err != nil {
			r.recorder.Record(diag.DiagErrorf("module %s, path %s does not exist", name, path))
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.parseModule(ctx, cache.NSN{Name: fmt.Sprintf("module.%s", filepath.Base(path))}, path)
		}()
	}
	wg.Wait()
}

func (r *kformparser) GetProviderRequirements(ctx context.Context) map[cache.NSN][]kformpkgmetav1alpha1.Provider {
	rootModule, err := r.modules.Get(r.rootModuleName)
	if err != nil {
		r.recorder.Record(diag.DiagErrorf("cannot validate provider requirements references, root module %s not found", r.rootModuleName.Name))
	}

	rootProviderConfigs := rootModule.ProviderConfigs.List()
	// delete the unreferenced provider configs from the provider configs
	unreferenceProviderConfigs := r.getUnReferencedProviderConfigs(ctx)
	for _, name := range unreferenceProviderConfigs {
		delete(rootProviderConfigs, cache.NSN{Name: name})
	}

	// we initialize all provider if they have aa req or not, if not the latest provider will be downloaded
	allprovreqs := map[cache.NSN][]kformpkgmetav1alpha1.Provider{}
	for nsn := range rootProviderConfigs {
		allprovreqs[nsn] = []kformpkgmetav1alpha1.Provider{}
	}

	for _, m := range r.modules.List() {
		provReqs := m.ProviderRequirements.List()
		for provNSN, provReq := range provReqs {
			if _, ok := rootProviderConfigs[provNSN]; ok {
				// since we initialized allprovreqs we dont need to check if the list is initialized
				allprovreqs[provNSN] = append(allprovreqs[provNSN], provReq)
			}
		}
	}
	return allprovreqs
}

func (r *kformparser) GenerateDAG(ctx context.Context) map[cache.NSN]dag.DAG[*types.VertexContext] {
	dags := map[cache.NSN]dag.DAG[*types.VertexContext]{}
	for nsn, m := range r.modules.List() {
		dags[nsn] = m.GenerateDAG(ctx)
	}
	return dags
}
