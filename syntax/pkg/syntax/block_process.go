package syntax

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/syntax/apis/k8sform/v1alpha1"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cctx"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func (r *parser) processConfigs(ctx context.Context) {
	var wg sync.WaitGroup
	for _, kform := range r.kforms {
		kform := kform
		for _, block := range kform.Blocks {
			block := block
			wg.Add(1)
			go func(block *v1alpha1.Block) {
				defer wg.Done()
				//ctx = context.Background()
				ctx = context.WithValue(ctx, CtxKeyFileName, kform.FileName)
				ctx = context.WithValue(ctx, CtxKeyLevel, 0)
				r.processBlock(ctx, block)
			}(&block)
		}
	}
	wg.Wait()
}

// walkTopBlock identifies the blockType
func (r *parser) processBlock(ctx context.Context, block *v1alpha1.Block) {
	blockType, block, err := getNextBlock(ctx, block)
	if err != nil {
		r.recorder.Record(diag.DiagFromErr(err))
		return
	}
	//fmt.Println("processBlock blockType", blockType)
	b, err := getBlock(blockType)
	if err != nil {
		r.recorder.Record(diag.DiagFromErr(err))
		return
	}
	b.WithRecorder(r.recorder)
	//b.WithProviders(r.providers)
	//b.WithVars(r.vars)
	ctx = context.WithValue(ctx, CtxKeyBlockType, b.GetName())
	// if ok we add the resource to the cache
	ctx = b.processBlock(ctx, block)
	b.addData(ctx)
}

func (r *bt) processBlock(ctx context.Context, block *v1alpha1.Block) context.Context {
	level := cctx.GetContextValue[int](ctx, CtxKeyLevel)
	if level < r.GetLevel() {
		// continue to walk
		// validate if attr or obj are present at the intermediate level
		r.validateAttrAndObjectAtIntermediateLevel(ctx, block)
		// validate the block prior to processing
		blockName, block, err := getNextBlock(ctx, block)
		if err != nil {
			r.recorder.Record(diag.DiagFromErr(err))
			return ctx
		}
		// process the next block
		level++
		ctx = r.addContext(ctx, blockName, level)
		ctx = context.WithValue(ctx, CtxKeyLevel, level)
		return r.processBlock(ctx, block)
	}
	/*
		if r.GetKind() == BlockKindAttributeOnly {
			// for blockkind = attributeOnly we do NOT expect an object
			if err := r.validateAttributeOnlyBlock(ctx, block); err != nil {
				r.recorder.Record(diag.DiagFromErr(err))
			}
		}
	*/
	// process attributes
	if block.Attributes != nil {
		//fmt.Println(block.Attributes)
		ctx = context.WithValue(ctx, CtxKeyAttributes, block.Attributes)
	}
	// process object
	if block.Object != nil {
		//fmt.Println(block.Object)
		ctx = context.WithValue(ctx, CtxKeyObject, block.Object)
	}
	fmt.Println("processed:",
		cctx.GetContextValue[string](ctx, CtxKeyBlockType),
		//cctx.GetContextValue[string](ctx, ctxKeyProvider),
		cctx.GetContextValue[string](ctx, CtxKeyVarType),
		cctx.GetContextValue[string](ctx, CtxKeyVarName),
	)
	return ctx
}

func getNextBlock(ctx context.Context, block *v1alpha1.Block) (string, *v1alpha1.Block, error) {
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
	level := cctx.GetContextValue[int](ctx, CtxKeyLevel)
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
	level := cctx.GetContextValue[int](ctx, CtxKeyLevel)
	if block.Object != nil {
		r.recorder.Record(diag.DiagWarnfWithContext(GetContext(ctx), "object at level %d present but ignored", level))
	}
	// for blockkind = attributeOnly we do not expect an object
	if len(block.Attributes) > 0 {
		r.recorder.Record(diag.DiagWarnfWithContext(GetContext(ctx), "attributes at level %d present but ignored", level))
	}
}

func (r *bt) addContext(ctx context.Context, blockName string, level int) context.Context {
	if level == r.Level {
		ctx = context.WithValue(ctx, CtxKeyVarName, blockName)
	}

	if r.Name == "resource" || r.Name == "data" {
		if level == r.Level-1 {
			ctx = context.WithValue(ctx, CtxKeyVarType, blockName)
		}
	}
	return ctx
}

func validateResourceIdentifier(pgvk string) (string, schema.GroupVersionKind, error) {
	split := strings.Split(pgvk, "_")

	if len(split) < 3 {
		return "", schema.GroupVersionKind{}, fmt.Errorf("format of the provider gvk should be <provider>_<apiVersion>_<kind>, got: %s", pgvk)
	}
	gvk := schema.GroupVersionKind{
		Version: split[len(split)-2],
		Kind:    split[len(split)-1],
	}
	provider := split[0]
	if len(split) > 3 {
		gvk.Group = strings.Join(split[1:len(split)-2], ".")
	}

	return provider, gvk, nil
}

type Renderer interface {
	GatherDependencies(x any) error
	GetDependencies() []string
}

func newRenderer() Renderer {
	return &renderer{
		deps: []string{},
	}
}

type renderer struct {
	deps []string
}

func (r *renderer) GetDependencies() []string {
	return r.deps
}

func (r *renderer) GatherDependencies(x any) error {
	switch x := x.(type) {
	case map[string]any:
		for _, v := range x {
			if err := r.GatherDependencies(v); err != nil {
				return err
			}
		}
	case []any:
		for _, t := range x {
			if err := r.GatherDependencies(t); err != nil {
				return err
			}
		}
	case string:
		if err := r.getExprDependencies(x); err != nil {
			return err
		}
	}
	return nil
}

func (r *renderer) getExprDependencies(expr string) error {
	depsSplit := strings.Split(expr, "$")
	if len(depsSplit) == 1 {
		return nil
	}
	for _, dep := range depsSplit[1:] {
		depSplit := strings.Split(dep, ".")
		if len(depSplit) < 2 {
			return fmt.Errorf("a dependency always need <resource-type>.<resource-identifier>, got: %s", dep)
		}
		r.deps = append(r.deps, strings.Join(depSplit[:2], "."))
	}
	return nil
}
