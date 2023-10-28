package parser

import (
	"context"
	"path/filepath"
	"sync"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/types"
	"github.com/henderiw/logger/log"
)

func (r *moduleparser) validate(ctx context.Context, kforms map[string]*types.Kform) {
	var wg sync.WaitGroup
	for path, kform := range kforms {
		path := path
		kform := kform
		for _, block := range kform.Blocks {
			block := block
			wg.Add(1)
			go func(block *types.KformBlock) {
				defer wg.Done()
				ctx = context.WithValue(ctx, types.CtxKeyFileName, filepath.Join(r.path, path))
				ctx = context.WithValue(ctx, types.CtxKeyLevel, 0)
				r.processBlock(ctx, block)
			}(&block)
		}
	}
	wg.Wait()
}

// walkTopBlock identifies the blockType
func (r *moduleparser) processBlock(ctx context.Context, block *types.KformBlock) {
	blockType, block, err := types.GetNextBlock(ctx, block)
	if err != nil {
		r.recorder.Record(diag.DiagFromErr(err))
		return
	}
	l := log.FromContext(ctx).With("blockType", blockType)
	l.Debug("processBlock")
	bt, err := types.GetBlock(ctx, blockType)
	if err != nil {
		r.recorder.Record(diag.DiagFromErr(err))
		return
	}
	ctx = context.WithValue(ctx, types.CtxKeyBlockType, bt.GetBlockType())
	// if ok we add the resource to the cache
	ctx = bt.ProcessBlock(ctx, block)
	bt.UpdateModule(ctx)
}
