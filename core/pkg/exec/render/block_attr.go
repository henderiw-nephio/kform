package render

import "github.com/google/cel-go/cel"

const (
	BlockAttrCount   = "count"
	BlockAttrForEach = "for_each"

	KeyCountIndex = "count.index"
	KeyForEachKey = "each.key"
	KeyForEachVal = "each.value"
)

var BlockAttr = map[string]*cel.Type{
	BlockAttrCount:   cel.IntType,
	BlockAttrForEach: cel.DynType,
}
