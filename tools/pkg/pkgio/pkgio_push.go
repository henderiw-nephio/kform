package pkgio

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/henderiw-nephio/kform/tools/pkg/fsys"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio/ignore"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio/oci"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio/registry"
)

type PkgPushReadWriter interface {
	Reader
	Writer
}

func NewPkgPushReadWriter(path, ref string) PkgPushReadWriter {

	// TBD do we add validation here
	// Ignore file processing should be done here
	fs := fsys.NewDiskFS(path)
	ignoreRules := ignore.Empty(IgnoreFileMatch[0])

	fmt.Println("path", path)

	return &pkgPushReadWriter{
		reader: &PkgReader{
			PathExists:     true,
			Fsys:           fsys.NewDiskFS(path),
			MatchFilesGlob: MatchAll,
			IgnoreRules:    ignoreRules,
		},
		writer: &pkgPushWriter{
			fsys:     fs,
			rootPath: path,
			pkgName:  filepath.Base(path),
			ref:      ref,
		},
	}
}

type pkgPushReadWriter struct {
	reader *PkgReader
	writer *pkgPushWriter
}

func (r *pkgPushReadWriter) Read(ctx context.Context, data *Data) (*Data, error) {
	return r.reader.Read(ctx, data)
}

func (r *pkgPushReadWriter) Write(ctx context.Context, data *Data) error {
	return r.writer.Write(ctx, data)
}

type pkgPushWriter struct {
	fsys     fsys.FS
	rootPath string
	pkgName  string
	ref      string
}

func (r *pkgPushWriter) Write(ctx context.Context, data *Data) error {
	/*
		tag, err := name.NewTag(r.tag)
		if err != nil {
			return err
		}

		img, err := tarball.ImageFromPath(
			filepath.Join(r.rootPath, fmt.Sprintf("%s.%s", r.pkgName, kformOciPkgExt)),
			nil)
		if err != nil {
			return err
		}
		return remote.Write(tag, img, remote.WithAuthFromKeychain(authn.DefaultKeychain))
	*/
	c, err := registry.NewClient()
	if err != nil {
		return err
	}
	schemaData, err := oci.BuildTgz(data.List())
	if err != nil {
		return err
	}

	result, err := c.Push(schemaData, r.ref)
	if err != nil {
		return err
	}
	fmt.Println(result)
	return nil
}
