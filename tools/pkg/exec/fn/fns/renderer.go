package fns

import (
	"context"
	"fmt"
	"reflect"

	"github.com/henderiw-nephio/kform/tools/pkg/exec/fn/render"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/vars"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
)

type renderer struct {
	vars cache.Cache[vars.Variable]
}

func (r *renderer) renderData(ctx context.Context, blockName string, x any, localVars map[string]any) error {
	renderer := render.Renderer{
		Vars:      r.vars,
		LocalVars: localVars,
	}
	d, err := renderer.Render(ctx, x)
	if err != nil {
		return fmt.Errorf("run output, render failed for blockName %s, err: %s", blockName, err.Error())
	}

	total, ok := localVars[render.LoopKeyItemsTotal]
	if !ok {
		total = 1
	}
	totalInt, ok := total.(int)
	if !ok {
		return fmt.Errorf("items.total must always be an int: got: %s", reflect.TypeOf(total))
	}

	index, ok := localVars[render.LoopKeyItemsIndex]
	if !ok {
		index = 0
	}
	indexInt, ok := index.(int)
	if !ok {
		return fmt.Errorf("items.index must always be an int: got: %s", reflect.TypeOf(index))
	}
	if indexInt >= totalInt {
		return fmt.Errorf("index cannot be bigger or equal to total index: %d, totol: %d", indexInt, totalInt)
	}

	// if the data already exists we can add the content to it
	v, err := r.vars.Get(cache.NSN{Name: blockName})
	if err != nil {
		// variable does not exist in the varCache
		v := vars.Variable{}
		v.Data = map[string][]any{vars.DummyKey: make([]any, totalInt)}
		v.Data[vars.DummyKey] = r.insert(v.Data[vars.DummyKey], indexInt, d)
		r.vars.Add(ctx, cache.NSN{Name: blockName}, v)
	} else {
		// variable exists in the varCache
		if len(v.Data) == 0 {
			v.Data = map[string][]any{vars.DummyKey: make([]any, totalInt)}
			v.Data[vars.DummyKey] = r.insert(v.Data[vars.DummyKey], indexInt, d)
		} else {
			if x, ok := v.Data[vars.DummyKey]; !ok {
				v.Data = map[string][]any{vars.DummyKey: make([]any, totalInt)}
				v.Data[vars.DummyKey] = r.insert(v.Data[vars.DummyKey], indexInt, d)
			} else {
				if len(x) == 0 {
					v.Data = map[string][]any{vars.DummyKey: make([]any, totalInt)}
					v.Data[vars.DummyKey] = r.insert(v.Data[vars.DummyKey], indexInt, d)
				} else {
					v.Data[vars.DummyKey] = r.insert(v.Data[vars.DummyKey], indexInt, d)
				}
			}
		}
		r.vars.Upsert(ctx, cache.NSN{Name: blockName}, v)
	}

	return nil
}

func (r *renderer) insert(slice []any, pos int, value any) []any {
	// Check if the position is out of bounds
	if pos < 0 || pos > len(slice) {
		// Should never happen
		return slice
	}
	slice[pos] = value
	return slice
}
