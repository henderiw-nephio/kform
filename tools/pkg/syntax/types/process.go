package types

import (
	"context"
	"fmt"
	"regexp"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/tools/apis/kform/block/v1alpha1"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/sctx"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cctx"
)

func (r *bt) ProcessBlock(ctx context.Context, block *v1alpha1.Block) context.Context {
	level := cctx.GetContextValue[int](ctx, sctx.CtxKeyLevel)
	if level < r.GetLevel() {
		// continue to walk
		// validate if attr or obj are present at the intermediate level
		r.validateAttrAndObjectAtIntermediateLevel(ctx, block)
		// validate the block prior to processing
		blockName, block, err := GetNextBlock(ctx, block)
		if err != nil {
			r.recorder.Record(diag.DiagFromErr(err))
			return ctx
		}
		// process the next block
		level++
		ctx = r.addContext(ctx, blockName, level)
		ctx = context.WithValue(ctx, sctx.CtxKeyLevel, level)
		return r.ProcessBlock(ctx, block)
	}
	// process attributes
	if block.Attributes != nil {
		//fmt.Println(block.Attributes)
		ctx = context.WithValue(ctx, sctx.CtxKeyAttributes, block.Attributes)
	}
	// process object
	if block.Object != nil {
		//fmt.Println(block.Object)
		ctx = context.WithValue(ctx, sctx.CtxKeyObject, block.Object)
	}

	fmt.Println("processed:",
		cctx.GetContextValue[string](ctx, sctx.CtxKeyBlockType),
		cctx.GetContextValue[string](ctx, sctx.CtxKeyVarType),
		cctx.GetContextValue[string](ctx, sctx.CtxKeyVarName),
	)
	return ctx
}

func GetNextBlock(ctx context.Context, block *v1alpha1.Block) (string, *v1alpha1.Block, error) {
	// validate the block prior to processing
	if err := validateBlock(ctx, block); err != nil {
		return "", nil, err
	}
	// process next level
	for blockName, block := range block.NestedBlock {
		block := block
		return blockName, &block, nil
	}
	// we should never get here
	return "", nil, fmt.Errorf("cannot have a block without a nested block")
}

func validateBlock(ctx context.Context, block *v1alpha1.Block) error {
	level := cctx.GetContextValue[int](ctx, sctx.CtxKeyLevel)
	// if there is no block assigned in the topBlock this is an invalid block
	if len(block.NestedBlock) == 0 {
		if level == 0 {
			return fmt.Errorf("cannot have a block without a block type: %v", block.NestedBlock)
		} else {
			return fmt.Errorf("cannot have a block without a nested block")
		}
	}
	// a block can only have 1 blocktype
	if len(block.NestedBlock) > 1 {
		return fmt.Errorf("cannot have more then 1 blocktype in a block, got: %v", block.NestedBlock)
	}
	return nil
}

func (r *bt) validateAttrAndObjectAtIntermediateLevel(ctx context.Context, block *v1alpha1.Block) {
	level := cctx.GetContextValue[int](ctx, sctx.CtxKeyLevel)
	if block.Object != nil {
		r.recorder.Record(diag.DiagWarnfWithContext(sctx.GetContext(ctx), "object at level %d present but ignored", level))
	}
	// for blockkind = attributeOnly we do not expect an object
	if len(block.Attributes) > 0 {
		r.recorder.Record(diag.DiagWarnfWithContext(sctx.GetContext(ctx), "attributes at level %d present but ignored", level))
	}
}

func (r *bt) addContext(ctx context.Context, blockName string, level int) context.Context {
	if level == r.Level {
		ctx = context.WithValue(ctx, sctx.CtxKeyVarName, blockName)
	}

	if r.Name == "resource" || r.Name == "data" {
		if level == r.Level-1 {
			ctx = context.WithValue(ctx, sctx.CtxKeyVarType, blockName)
		}
	}
	return ctx
}

// validateResourceSyntax validates the syntax of the resource kind
// resource Type must starts with a letter
// resource Type can container letters in lower and upper case, numbers and '-', '_'
func validateResourceSyntax(kind string, name string) error {
	re := regexp.MustCompile(`^[a-zA-Z]+[a-zA-Z0-9_-]*$`)
	if !re.Match([]byte(name)) {
		return fmt.Errorf("syntax error a %s starts with a letter and can container letters in lower and upper case, numbers and '-', '_', got: %s", kind, name)
	}
	return nil
}
