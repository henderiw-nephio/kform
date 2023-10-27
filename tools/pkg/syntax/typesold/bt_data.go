package types

/*
import (
	"context"
	"fmt"
	"strings"

	"github.com/GoogleContainerTools/kpt-functions-sdk/go/fn"
	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	blockv1alpha1 "github.com/henderiw-nephio/kform/tools/apis/kform/block/v1alpha1"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/exttypes"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cctx"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/yaml"
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
	resourceType := cctx.GetContextValue[string](ctx, blockv1alpha1.CtxKeyVarType)
	if err := validateResourceSyntax("resourceType", resourceType); err != nil {
		r.recorder.Record(diag.DiagFromErrWithContext(blockv1alpha1.GetContext(ctx), err))
		return
	}
	split := strings.Split(resourceType, "_")
	provider := split[0]

	attrs := cctx.GetContextValue[map[string]any](ctx, blockv1alpha1.CtxKeyAttributes)
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
	instances := cctx.GetContextValue[[]any](ctx, blockv1alpha1.CtxKeyInstances)
	if instances != nil {
		deps := r.getDependencies(ctx, instances)
		if len(deps) > 0 {
			allDeps = append(allDeps, deps...)
		}
	}
	var gvk schema.GroupVersionKind
	if instances == nil {
		r.recorder.Record(diag.DiagErrorf("cannot have a resource with no objectData"))
	} else {
		b, err := yaml.Marshal(instances)
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

	execCfg := cctx.GetContextValue[exttypes.ExecConfig](ctx, blockv1alpha1.CtxExecConfig)
	if execCfg == nil {
		r.recorder.Record(diag.DiagErrorf("cannot add data without execConfig"))
	}

	name := fmt.Sprintf("%s.%s", resourceType, cctx.GetContextValue[string](ctx, blockv1alpha1.CtxKeyVarName))

	if execCfg.GetVars().VertexExists(name) {
		// we can ignore the error since Exists does the same check
		v, _ := execCfg.GetVars().GetVertex(name)
		r.recorder.Record(diag.DiagFromErrWithContext(blockv1alpha1.GetContext(ctx), fmt.Errorf("duplicate resource with fileName: %s, name: %s, type: %s", v.FileName, name, string(v.BlockType))))
		return
	}

	execCfg.GetVars().AddVertex(ctx, name, blockv1alpha1.Variable{
		FileName:     cctx.GetContextValue[string](ctx, blockv1alpha1.CtxKeyFileName),
		BlockType:    blockv1alpha1.BlockTypeData,
		Provider:     provider,
		GVK:          gvk,
		Attributes:   attrs,
		Instances:    instances,
		Dependencies: allDeps,
	})
}
*/
