package parser

import (
	"context"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/address"
)

// validateAndOrInstallProviders looks at the provider requirements
// 1. convert provider requirements to packages
// 2.
func (r *kformparser) validateAndOrInstallProviders(ctx context.Context, init bool) {
	for nsn, reqs := range r.GetProviderRequirements(ctx) {
		// convert provider requirements to a package
		// the source was validated to be aligned before so we can just pick the first one.
		pkg, err := address.GetPackage(nsn, reqs[0].Source)
		if err != nil {
			r.recorder.Record(diag.DiagFromErr(err))
			return
		}
		// retrieve the available releases/versions for this provider
		if err := pkg.GetReleases(); err != nil {
			r.recorder.Record(diag.DiagFromErr(err))
			return
		}
		// append the requirements together
		for _, req := range reqs {
			pkg.AddConstraints(req.Version)
		}

		// generate the candidate versions by looking at the available
		// versions and applying the constraints on them
		if err := pkg.GenerateCandidates(); err != nil {
			r.recorder.Record(diag.DiagFromErr(err))
			return
		}
		r.providers.Add(ctx, nsn, pkg)
	}

	pkgrw := pkgio.NewPkgProviderReadWriter(r.rootModulePath, r.providers)
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
