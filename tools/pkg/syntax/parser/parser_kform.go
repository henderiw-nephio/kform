package parser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/types"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cctx"
)

type KformParser interface {
	Parse(ctx context.Context)
	GetModules() map[cache.NSN]*types.Module
}

func NewKformParser(ctx context.Context, path string) (KformParser, error) {
	recorder := cctx.GetContextValue[diag.Recorder](ctx, types.CtxKeyRecorder)
	if recorder == nil {
		return nil, fmt.Errorf("cannot parse without a recorder")
	}
	return &kformparser{
		rootPath: path,
		recorder: recorder,
		modules:  cache.New[*types.Module](),
	}, nil
}

type kformparser struct {
	rootPath string
	recorder diag.Recorder

	modules cache.Cache[*types.Module]
}

func (r *kformparser) GetModules() map[cache.NSN]*types.Module {
	return r.modules.List()
}

func (r *kformparser) Parse(ctx context.Context) {
	// we start by parsing the root module
	// if there are child modules they will be resolved concurrently
	r.parseModule(ctx, cache.NSN{Name: fmt.Sprintf("module.%s",filepath.Base(r.rootPath))}, r.rootPath)
}

func (r *kformparser) parseModule(ctx context.Context, nsn cache.NSN, path string) {
	ctx = context.WithValue(ctx, types.CtxKeyModuleName, nsn)
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

	var wg sync.WaitGroup
	for name, module := range m.ModuleCalls.List() {
		source := module.GetAttributes().GetSource()
		// TODO check local or remote module
		// The recursive modules always reference from the rootModule
		path := fmt.Sprintf("./%s", filepath.Join(".", r.rootPath, source))
		if _, err := os.Stat(path); err != nil {
			r.recorder.Record(diag.DiagErrorf("module %s, path %s does not exist", name, path))
			continue
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			r.parseModule(ctx, cache.NSN{Name: fmt.Sprintf("module.%s",filepath.Base(path))}, path)
		}()

	}
	wg.Wait()
}
