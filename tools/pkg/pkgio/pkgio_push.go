package pkgio

import (
	"context"
	"path/filepath"

	"github.com/henderiw-nephio/kform/tools/apis/kform/pkg/meta/v1alpha1"
	"github.com/henderiw-nephio/kform/tools/pkg/fsys"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio/ignore"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio/oci"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio/oras"
	"github.com/henderiw/logger/log"
	"gopkg.in/yaml.v2"
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
	return r.writer.write(ctx, data)
}

type pkgPushWriter struct {
	fsys     fsys.FS
	rootPath string
	pkgName  string
	ref      string
}

func (r *pkgPushWriter) write(ctx context.Context, data *Data) error {
	log := log.FromContext(ctx)
	// get the kform file to determine is this a provider or a module
	d, err := data.Get(PkgFileMatch[0])
	if err != nil {
		return err
	}
	kformFile := v1alpha1.KformFile{}
	if err := yaml.Unmarshal([]byte(d), &kformFile); err != nil {
		return err
	}
	if err := kformFile.Spec.Kind.Validate(); err != nil {
		return err
	}

	// build a zipped tar bal from the data in the pkg
	pkgData, err := oci.BuildTgz(data.List())
	if err != nil {
		return err
	}

	var imgData []byte
	if kformFile.Spec.Kind == v1alpha1.PkgKindProvider {
		// TODO get image
	}

	// get a client to the registry
	/*
		var c *registry.Client
		if r.authorizer != nil {
			c, err = registry.NewClient(registry.ClientOptAuth(r.authorizer))
			if err != nil {
				return err
			}
		} else {
			c, err = registry.NewClient()
			if err != nil {
				return err
			}
		}
	*/

	if err := oras.Push(ctx, kformFile.Spec.Kind, r.ref, pkgData, imgData); err != nil {
		return err
	}
	log.Info("push succeeded")

	return nil
}
