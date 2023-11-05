package record

import (
	"fmt"
	"strings"
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

func (r Record) GetDetails() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf("duration=%v", r.Stop.Sub(r.Start)))
	if r.Severity == kfplugin1.Severity_UNDEFINED {
		sb.WriteString(", severity=SUCCESS")
	} else {
		sb.WriteString(fmt.Sprintf(", severity=%s", r.Severity))
	}

	sb.WriteString(fmt.Sprintf(", context=%s", r.Context))
	if r.Detail != "" {
		sb.WriteString(fmt.Sprintf(", detail=%s", r.Detail))
	}

	return sb.String()
}
