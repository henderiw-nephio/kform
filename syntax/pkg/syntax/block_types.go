package syntax

import (
	"context"
	"fmt"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/syntax/apis/k8sform/v1alpha1"
	kformtypes "github.com/henderiw-nephio/kform/syntax/pkg/dag/types"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cctx"
)

var blockTypes = map[kformtypes.BlockType]BlockInitializer{
	kformtypes.BlockTypeProvider: newProvider,
	kformtypes.BlockTypeVariable: newVariable,
	kformtypes.BlockTypeResource: newResource,
	kformtypes.BlockTypeData:     newData,
}

type BlockInitializer func(n string) Block

func getBlockTypeNames() []string {
	s := make([]string, 0, len(blockTypes))
	for n := range blockTypes {
		s = append(s, string(n))
	}
	return s
}

func getBlock(n string) (Block, error) {
	bi, ok := blockTypes[kformtypes.GetBlockType(n)]
	if !ok {
		return nil, fmt.Errorf("cannot get blockType for %s, supported blocktypes %v", n, getBlockTypeNames())
	}
	return bi(n), nil
}

type Block interface {
	WithRecorder(diag.Recorder)
	GetName() string
	GetLevel() int
	processBlock(context.Context, *v1alpha1.Block) context.Context
	addData(context.Context)
}

type bt struct {
	Level    int
	Name     string
	recorder diag.Recorder
}

func (r *bt) GetName() string { return r.Name }

func (r *bt) GetLevel() int { return r.Level }

func (r *bt) WithRecorder(rec diag.Recorder) { r.recorder = rec }

func (r *bt) getDependencies(ctx context.Context, x any) []string {
	rn := newRenderer()
	if err := rn.GatherDependencies(x); err != nil {
		r.recorder.Record(diag.DiagFromErrWithContext(GetContext(ctx), err))
		return []string{}
	}
	return rn.GetDependencies()
}

type provider struct{ bt }

func newProvider(n string) Block {
	return &provider{
		bt{
			Level: 1,
			Name:  n,
		},
	}
}

func (r *provider) addData(ctx context.Context) {
	provider := cctx.GetContextValue[string](ctx, CtxKeyVarName)
	attrs := cctx.GetContextValue[map[string]any](ctx, CtxKeyAttributes)
	for k, v := range attrs {
		if k == "alias" {
			if alias, ok := v.(string); ok {
				provider = fmt.Sprintf("%s.%s", provider, alias)
			}
		}
	}

	execCfg := cctx.GetContextValue[ExecConfig](ctx, CtxExecConfig)
	if execCfg == nil {
		r.recorder.Record(diag.DiagErrorf("cannot add provider without execConfig"))
	}

	if execCfg.GetProviders().VertexExists(provider) {
		// we can ignore the error since Exists does the same check
		v, _ := execCfg.GetProviders().GetVertex(provider)
		r.recorder.Record(diag.DiagFromErrWithContext(GetContext(ctx), fmt.Errorf("duplicate resource with fileName: %s, name: %s, type: %s", v.FileName, provider, string(v.BlockType))))
		return
	}

	execCfg.GetProviders().AddVertex(ctx, provider, kformtypes.Provider{
		FileName:   cctx.GetContextValue[string](ctx, CtxKeyFileName),
		BlockType:  kformtypes.BlockTypeProvider,
		Attributes: attrs,
		Object:     cctx.GetContextValue[any](ctx, CtxKeyObject),
	})
}

func newResource(n string) Block {
	return &resource{
		bt{
			Level: 2,
			Name:  n,
		},
	}
}

type resource struct{ bt }

