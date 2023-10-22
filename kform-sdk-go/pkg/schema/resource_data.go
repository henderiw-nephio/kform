package schema

import "github.com/henderiw-nephio/kform/kform-plugin/kfprotov1/kfplugin1"

type ResourceData struct {
	scope kfplugin1.Scope
	data  []byte // resource data in json format
}

func (r *ResourceData) GetScope() kfplugin1.Scope {
	return r.scope
}

func (r *ResourceData) GetData() []byte {
	return r.data
}
