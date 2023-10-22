package render

import (
	"regexp"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types/ref"
	"github.com/google/cel-go/common/types/traits"
)

const specialCharExpr = "[$&+,:;=?@#|'<>-^*()%!]"

func IsCelExpression(s string) (bool, error) {
	return regexp.MatchString(specialCharExpr, s)
}

func getCelEnv(vars map[string]any) (*cel.Env, error) {
	var opts []cel.EnvOption
	opts = append(opts, cel.HomogeneousAggregateLiterals())
	//opts = append(opts, cel.EagerlyValidateDeclarations(true), cel.DefaultUTCTimeZone(true))
	//opts = append(opts, library.ExtensionLibs...)
	opts = append(opts, cel.Function("concat",
		cel.MemberOverload(
			"size_list",
			[]*cel.Type{cel.StringType},
			cel.StringType,
			cel.FunctionBinding(concat)),
	))

	for k := range vars {
		if ct, ok := BlockAttr[k]; ok {
			// for builtin variables like count, for_each we know the type
			// this provide the type safety
			opts = append(opts, cel.Variable(k, ct))
		} else {
			opts = append(opts, cel.Variable(k, cel.DynType))
		}
	}

	return cel.NewEnv(opts...)
}

func concat(args ...ref.Val) ref.Val {
	var v ref.Val
	for _, arg := range args {
		v = arg.(traits.Adder).Add(v)
	}
	return v
}
