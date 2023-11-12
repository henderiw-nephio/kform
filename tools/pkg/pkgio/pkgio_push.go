package pkgio

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/google/go-containerregistry/pkg/v1/remote"
	"github.com/google/go-containerregistry/pkg/v1/tarball"
	"github.com/henderiw-nephio/kform/tools/pkg/fsys"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio/ignore"
)

type PkgPushReadWriter interface {
	Reader
	Writer
}

func NewPkgPushReadWriter(path, tag string) PkgPushReadWriter {

	// TBD do we add validation here
	// Ignore file processing should be done here
	fs := fsys.NewDiskFS(path)
	ignoreRules := ignore.Empty(IgnoreFileMatch[0])

	return &pkgPushReadWriter{
		reader: &PkgReader{
			PathExists:     true,
			Fsys:           fsys.NewDiskFS(path),
			MatchFilesGlob: PkgMatch,
			IgnoreRules:    ignoreRules,
		},
		writer: &pkgPushWriter{
			fsys:     fs,
			rootPath: path,
			pkgName:  filepath.Base(path),
			tag:      tag,
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
	tag      string
}

func (r *pkgPushWriter) Write(ctx context.Context, data *Data) error {
	tag, err := name.NewTag(r.tag)
	if err != nil {
		return err
	}

	// TODO local packageName
	img, err := tarball.ImageFromPath(
		filepath.Join(r.rootPath, fmt.Sprintf("%s.%s", r.pkgName, kformOciPkgExt)),
		nil)
	if err != nil {
		return err
	}
	/*
		if r.local {
			f, err := r.fsys.Create(r.pkgName + ".tgz")
			if err != nil {
				return err
			}
			defer f.Close()
			reg, err := name.NewRegistry("local")
			if err != nil {
				return err
			}
			return tarball.Write(name.Tag{
				Repository: name.Repository{
					Registry: reg,
				},
			}, img, f)
		}
	*/
	return remote.Write(tag, img, remote.WithAuthFromKeychain(authn.DefaultKeychain))
}
