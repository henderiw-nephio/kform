package kformpkg

import (
	"context"
	"errors"
)

const (
	ReadmeFile = "README.md"
	IgnoreFile = ".kformignore"
)

var ErrNoPkgType = errors.New("cannot initialize a package without a packageType in the context")
var ErrNoPkgPath = errors.New("cannot initialize a package without a packagePath in the context")

type Initializer interface {
	Initialize(ctx context.Context, dir string) error
}
