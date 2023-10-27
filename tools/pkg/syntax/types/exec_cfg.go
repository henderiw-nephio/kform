package types

import (
	"github.com/henderiw-nephio/kform/tools/pkg/dag"
)

type ExecConfig interface {
	GetVars() dag.DAG[Block]
	GetProviders() dag.DAG[Block]
	GetModules() dag.DAG[Block]
	GetBackends() dag.DAG[Block]
}

func NewExecConfig() ExecConfig {
	return &execConfig{
		providers: dag.New[Block](),
		modules:   dag.New[Block](),
		vars:      dag.New[Block](),
		backends:  dag.New[Block](),
	}
}

type execConfig struct {
	backends  dag.DAG[Block]
	modules   dag.DAG[Block]
	providers dag.DAG[Block]
	vars      dag.DAG[Block]
}

func (r *execConfig) GetVars() dag.DAG[Block] {
	return r.vars
}

func (r *execConfig) GetProviders() dag.DAG[Block] {
	return r.providers
}

func (r *execConfig) GetModules() dag.DAG[Block] {
	return r.modules
}

func (r *execConfig) GetBackends() dag.DAG[Block] {
	return r.backends
}
