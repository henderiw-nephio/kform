package render

import (
	"context"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/henderiw-nephio/kform/tools/pkg/dag"
	"github.com/henderiw/logger/log"
	blockv1alpha1 "github.com/henderiw-nephio/kform/tools/apis/kform/block/v1alpha1"
)

type Renderer interface {
	Render(context.Context, any) (any, error)
}

func New(varStore dag.DAG[blockv1alpha1.Variable], initVars map[string]any) (Renderer, error) {
	vars, err := newVars(varStore, initVars)
	if err != nil {
		return nil, err
	}
	return &renderer{
		vars: vars,
	}, nil
}

type renderer struct {
	vars Vars
}

func (r *renderer) Render(ctx context.Context, v any) (any, error) {
	log := log.FromContext(ctx)

	var err error
	switch x := v.(type) {
	case map[string]any:
		for k, v := range x {
			x[k], err = r.Render(ctx, v)
			if err != nil {
				return nil, err
			}
		}
	case []any:
		for i, v := range x {
			x[i], err = r.Render(ctx, v)
			if err != nil {
				return nil, err
			}
		}
	case string:
		isCelExpr, err := IsCelExpression(x)
		if err != nil {
			return nil, err
		}
		if isCelExpr {

			// get the variables from the expression
			varsForExpr, err := r.vars.GetVarsFromExpression(x)
			if err != nil {
				return 0, err
			}

			// remove the $ from the string, otherwise cell will complain
			x = strings.ReplaceAll(x, "$", "")
			//x = strings.ReplaceAll(x, "\"", "")
			log.Info("expression", "expr", x)
			env, err := getCelEnv(varsForExpr)
			if err != nil {
				log.Error("cel environment failed", "error", err)
				return nil, err
			}
			ast, iss := env.Compile(x)
			if iss.Err() != nil {
				log.Error("compile env to ast failed", "error", iss.Err())
				return nil, err
			}
			_, err = cel.AstToCheckedExpr(ast)
			if err != nil {
				log.Error("ast to checked expression failed", "error", err)
				return nil, err
			}
			prog, err := env.Program(ast,
				cel.EvalOptions(cel.OptOptimize),
				// TODO: uncomment after updating to latest k8s
				//cel.OptimizeRegex(library.ExtensionLibRegexOptimizations...),
			)
			if err != nil {
				log.Error("env program failed", "error", iss.Err())
				return nil, err
			}

			val, _, err := prog.Eval(varsForExpr)
			if err != nil {
				log.Error("evaluate program failed", "error", iss.Err())
				return nil, err
			}

			/*
				result, err := val.ConvertToNative(reflect.TypeOf(""))
				if err != nil {
					log.Error("value conversion failed", "error", iss.Err())
					return nil, err
				}

				s, ok := result.(string)
				if !ok {
					return nil, fmt.Errorf("expression returned non-string value: %v", result)
				}
			*/
			return val.Value(), nil
		}
	}
	return v, err
}
