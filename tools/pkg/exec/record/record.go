package record

import (
	"time"

	"github.com/henderiw-nephio/kform/kform-plugin/kfprotov1/kfplugin1"
)

type Record struct {
	Start time.Time
	Stop  time.Time
	*kfplugin1.Diagnostic
}

func (r Record) GetSeverity() kfplugin1.Severity {
	return r.Severity
}

func (r Record) GetDetail() string {
	return r.Detail
}

func (r Record) GetContext() string {
	return r.Context
}
