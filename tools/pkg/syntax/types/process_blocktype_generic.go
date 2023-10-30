package types

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cctx"
	"github.com/henderiw/logger/log"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

const (
	ResourceType = "resourceType"
	ResourceID   = "resourceID"
)

type config struct {
	level              int
	blockType          BlockType
	expectedKeywords   map[BlockContextKey]bool
	expectedAttributes map[string]bool
	recorder           diag.Recorder

	// dynamic config
	fileName   string
	moduleName string
	gvk        schema.GroupVersionKind
	KformBlockContext
	dependencies    map[string]string
	modDependencies map[string]string
}

var mandatory = true
var optional = false

func (r *config) GetBlockType() string { return string(r.blockType) }

func (r *config) GetLevel() int { return r.level }

func (r *config) getDependencies(ctx context.Context) {
	rn := NewRenderer()

	r.getAttributeDependencies(ctx, rn)
	if r.KformBlockContext.Attributes != nil {
		if err := rn.GatherDependencies(ctx, r.KformBlockContext.Attributes); err != nil {
			r.recorder.Record(diag.DiagFromErrWithContext(GetContext(ctx), err))
		}
	}
	if r.KformBlockContext.Instances != nil {
		if err := rn.GatherDependencies(ctx, r.KformBlockContext.Instances); err != nil {
			r.recorder.Record(diag.DiagFromErrWithContext(GetContext(ctx), err))
		}
	}
	if r.KformBlockContext.Default != nil {
		if err := rn.GatherDependencies(ctx, r.KformBlockContext.Default); err != nil {
			r.recorder.Record(diag.DiagFromErrWithContext(GetContext(ctx), err))
		}
	}
	if r.KformBlockContext.InputParams != nil {
		if err := rn.GatherDependencies(ctx, r.KformBlockContext.InputParams); err != nil {
			r.recorder.Record(diag.DiagFromErrWithContext(GetContext(ctx), err))
		}
	}
	if r.KformBlockContext.Config != nil {
		if err := rn.GatherDependencies(ctx, r.KformBlockContext.Config); err != nil {
			r.recorder.Record(diag.DiagFromErrWithContext(GetContext(ctx), err))
		}
	}
	r.dependencies = rn.GetDependencies()
	r.modDependencies = rn.GetModuleOutputDependencies()
}

func (r *config) getAttributeDependencies(ctx context.Context, rn Renderer) {
	if r.KformBlockContext.Attributes != nil {
		b, err := json.Marshal(r.KformBlockContext.Attributes)
		if err != nil {
			r.recorder.Record(diag.DiagErrorfWithContext(GetContext(ctx), "err marshaling kform block context, err: %s", err.Error()))
			return
		}
		attributes := map[string]any{}
		if err := json.Unmarshal(b, &attributes); err != nil {
			r.recorder.Record(diag.DiagErrorfWithContext(GetContext(ctx), "err unmarshaling kform block context, err: %s", err.Error()))
			return
		}
		if err := rn.GatherDependencies(ctx, attributes); err != nil {
			r.recorder.Record(diag.DiagFromErrWithContext(GetContext(ctx), err))
		}
	}
}

func (r *config) ProcessBlock(ctx context.Context, block *KformBlock) context.Context {
	recorder := cctx.GetContextValue[diag.Recorder](ctx, CtxKeyRecorder)
	if recorder == nil {
		//r.recorder.Record(diag.DiagErrorfWithContext(GetContext(ctx), "cannot parse without a recorder"))
		return ctx
	}

	level := cctx.GetContextValue[int](ctx, CtxKeyLevel)
	if level < r.GetLevel() {
		// continue to walk
		// validate if attr or obj are present at the intermediate level
		r.validateAttrAndObjectAtIntermediateLevel(ctx, block)
		// validate the block prior to processing
		blockName, block, err := GetNextBlock(ctx, block)
		if err != nil {
			r.recorder.Record(diag.DiagFromErr(err))
			return ctx
		}
		// process the next block
		level++
		ctx = r.addContext(ctx, blockName, level)
		ctx = context.WithValue(ctx, CtxKeyLevel, level)
		return r.ProcessBlock(ctx, block)
	}
	ctx = context.WithValue(ctx, CtxKeyKformContext, block.KformBlockContext)

	log := log.FromContext(ctx)
	log.Debug("processed:",
		"blockType", cctx.GetContextValue[string](ctx, CtxKeyBlockType),
		"type", cctx.GetContextValue[string](ctx, CtxKeyVarType),
		"name", cctx.GetContextValue[string](ctx, CtxKeyVarName),
	)

	return ctx
}

func GetNextBlock(ctx context.Context, block *KformBlock) (string, *KformBlock, error) {
	// validate the block prior to processing
	if err := validateBlock(ctx, block); err != nil {
		return "", nil, err
	}
	// process next level
	for blockName, block := range block.NestedBlock {
		block := block
		return blockName, &block, nil
	}
	// we should never get here
	return "", nil, fmt.Errorf("cannot have a block without a nested block")
}

