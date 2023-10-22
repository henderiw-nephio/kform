package render

import (
	"errors"
	"fmt"
	"regexp"
	"strings"

	blockv1alpha1 "github.com/henderiw-nephio/kform/tools/apis/kform/block/v1alpha1"
	"github.com/henderiw-nephio/kform/tools/pkg/dag"
)

type Vars interface {
	// GetVarsFromExpression return the variables with the real data from the variable store
	// if the variables or not found in the variable store an error is returned
	GetVarsFromExpression(expr string) (map[string]any, error)
}

func newVars(varStore dag.DAG[blockv1alpha1.Variable], initVars map[string]any) (Vars, error) {
	if varStore == nil {
		return nil, errors.New("cannot initialize vars w/o varStore")
	}
	if len(initVars) == 0 {
		initVars = map[string]any{}
	}

	return &vars{
		varStore: varStore,
		initVars: initVars,
	}, nil
}

type vars struct {
	varStore dag.DAG[blockv1alpha1.Variable]
	initVars map[string]any
}

func (r *vars) GetVarsFromExpression(expr string) (map[string]any, error) {
	// we get the variables from the expression by parsing the information
	// from the $ variables
	varKeys, err := getVarsFromExpr(expr)
	if err != nil {
		return nil, err
	}
	// we create a new copy to ensure we dont loose the init values
	newVars := make(map[string]any, len(varKeys)+len(r.initVars))
	for _, k := range varKeys {
		// first lookup the vars in the initVars, which are the local vars for count
		// for_each, etc
		v, ok := r.initVars[k]
		if !ok {
			// lookup in the local var failed, so lookup the real vars
			varVal, err := r.varStore.GetVertex(k)
			if err != nil {
				return nil, err
			}
			v = varVal.RenderedObject
		}
		newVars[k] = v
	}
	for k, v := range r.initVars {
		newVars[k] = v
	}
	return newVars, nil
}

func getVarsFromExpr(expr string) ([]string, error) {
	vars := []string{}
	depsSplit := strings.Split(expr, "$")
	if len(depsSplit) == 1 {
		return vars, nil
	}
	for _, dep := range depsSplit[1:] {
		depSplit := strings.Split(dep, ".")
		if len(depSplit) < 2 {
			return vars, fmt.Errorf("a dependency always need <resource-type>.<resource-identifier>, got: %s", dep)
		}

		depSplit[0] = clearString(depSplit[0])
		depSplit[1] = clearString(depSplit[1])
		vars = append(vars, strings.Join(depSplit[:2], "."))
	}
	return vars, nil
}

var nonAlphanumericRegex = regexp.MustCompile(`[^a-zA-Z0-9 ]+`)

func clearString(str string) string {
	return nonAlphanumericRegex.ReplaceAllString(str, "")
}
