package pkgio

import (
	"io/fs"
	"path/filepath"
	"sync"

	kformpkgmetav1alpha1 "github.com/henderiw-nephio/kform/tools/apis/kform/pkg/meta/v1alpha1"
	"github.com/henderiw-nephio/kform/tools/pkg/fsys"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio/ignore"
)

type pkgReader struct {
	fsys           fsys.FS
	rootPath       string
	parentPkgPath  string
	pkgName        string
	pkgKind        kformpkgmetav1alpha1.PkgKind
	matchFilesGlob []string
	ignoreRules    *ignore.Rules
	skipDir        bool
}

func (r *pkgReader) Read(data *Data) (*Data, error) {
	paths, err := r.getPaths()
	if err != nil {
		return data, err
	}
	return r.readFileContent(paths, data)
}

func (r *pkgReader) getPaths() ([]string, error) {
	// collect the paths
	paths := []string{}
	if err := r.fsys.Walk(".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			// Directory-based ignore rules involve skipping the entire
			// contents of that directory.
			if r.ignoreRules.Ignore(path, d) {
				return filepath.SkipDir
			}
			if r.skipDir && d.Name() != "." {
				return filepath.SkipDir
			}
			return nil
		}
		if r.ignoreRules.Ignore(path, d) {
			return nil
		}
		// process glob
		if match, err := r.shouldSkipFile(path); err != nil {
			return err
		} else if match {
			// skip the file
			return nil
		}
		paths = append(paths, path)
		return nil
	}); err != nil {
		return nil, err
	}
	return paths, nil
}

func (r *pkgReader) readFileContent(paths []string, data *Data) (*Data, error) {
	var wg sync.WaitGroup
	for _, path := range paths {
		path := path
		wg.Add(1)
		var err error
		go func() {
			defer wg.Done()
			var d []byte
			d, err = r.fsys.ReadFile(path)
			if err != nil {
				return
			}
			data.Add(path, d)
		}()
		if err != nil {
			return nil, err
		}
	}
	wg.Wait()

	return data, nil
}

func (r *pkgReader) Write(*Data) error {
	return nil
}

func (r *pkgReader) shouldSkipFile(path string) (bool, error) {
	for _, g := range r.matchFilesGlob {
		if match, err := filepath.Match(g, filepath.Base(path)); err != nil {
			// if err we should skip the file
			return true, err
		} else if match {
			// if matchw e should include the file
			return false, nil
		}
	}
	// if no match we should skip the file
	return true, nil
}
