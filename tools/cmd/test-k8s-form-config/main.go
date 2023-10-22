package main

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/henderiw-nephio/kform/syntax/apis/k8sform/v1alpha1"
	"github.com/henderiw-nephio/kform/syntax/pkg/syntax"
	"gopkg.in/yaml.v3"
)

const dir = "examples/config"

func main() {
	fsEntries, err := os.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	kforms := []*v1alpha1.K8sFormCtx{}
	for _, fsEntry := range fsEntries {
		if !fsEntry.IsDir() {
			//fmt.Println(fsEntry.Name())
			fName := filepath.Join(dir, fsEntry.Name())
			b, err := os.ReadFile(fName)
			if err != nil {
				panic(err)
			}
			//fmt.Println(string(b))
			kform := v1alpha1.K8sForm{}
			if err := yaml.Unmarshal(b, &kform); err != nil {
				panic(err)
			}
			kforms = append(kforms, &v1alpha1.K8sFormCtx{
				FileName: fName,
				K8sForm:  kform,
			})
			/*
				b, err = json.MarshalIndent(k8sFormConfig, "", "  ")
				if err != nil {
					panic(err)
				}
				fmt.Println(string(b))
			*/
		}
	}
	ctx := context.Background()
	for _, kform := range kforms {
		for _, block := range kform.Blocks {
			for blockType := range block.NestedBlock {
				fmt.Println("blockType", blockType)
			}
		}
	}
	p := syntax.NewParser(ctx, kforms)
	execCfg, diags := p.Parse(ctx)
	if diags.Error() != nil {
		panic(err)
	}

	fmt.Println("diags...")
	for _, diag := range diags {
		fmt.Printf("  %v\n", diag)
	}
	fmt.Println("providers...")
	for name, provider := range execCfg.GetProviders().GetVertices() {
		fmt.Printf("  name: %v, provider: %v\n", name, provider)
	}
	fmt.Println("vars...")
	for name, v := range execCfg.GetVars().GetVertices() {
		fmt.Printf("  name: %v\n", name)
		if len(v.Attributes) != 0 {
			fmt.Printf("    attributes: %v\n", v.Attributes)
		}
		if v.Provider != "" {
			fmt.Printf("    provider: %v\n", v.Provider)
			fmt.Printf("    gvk: %s\n", v.GVK.String())
		}
		if len(v.Dependencies) > 0 {
			fmt.Printf("    dep: %v\n", v.Dependencies)
		}

		down := execCfg.GetVars().GetDownVertexes(name)
		if len(down) != 0 {
			fmt.Printf("    down: %v\n", down)
		}
		up := execCfg.GetVars().GetUpVertexes(name)
		if len(up) != 0 {
			fmt.Printf("    up: %v\n", up)
		}
	}
}
