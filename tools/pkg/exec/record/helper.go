package record

import (
	"fmt"
	"time"

	"github.com/henderiw-nephio/kform/kform-plugin/kfprotov1/kfplugin1"
)

func FromErr(ctx string, start, stop time.Time, err error, d ...string) Record {
	if err == nil {
		detail := ""
		if len(d) > 0 {
			detail = d[0]
		}
		return Record{
			Start: start,
			Stop:  stop,
			Diagnostic: &kfplugin1.Diagnostic{
				Context: ctx,
				Detail:  detail,
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

func Success(ctx string, start, stop time.Time, d ...string) Record {
	detail := ""
	if len(d) > 0 {
		detail = d[0]
	}
	return Record{
		Start: start,
		Stop:  stop,
		Diagnostic: &kfplugin1.Diagnostic{
			Context: ctx,
			Detail:  detail,
		},
	}
}
