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

func FromErr(err error) []Diagnostic {
	if err == nil {
		return nil
	}
	return []Diagnostic{
		DiagFromErr(err),
	}
}

func FromErrWithContext(ctx string, err error) []Diagnostic {
	if err == nil {
		return nil
	}
	return []Diagnostic{
		DiagFromErrWithContext(ctx, err),
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

func Errorf(format string, a ...interface{}) []Diagnostic {
	return []Diagnostic{
		DiagErrorf(format, a...),
	}
}

func ErrorfWithContext(ctx string, format string, a ...interface{}) []Diagnostic {
	return []Diagnostic{
		DiagErrorfWithContext(ctx, format, a...),
	}
}
