package kformpkg

import "fmt"

type CtxKey string

func (c CtxKey) String() string {
	return fmt.Sprintf("context key %s", string(c))
}

const (
	CtxkeyPkgType CtxKey = "packageType"
	CtxkeyPath    CtxKey = "packagePath"
)
