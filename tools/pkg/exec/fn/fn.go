package fn

import (
	"context"

	"github.com/henderiw-nephio/kform/tools/pkg/syntax/types"
)

type BlockRunner interface {
	Run(ctx context.Context, vCtx *types.VertexContext) error
}

type BlockInstanceRunner interface {
	Run(ctx context.Context, vCtx *types.VertexContext, localVars map[string]any) error
}

type BlockRunnerOption func(BlockRunner)
