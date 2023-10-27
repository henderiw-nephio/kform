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

type Parser interface {
	Parse(ctx context.Context) *types.Module
}

func NewModuleParser(ctx context.Context, path string) (Parser, error) {
	recorder := cctx.GetContextValue[diag.Recorder](ctx, types.CtxKeyRecorder)
	if recorder == nil {
		return nil, fmt.Errorf("cannot parse without a recorder")
	}
	return &parser{
		path:     path,
		fsys:     fsys.NewDiskFS(path),
		recorder: recorder,
	}, nil
}

type parser struct {
	path     string
	fsys     fsys.FS
	recorder diag.Recorder
}

// Parse
func (r *parser) Parse(ctx context.Context) *types.Module {

	kf, kforms, err := r.getKforms(ctx)
	if err != nil {
		r.recorder.Record(diag.DiagErrorf("cannot get kfile and/or kforms for this path: %s, err: %s", r.path, err.Error()))
		return nil
	}
	m := types.NewModule(r.recorder)
	// add the required providers in the module
	for providerName, provider := range kf.Spec.RequiredProviders {
		if err := m.ProviderRequirements.Add(
			ctx,
			cache.NSN{Name: providerName},
			provider,
		); err != nil {
			r.recorder.Record(diag.DiagErrorf("cannot add provider %s in required providers, err: %s", providerName, err.Error()))
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
