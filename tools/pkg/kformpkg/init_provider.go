package kformpkg

import (
	"context"
	"errors"
	"fmt"
	"io/fs"

	kformpkgmetav1alpha1 "github.com/henderiw-nephio/kform/tools/apis/kform/pkg/meta/v1alpha1"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cctx"
)

// ProviderInitializer implements Initializer interface.
type ProviderInitializer struct{}

func (i *ProviderInitializer) Initialize(ctx context.Context, fsys fs.FS) error {
	pkgKind := kformpkgmetav1alpha1.PkgKindProvider
	pkgPath := cctx.GetContextValue[string](ctx, CtxkeyPath)
	if pkgPath == "" {
		return ErrNoPkgPath
	}

	fmt.Println("pkgKind", pkgKind, "pkgPath", pkgPath)
	pkg, err := New(ctx, fsys, pkgKind)
	if err != nil {
		return err
	}
	if werr := pkg.WriteKformfile(); werr != nil {
		err = errors.Join(err, werr)
	}
	if werr := pkg.WriteReadmeFile(); werr != nil {
		err = errors.Join(err, werr)
	}
	if werr := pkg.WriteIgnoreFile(); werr != nil {
		err = errors.Join(err, werr)
	}
	return err
}
