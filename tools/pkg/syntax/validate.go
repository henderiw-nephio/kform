package syntax

import (
	"context"
	"sync"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	blockv1alpha1 "github.com/henderiw-nephio/kform/tools/apis/kform/block/v1alpha1"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/sctx"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/types"
	"github.com/henderiw/logger/log"
)

func (r *parser) Validate(ctx context.Context) {
	var wg sync.WaitGroup
	for path, kform := range r.kforms {
		path := path
		kform := kform
		for _, block := range kform.Blocks {
			block := block
			wg.Add(1)
			go func(block *blockv1alpha1.Block) {
				defer wg.Done()
				ctx = context.WithValue(ctx, sctx.CtxKeyFileName, path)
				ctx = context.WithValue(ctx, sctx.CtxKeyLevel, 0)
				r.processBlock(ctx, block)
			}(&block)
		}
	}
	wg.Wait()
}

// walkTopBlock identifies the blockType
func (r *parser) processBlock(ctx context.Context, block *blockv1alpha1.Block) {
	blockType, block, err := types.GetNextBlock(ctx, block)
	if err != nil {
		r.recorder.Record(diag.DiagFromErr(err))
		return
	}
	log := log.FromContext(ctx).With("blockType", blockType)
	log.Debug("processBlock")
	bt, err := types.GetBlock(blockType)
	if err != nil {
		r.recorder.Record(diag.DiagFromErr(err))
		return
	}
	bt.WithRecorder(r.recorder)
	ctx = context.WithValue(ctx, sctx.CtxKeyBlockType, bt.GetName())
	// if ok we add the resource to the cache
	ctx = bt.ProcessBlock(ctx, block)
	bt.AddData(ctx)
}
