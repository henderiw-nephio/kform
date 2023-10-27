package types

import (
	"fmt"
	"regexp"
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
			return fmt.Errorf("a dependency always need <namespace>.<name>, got: %s", dep)
		}

		r.deps = append(r.deps, parsedependencyString(strings.Join(depSplit[:2], ".")))
	}
	return nil
}

func parsedependencyString(inputString string) string {
	// Define a regular expression pattern to match special characters
	specialCharPattern := "[$&+,:;=?@#|'<>-^*()%!]"

	// Compile the regular expression
	regex := regexp.MustCompile(specialCharPattern)

	// Find all matches in the input string
	matches := regex.FindAllStringIndex(inputString, -1)

	if matches == nil {
		//fmt.Println("No special characters found.")
		return inputString
	} else {
		/*
			for _, match := range matches {
				start, end := match[0], match[1]
				fmt.Printf("Special character found at positions %d to %d: %s\n", start, end-1, inputString[start:end])
			}
		*/
		return inputString[0:matches[0][0]]
	}
}
