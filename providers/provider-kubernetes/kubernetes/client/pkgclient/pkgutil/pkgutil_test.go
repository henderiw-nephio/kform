package pkgutil

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/henderiw-nephio/kform/providers/provider-kubernetes/kubernetes/client/pkgclient/testutil"
	"github.com/henderiw-nephio/kform/providers/provider-kubernetes/kubernetes/client/pkgclient/testutil/pkgbuilder"
	"github.com/stretchr/testify/assert"
)

func TestWalkPackage(t *testing.T) {
	testCases := map[string]struct {
		pkg      *pkgbuilder.RootPkg
		expected []string
	}{
		"walks subdirectories of a package": {
			pkg: pkgbuilder.NewRootPkg().
				WithFile("abc.yaml", "42").
				WithFile("test.txt", "Hello, World!").
				WithSubPackages(
					pkgbuilder.NewSubPkg("foo").
						WithFile("def.yaml", "123"),
				),
			expected: []string{
				".",
				"abc.yaml",
				"foo",
				"foo/def.yaml",
				"test.txt",
			},
		},
		"ignores .git folder": {
			pkg: pkgbuilder.NewRootPkg().
				WithFile("abc.yaml", "42").
				WithSubPackages(
					pkgbuilder.NewSubPkg(".git").
						WithFile("INDEX", "ABC123"),
				),
			expected: []string{
				".",
				"abc.yaml",
			},
		},
	}

	for tn, tc := range testCases {
		t.Run(tn, func(t *testing.T) {
			pkgPath := tc.pkg.ExpandPkg(t, testutil.EmptyReposInfo)
			fmt.Println(pkgPath)

			var visited []string
			if err := WalkPackage(pkgPath, func(s string, info os.FileInfo, err error) error {
				if err != nil {
					return err
				}

				relPath, err := filepath.Rel(pkgPath, s)
				if err != nil {
					return err
				}
				visited = append(visited, relPath)
				return nil
			}); !assert.NoError(t, err) {
				t.FailNow()
			}

			sort.Strings(visited)

			assert.Equal(t, tc.expected, visited)
		})
	}
}
