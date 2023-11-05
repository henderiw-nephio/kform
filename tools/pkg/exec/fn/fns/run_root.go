package fns

import (
	"context"

	"github.com/henderiw-nephio/kform/tools/pkg/exec/fn"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/vctx"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/types"
	"github.com/henderiw/logger/log"
)

func NewRootFn(cfg *Config) fn.BlockInstanceRunner {
	return &root{
		rootModuleName: cfg.RootModuleName,
	}
}

type root struct {
	rootModuleName string
}

func (r *root) Run(ctx context.Context, vCtx *types.VertexContext, localVars map[string]any) error {
	log := log.FromContext(ctx).With("vertexContext", vctx.GetContext(r.rootModuleName, vCtx))
	log.Info("run block instance start...")
	log.Info("run block instance finished...")
	return nil
}
