package vctx

import (
	"github.com/henderiw-nephio/kform/tools/pkg/dag"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/types"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type VertexContext struct {
	// FileName and Module provide context in which this
	FileName   string
	ModuleName string
	// BlockType determines which function we need to execute
	BlockType types.BlockType
	// BlockName has syntx <namespace>.<name>
	BlockName string
	// schema relevant for this blockType
	GVK schema.GroupVersionKind
	// provides the contextual data
	BlockContext    types.KformBlockContext
	Dependencies    map[string]string
	ModDependencies map[string]string
	// only relevaant for blocktype resource and data
	Provider string
	// only relevant for blocktype module
	DAG dag.DAG[*VertexContext]
}

func (r *VertexContext) AddDAG(d dag.DAG[*VertexContext]) {
	r.DAG = d
}

func (r *VertexContext) GetDependencies() map[string]string {
	return r.Dependencies
}
