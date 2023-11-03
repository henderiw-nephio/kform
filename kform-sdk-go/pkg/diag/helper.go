package diag

import (
	"fmt"

	"github.com/henderiw-nephio/kform/kform-plugin/kfprotov1/kfplugin1"
)

func DiagFromErr(err error) Diagnostic {
	if err == nil {
		return Diagnostic{}
	}
	return Diagnostic{
		Diagnostic: &kfplugin1.Diagnostic{
			Severity: kfplugin1.Severity_ERROR,
			Detail:   err.Error(),
		},
	}
}

func DiagFromErrWithContext(ctx string, err error) Diagnostic {
	if err == nil {
		return Diagnostic{}
	}
	return Diagnostic{
		Diagnostic: &kfplugin1.Diagnostic{
			Severity: kfplugin1.Severity_ERROR,
			Detail:   err.Error(),
			Context:  ctx,
		},
	}
}

func DiagErrorf(format string, a ...interface{}) Diagnostic {
	return Diagnostic{
		Diagnostic: &kfplugin1.Diagnostic{
			Severity: kfplugin1.Severity_ERROR,
			Detail:   fmt.Sprintf(format, a...),
		},
	}
}

func DiagWarnf(format string, a ...interface{}) Diagnostic {
	return Diagnostic{
		Diagnostic: &kfplugin1.Diagnostic{
			Severity: kfplugin1.Severity_WARNING,
			Detail:   fmt.Sprintf(format, a...),
		},
	}
}

func DiagErrorfWithContext(ctx string, format string, a ...interface{}) Diagnostic {
	return Diagnostic{
		Diagnostic: &kfplugin1.Diagnostic{
			Severity: kfplugin1.Severity_ERROR,
			Detail:   fmt.Sprintf(format, a...),
			Context:  ctx,
		},
	}
}

func DiagWarnfWithContext(ctx string, format string, a ...interface{}) Diagnostic {
	return Diagnostic{
		Diagnostic: &kfplugin1.Diagnostic{
			Severity: kfplugin1.Severity_WARNING,
			Detail:   fmt.Sprintf(format, a...),
			Context:  ctx,
		},
	}
}

func FromErr(err error) Diagnostics {
	if err == nil {
		return nil
	}
	return Diagnostics{
		DiagFromErr(err).Get(),
	}
}

func FromErrWithContext(ctx string, err error) Diagnostics {
	if err == nil {
		return nil
	}
	return Diagnostics{
		DiagFromErrWithContext(ctx, err).Get(),
	}
}

func Errorf(format string, a ...interface{}) Diagnostics {
	return Diagnostics{
		DiagErrorf(format, a...).Get(),
	}
}

func ErrorfWithContext(ctx string, format string, a ...interface{}) Diagnostics {
	return Diagnostics{
		DiagErrorfWithContext(ctx, format, a...).Get(),
	}
}
