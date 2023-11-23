package pkgio

import (
	"context"
	"path/filepath"
	"runtime"
	"strings"

	kformpkgmetav1alpha1 "github.com/henderiw-nephio/kform/tools/apis/kform/pkg/meta/v1alpha1"
	"github.com/henderiw-nephio/kform/tools/pkg/fsys"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio/data"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio/oci"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio/oras"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/address"
	"github.com/henderiw/logger/log"
)

type PkgPullReadWriter interface {
	Reader
	Writer
}

func NewPkgPullReadWriter(dstPath string, pkg *address.Package, pkgKind kformpkgmetav1alpha1.PkgKind) PkgPullReadWriter {
	// TBD do we add validation here
	// Ignore file processing should be done here
	fs := fsys.NewDiskFS(dstPath)
	//ignoreRules := ignore.Empty(IgnoreFileMatch[0])
	return &pkgPullReadWriter{
		reader: &pkgPullReader{
			pkg:     pkg,
			pkgKind: pkgKind,
			dstPath: dstPath,
			//PathExists:     true,
			//Fsys:           fsys.NewDiskFS(srcPath),
			//MatchFilesGlob: MatchAll,
			//IgnoreRules:    ignoreRules,
		},
		writer: &pkgPullWriter{
			fsys: fs,
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

func (r *pkgPullReadWriter) Read(ctx context.Context, data *data.Data) (*data.Data, error) {
	return r.reader.Read(ctx, data)
}

func (r *pkgPullReadWriter) Write(ctx context.Context, data *data.Data) error {
	return r.writer.write(ctx, data)
}

type pkgPullReader struct {
	pkg     *address.Package
	pkgKind kformpkgmetav1alpha1.PkgKind
	dstPath string
}

func (r *pkgPullReader) Read(ctx context.Context, data *data.Data) (*data.Data, error) {
	// add the runtime environment in case the package is a provider
	// for module the os/arch is not required
	if r.pkgKind == kformpkgmetav1alpha1.PkgKindProvider {
		r.pkg.Platform = &address.Platform{
			OS:   runtime.GOOS,
			Arch: runtime.GOARCH,
		}
	}
	if err := oras.Pull(ctx, r.pkg.GetVersionRef(), data); err != nil {
		return data, err
	}
	return data, nil
}

type pkgPullWriter struct {
	fsys fsys.FS
	//rootPath string
	//pkgName  string
	//pkg   *address.Package
	//local bool
}

func (r *pkgPullWriter) write(ctx context.Context, data *data.Data) error {
	log := log.FromContext(ctx)
	for path, b := range data.List() {
		if strings.HasSuffix(path, ".tar.gz") {
			b, err := oci.ReadTgz([]byte(b))
			if err != nil {
				return err
			}
			// add the plain file
			data.Add(strings.TrimSuffix(path, ".tar.gz"), b)
			// delete the tgz file
			data.Delete(path)
		}
	}

	// write files to fsys
	for path, b := range data.List() {
		if err := r.fsys.MkdirAll(filepath.Dir(path)); err != nil {
			log.Error("cannot create dir", "path", filepath.Dir(path), "err", err.Error())
			continue
		}

		if err := r.fsys.WriteFile(path, []byte(b)); err != nil {
			log.Error("cannot create file", "path", path, "err", err.Error())
			continue
		}
	}
	return nil
}
