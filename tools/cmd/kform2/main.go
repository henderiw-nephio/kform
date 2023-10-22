package main

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
	extv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

const (
	crdDir = "examples/crd"
)

func main() {
	fsEntries, err := os.ReadDir(crdDir)
	if err != nil {
		panic(err)
	}

	for _, fsEntry := range fsEntries {
		if !fsEntry.IsDir() {
			fName := filepath.Join(crdDir, fsEntry.Name())
			b, err := os.ReadFile(fName)
			if err != nil {
				panic(err)
			}

			crd := &extv1.CustomResourceDefinition{}
			if err := yaml.Unmarshal(b, crd); err != nil {

			}
		}
	}
}
