package record

import (
	"fmt"
	"time"

	"github.com/henderiw-nephio/kform/kform-plugin/kfprotov1/kfplugin1"
)

func FromErr(ctx string, start, stop time.Time, err error) Record {
	if err == nil {
		return Record{
			Start: start,
			Stop:  stop,
			Diagnostic: &kfplugin1.Diagnostic{
				Context: ctx,
			},
		}
	}
	return Record{
		Start: start,
		Stop:  stop,
		Diagnostic: &kfplugin1.Diagnostic{
			Severity: kfplugin1.Severity_ERROR,
			Detail:   err.Error(),
			Context:  ctx,
		},
	}
}

func Errorf(ctx string, start, stop time.Time, format string, a ...interface{}) Record {
	return Record{
		Start: start,
		Stop:  stop,
		Diagnostic: &kfplugin1.Diagnostic{
			Severity: kfplugin1.Severity_ERROR,
			Detail:   fmt.Sprintf(format, a...),
			Context:  ctx,
		},
	}
}

func Success(ctx string, start, stop time.Time) Record {
	return Record{
		Start: start,
		Stop:  stop,
		Diagnostic: &kfplugin1.Diagnostic{
			Context: ctx,
		},
	}
}
