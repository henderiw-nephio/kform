package fns

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"time"

	"github.com/henderiw-nephio/kform/tools/pkg/exec/fn/render"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/record"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/vars"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/vctx"
	"github.com/henderiw-nephio/kform/tools/pkg/recorder"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/types"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
	"github.com/henderiw-nephio/kform/tools/pkg/util/sets"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// kform handler executes a single Kform BlockType aka Vertex
// It uses a fnmap to execute the particular vertex BlockType
// Before calling the specific blockTYpe

type ExecHandler struct {
	RootModuleName string
	ModuleName     string
	BlockName      string
	FnsMap         Map
	Vars           cache.Cache[vars.Variable]
	Recorder       recorder.Recorder[record.Record]
}

// PostRun records the overall result of the module
func (r *ExecHandler) PostRun(ctx context.Context, start, stop time.Time, success bool) {
	recorder := r.Recorder
	if success {
		recorder.Record(record.Success(vctx.GetContextFromModule(r.RootModuleName, r.ModuleName), start, stop))
	} else {
		recorder.Record(record.FromErr(vctx.GetContextFromModule(r.RootModuleName, r.ModuleName), start, stop, fmt.Errorf("failed module execution")))
	}
}

func (r *ExecHandler) BlockRun(ctx context.Context, vertexName string, vCtx *vctx.VertexContext) bool {
	//log := log.FromContext(ctx).With("rootModuleName", r.RootModuleName, "moduleName", r.ModuleName, "blockName", vertexName)
	recorder := r.Recorder
	start := time.Now()
	success := true
	if err := r.runInstances(ctx, vCtx); err != nil {
		recorder.Record(record.FromErr(vctx.GetContextFromModule(r.RootModuleName, r.ModuleName), start, time.Now(), fmt.Errorf("failed module instances execution")))
		return !success
	}
	recorder.Record(record.Success(vctx.GetContextFromModule(r.RootModuleName, r.ModuleName), start, time.Now()))
	return success
}

func (r *ExecHandler) runInstances(ctx context.Context, vCtx *vctx.VertexContext) error {
	recorder := r.Recorder
	isForEach, items, err := r.getLoopItems(ctx, vCtx.BlockContext.Attributes)
	if err != nil {
		return err
	}
	g, ctx := errgroup.WithContext(ctx)
	for _, item := range items.UnsortedList() {
		localVars := map[string]any{}
		item := item
		if isForEach {
			localVars[render.LoopKeyForEachKey] = item.key
			localVars[render.LoopKeyForEachVal] = item.val
		} else {
			// we treat a singleton in the same way as count -> count.index will not be used based on our syntax checks
			localVars[render.LoopKeyCountIndex] = item.key
		}

		g.Go(func() error {
			start := time.Now()
			// lookup the blockType in the map
			if err := r.FnsMap.Run(ctx, vCtx, localVars); err != nil {
				recorder.Record(record.FromErr(vctx.GetContextFromName(r.BlockName), start, time.Now(), err))
				return err
			}
			recorder.Record(record.Success(vctx.GetContext(vCtx), start, time.Now()))
			select {
			case <-ctx.Done():
				return ctx.Err()
			default:
				return nil
			}
		})

	}
	return g.Wait()
}

type item struct {
	key any
	val any
}

func (r *ExecHandler) getLoopItems(ctx context.Context, attrs *types.KformBlockAttributes) (bool, sets.Set[item], error) {
	renderer := &render.Renderer{
		Vars: r.Vars,
	}
	isForEach := false
	items := sets.New[item]()
	// forEach and count cannot be used together
	if attrs != nil && attrs.ForEach != nil {
		isForEach = true
		v, err := renderer.Render(ctx, *attrs.ForEach)
		if err != nil {
			return isForEach, items, errors.Wrap(err, "render loop forEach failed")
		}
		switch v := v.(type) {
		case []any:
			// in a list we return key = int, val = any
			for k, v := range v {
				sets.Insert[item](items, item{key: k, val: v})
			}
		case map[any]any:
			// in a list we return key = any, val = any
			for k, v := range v {
				sets.Insert[item](items, item{key: k, val: v})
			}
		default:
			// in a regular value we return key = int, val = any
			sets.Insert[item](items, item{key: 0, val: v})
		}
		return isForEach, items, nil

	}
	if attrs != nil && attrs.Count != nil {
		v, err := renderer.Render(ctx, *attrs.Count)
		if err != nil {
			return isForEach, items, errors.Wrap(err, "render count failed")
		}
		switch v := v.(type) {
		case string:
			c, err := strconv.Atoi(v)
			if err != nil {
				return isForEach, items, fmt.Errorf("render count returned a string that cannot be converted to an int, got: %s", v)
			}
			items = getSetWithInt(c)
			return isForEach, items, nil
		case int64:
			items = getSetWithInt(int(v))
			return isForEach, items, nil
		default:
			return isForEach, items, errors.Errorf("render count return an unsupported type; support [int64, string], got: %s", reflect.TypeOf(v))
		}

	}
	items = getSetWithInt(1)
	return isForEach, items, nil
}

func getSetWithInt(i int) sets.Set[item] {
	items := sets.New[item]()
	for idx := 0; idx <= i; idx++ {
		sets.Insert[item](items, item{key: idx, val: idx})
	}
	return items
}
