package types

import "github.com/henderiw-nephio/kform/tools/pkg/dag"

type Kform struct {
	Blocks []KformBlock `json:"spec" yaml:"spec"`
}

type KformBlock struct {
	KformBlockContext `json:",inline" yaml:",inline"`
	NestedBlock       map[string]KformBlock `json:",inline" yaml:",inline"`
}

type KformBlockContext struct {
	Attributes *KformBlockAttributes `json:"attributes,omitempty" yaml:"attributes,omitempty"`
	//Instances  []any                 `json:"instances,omitempty" yaml:"instances,omitempty"`
	Default    []any                 `json:"default,omitempty" yaml:"default,omitempty"`
	// NOTE: had to change to inputParams as input is a blocktype name
	InputParams map[string]any `json:"inputParams,omitempty" yaml:"inputParams,omitempty"`
	Config      any            `json:"config,omitempty" yaml:"config,omitempty"`
	Value       any            `json:"value,omitempty" yaml:"value,omitempty"`
}

type KformBlockAttributes struct {
	Schema *KformBlockSchema `json:"schema,omitempty" yaml:"schema,omitempty"`
	Source *string           `json:"source,omitempty" yaml:"source,omitempty"`
	Alias  *string           `json:"alias,omitempty" yaml:"alias,omitempty"`
	// should be an int actually
	Count         *string           `json:"count,omitempty" yaml:"count,omitempty"`
	ForEach       *string           `json:"forEach,omitempty" yaml:"forEach,omitempty"`
	Provider      *string           `json:"provider,omitempty" yaml:"provider,omitempty"`
	Providers     map[string]string `json:"providers,omitempty" yaml:"providers,omitempty"`
	DependsOn     *string           `json:"dependsOn,omitempty" yaml:"dependsOn,omitempty"`
	Description   *string           `json:"description,omitempty" yaml:"description,omitempty"`
	Sensitive     *bool             `json:"sensitive,omitempty" yaml:"sensitive,omitempty"`
	Validation    *string           `json:"validation,omitempty" yaml:"validation,omitempty"`
	Lifecycle     *string           `json:"lifecycle,omitempty" yaml:"lifecycle,omitempty"`
	PreCondition  *string           `json:"preCondition,omitempty" yaml:"preCondition,omitempty"`
	PostCondition *string           `json:"postCondition,omitempty" yaml:"postCondition,omitempty"`
	Provisioner   *string           `json:"provisioner,omitempty" yaml:"provisioner,omitempty"`
	Connection    *string           `json:"connection,omitempty" yaml:"connection,omitempty"`
	Hostname      *string           `json:"hostname,omitempty" yaml:"hostname,omitempty"`
	Organization  *string           `json:"organization,omitempty" yaml:"organization,omitempty"`
	Workspaces    map[string]string `json:"workspaces,omitempty" yaml:"workspaces,omitempty"`
}

type KformBlockSchema struct {
	ApiVersion string `json:"apiVersion,omitempty" yaml:"apiVersion,omitempty"`
	Kind       string `json:"kind,omitempty" yaml:"kind,omitempty"`
}

type BlockType string

const (
	BlockTypeUnknown           BlockType = "unknown"
	BlockTypeBackend           BlockType = "backend"
	BlockTypeRequiredProviders BlockType = "requiredProviders"
	BlockTypeProvider          BlockType = "provider"
	BlockTypeModule            BlockType = "module"
	BlockTypeInput             BlockType = "input"
	BlockTypeOutput            BlockType = "output"
	BlockTypeLocal             BlockType = "local"
	BlockTypeResource          BlockType = "resource"
	BlockTypeData              BlockType = "data"
	BlockTypeRoot              BlockType = dag.Root
)

func GetBlockType(n string) BlockType {
	switch n {
	case "backend":
		return BlockTypeBackend
	case "requiredProviders":
		return BlockTypeRequiredProviders
	case "provider":
		return BlockTypeProvider
	case "module":
		return BlockTypeModule
	case "input":
		return BlockTypeInput
	case "output":
		return BlockTypeOutput
	case "local":
		return BlockTypeLocal
	case "resource":
		return BlockTypeResource
	case "data":
		return BlockTypeData
	default:
		return BlockTypeUnknown
	}
}

type BlockContextKey string

const (
	BlockContextKeyUnknown     BlockContextKey = "unknown"
	BlockContextKeyAttributes  BlockContextKey = "attributes"
	//BlockContextKeyInstances   BlockContextKey = "instances"
	BlockContextKeyDefault     BlockContextKey = "default"
	BlockContextKeyInputParams BlockContextKey = "inputParams"
	BlockContextKeyConfig      BlockContextKey = "config"
	BlockContextKeyValue       BlockContextKey = "value"
)

type MetaArgument string

const (
	MetaArgumentUnknown MetaArgument = "unknown"
	MetaArgumentSchema  MetaArgument = "schema"
	MetaArgumentSource  MetaArgument = "source"
	//MetaArgumentAlias         MetaArgument = "alias"
	//MetaArgumentAliases       MetaArgument = "aliases"
	MetaArgumentCount         MetaArgument = "count"
	MetaArgumentForEach       MetaArgument = "forEach"
	MetaArgumentProvider      MetaArgument = "provider"
	MetaArgumentProviders     MetaArgument = "providers"
	MetaArgumentDependsOn     MetaArgument = "dependsOn"
	MetaArgumentLifecycle     MetaArgument = "lifecycle"
	MetaArgumentPrecondition  MetaArgument = "precondition"
	MetaArgumentPostcondition MetaArgument = "postcondition"
	MetaArgumentConnection    MetaArgument = "connection"
	MetaArgumentProvisioner   MetaArgument = "provisioner"
	MetaArgumentDescription   MetaArgument = "description"
	MetaArgumentSensitive     MetaArgument = "sensitive"
	MetaArgumentValidation    MetaArgument = "validation"
	MetaArgumentHostname      MetaArgument = "hostname"
	MetaArgumentOrganization  MetaArgument = "organization"
	MetaArgumentWorkspaces    MetaArgument = "workspaces"
)