func validateBlock(ctx context.Context, block *KformBlock) error {
	level := cctx.GetContextValue[int](ctx, CtxKeyLevel)
	// if there is no block assigned in the topBlock this is an invalid block
	if len(block.NestedBlock) == 0 {
		if level == 0 {
			return fmt.Errorf("cannot have a block without a block type: %v ctx: %s", block.NestedBlock, GetContext(ctx))
		} else {
			return fmt.Errorf("cannot have a block without a nested block ctx: %s", GetContext(ctx))
		}
	}
	// a block can only have 1 blocktype
	if len(block.NestedBlock) > 1 {
		return fmt.Errorf("cannot have more then 1 blocktype in a block, got: %v ctx: %s", block.NestedBlock, GetContext(ctx))
	}
	return nil
}

func (r *config) validateAttrAndObjectAtIntermediateLevel(ctx context.Context, block *KformBlock) {
	level := cctx.GetContextValue[int](ctx, CtxKeyLevel)
	if len(block.Instances) > 0 {
		r.recorder.Record(diag.DiagWarnfWithContext(GetContext(ctx), "instances at level %d present but ignored", level))
	}
	if block.Attributes != nil {
		r.recorder.Record(diag.DiagWarnfWithContext(GetContext(ctx), "attributes at level %d present but ignored", level))
	}
	if len(block.InputParams) > 0 {
		r.recorder.Record(diag.DiagWarnfWithContext(GetContext(ctx), "input at level %d present but ignored", level))
	}
	if len(block.Default) > 0 {
		r.recorder.Record(diag.DiagWarnfWithContext(GetContext(ctx), "default at level %d present but ignored", level))
	}
	if block.Config != nil {
		r.recorder.Record(diag.DiagWarnfWithContext(GetContext(ctx), "config at level %d present but ignored", level))
	}
}

func (r *config) addContext(ctx context.Context, blockName string, level int) context.Context {
	if level == r.level {
		ctx = context.WithValue(ctx, CtxKeyVarName, blockName)
	}

	// for blockTypes with a level == 2 (resource/data) we also want to capture the blockName
	//if r.blockType == BlockTypeResource || r.blockType == BlockTypeData {
	if r.level > 1 && level == r.level-1 {
		ctx = context.WithValue(ctx, CtxKeyVarType, blockName)
	}
	return ctx
}

// validateResourceSyntax validates the syntax of the resource kind
// resource Type must starts with a letter
// resource Type can container letters in lower and upper case, numbers and '-', '_'
func validateResourceSyntax(kind string, name string) error {
	re := regexp.MustCompile(`^[a-zA-Z]+[a-zA-Z0-9_-]*$`)
	if !re.Match([]byte(name)) {
		return fmt.Errorf("syntax error a %s starts with a letter and can container letters in lower and upper case, numbers and '-', '_', got: %s", kind, name)
	}
	return nil
}

func (r *config) GetFileName() string {
	return r.fileName
}

func (r *config) GetModuleName() string {
	return r.moduleName
}

func (r *config) GetAttributes() *KformBlockAttributes {
	return r.Attributes
}

func (r *config) GetInstances() []any {
	return r.Instances
}

func (r *config) GetInputParams() map[string]any {
	return r.InputParams
}

func (r *config) GetDefault() []any {
	return r.Default
}

func (r *config) GetConfig() any {
	return r.Config
}

func (r *config) GetDependencies() map[string]string {
	return r.dependencies
}

func (r *config) GetModDependencies() map[string]string {
	return r.modDependencies
}

func (r *config) GetContext(n string) string {
	return getContext(r.GetFileName(), r.GetModuleName(), n, BlockType(r.GetBlockType()))
}

func getContext(fileName, moduleName, name string, blockType BlockType) string {
	return fmt.Sprintf("fileName=%s, moduleName=%s name=%s, blockType=%s", fileName, moduleName, name, string(blockType))
}

func (r *config) getSchema(ctx context.Context, kfctx KformBlockContext) {
	//var gvk schema.GroupVersionKind
	if kfctx.Attributes != nil {
		if kfctx.Attributes.Schema != nil {
			if kfctx.Attributes.Schema.ApiVersion == "" {
				r.recorder.Record(diag.DiagErrorfWithContext(GetContext(ctx), "schema requires apiVersion but not present in schema attribute"))
			} else {
				split := strings.Split(kfctx.Attributes.Schema.ApiVersion, "/")
				switch len(split) {
				case 1:
					r.gvk.Version = split[0]
				case 2:
					r.gvk.Version = split[1]
					r.gvk.Group = split[0]
				default:
					r.recorder.Record(diag.DiagErrorfWithContext(GetContext(ctx), "schema apiVersion expected syntax is <group>/<version>, got: %s", kfctx.Attributes.Schema.ApiVersion))
				}
			}
			if kfctx.Attributes.Schema.Kind == "" {
				r.recorder.Record(diag.DiagErrorfWithContext(GetContext(ctx), "schema requires kind but not present in schema attribute"))
			} else {
				r.gvk.Kind = kfctx.Attributes.Schema.Kind
			}
		}
		// validation of schema presence was already done
	}
	//r.gvk = gvk
}

