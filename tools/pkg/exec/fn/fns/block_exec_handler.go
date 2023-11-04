package fns

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"sync"
	"time"

	"github.com/henderiw-nephio/kform/tools/pkg/exec/fn/render"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/record"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/vars"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/vctx"
	"github.com/henderiw-nephio/kform/tools/pkg/recorder"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/types"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
	"github.com/henderiw/logger/log"
	"github.com/pkg/errors"
	"golang.org/x/sync/errgroup"
)

// kform handler executes a single Kform BlockType aka Vertex
// It uses a fnmap to execute the particular vertex BlockType
// Before calling the specific blockTYpe

type EHConfig struct {
	RootModuleName string
	ModuleName     string
	BlockName      string
	Vars           cache.Cache[vars.Variable]
	Recorder       recorder.Recorder[record.Record]
}

func NewExecHandler(ctx context.Context, cfg *EHConfig) *ExecHandler {
	return &ExecHandler{
		RootModuleName: cfg.RootModuleName,
		ModuleName:     cfg.ModuleName,
		BlockName:      cfg.BlockName,
		Vars:           cfg.Vars,
		Recorder:       cfg.Recorder,
		fnsMap: NewMap(ctx, &Config{
			RootModuleName: cfg.RootModuleName,
			Vars:           cfg.Vars,
			Recorder:       cfg.Recorder}),
	}
}

type ExecHandler struct {
	RootModuleName string
	ModuleName     string
	BlockName      string
	Vars           cache.Cache[vars.Variable]
	Recorder       recorder.Recorder[record.Record]
	fnsMap         Map
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

func (r *ExecHandler) BlockRun(ctx context.Context, vertexName string, vCtx *types.VertexContext) bool {
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

func (r *ExecHandler) runInstances(ctx context.Context, vCtx *types.VertexContext) error {
	recorder := r.Recorder
	isForEach, items, err := r.getLoopItems(ctx, vCtx.BlockContext.Attributes)
	if err != nil {
		return err
	}
	g, ctx := errgroup.WithContext(ctx)
	for idx, item := range items.List() {
		localVars := map[string]any{}
		item := item
		localVars[render.LoopKeyItemsTotal] = items.Len()
		localVars[render.LoopKeyItemsIndex] = idx
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
			if err := r.fnsMap.Run(ctx, vCtx, localVars); err != nil {
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

func (r *ExecHandler) getLoopItems(ctx context.Context, attrs *types.KformBlockAttributes) (bool, *items, error) {
	log := log.FromContext(ctx)
	log.Info("getLoopItems", "attrs", attrs)
	renderer := &render.Renderer{
		Vars: r.Vars,
	}
	isForEach := false
	items := &items{}
	// forEach and count cannot be used together
	if attrs != nil && attrs.ForEach != nil {
		isForEach = true
		v, err := renderer.Render(ctx, *attrs.ForEach)
		if err != nil {
			return isForEach, items, errors.Wrap(err, "render loop forEach failed")
		}
		log.Info("getLoopItems forEach render output", "value type", reflect.TypeOf(v), "value", v)
		switch v := v.(type) {
		case []any:
			// in a list we return key = int, val = any
			for k, v := range v {
				log.Info("getLoopItems forEach insert item", "k", k, "v", v)
				items.Add(k, item{key: k, val: v})
			}
		case map[any]any:
			// in a list we return key = any, val = any
			idx := 0
			for k, v := range v {
				items.Add(idx, item{key: k, val: v})
				idx++
			}
		default:
			// in a regular value we return key = int, val = any
			items.Add(0, item{key: 0, val: v})
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

func getSetWithInt(i int) *items {
	items := &items{}
	for idx := 0; idx < i; idx++ {
		items.Add(idx, item{key: idx, val: idx})

	}
	return items
}

type items struct {
	m     sync.RWMutex
	items map[any]item
}

func (r *items) Add(k any, v item) {
	r.m.Lock()
	defer r.m.Unlock()
	r.items[k] = v
}

func (r *items) List() map[any]item {
	r.m.RLock()
	defer r.m.RUnlock()
	x := map[any]item{}
	for k, v := range r.items {
		x[k] = v
	}
	return x
}

func (r *items) Len() int {
	r.m.RLock()
	defer r.m.RUnlock()
	return len(r.items)
}
