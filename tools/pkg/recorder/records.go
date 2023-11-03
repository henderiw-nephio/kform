package recorder

import (
	"errors"
	"fmt"

	"github.com/henderiw-nephio/kform/kform-plugin/kfprotov1/kfplugin1"
)

type Records interface {
	HasError() bool
	Error() error
}

type records[T Record] []T

func (r records[T]) HasError() bool {
	for _, d := range r {
		d := d
		if d.GetSeverity() == kfplugin1.Severity_ERROR {
			return true
		}
	}
	return false
}

func (r records[T]) Error() error {
	var err error
	for _, d := range r {
		d := d
		if d.GetSeverity() == kfplugin1.Severity_ERROR {
			err = errors.Join(err, fmt.Errorf("ctx: %s, detail: %s", d.GetContext(), d.GetDetail()))
		}
	}
	return err
}
