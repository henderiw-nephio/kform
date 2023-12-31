package types

import (
	"context"
	"encoding/json"

	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cctx"
)

type CtxKey string

func (c CtxKey) String() string {
	return "context key " + string(c)
}

const (
	//CtxExecConfig      CtxKey = "execConfig"
	CtxKeyRecorder     CtxKey = "recorder"
	CtxKeyModule       CtxKey = "module"
	CtxKeyModuleName   CtxKey = "moduleName"
	CtxKeyModuleKind   CtxKey = "moduleKind"
	CtxKeyFileName     CtxKey = "fileName"
	CtxKeyLevel        CtxKey = "level"
	CtxKeyBlockType    CtxKey = "blockType"
	CtxKeyVarName      CtxKey = "varName"
	CtxKeyVarType      CtxKey = "varType"
	CtxKeyKformContext CtxKey = "kformContext"
	//CtxKeyAttributes CtxKey = "attributes"
	//CtxKeyInstances  CtxKey = "instances"
	//CtxKeyInput      CtxKey = "input"
	//CtxKeyDefault    CtxKey = "default"
)

type ModuleKind = string

const (
	ModuleKindRoot  ModuleKind = "root"
	ModuleKindChild ModuleKind = "child"
)

type Context struct {
	FileName   string  `json:"fileName"`
	ModuleKind string  `json:"moduleKind"`
	Module     string  `json:"module"`
	BlockType  *string `json:"blockType,omitempty"`
	Level      int     `json:"level"`
	VarName    *string `json:"varName,omitempty"`
	VarType    *string `json:"varType,omitempty"`
	//Provider  *string `json:"provider,omitempty"`
}

func GetContext(ctx context.Context) string {
	c := Context{}

	blockType := cctx.GetContextValue[string](ctx, CtxKeyBlockType)
	if blockType != "" {
		c.BlockType = &blockType
	}
	varName := cctx.GetContextValue[string](ctx, CtxKeyVarName)
	if varName != "" {
		c.VarName = &varName
	}
	varType := cctx.GetContextValue[string](ctx, CtxKeyVarType)
	if varName != "" {
		c.VarType = &varType
	}
	moduleName := cctx.GetContextValue[cache.NSN](ctx, CtxKeyModuleName)
	c.Module = moduleName.Name
	c.FileName = cctx.GetContextValue[string](ctx, CtxKeyFileName)
	c.Level = cctx.GetContextValue[int](ctx, CtxKeyLevel)

	b, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return string(b)
}
