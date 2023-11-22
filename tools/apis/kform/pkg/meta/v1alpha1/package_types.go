package v1alpha1

import (
	"fmt"
	"strings"
)

type PkgKind string

const (
	PkgKindProvider PkgKind = "provider"
	PkgKindModule   PkgKind = "module"
)

func ValidatePackageType(pt string) error {
	switch strings.ToLower(pt) {
	case string(PkgKindProvider):
		return nil
	case string(PkgKindModule):
		return nil
	default:
		return fmt.Errorf("unsupported packageType, expecting %s or %s, got %s", PkgKindProvider, PkgKindModule, strings.ToLower(pt))
	}
}

func (r PkgKind) Validate() error {
	switch r {
	case PkgKindProvider, PkgKindModule:
		return nil
	default:
		return fmt.Errorf("unsupported packageType, expecting %s or %s, got %s", PkgKindProvider, PkgKindModule, strings.ToLower(string(r)))
	}
}
