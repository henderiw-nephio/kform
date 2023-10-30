package exec

import (
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/types"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

type VertexContext struct {
	FileName     string
	Module       string
	BlockType    types.BlockType
	Provider     string
	GVK          schema.GroupVersionKind
	BlockContext types.KformBlockContext
	Dependencies map[string]string
}
