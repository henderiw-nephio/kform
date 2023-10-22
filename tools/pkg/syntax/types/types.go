package types

import (
	"context"
	"fmt"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/tools/apis/kform/block/v1alpha1"
	blockv1alpha1 "github.com/henderiw-nephio/kform/tools/apis/kform/block/v1alpha1"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/sctx"
)

var BlockTypes = map[blockv1alpha1.BlockType]BlockInitializer{
	blockv1alpha1.BlockTypeProvider: newProvider,
	blockv1alpha1.BlockTypeVariable: newVariable,
	blockv1alpha1.BlockTypeResource: newResource,
	blockv1alpha1.BlockTypeData:     newData,
}

type BlockInitializer func(n string) Block

func GetBlockTypeNames() []string {
	s := make([]string, 0, len(BlockTypes))
	for n := range BlockTypes {
		s = append(s, string(n))
	}
	return s
}

func GetBlock(n string) (Block, error) {
	bi, ok := BlockTypes[blockv1alpha1.GetBlockType(n)]
	if !ok {
		return nil, fmt.Errorf("cannot get blockType for %s, supported blocktypes %v", n, GetBlockTypeNames())
	}
	return bi(n), nil
}

type Block interface {
	WithRecorder(diag.Recorder)
	GetName() string
	GetLevel() int
	ProcessBlock(context.Context, *v1alpha1.Block) context.Context
	AddData(context.Context)
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
	rn := NewRenderer()
	if err := rn.GatherDependencies(x); err != nil {
		r.recorder.Record(diag.DiagFromErrWithContext(sctx.GetContext(ctx), err))
		return []string{}
	}
	return rn.GetDependencies()
}
