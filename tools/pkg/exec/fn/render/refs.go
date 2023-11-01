package render

import (
	"fmt"
	"strings"

	"github.com/henderiw-nephio/kform/tools/pkg/syntax/types"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
	"github.com/henderiw-nephio/kform/tools/pkg/util/sets"
)

func (r *Renderer) getRefsFromExpression(expr string) (map[string]any, error) {
	// we get the variables from the expression by parsing the information
	// from the $ variables
	refs, err := getRefsFromExpr(expr)
	if err != nil {
		return nil, err
	}
	fmt.Println("refs", refs.UnsortedList())
	// we create a new copy to ensure we dont loose the init values
	newVars := make(map[string]any, refs.Len()+len(r.LocalVars))
	for _, ref := range refs.UnsortedList() {
		// first lookup the vars in the initVars, which are the local vars for count
		// for_each, etc
		v, ok := r.LocalVars[ref]
		if !ok {
			// lookup in the local var failed, so lookup the real vars
			varVal, err := r.Vars.Get(cache.NSN{Name: ref})
			if err != nil {
				return nil, err
			}
			v = varVal.Data
		}
		newVars[ref] = v
	}
	for ref, v := range r.LocalVars {
		newVars[ref] = v
	}
	return newVars, nil
}

func getRefsFromExpr(expr string) (sets.Set[string], error) {
	//vars := []string{}
	vars := sets.New[string]()
	depsSplit := strings.Split(expr, "$")
	if len(depsSplit) == 1 {
		return vars, nil
	}
	for _, dep := range depsSplit[1:] {
		depSplit := strings.Split(dep, ".")
		if len(depSplit) < 2 {
			return vars, fmt.Errorf("a dependency always need <resource-type>.<resource-identifier>, got: %s", dep)
		}
		vars.Insert(types.ParseReferenceString(strings.Join(depSplit[:2], ".")))
		//vars = append(vars, types.ParseReferenceString(strings.Join(depSplit[:2], ".")))
	}
	return vars, nil
}
