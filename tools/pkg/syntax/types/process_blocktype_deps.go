package types

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"sync"
)

type Renderer interface {
	GatherDependencies(ctx context.Context, x any) error
	GetDependencies() map[string]string
	GetModuleOutputDependencies() map[string]string
}

func NewRenderer() Renderer {
	return &renderer{
		deps:    map[string]string{},
		modDeps: map[string]string{},
	}
}

type renderer struct {
	m       sync.RWMutex
	deps    map[string]string
	modDeps map[string]string
}

func (r *renderer) GetDependencies() map[string]string {
	r.m.RLock()
	defer r.m.RLock()
	d := make(map[string]string, len(r.deps))
	for k, v := range r.deps {
		d[k] = v
	}
	return d
}

func (r *renderer) GetModuleOutputDependencies() map[string]string {
	r.m.RLock()
	defer r.m.RLock()
	d := make(map[string]string, len(r.modDeps))
	for k, v := range r.modDeps {
		d[k] = v
	}
	return d
}

func (r *renderer) GatherDependencies(ctx context.Context, x any) error {
	/*
		blockType := cctx.GetContextValue[string](ctx, CtxKeyBlockType)
		if blockType == string(BlockTypeOutput) {
			fmt.Println("gatherdeps", x)
		}
	*/
	switch x := x.(type) {
	case map[string]any:
		for _, v := range x {
			if err := r.GatherDependencies(ctx, v); err != nil {
				return err
			}
		}
	case map[any]any:
		for _, v := range x {
			if err := r.GatherDependencies(ctx, v); err != nil {
				return err
			}
		}
	case []any:
		for _, t := range x {
			if err := r.GatherDependencies(ctx, t); err != nil {
				return err
			}
		}
	case string:
		if err := r.getRefsFromExpr(ctx, x); err != nil {
			return err
		}
	default:
		/*
			if blockType == string(BlockTypeOutput) {
				fmt.Println("gatherdeps default", reflect.TypeOf(x))
			}
		*/
	}
	return nil
}

func (r *renderer) getRefsFromExpr(ctx context.Context, expr string) error {
	depsSplit := strings.Split(expr, "$")
	if len(depsSplit) == 1 {
		return nil
	}
	for _, dep := range depsSplit[1:] {
		depSplit := strings.Split(dep, ".")
		if len(depSplit) < 2 {
			return fmt.Errorf("a dependency always need <namespace>.<name>, got: %s", dep)
		}

		r.addDependency(ParseReferenceString(strings.Join(depSplit[:2], ".")), GetContext(ctx))

		if depSplit[0] == "module" {
			if len(depSplit) < 3 {
				return fmt.Errorf("a module dependency always need <namespace>.<name>.<output>, got: %s", dep)
			}
			r.addModDependency(ParseReferenceString(strings.Join(depSplit[:3], ".")), GetContext(ctx))
			//r.modDeps = append(r.modDeps, parsedependencyString(strings.Join(depSplit[:3], ".")))
		}
	}
	return nil
}

func ParseReferenceString(inputString string) string {
	// Define a regular expression pattern to match special characters
	specialCharPattern := "[$&+,:;=?@#|'<>-^*()%!]"

	// Compile the regular expression
	regex := regexp.MustCompile(specialCharPattern)

	// Find all matches in the input string
	matches := regex.FindAllStringIndex(inputString, -1)

	//fmt.Println("inputString", inputString)
	if matches == nil {
		//fmt.Println("No special characters found.")
		//fmt.Println("outputString no match", inputString)
		return inputString
	} else {
		for matchIdx, match := range matches {
			start, end := match[0], match[1]
			//fmt.Printf("Special character found at positions %d to %d: %s\n", start, end-1, inputString[start:end])
			// if the special char is a lowercase/upercase letter or - or _ we continue
			re := regexp.MustCompile(`^[a-zA-Z]+[a-zA-Z0-9_-]*$`)
			if re.Match([]byte(inputString[start:end])) {
				continue
			}
			//fmt.Println("outputString match", inputString)
			return inputString[0:matches[matchIdx][0]]
		}
		//fmt.Println("outputString match out", inputString)
		return inputString
	}
}

func (r *renderer) addDependency(k string, v string) {
	r.m.Lock()
	defer r.m.Unlock()
	r.deps[k] = v
}

func (r *renderer) addModDependency(k string, v string) {
	r.m.Lock()
	defer r.m.Unlock()
	r.modDeps[k] = v
}
