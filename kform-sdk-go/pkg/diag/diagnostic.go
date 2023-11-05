package diag

import (
	"fmt"
	"strings"

	"github.com/henderiw-nephio/kform/kform-plugin/kfprotov1/kfplugin1"
)

type Diagnostic struct {
	*kfplugin1.Diagnostic
}

var severities = []kfplugin1.Severity{kfplugin1.Severity_ERROR, kfplugin1.Severity_WARNING}

func (r Diagnostic) Get() *kfplugin1.Diagnostic {
	if r.Diagnostic != nil {
		return r.Diagnostic
	}
	return &kfplugin1.Diagnostic{}
}

func (r Diagnostic) GetSeverity() kfplugin1.Severity {
	return r.Severity
}

func (r Diagnostic) GetDetail() string {
	return r.Detail
}

func (r Diagnostic) GetContext() string {
	return r.Context
}

func (r Diagnostic) Validate() error {
	var valid bool
	for _, sev := range severities {
		if r.Severity == sev {
			valid = true
			break
		}
	}
	if !valid {
		return fmt.Errorf("invalid severity: %v", r.Severity)
	}
	return nil
}

func (r Diagnostic) GetDetails() string {
	var sb strings.Builder

	sb.WriteString(fmt.Sprintf(", severity=%s", r.Severity))
	sb.WriteString(fmt.Sprintf(", context=%s", r.Context))
	if r.Detail != "" {
		sb.WriteString(fmt.Sprintf(", detail=%s", r.Detail))
	}

	return sb.String()
}
