package types

import (
	"fmt"

	"k8s.io/apimachinery/pkg/runtime/schema"
)

type Variable struct {
	FileName       string
	BlockType      BlockType
	Provider       string
	GVK            schema.GroupVersionKind
	Attributes     map[string]any
	Object         any
	Dependencies   []string
	RenderedObject any
}

func (r Variable) GetContext(name string) string {
	return fmt.Sprintf("fileName=%s, name=%s, blockType=%s", r.FileName, name, string(r.BlockType))
}