func (r *config) validateKeyWords(ctx context.Context, kfctx KformBlockContext) {
	if kfctx.Attributes != nil {
		b, err := json.Marshal(kfctx.Attributes)
		if err != nil {
			r.recorder.Record(diag.DiagErrorfWithContext(GetContext(ctx), "err marshaling kform block context, err: %s", err.Error()))
			return
		}
		attributes := map[string]any{}
		if err := json.Unmarshal(b, &attributes); err != nil {
			r.recorder.Record(diag.DiagErrorfWithContext(GetContext(ctx), "err unmarshaling kform block context, err: %s", err.Error()))
			return
		}
		r.validateKeyWord(ctx, attributes, BlockContextKeyAttributes)
	} else {
		r.validateKeyWord(ctx, nil, BlockContextKeyAttributes)
	}

	if len(kfctx.Instances) == 0 {
		r.validateKeyWord(ctx, nil, BlockContextKeyInstances)
	} else {
		r.validateKeyWord(ctx, kfctx.Instances, BlockContextKeyInstances)
	}

	if len(kfctx.Default) == 0 {
		r.validateKeyWord(ctx, nil, BlockContextKeyDefault)
	} else {
		r.validateKeyWord(ctx, kfctx.Default, BlockContextKeyDefault)
	}

	if len(kfctx.InputParams) == 0 {
		r.validateKeyWord(ctx, nil, BlockContextKeyInputParams)
	} else {
		r.validateKeyWord(ctx, kfctx.InputParams, BlockContextKeyInputParams)
	}
	r.validateKeyWord(ctx, kfctx.Config, BlockContextKeyConfig)
}

func (r *config) validateKeyWord(ctx context.Context, value any, keyword BlockContextKey) {
	//fmt.Println("validate keyword", GetContext(ctx), keyword, reflect.TypeOf(value), r.expectedAttributes)
	if value != nil {
		// validate if keywords are expected, if present and we dont expect it we create a warning
		if _, ok := r.expectedKeywords[keyword]; !ok {
			// unexpected keyword
			r.recorder.Record(diag.DiagWarnfWithContext(GetContext(ctx), "keyword %s present but ignored", keyword))
		}
	} else {
		//fmt.Println("validate keyword value nil", GetContext(ctx), keyword, r.expectedAttributes)
		// validate if the required keywords are present
		if mandatory, ok := r.expectedKeywords[keyword]; ok {
			if mandatory {
				// keyword expected but not present
				r.recorder.Record(diag.DiagErrorfWithContext(GetContext(ctx), "expected keyword %s but not present", keyword))
			}
		}
	}
}

func (r *config) validateAttributes(ctx context.Context, kfctx KformBlockContext) {
	if kfctx.Attributes != nil {
		b, err := json.Marshal(kfctx.Attributes)
		if err != nil {
			r.recorder.Record(diag.DiagErrorfWithContext(GetContext(ctx), "err marshaling kform block context, err: %s", err.Error()))
			return
		}
		attributes := map[string]any{}
		if err := json.Unmarshal(b, &attributes); err != nil {
			r.recorder.Record(diag.DiagErrorfWithContext(GetContext(ctx), "err unmarshaling kform block context, err: %s", err.Error()))
			return
		}
		// TBD do we have to validate the data type of the attribute
		// validate if attributes are expected
		for k := range attributes {
			if _, ok := r.expectedAttributes[k]; !ok {
				// unexpected keyword
				r.recorder.Record(diag.DiagWarnfWithContext(GetContext(ctx), "attribute %s present but ignored", k))
			}
		}
		// validate if the required attributes are present
		for k, mandatory := range r.expectedAttributes {
			if mandatory {
				if _, ok := attributes[k]; !ok {
					r.recorder.Record(diag.DiagErrorfWithContext(GetContext(ctx), "expected attribute %s but not present", k))
				}
			}
		}
	}
	// we dont have to report an error since validation already happened
}

func (r *config) initAndValidateBlockConfig(ctx context.Context) {
	r.fileName = cctx.GetContextValue[string](ctx, CtxKeyFileName)
	r.moduleName = cctx.GetContextValue[string](ctx, CtxKeyModuleName)
	//r.getFileName(ctx)
	//r.getModuleName(ctx)
	r.validateKeyWordsAndAttributes(ctx)
	r.getDependencies(ctx)
	r.getSchema(ctx, r.KformBlockContext)
}

func (r *config) validateKeyWordsAndAttributes(ctx context.Context) {
	r.KformBlockContext = cctx.GetContextValue[KformBlockContext](ctx, CtxKeyKformContext)
	r.validateKeyWords(ctx, r.KformBlockContext)
	r.validateAttributes(ctx, r.KformBlockContext)
}
