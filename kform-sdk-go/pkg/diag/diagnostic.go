package diag

import (
	"fmt"

	"github.com/henderiw-nephio/kform/kform-plugin/kfprotov1/kfplugin1"
)

type Diagnostic struct {
	*kfplugin1.Diagnostic
}

var severities = []kfplugin1.Diagnostic_Severity{kfplugin1.Diagnostic_ERROR, kfplugin1.Diagnostic_WARNING}

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
