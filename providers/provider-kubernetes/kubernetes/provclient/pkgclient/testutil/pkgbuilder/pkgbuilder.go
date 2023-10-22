package pkgbuilder

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// RootPkg is a package without any parent package.
type RootPkg struct {
	pkg *pkg
}

// NewRootPkg creates a new package for testing.
func NewRootPkg() *RootPkg {
	return &RootPkg{
		pkg: &pkg{
			files: make(map[string]string),
		},
	}
}

// WithFile configures the package to contain a file with the provided name
// and the given content.
func (rp *RootPkg) WithFile(name, content string) *RootPkg {
	rp.pkg.withFile(name, content)
	return rp
}

// WithSubPackages adds the provided packages as subpackages to the current
// package
func (rp *RootPkg) WithSubPackages(ps ...*SubPkg) *RootPkg {
	rp.pkg.withSubPackages(ps...)
	return rp
}

// ExpandPkg writes the provided package to disk. The name of the root package
// will just be set to "base".
func (rp *RootPkg) ExpandPkg(t *testing.T, reposInfo ReposInfo) string {
	return rp.ExpandPkgWithName(t, "base", reposInfo)
}

// ExpandPkgWithName writes the provided package to disk and uses the given
// rootName to set the value of the package directory and the metadata.name
// field of the root package.
func (rp *RootPkg) ExpandPkgWithName(t *testing.T, rootName string, reposInfo ReposInfo) string {
	dir, err := os.MkdirTemp("", "test-kpt-builder-")
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	err = rp.Build(dir, rootName, reposInfo)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	return filepath.Join(dir, rootName)
}

// Build outputs the current data structure as a set of (nested) package
// in the provided path.
func (rp *RootPkg) Build(path string, pkgName string, reposInfo ReposInfo) error {
	pkgPath := filepath.Join(path, pkgName)
	err := os.Mkdir(pkgPath, 0700)
	if err != nil {
		return err
	}
	if rp == nil {
		return nil
	}
	err = buildPkg(pkgPath, rp.pkg, pkgName, reposInfo)
	if err != nil {
		return err
	}
	for i := range rp.pkg.subPkgs {
		subPkg := rp.pkg.subPkgs[i]
		err := buildSubPkg(pkgPath, subPkg, reposInfo)
		if err != nil {
			return err
		}
	}
	return nil
}

type ReposInfo interface {
	ResolveRepoRef(repoRef string) (string, bool)
	ResolveCommitIndex(repoRef string, index int) (string, bool)
}
