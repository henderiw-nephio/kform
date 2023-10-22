package types

import (
	"fmt"
	"strings"
)

type Renderer interface {
	GatherDependencies(x any) error
	GetDependencies() []string
}

func NewRenderer() Renderer {
	return &renderer{
		deps: []string{},
	}
}

type renderer struct {
	deps []string
}

func (r *renderer) GetDependencies() []string {
	return r.deps
}

func (r *renderer) GatherDependencies(x any) error {
	switch x := x.(type) {
	case map[string]any:
		for _, v := range x {
			if err := r.GatherDependencies(v); err != nil {
				return err
			}
		}
	case []any:
		for _, t := range x {
			if err := r.GatherDependencies(t); err != nil {
				return err
			}
		}
	case string:
		if err := r.getExprDependencies(x); err != nil {
			return err
		}
	}
	return nil
}

func (r *renderer) getExprDependencies(expr string) error {
	depsSplit := strings.Split(expr, "$")
	if len(depsSplit) == 1 {
		return nil
	}
	for _, dep := range depsSplit[1:] {
		depSplit := strings.Split(dep, ".")
		if len(depSplit) < 2 {
			return fmt.Errorf("a dependency always need <resource-type>.<resource-identifier>, got: %s", dep)
		}
		r.deps = append(r.deps, strings.Join(depSplit[:2], "."))
	}
	return nil
}
