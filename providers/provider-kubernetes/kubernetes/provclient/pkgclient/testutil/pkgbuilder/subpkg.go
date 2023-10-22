package pkgbuilder

import (
	"os"
	"path/filepath"
)

type SubPkg struct {
	pkg *pkg

	Name string
}

// NewSubPkg returns a new subpackage for testing.
func NewSubPkg(name string) *SubPkg {
	return &SubPkg{
		pkg: &pkg{
			files: make(map[string]string),
		},
		Name: name,
	}
}

// WithFile configures the package to contain a file with the provided name
// and the given content.
func (sp *SubPkg) WithFile(name, content string) *SubPkg {
	sp.pkg.withFile(name, content)
	return sp
}

// WithSubPackages adds the provided packages as subpackages to the current
// package
func (sp *SubPkg) WithSubPackages(ps ...*SubPkg) *SubPkg {
	sp.pkg.withSubPackages(ps...)
	return sp
}

func buildSubPkg(path string, pkg *SubPkg, reposInfo ReposInfo) error {
	pkgPath := filepath.Join(path, pkg.Name)
	err := os.Mkdir(pkgPath, 0700)
	if err != nil {
		return err
	}
	err = buildPkg(pkgPath, pkg.pkg, pkg.Name, reposInfo)
	if err != nil {
		return err
	}
	for i := range pkg.pkg.subPkgs {
		subPkg := pkg.pkg.subPkgs[i]
		err := buildSubPkg(pkgPath, subPkg, reposInfo)
		if err != nil {
			return err
		}
	}
	return nil
}
