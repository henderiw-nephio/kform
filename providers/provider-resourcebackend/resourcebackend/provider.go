package resourcebackend

import (
	"context"
	"encoding/json"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/schema"
	"github.com/henderiw-nephio/kform/providers/provider-resourcebackend/resourcebackend/api/v1alpha1"
	"github.com/nokia/k8s-ipam/pkg/proxy/beclient"
)

func Provider() *schema.Provider {
	p := &schema.Provider{
		//Schema:         provSchema,
		ResourceMap: map[string]*schema.Resource{
			"resourcebackend_ipclaim":   resourceResourceBackendIPClaim(),
			"resourcebackend_vlanclaim": resourceResourceBackendVLANClaim(),
		},
		DataSourcesMap: map[string]*schema.Resource{
			"resourcebackend_ipclaim":   dataSourceResourceBackendIPClaim(),
			"resourcebackend_vlanclaim": dataSourceResourceBackendVLANClaim(),
		},
		//ListDataSourcesMap: map[string]*schema.Resource{
		//	"resourcebackend_ipclaim": dataSourcesResourceBackendIPClaim(),
		//},
	}
	p.ConfigureContextFunc = func(ctx context.Context, d []byte) (any, diag.Diagnostics) {
		return providerConfigure(ctx, d, p.Version)
	}
	return p
}

func providerConfigure(ctx context.Context, d []byte, _ string) (any, diag.Diagnostics) {
	providerConfig := &v1alpha1.ProviderConfig{}
	if err := json.Unmarshal(d, providerConfig); err != nil {
		return nil, diag.FromErr(err)
	}

	if !providerConfig.Spec.IsKindValid() {
		return nil, diag.Errorf("invalid provider kind, got: %s, expected: %v", providerConfig.Spec.Kind, v1alpha1.ExpectedProviderKinds)
	}

	if providerConfig.Spec.Kind == v1alpha1.ProviderKindMock {
		return beclient.NewMock(), diag.Diagnostics{}
	}

	return beclient.New(ctx, providerConfig.Spec.Address), diag.Diagnostics{}
}
