package pkgio

import (
	"io/fs"
	"path/filepath"
	"reflect"
	"sort"
	"testing"
	"testing/fstest"

	kformpkgmetav1alpha1 "github.com/henderiw-nephio/kform/tools/apis/kform/pkg/meta/v1alpha1"
	"github.com/henderiw-nephio/kform/tools/pkg/fsys"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio/ignore"
	"github.com/stretchr/testify/assert"
)

func TestPkgReadInitRead(t *testing.T) {
	fulldata := map[string]string{
		"README.md":      ("package foo\n"),
		".kformignore":   ("\n"),
		"KformFile.yaml": ("\n"),
	}
	fullFiles := fstest.MapFS{}
	for path, data := range fulldata {
		fullFiles[path] = &fstest.MapFile{Data: []byte(data)}
	}
	fullFS := fsys.NewMemFS("", fullFiles)

	emptyFiles := fstest.MapFS{}
	emptyFS := fsys.NewMemFS("", emptyFiles)

	pkgPath := "example/module"
	cases := map[string]struct {
		fsys          fsys.FS
		reader        Reader
		writer        Writer
		expectedData  map[string]string
		expectedFiles []string
	}{
		"Empty": {
			fsys: emptyFS,
			reader: &PkgReader{
				Fsys:           emptyFS,
				MatchFilesGlob: []string{IgnoreFileMatch[0], ReadmeFileMatch[0], PkgFileMatch[0]},
				IgnoreRules:    ignore.Empty(""),
			},
			writer: &pkgInitWriter{
				fsys:          emptyFS,
				rootPath:      pkgPath,
				parentPkgPath: filepath.Dir(pkgPath),
				pkgName:       filepath.Base(pkgPath),
			},
			expectedData: map[string]string{
				"README.md":      (""),
				".kformignore":   (""),
				"KformFile.yaml": (""),
			},
			expectedFiles: []string{IgnoreFileMatch[0], ReadmeFileMatch[0], PkgFileMatch[0]},
		},
		"Full": {
			fsys: fullFS,
			reader: &PkgReader{
				Fsys:           fullFS,
				MatchFilesGlob: []string{IgnoreFileMatch[0], ReadmeFileMatch[0], PkgFileMatch[0]},
				IgnoreRules:    ignore.Empty(""),
			},
			writer: &pkgInitWriter{
				fsys:          emptyFS,
				rootPath:      pkgPath,
				parentPkgPath: filepath.Dir(pkgPath),
				pkgName:       filepath.Base(pkgPath),
				pkgKind:       kformpkgmetav1alpha1.PkgKindModule,
			},
			expectedData:  fulldata,
			expectedFiles: []string{IgnoreFileMatch[0], ReadmeFileMatch[0], PkgFileMatch[0]},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {

			p := Pipeline{
				Inputs:  []Reader{tc.reader},
				Outputs: []Writer{tc.writer},
			}
			err := p.Execute()
			assert.NoError(t, err)

			got := map[string]string{}
			gotFiles := []string{}
			if err := tc.fsys.Walk(".", func(path string, d fs.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if d.IsDir() {
					return nil
				}
				b, err := tc.fsys.ReadFile(path)
				if err != nil {
					return err
				}
				gotFiles = append(gotFiles, path)
				got[path] = string(b)
				return nil
			}); err != nil {
				assert.NoError(t, err)
			}

			sort.Strings(tc.expectedFiles)
			sort.Strings(gotFiles)

			if !reflect.DeepEqual(tc.expectedFiles, gotFiles) {
				t.Errorf("want: %v, got: %v", tc.expectedFiles, gotFiles)
			}

		})
	}

}
