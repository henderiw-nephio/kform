package types

import (
	"context"
	"fmt"
	"strings"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	blockv1alpha1 "github.com/henderiw-nephio/kform/tools/apis/kform/block/v1alpha1"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/exttypes"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/sctx"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cctx"
	"sigs.k8s.io/yaml"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func newData(n string) Block {
	return &data{
		bt{
			Level: 2,
			Name:  n,
		},
	}
}

type data struct{ bt }

func (r *data) AddData(ctx context.Context) {
	// process -> provider, gvk, name
	resourceType := cctx.GetContextValue[string](ctx, sctx.CtxKeyVarType)
	if err := validateResourceSyntax("resourceType", resourceType); err != nil {
		r.recorder.Record(diag.DiagFromErrWithContext(sctx.GetContext(ctx), err))
		return
	}
	split := strings.Split(resourceType, "_")
	provider := split[0]

	attrs := cctx.GetContextValue[map[string]any](ctx, sctx.CtxKeyAttributes)
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
	obj := cctx.GetContextValue[any](ctx, sctx.CtxKeyObject)
	if obj != nil {
		deps := r.getDependencies(ctx, obj)
		if len(deps) > 0 {
			allDeps = append(allDeps, deps...)
		}
	}
	var gvk schema.GroupVersionKind
	if obj == nil {
		r.recorder.Record(diag.DiagErrorf("cannot have a resource with no objectData"))
	} else {
		b, err := yaml.Marshal(obj)
		if err != nil {
			r.recorder.Record(diag.DiagErrorf("cannot yaml marshal object err: %s", err))
		} else {
			ko, err := fn.ParseKubeObject(b)
			if err != nil {
				r.recorder.Record(diag.DiagErrorf("cannot parse kubeobject object err: %s", err))
			} else {
				gvk.Kind = ko.GetKind()
				split = strings.Split(ko.GetAPIVersion(), "/")
				switch len(split) {
				case 1:
					gvk.Version = split[0]
				case 2:
					gvk.Group = split[0]
					gvk.Version = split[1]
				default:
					r.recorder.Record(diag.DiagErrorf("invalid api version, got %s", ko.GetAPIVersion()))
				}
			}
		}
	}

	execCfg := cctx.GetContextValue[exttypes.ExecConfig](ctx, sctx.CtxExecConfig)
	if execCfg == nil {
		r.recorder.Record(diag.DiagErrorf("cannot add data without execConfig"))
	}

	name := fmt.Sprintf("%s.%s", resourceType, cctx.GetContextValue[string](ctx, sctx.CtxKeyVarName))

	if execCfg.GetVars().VertexExists(name) {
		// we can ignore the error since Exists does the same check
		v, _ := execCfg.GetVars().GetVertex(name)
		r.recorder.Record(diag.DiagFromErrWithContext(sctx.GetContext(ctx), fmt.Errorf("duplicate resource with fileName: %s, name: %s, type: %s", v.FileName, name, string(v.BlockType))))
		return
	}

	execCfg.GetVars().AddVertex(ctx, name, blockv1alpha1.Variable{
		FileName:     cctx.GetContextValue[string](ctx, sctx.CtxKeyFileName),
		BlockType:    blockv1alpha1.BlockTypeData,
		Provider:     provider,
		GVK:          gvk,
		Attributes:   attrs,
		Object:       obj,
		Dependencies: allDeps,
	})
}
