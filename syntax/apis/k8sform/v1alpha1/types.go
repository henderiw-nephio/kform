package v1alpha1

type K8sFormCtx struct {
	FileName string `json:"fileName" yaml:"fileName"`
	K8sForm  `json:",inline" yaml:",inline"`
}

type K8sForm struct {
	Blocks []Block `json:"spec" yaml:"spec"`
}

type Block struct {
	BlockData   `json:",inline" yaml:",inline"`
	NestedBlock map[string]Block `json:",inline" yaml:",inline"`
}

type BlockData struct {
	Attributes map[string]any `json:"attributes,omitempty" yaml:"attributes,omitempty"`
	Object     any            `json:"object,omitempty" yaml:"object,omitempty"`
}

/*
type ControllerConfigSpec struct {
	// key represents the variable
	For map[string]*GvkObject `json:"for" yaml:"for"`
	// key represents the variable
	Own map[string]*GvkObject `json:"own,omitempty" yaml:"own,omitempty"`
	// key represents the variable
	Watch map[string]*GvkObject `json:"watch,omitempty" yaml:"watch,omitempty"`
	// key respresents the variable
	//Functions map[string]ControllerConfigFunctionBlock `json:",inline" yaml:",inline"`
	Pipelines []*Pipeline `json:"pipelines,omitempty" yaml:"pipelines,omitempty"`

	Services map[string]*Function `json:"services,omitempty" yaml:"services,omitempty"`
}

type GvkObject struct {
	Resource          runtime.RawExtension `json:"resource,omitempty" yaml:"resource,omitempty"`
	ApplyPipelineRef  string               `json:"applyPipelineRef,omitempty" yaml:"applyPipelineRef,omitempty"`
	DeletePipelineRef string               `json:"deletePipelineRef,omitempty" yaml:"deletePipelineRef,omitempty"`
}

type Pipeline struct {
	Name  string                      `json:"name" yaml:"name"`
	Vars  map[string]*FunctionElement `json:"vars,omitempty" yaml:"vars,omitempty"`
	Tasks map[string]*FunctionElement `json:"tasks,omitempty" yaml:"tasks,omitempty"`
}

type Block struct {
	Range     *RangeValue          `json:"range,omitempty" yaml:"range,omitempty"`
	Condition *ConditionExpression `json:"condition,omitempty" yaml:"condition,omitempty"`
}

type RangeValue struct {
	Value string `json:"value" yaml:"value"`
	Block `json:",inline" yaml:",inline"`
}

type ConditionExpression struct {
	Expression string `json:"expression" yaml:"expression"`
	Block      `json:",inline" yaml:",inline"`
}

type FunctionType string

const (
	RootType       FunctionType = "root"
	QueryType      FunctionType = "query"
	SliceType      FunctionType = "slice"
	MapType        FunctionType = "map"
	JQType         FunctionType = "jq"
	ContainerType  FunctionType = "container"
	WasmType       FunctionType = "wasm"
	GoTemplateType FunctionType = "gotemplate"
	BlockType      FunctionType = "block"
)

type FunctionElement struct {
	Function      `json:",inline" yaml:",inline"`
	FunctionBlock map[string]*FunctionElement `json:"block,omitempty" yaml:"block,omitempty"`
}

type Function struct {
	Block    `json:",inline" yaml:",inline"`
	Executor `json:",inline" yaml:",inline"`
	// Vars define the local variables in the function
	// The Key respresents the local variable name
	// The Value represents the jq expression
	Vars   map[string]string `json:"vars,omitempty" yaml:"vars,omitempty"`
	Type   FunctionType      `json:"type,omitempty" yaml:"type,omitempty"`
	Config string            `json:"config,omitempty" yaml:"config,omitempty"`
	// input is always a GVK of some sort
	Input *Input `json:"input,omitempty" yaml:"input,omitempty"`
	// key = variableName, value is gvr format or not -> gvr format is needed for external resources
	Output    map[string]*Output `json:"output,omitempty" yaml:"output,omitempty"`
	DependsOn []string           `json:"dependsOn,omitempty" yaml:"dependsOn,omitempty"`
}

type Output struct {
	Internal    bool                 `json:"internal" yaml:"internal"`
	Conditioned bool                 `json:"conditioned" yaml:"conditioned"`
	Resource    runtime.RawExtension `json:"resource" yaml:"resource"`
}

type Input struct {
	Selector     *metav1.LabelSelector `json:"selector,omitempty" yaml:"selector,omitempty"`
	Key          string                `json:"key,omitempty" yaml:"key,omitempty"`
	Value        string                `json:"value,omitempty" yaml:"value,omitempty"`
	GenericInput map[string]string     `json:",inline" yaml:",inline"`
	Expression   string                `json:"expression,omitempty" yaml:"expression,omitempty"`
	Resource     runtime.RawExtension  `json:"resource,omitempty" yaml:"resource,omitempty"`
	Template     string                `json:"template,omitempty" yaml:"template,omitempty"`
}

type Executor struct {
	Image string `json:"image,omitempty" yaml:"image,omitempty"`
	Exec  string `json:"exec,omitempty" yaml:"exec,omitempty"`
}
*/
