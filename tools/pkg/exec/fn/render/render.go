package render

import (
	"context"
	"strings"

	"github.com/google/cel-go/cel"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/vars"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
	"github.com/henderiw/logger/log"
)

type Renderer struct {
	Vars      cache.Cache[vars.Variable]
	LocalVars map[string]any
}

func (r *Renderer) Render(ctx context.Context, v any) (any, error) {
	var err error
	switch x := v.(type) {
	case map[string]any:
		for k, v := range x {
			x[k], err = r.Render(ctx, v)
			if err != nil {
				return nil, err
			}
		}
		return v, err
	case map[any]any:
		for k, v := range x {
			if x, ok := k.(string); ok {
				k, err = r.handleString(ctx, x)
				if err != nil {
					return nil, err
				}
			}
			x[k], err = r.Render(ctx, v)
			if err != nil {
				return nil, err
			}
		}
		return v, err
	case []any:
		for i, v := range x {
			x[i], err = r.Render(ctx, v)
			if err != nil {
				return nil, err
			}
		}
		return v, err
	case string:
		return r.handleString(ctx, x)
	default:
		return v, err
	}
}

func (r *Renderer) handleString(ctx context.Context, x string) (any, error) {
	log := log.FromContext(ctx)
	isCelExpr, err := IsCelExpression(x)
	if err != nil {
		return nil, err
	}
	if isCelExpr {
		// get the variables from the expression
		varsForExpr, err := r.getRefsFromExpression(x)
		if err != nil {
			return nil, err
		}
		// remove the $ from the string, otherwise cell will complain
		x = strings.ReplaceAll(x, "$", "")
		// replace reference . to _ otherwise cell complains to do json lookups of a struct
		for origRef := range varsForExpr {
			newRef := strings.ReplaceAll(origRef, ".", "_")
			//fmt.Println("origRef", origRef)
			//fmt.Println("newRef", newRef)
			x = strings.ReplaceAll(x, origRef, newRef)
		}
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
			log.Error("env program failed", "expression", x, "error", err)
			return nil, err
		}

		// replace the reference since cel does not deal with . for json references
		newVarsForExpr := map[string]any{}
		for k, v := range varsForExpr {
			newVarsForExpr[strings.ReplaceAll(k, ".", "_")] = v
		}
		//fmt.Println("newVarsForExpr", newVarsForExpr)

		val, _, err := prog.Eval(newVarsForExpr)
		if err != nil {
			log.Error("evaluate program failed", "expression", x, "error", err)
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
		// provide a return
	}
	return x, nil
}
