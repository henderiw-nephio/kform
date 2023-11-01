package fn

import (
	"context"

	"github.com/henderiw-nephio/kform/tools/pkg/exec/vctx"
)

type BlockRunner interface {
	Run(ctx context.Context, vCtx *vctx.VertexContext) error
}

type BlockInstanceRunner interface {
	Run(ctx context.Context, vCtx *vctx.VertexContext, localVars map[string]any) error
}

type BlockRunnerOption func(BlockRunner)
