package schema

import "github.com/henderiw-nephio/kform/kform-plugin/kfprotov1/kfplugin1"

type ResourceObject struct {
	scope kfplugin1.Scope
	obj   []byte // new resource obj in json format
}

func (r *ResourceObject) GetScope() kfplugin1.Scope {
	return r.scope
}

func (r *ResourceObject) GetObject() []byte {
	return r.obj
}
