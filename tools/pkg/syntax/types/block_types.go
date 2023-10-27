package types

import (
	"context"
	"fmt"
)

var BlockTypes = map[BlockType]BlockInitializer{
	BlockTypeBackend:  newBackend,
	BlockTypeProvider: newProvider,
	BlockTypeModule:   newModule,
	BlockTypeInput:    newInput,
	BlockTypeOutput:   newOutput,
	BlockTypeLocal:    newLocal,
	BlockTypeResource: newResource,
	BlockTypeData:     newResource,
}

type BlockInitializer func(ctx context.Context, n string) Block

type Block interface {
	// implemented generically
	//WithRecorder(diag.Recorder)
	GetBlockType() string
	GetLevel() int
	ProcessBlock(context.Context, *KformBlock) context.Context

	// dynamicData
	//GetBlockName() string
	//GetFileName() string
	//GetAttributes() *KformBlockAttributes
	//GetInstances() []any
	//GetInput() map[string]any
	//GetDefault() []any
	//GetConfig() any
	//GetDependencies() []string
	//GetContext(string) string

	// specific implementation
	UpdateModule(context.Context)

	// specifics
	//GetSource() string
	//GetProvider() string
}

func GetBlockTypeNames() []string {
	s := make([]string, 0, len(BlockTypes))
	for n := range BlockTypes {
		s = append(s, string(n))
	}
	return s
}

func GetBlock(ctx context.Context, n string) (Block, error) {
	bi, ok := BlockTypes[GetBlockType(n)]
	if !ok {
		return nil, fmt.Errorf("cannot get blockType for %s, supported blocktypes %v", n, GetBlockTypeNames())
	}
	return bi(ctx, n), nil
}
