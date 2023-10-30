package parser

import (
	"context"
	"fmt"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/tools/pkg/fsys"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/types"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cctx"
)

type ModuleParser interface {
	Parse(ctx context.Context) *types.Module
}

// TODO moduleName
func NewModuleParser(ctx context.Context, path string) (ModuleParser, error) {
	recorder := cctx.GetContextValue[diag.Recorder](ctx, types.CtxKeyRecorder)
	if recorder == nil {
		return nil, fmt.Errorf("cannot parse without a recorder")
	}
	return &moduleparser{
		nsn:      cctx.GetContextValue[cache.NSN](ctx, types.CtxKeyModuleName),
		kind:     cctx.GetContextValue[types.ModuleKind](ctx, types.CtxKeyModuleKind),
		path:     path,
		fsys:     fsys.NewDiskFS(path),
		recorder: recorder,
	}, nil
}

type moduleparser struct {
	nsn      cache.NSN
	kind     types.ModuleKind
	path     string
	fsys     fsys.FS
	recorder diag.Recorder
}

// Parse
func (r *moduleparser) Parse(ctx context.Context) *types.Module {

	kf, kforms, err := r.getKforms(ctx)
	if err != nil {
		r.recorder.Record(diag.DiagErrorf("cannot get kfile and/or kforms for this path: %s, err: %s", r.path, err.Error()))
		return nil
	}
	if kf == nil {
		r.recorder.Record(diag.DiagErrorf("cannot parse module with a kform file"))
		return nil
	}
	m := types.NewModule(
		cctx.GetContextValue[cache.NSN](ctx, types.CtxKeyModuleName),
		cctx.GetContextValue[types.ModuleKind](ctx, types.CtxKeyModuleKind),
		r.recorder)
	// add the required providers in the module
	for providerRawName, providerReq := range kf.Spec.ProviderRequirements {
		if err := m.ProviderRequirements.Add(
			ctx,
			cache.NSN{Name: providerRawName},
			providerReq,
		); err != nil {
			r.recorder.Record(diag.DiagErrorf("cannot add provider %s in provider requirements, err: %s", providerRawName, err.Error()))
		}
	}

	ctx = context.WithValue(ctx, types.CtxKeyModule, m)
	r.validate(ctx, kforms)
	if r.recorder.Get().HasError() {
		return nil
	}
	r.resolve(ctx, m)
	if r.recorder.Get().HasError() {
		return nil
	}
	return m
}
