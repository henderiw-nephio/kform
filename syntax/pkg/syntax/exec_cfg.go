package syntax

import (
	"github.com/henderiw-nephio/kform/syntax/pkg/dag"
	kformtypes "github.com/henderiw-nephio/kform/syntax/pkg/dag/types"
)

type ExecConfig interface {
	GetVars() dag.DAG[kformtypes.Variable]
	GetProviders() dag.DAG[kformtypes.Provider]
}

func NewExecConfig() ExecConfig {
	return &execConfig{
		providers: dag.New[kformtypes.Provider](),
		vars:      dag.New[kformtypes.Variable](),
	}
}

type execConfig struct {
	providers dag.DAG[kformtypes.Provider]
	vars      dag.DAG[kformtypes.Variable]
}

func (r *execConfig) GetVars() dag.DAG[kformtypes.Variable] {
	return r.vars
}

func (r *execConfig) GetProviders() dag.DAG[kformtypes.Provider] {
	return r.providers
}
