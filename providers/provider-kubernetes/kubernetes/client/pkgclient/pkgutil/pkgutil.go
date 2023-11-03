package pkgutil

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/copyutil"
	"sigs.k8s.io/kustomize/kyaml/kio"
	"sigs.k8s.io/kustomize/kyaml/kio/kioutil"
)

// WalkPackage walks the package defined at src and provides a callback for
// every folder and file. .git folder are excluded.
func WalkPackage(src string, c func(string, os.FileInfo, error) error) error {
	excludedDirs := make(map[string]bool)
	return filepath.Walk(src, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return c(path, info, err)
		}
		// don't walk the .git dir
		if path != src {
			rel := strings.TrimPrefix(path, src)
			if copyutil.IsDotGitFolder(rel) {
				return nil
			}
		}

		for dir := range excludedDirs {
			if strings.HasPrefix(path, dir) {
				return nil
			}
		}

		return c(path, info, err)
	})
}

func GetPackage(src string, m ...string) (*kio.PackageBuffer, error) {
	inputs := []kio.Reader{}
	if err := WalkPackage(src, func(s string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, s)
		if err != nil {
			return err
		}

		// exclude directories
		if !info.IsDir() {
			if includeFile(relPath, m) {
				yamlFile, err := os.ReadFile(s)
				if err != nil {
					return err
				}
				inputs = append(inputs, &kio.ByteReader{
					Reader: strings.NewReader(string(yamlFile)),
					SetAnnotations: map[string]string{
						kioutil.PathAnnotation: relPath,
					},
					DisableUnwrapping: true,
				})
			}
		}
		return nil
	}); err != nil {
		return nil, err
	}
	var pb kio.PackageBuffer
	err := kio.Pipeline{
		Inputs:  inputs,
		Filters: []kio.Filter{},
		Outputs: []kio.Writer{&pb},
	}.Execute()
	if err != nil {
		return nil, fmt.Errorf("kio error: %s", err)
	}

	return &pb, nil
}

func includeFile(path string, match []string) bool {
	for _, m := range match {
		file := filepath.Base(path)
		if matched, err := filepath.Match(m, file); err == nil && matched {
			return true
		}
	}
	return false
}