func (r *resource) addData(ctx context.Context) {
	// process -> provider, gvk, name
	pgvk := cctx.GetContextValue[string](ctx, CtxKeyVarType)
	provider, gvk, err := validateResourceIdentifier(pgvk)
	if err != nil {
		r.recorder.Record(diag.DiagFromErrWithContext(GetContext(ctx), err))
		return
	}

	attrs := cctx.GetContextValue[map[string]any](ctx, CtxKeyAttributes)
	for k, v := range attrs {
		if k == "provider" {
			if p, ok := v.(string); ok {
				provider = p
			}
		}
	}

	allDeps := []string{}
	for _, attr := range attrs {
		deps := r.getDependencies(ctx, attr)
		if len(deps) > 0 {
			allDeps = append(allDeps, deps...)
		}
	}
	obj := cctx.GetContextValue[any](ctx, CtxKeyObject)
	if obj != nil {
		deps := r.getDependencies(ctx, obj)
		if len(deps) > 0 {
			allDeps = append(allDeps, deps...)
		}
	}

	execCfg := cctx.GetContextValue[ExecConfig](ctx, CtxExecConfig)
	if execCfg == nil {
		r.recorder.Record(diag.DiagErrorf("cannot add data without execConfig"))
	}

	name := fmt.Sprintf("%s.%s", pgvk, cctx.GetContextValue[string](ctx, CtxKeyVarName))

	if execCfg.GetVars().VertexExists(name) {
		// we can ignore the error since Exists does the same check
		v, _ := execCfg.GetVars().GetVertex(name)
		r.recorder.Record(diag.DiagFromErrWithContext(GetContext(ctx), fmt.Errorf("duplicate resource with fileName: %s, name: %s, type: %s", v.FileName, name, string(v.BlockType))))
		return
	}

	execCfg.GetVars().AddVertex(ctx, name, kformtypes.Variable{
		FileName:     cctx.GetContextValue[string](ctx, CtxKeyFileName),
		BlockType:    kformtypes.BlockTypeResource,
		Provider:     provider,
		GVK:          gvk,
		Attributes:   attrs,
		Object:       obj,
		Dependencies: allDeps,
	})
}

func newData(n string) Block {
	return &data{
		bt{
			Level: 2,
			Name:  n,
		},
	}
}

type data struct{ bt }

func (r *data) addData(ctx context.Context) {
	// process -> provider, gvk, name
	pgvk := cctx.GetContextValue[string](ctx, CtxKeyVarType)
	provider, gvk, err := validateResourceIdentifier(pgvk)
	if err != nil {
		r.recorder.Record(diag.DiagFromErrWithContext(GetContext(ctx), err))
		return
	}

	attrs := cctx.GetContextValue[map[string]any](ctx, CtxKeyAttributes)
	for k, v := range attrs {
		if k == "provider" {
			if p, ok := v.(string); ok {
				provider = p
			}
		}
	}

	allDeps := []string{}
	for _, attr := range attrs {
		deps := r.getDependencies(ctx, attr)
		if len(deps) > 0 {
			allDeps = append(allDeps, deps...)
		}
	}
	obj := cctx.GetContextValue[any](ctx, CtxKeyObject)
	if obj != nil {
		deps := r.getDependencies(ctx, obj)
		if len(deps) > 0 {
			allDeps = append(allDeps, deps...)
		}
	}

	execCfg := cctx.GetContextValue[ExecConfig](ctx, CtxExecConfig)
	if execCfg == nil {
		r.recorder.Record(diag.DiagErrorf("cannot add data without execConfig"))
	}

	name := fmt.Sprintf("%s.%s", pgvk, cctx.GetContextValue[string](ctx, CtxKeyVarName))

	if execCfg.GetVars().VertexExists(name) {
		// we can ignore the error since Exists does the same check
		v, _ := execCfg.GetVars().GetVertex(name)
		r.recorder.Record(diag.DiagFromErrWithContext(GetContext(ctx), fmt.Errorf("duplicate resource with fileName: %s, name: %s, type: %s", v.FileName, name, string(v.BlockType))))
		return
	}

	execCfg.GetVars().AddVertex(ctx, name, kformtypes.Variable{
		FileName:     cctx.GetContextValue[string](ctx, CtxKeyFileName),
		BlockType:    kformtypes.BlockTypeData,
		Provider:     provider,
		GVK:          gvk,
		Attributes:   attrs,
		Object:       obj,
		Dependencies: allDeps,
	})
}

func newVariable(n string) Block {
	return &variable{
		bt{
			Level: 1,
			Name:  n,
		},
	}
}

type variable struct{ bt }

func (r *variable) addData(ctx context.Context) {
	attrs := cctx.GetContextValue[map[string]any](ctx, CtxKeyAttributes)

	allDeps := []string{}
	for _, attr := range attrs {
		deps := r.getDependencies(ctx, attr)
		if len(deps) > 0 {
			allDeps = append(allDeps, deps...)
		}
	}

	execCfg := cctx.GetContextValue[ExecConfig](ctx, CtxExecConfig)
	if execCfg == nil {
		r.recorder.Record(diag.DiagErrorf("cannot add variable without execConfig"))
	}

	name := fmt.Sprintf("var.%s", cctx.GetContextValue[string](ctx, CtxKeyVarName))

	execCfg.GetVars().AddVertex(ctx, name, kformtypes.Variable{
		FileName:     cctx.GetContextValue[string](ctx, CtxKeyFileName),
		BlockType:    kformtypes.BlockTypeVariable,
		Attributes:   attrs,
		Dependencies: allDeps,
	})
}
