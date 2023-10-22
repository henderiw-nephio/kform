package schema

import (
	"context"
	"log/slog"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	//apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
)

type Provider struct {
	//Schema               *apiext.JSONSchemaProps
	ResourceMap          map[string]*Resource
	DataSourcesMap       map[string]*Resource
	ListDataSourcesMap   map[string]*Resource
	ConfigureContextFunc ConfigureContextFunc

	// configured is enabled after a Configure() call
	configured bool

	providerMetaConfig any

	Version string
}

type ConfigureContextFunc func(ctx context.Context, c []byte) (any, diag.Diagnostics)

func (r *Provider) Configure(ctx context.Context, c []byte) diag.Diagnostics {
	// No configuration
	if r.ConfigureContextFunc == nil {
		return nil
	}

	// A schema is required when the provider is to be configured
	/*
		if r.Schema == nil {
			return diag.Diagnostics{
				{
					Severity: diag.Error,
					Detail:   "A provider with configuration needs a schema for validation",
				},
			}
		}
	*/

	// A provider is getting reconfigured -> this can cause issues
	if r.configured {
		slog.Warn("Previously configured provider being re-configured, This can cause issues")
	}
	// get a new schema
	/*
		v, _, err := NewSchemaValidator(r.Schema)
		if err != nil {
			diag.FromErr(err)
		}
		valResult := v.Validate(c)
		if !valResult.IsValid() {
			return diag.Diagnostics{
				{
					Severity: diag.Error,
					Detail:   fmt.Sprintf("provider config validation failed errors: %v, warnings: %v", valResult.Errors, valResult.Warnings),
				},
			}
		}
	*/
	// initialize diags
	diags := diag.Diagnostics{}
	// call the provider configure fn to initialize the provider config
	providerMetaConfig, configureDiags := r.ConfigureContextFunc(ctx, c)
	diags = append(diags, configureDiags...)
	if diags.HasError() {
		return diags
	}
	r.providerMetaConfig = providerMetaConfig // this is the providerConfig
	// indicates the provider is configured
	r.configured = true

	return diags
}

func (r *Provider) getDataSources() []string {
	s := make([]string, 0, len(r.DataSourcesMap))
	for n := range r.DataSourcesMap {
		s = append(s, n)
	}
	return s
}

func (r *Provider) getListDataSources() []string {
	s := make([]string, 0, len(r.ListDataSourcesMap))
	for n := range r.ListDataSourcesMap {
		s = append(s, n)
	}
	return s
}

func (r *Provider) getResources() []string {
	s := make([]string, 0, len(r.ResourceMap))
	for n := range r.ResourceMap {
		s = append(s, n)
	}
	return s
}
