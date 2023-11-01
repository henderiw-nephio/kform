package recorder

import "github.com/henderiw-nephio/kform/kform-plugin/kfprotov1/kfplugin1"

type Record interface {
	GetSeverity() kfplugin1.Severity
	GetDetail() string
	GetContext() string
}
