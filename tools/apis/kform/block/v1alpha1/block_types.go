package v1alpha1

type K8sForm struct {
	Blocks []Block `json:"spec" yaml:"spec"`
}

type Block struct {
	BlockData   `json:",inline" yaml:",inline"`
	NestedBlock map[string]Block `json:",inline" yaml:",inline"`
}

type BlockData struct {
	Attributes map[string]any `json:"attributes,omitempty" yaml:"attributes,omitempty"`
	Object     any            `json:"object,omitempty" yaml:"object,omitempty"`
}

type BlockType string

const (
        BlockTypeUnknown  BlockType = "unknown"
        BlockTypeProvider BlockType = "provider"
        BlockTypeResource BlockType = "resource"
        BlockTypeData     BlockType = "data"
        BlockTypeVariable BlockType = "variable"
)

func GetBlockType(n string) BlockType {
        switch n {
        case "provider":
                return BlockTypeProvider
        case "resource":
                return BlockTypeResource
        case "data":
                return BlockTypeData
        case "variable":
                return BlockTypeVariable
        default:
                return BlockTypeUnknown
        }
}