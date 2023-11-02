package types

import (
	"fmt"

	"github.com/henderiw-nephio/kform/tools/pkg/dag"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	LoopKeyCountIndex = "count.index"
	LoopKeyForEachKey = "each.key"
	LoopKeyForEachVal = "each.value"
)

var LocalVars = map[string]struct{}{
	LoopKeyCountIndex: {},
	LoopKeyForEachKey: {},
	LoopKeyForEachVal: {},
}

type VertexContext struct {
	// FileName and Module provide context in which this
	FileName   string
	ModuleName string
	// BlockType determines which function we need to execute
	BlockType BlockType
	// BlockName has syntx <namespace>.<name>
	BlockName string
	// schema relevant for this blockType
	GVK schema.GroupVersionKind
	// provides the contextual data
	BlockContext    KformBlockContext
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

func (r *VertexContext) GetBlockDependencies() map[string]string {
	blockDeps := map[string]string{}
	fmt.Println("block dependencies", r.BlockType, r.Dependencies)
	for k, v := range r.Dependencies {
		// filter out the dependencies that refer to loop variables
		if _, ok := LocalVars[k]; !ok {
			blockDeps[k] = v
		}
	}
	return blockDeps
}
