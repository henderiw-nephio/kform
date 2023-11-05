package vctx

import (
	"fmt"
	"strings"

	"github.com/henderiw-nephio/kform/tools/pkg/syntax/types"
)

func GetContextFromName(vertexName string) string {
	return fmt.Sprintf("blockName=%s", vertexName)
}

func GetContext(rootModuleName string, vctx *types.VertexContext) string {
	if vctx == nil {
		return ""
	}
	var sb strings.Builder

	//sb.WriteString(fmt.Sprintf("fileName=%s", vctx.FileName))
	sb.WriteString(fmt.Sprintf("rootModule/moduleName=%s/%s", rootModuleName, vctx.ModuleName))
	sb.WriteString(fmt.Sprintf(", blockType=%s", vctx.BlockType))
	sb.WriteString(fmt.Sprintf(", blockName=%s", vctx.BlockName))

	return sb.String()
}

func GetContextFromModule(rootModuleName, moduleName string) string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("rootModule/moduleName=%s/%s", rootModuleName, moduleName))
	sb.WriteString(fmt.Sprintf(", blockType=%s", "root"))

	return sb.String()
}
