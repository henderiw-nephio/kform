package exttypes

import (
	blockv1alpha1 "github.com/henderiw-nephio/kform/tools/apis/kform/block/v1alpha1"
	"github.com/henderiw-nephio/kform/tools/pkg/dag"
)

type ExecConfig interface {
	GetVars() dag.DAG[blockv1alpha1.Variable]
	GetProviders() dag.DAG[blockv1alpha1.Provider]
}

func NewExecConfig() ExecConfig {
	return &execConfig{
		providers: dag.New[blockv1alpha1.Provider](),
		vars:      dag.New[blockv1alpha1.Variable](),
	}
}

type execConfig struct {
	providers dag.DAG[blockv1alpha1.Provider]
	vars      dag.DAG[blockv1alpha1.Variable]
}

func (r *execConfig) GetVars() dag.DAG[blockv1alpha1.Variable] {
	return r.vars
}

func (r *execConfig) GetProviders() dag.DAG[blockv1alpha1.Provider] {
	return r.providers
}
