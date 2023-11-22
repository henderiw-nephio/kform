package pkgio

import (
	"context"
	"path/filepath"

	"github.com/henderiw-nephio/kform/tools/pkg/fsys"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio/oras"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/address"
)

type PkgPullReadWriter interface {
	Reader
	Writer
}

func NewPkgPullReadWriter(srcPath string, pkg *address.Package) PkgPullReadWriter {
	// TBD do we add validation here
	// Ignore file processing should be done here
	fs := fsys.NewDiskFS(srcPath)
	//ignoreRules := ignore.Empty(IgnoreFileMatch[0])
	return &pkgPullReadWriter{
		reader: &pkgPullReader{
			pkg: pkg,
			//PathExists:     true,
			//Fsys:           fsys.NewDiskFS(srcPath),
			//MatchFilesGlob: MatchAll,
			//IgnoreRules:    ignoreRules,
		},
		writer: &pkgPullWriter{
			fsys:     fs,
			rootPath: srcPath,
			//pkgName:  filepath.Base(srcPath),
			//pkg: pkg,
			//local: local,
		},
	}
}

type pkgPullReadWriter struct {
	reader *pkgPullReader
	writer *pkgPullWriter
}

func (r *pkgPullReadWriter) Read(ctx context.Context, data *Data) (*Data, error) {
	return r.reader.Read(ctx, data)
}

func (r *pkgPullReadWriter) Write(ctx context.Context, data *Data) error {
	return r.writer.write(ctx, data)
}

type pkgPullReader struct {
	pkg *address.Package
}

func (r *pkgPullReader) Read(ctx context.Context, data *Data) (*Data, error) {
	if err := oras.Pull(ctx, r.pkg.GetRef()); err != nil {
		return data, err
	}
	return data, nil
}

type pkgPullWriter struct {
	fsys     fsys.FS
	rootPath string
	//pkgName  string
	//pkg   *address.Package
	//local bool
}

func (r *pkgPullWriter) write(ctx context.Context, data *Data) error {
	for path, b := range data.List() {
		r.fsys.MkdirAll(filepath.Dir(filepath.Join(r.rootPath, path)))
		r.fsys.WriteFile(path, []byte(b))
	}
	return nil
}
