package types

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
