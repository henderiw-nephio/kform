package v1alpha1

import "fmt"

type Provider struct {
	FileName   string
	BlockType  BlockType
	Attributes map[string]any
	Object     any
}

func (r Provider) GetContext(name string) string {
	return fmt.Sprintf("fileName=%s, name=%s, blockType=%s", r.FileName, name, string(r.BlockType))
}
