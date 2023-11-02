package fns

import (
	"context"
	"fmt"
	"sort"
	"sync"

	"github.com/henderiw-nephio/kform/tools/pkg/exec/fn"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/record"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/vars"
	"github.com/henderiw-nephio/kform/tools/pkg/recorder"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/types"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
)

type Initializer func(*Config) fn.BlockInstanceRunner

type Map interface {
	fn.BlockInstanceRunner
}

type Config struct {
	RootModuleName string
	Vars           cache.Cache[vars.Variable]
	Recorder       recorder.Recorder[record.Record]
}

func NewMap(ctx context.Context, cfg *Config) Map {
	if cfg == nil {
		cfg = &Config{}
	}
	return &fnMap{
		cfg: *cfg,
		fns: map[types.BlockType]Initializer{
			types.BlockTypeModule:   NewModuleFn,
			types.BlockTypeInput:    NewInputFn,
			types.BlockTypeOutput:   NewLocalOrOutputFn,
			types.BlockTypeLocal:    NewLocalOrOutputFn,
			types.BlockTypeResource: NewResourceFn,
			types.BlockTypeData:     NewResourceFn,
			types.BlockTypeRoot:     NewRootFn,
		},
	}
}

type fnMap struct {
	cfg Config
	m   sync.RWMutex
	fns map[types.BlockType]Initializer
}

func (r *fnMap) getInitializedBlockTypes() []string {
	//r.m.Lock()
	//defer r.m.Unlock()
	rfns := make([]string, 0, len(r.fns))
	for blockType := range r.fns {
		rfns = append(rfns, string(blockType))
	}
	sort.Strings(rfns)
	return rfns
}

func (r *fnMap) init(blockType types.BlockType) (fn.BlockInstanceRunner, error) {
	//r.m.RLock()
	//defer r.m.RUnlock()
	initFn, ok := r.fns[blockType]
	if !ok {
		return nil, fmt.Errorf("blockType not initialized, got %s, initialized blocktypes: %v", blockType, r.getInitializedBlockTypes())
	}
	return initFn(&r.cfg), nil

}

func (r *fnMap) Run(ctx context.Context, vctx *types.VertexContext, localVars map[string]any) error {
	r.m.RLock()
	defer r.m.RUnlock()
	fn, err := r.init(vctx.BlockType)
	if err != nil {
		return err
	}
	return fn.Run(ctx, vctx, localVars)
}
