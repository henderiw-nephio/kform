package pkgbuilder

import (
	"fmt"
	"os"
	"path/filepath"
)

// Pkg represents a package that can be created on the file system
// by using the Build function
type pkg struct {
	files map[string]string

	subPkgs []*SubPkg
}

// withFile configures the package to contain a file with the provided name
// and the given content.
func (p *pkg) withFile(name, content string) {
	p.files[name] = content
}

// withSubPackages adds the provided packages as subpackages to the current
// package
func (p *pkg) withSubPackages(ps ...*SubPkg) {
	p.subPkgs = append(p.subPkgs, ps...)
}

func buildPkg(pkgPath string, pkg *pkg, pkgName string, reposInfo ReposInfo) error {
	for name, content := range pkg.files {
		filePath := filepath.Join(pkgPath, name)
		_, err := os.Stat(filePath)
		if err != nil && !os.IsNotExist(err) {
			return err
		}
		if !os.IsNotExist(err) {
			return fmt.Errorf("file %s already exists", name)
		}
		err = os.WriteFile(filePath, []byte(content), 0600)
		if err != nil {
			return err
		}
	}
	return nil
}
