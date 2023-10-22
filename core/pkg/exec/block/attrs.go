package block

import "github.com/henderiw-nephio/kform/core/pkg/exec/render"

type Attrs map[string]any

func (r Attrs) isCountAttrPresent() bool {
	if _, ok := r[render.BlockAttrCount]; ok {
		return true
	}
	return false
}

func (r Attrs) isForEachAttrPresent() bool {
	if _, ok := r[render.BlockAttrForEach]; ok {
		return true
	}
	return false
}

func (r Attrs) isLoopAttrPresent() bool {
	return r.isCountAttrPresent() || r.isForEachAttrPresent()
}
