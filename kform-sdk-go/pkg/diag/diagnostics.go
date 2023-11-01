package diag

/*
import (
	"errors"
	"fmt"

	"github.com/henderiw-nephio/kform/kform-plugin/kfprotov1/kfplugin1"
)

type Diagnostics []*kfplugin1.Diagnostic

func (r Diagnostics) HasError() bool {
	for _, d := range r {
		d := d
		if d.Severity == kfplugin1.Severity_ERROR {
			return true
		}
	}
	return false
}

func (r Diagnostics) Error() error {
	var err error
	for _, d := range r {
		d := d
		if d.Severity == kfplugin1.Severity_ERROR {
			err = errors.Join(err, fmt.Errorf("ctx: %s, detail: %s", d.Context, d.Detail))
		}
	}
	return err
}
*/