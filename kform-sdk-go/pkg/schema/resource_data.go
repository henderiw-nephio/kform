package schema

import "github.com/henderiw-nephio/kform/kform-plugin/kfprotov1/kfplugin1"

type ResourceObject struct {
	Scope  kfplugin1.Scope
	Obj    []byte // new resource obj in json format
	OldObj []byte // old resource obj in json format
}

func (r *ResourceObject) GetScope() kfplugin1.Scope {
	return r.Scope
}

func (r *ResourceObject) GetObject() []byte {
	return r.Obj
}

func (r *ResourceObject) GetOldObject() []byte {
	return r.OldObj
}
