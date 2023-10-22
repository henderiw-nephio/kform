package block

import (
	"context"
	"fmt"
	"reflect"
	"sort"
	"strconv"

	"github.com/henderiw-nephio/kform/core/pkg/exec/render"
	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/syntax/pkg/dag"
	kformtypes "github.com/henderiw-nephio/kform/syntax/pkg/dag/types"
)

type Block interface {
}

func New(varStore dag.DAG[kformtypes.Variable]) Block {
	return &block{
		recorder: diag.NewRecorder(),
		varStore: varStore,
	}
}

type block struct {
	recorder diag.Recorder
	varStore dag.DAG[kformtypes.Variable]
}

func (r *block) Run(ctx context.Context, blockName string) diag.Diagnostics {
	// check if the provider is running, if not start it

	// need to lookup kind -> substitution with schema is better for type checking

	// parse loop attributes count/for_each and propagate the variables
	d := r.propagateVariables(ctx, blockName)
	if d.HasError() {
		return d
	}

	// resource validation of the full resource

	// interact with the provider

	//

	return d
}

func (r *block) propagateVariables(ctx context.Context, blockName string) diag.Diagnostics {
	v, err := r.varStore.GetVertex(blockName)
	if err != nil {
		return diag.FromErr(err)
	}

	d := diag.Diagnostics{}
	if Attrs(v.Attributes).isLoopAttrPresent() {
		evalCtx, diags := r.evalAttrs(ctx, v.Attributes)
		if diags.HasError() {
			return diags
		}
		d = append(d, diags...)
		newObj := make([]any, 0, evalCtx.Total)
		for idx := 0; idx < evalCtx.Total; idx++ {
			// The index is always set, but could be just 1 w/o a count present
			initVars := make(map[string]any, evalCtx.Total)
			if evalCtx.Count > 0 {
				initVars[render.KeyCountIndex] = idx
			}
			if len(evalCtx.ForEaches) > 0 {
				initVars[render.KeyForEachKey] = evalCtx.ForEaches[idx].Key
				initVars[render.KeyForEachVal] = evalCtx.ForEaches[idx].Value
			}
			rdr, err := render.New(r.varStore, initVars)
			if err != nil {
				return diag.FromErr(err)
			}
			obj, err := DeepCopy(v.Object)
			if err != nil {
				return diag.FromErr(err)
			}
			newObject, err := rdr.Render(ctx, obj)
			if err != nil {
				return diag.FromErr(err)
			}
			newObj = append(newObj, newObject)
		}
		v.RenderedObject = newObj

	} else {
		// no loopAttrs present
		rdr, err := render.New(r.varStore, nil)
		if err != nil {
			return diag.FromErr(err)
		}
		newObject, err := rdr.Render(ctx, v.Object)
		if err != nil {
			return diag.FromErr(err)
		}
		v.RenderedObject = newObject
	}
	if err := r.varStore.UpdateVertex(ctx, blockName, v); err != nil {
		return diag.FromErr(err)
	}
	return d
}

type evalAttrCtx struct {
	Total     int
	Count     int
	ForEaches []ForEach
}

func (r *block) evalAttrs(ctx context.Context, attrs Attrs) (evalAttrCtx, diag.Diagnostics) {
	count, err := r.evalCount(ctx, attrs)
	if err != nil {
		r.recorder.Record(diag.DiagFromErr(err))
		// set c = 0 so that we know later it should not be used
		count = 0
	}
	forEaches, err := r.evalForEach(ctx, attrs)
	if err != nil {
		r.recorder.Record(diag.DiagFromErr(err))
	}
	total := 0
	switch {
	case count > 0 && len(forEaches) > 0:
		total = min(count, len(forEaches))
	case count > 0 || len(forEaches) > 0:
		total = max(count, len(forEaches))
	}

	return evalAttrCtx{Total: total, Count: count, ForEaches: forEaches}, r.recorder.Get()

}

func (r *block) evalCount(ctx context.Context, attr Attrs) (int, error) {
	if x, ok := attr[render.BlockAttrCount]; ok {
		switch x := x.(type) {
		case int:
			return x, nil
		case string:
			isCelExpr, err := render.IsCelExpression(x)
			if err != nil {
				return 0, err
			}
			if isCelExpr {
				rdr, err := render.New(r.varStore, nil)
				if err != nil {
					return 0, err
				}
				x, err := rdr.Render(ctx, x)
				if err != nil {
					return 0, err
				}
				switch c := x.(type) {
				case int:
					return c, nil
				case string:
					return strconv.Atoi(c)
				default:
					return 0, fmt.Errorf("unexpected type after cel expression evaluation expected string or int got: %s", reflect.TypeOf(c))
				}
			} else {
				// when there is no cell expression we try to convert the string to an int
				return strconv.Atoi(x)
			}
		default:
			return 0, fmt.Errorf("unexpected type expected string or int got: %s", reflect.TypeOf(x))
		}
	}
	return 0, nil
}

type ForEach struct {
	Key   any
	Value any
}

func (r *block) evalForEach(ctx context.Context, attr Attrs) ([]ForEach, error) {
	forEaches := []ForEach{}
	if x, ok := attr[render.BlockAttrForEach]; ok {
		switch x := x.(type) {
		case map[string]any:
			keys := make([]string, 0, len(x))
			for k := range x {
				keys = append(keys, k)
			}
			sort.Strings(keys)
			for _, k := range keys {
				forEaches = append(forEaches, ForEach{Key: k, Value: x[k]})
			}
			return forEaches, nil
		case []any:
			for k, v := range x {
				forEaches = append(forEaches, ForEach{Key: k, Value: v})
			}
			return forEaches, nil
		case string:
			isCelExpr, err := render.IsCelExpression(x)
			if err != nil {
				return forEaches, err
			}
			if isCelExpr {
				// render the expression
				rdr, err := render.New(r.varStore, nil)
				if err != nil {
					return forEaches, err
				}
				x, err := rdr.Render(ctx, x)
				if err != nil {
					return forEaches, err
				}
				switch x := x.(type) {
				case map[any]string:
					for k, v := range x {
						forEaches = append(forEaches, ForEach{Key: k, Value: v})
					}
					return forEaches, nil
				case []any:
					for k, v := range x {
						forEaches = append(forEaches, ForEach{Key: k, Value: v})
					}
					return forEaches, nil
				default:
					return forEaches, fmt.Errorf("unexpected type expected slice or map got: %s", reflect.TypeOf(x).Name())
				}

			} else {
				return forEaches, fmt.Errorf("unexpected type expected slice or map got: %s", reflect.TypeOf(x).Name())
			}
		default:
			return forEaches, fmt.Errorf("unexpected type expected slice or map got: %s", reflect.TypeOf(x).Name())
		}
	}
	return forEaches, nil
}
