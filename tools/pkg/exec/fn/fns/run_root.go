package fns

import (
	"context"

	"github.com/henderiw-nephio/kform/tools/pkg/exec/fn"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/vctx"
	"github.com/henderiw/logger/log"
)

func NewRootFn(cfg *Config) fn.BlockInstanceRunner {
	return &root{}
}

type root struct {
}

func (r *root) Run(ctx context.Context, vCtx *vctx.VertexContext, localVars map[string]any) error {
	log := log.FromContext(ctx).With("vertexContext", vctx.GetContext(vCtx))
	log.Info("run instance")
	return nil
}
