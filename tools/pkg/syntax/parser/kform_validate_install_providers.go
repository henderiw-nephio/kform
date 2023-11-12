package parser

import (
	"context"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/address"
)

func (r *kformparser) validateAndOrInstallProviders(ctx context.Context, init bool) {
	// install providers
	pkgs := []*address.Package{}
	for nsn, reqs := range r.GetProviderRequirements(ctx) {
		pkg, err := address.GetPackage(nsn, reqs)
		if err != nil {
			r.recorder.Record(diag.DiagFromErr(err))
			return
		}
		pkgs = append(pkgs, pkg)
	}

	pkgrw := pkgio.NewPkgProviderReadWriter(r.rootModulePath, pkgs)
	p := pkgio.Pipeline{}
	if init {
		p = pkgio.Pipeline{
			Inputs:     []pkgio.Reader{pkgrw},
			Processors: []pkgio.Process{pkgrw},
			Outputs:    []pkgio.Writer{pkgrw},
		}
	} else {
		p = pkgio.Pipeline{
			Inputs:     []pkgio.Reader{pkgrw},
			Processors: []pkgio.Process{pkgrw},
			Outputs:    []pkgio.Writer{pkgio.NewPkgValidator()},
		}
	}

	if err := p.Execute(ctx); err != nil {
		r.recorder.Record(diag.DiagFromErr(err))
		return
	}
}
