package pkgio

import (
	"fmt"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/henderiw-nephio/kform/tools/pkg/fsys"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio/ignore"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio/oci"
)

type PkgBuildReadWriter interface {
	Reader
	Writer
}

func NewPkgBuildReadWriter(path string) PkgBuildReadWriter {

	// TBD do we add validation here
	// Ignore file processing should be done here
	fs := fsys.NewDiskFS(path)
	ignoreRules := ignore.Empty(IgnoreFileMatch[0])
	f, err := fs.Open(IgnoreFileMatch[0])
	if err == nil {
		// if an error is return the rules is empty, so we dont have to worry about the error
		ignoreRules, _ = ignore.Parse(f)
	}

	return &pkgBuildReadWriter{
		reader: &PkgReader{
			Fsys: fsys.NewDiskFS(path),
			//matchFilesGlob: YAMLMatch,
			IgnoreRules: ignoreRules,
		},
		writer: &pkgBuildWriter{
			fsys:     fs,
			rootPath: path,
			pkgName:  filepath.Base(path),
		},
	}
}

type pkgBuildReadWriter struct {
	reader *PkgReader
	writer *pkgBuildWriter
}

func (r *pkgBuildReadWriter) Read(data *Data) (*Data, error) {
	return r.reader.Read(data)
}

func (r *pkgBuildReadWriter) Write(data *Data) error {
	return r.writer.Write(data)
}

type pkgBuildWriter struct {
	fsys     fsys.FS
	rootPath string
	pkgName  string
}

func (r *pkgBuildWriter) Write(data *Data) error {
	img, err := oci.Build(data.Get())
	if err != nil {
		fmt.Println(err)
		return err
	}

	f, err := r.fsys.Create(fmt.Sprintf("%s.%s", r.pkgName, kformOciPkgExt))
	if err != nil {
		return err
	}
	defer f.Close()
	return tarball.Write(nil, img, f)
}
