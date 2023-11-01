package render

import (
	"regexp"
	"strings"

	"github.com/google/cel-go/cel"
)

const (
	LoopAttrCount   = "count"
	LoopAttrForEach = "forEach"

	LoopKeyCountIndex = "count.index"
	LoopKeyForEachKey = "each.key"
	LoopKeyForEachVal = "each.value"
)

var LocalVars = map[string]struct{}{
	LoopKeyCountIndex: {},
	LoopKeyForEachKey: {},
	LoopKeyForEachVal: {},
}

var LoopAttr = map[string]*cel.Type{
	LoopAttrCount:   cel.IntType,
	LoopAttrForEach: cel.DynType,
}

const specialCharExpr = "[$&+,:;=?@#|'<>-^*()%!]"

func IsCelExpression(s string) (bool, error) {
	return regexp.MatchString(specialCharExpr, s)
}

func getCelEnv(vars map[string]any) (*cel.Env, error) {
	// replace reference . to _ otherwise cell complains to do json lookups of a struct
	newvars := map[string]any{}
	for k, v := range vars {
		newvars[strings.ReplaceAll(k, ".", "_")] = v
	}
	//fmt.Println("cellvars", newvars)

	var opts []cel.EnvOption
	opts = append(opts, cel.HomogeneousAggregateLiterals())
	//opts = append(opts, cel.EagerlyValidateDeclarations(true), cel.DefaultUTCTimeZone(true))
	//opts = append(opts, library.ExtensionLibs...)

	for k := range newvars {
		//fmt.Println("cellvar", k)
		// for builtin variables like count, forEach we know the type
		// this provide more type safety
		if ct, ok := LoopAttr[k]; ok {
			opts = append(opts, cel.Variable(k, ct))
		} else {
			opts = append(opts, cel.Variable(k, cel.DynType))
		}
	}
	return cel.NewEnv(opts...)
}
