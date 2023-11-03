package resourcebackend

import (
	"context"
	"encoding/json"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/schema"
	"github.com/henderiw-nephio/kform/providers/provider-resourcebackend/resourcebackend/api"

	//apiext "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
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

func providerConfigure(ctx context.Context, d []byte, version string) (any, diag.Diagnostics) {
	providerAPIConfig := &api.ProviderAPI{}
	if err := json.Unmarshal(d, providerAPIConfig); err != nil {
		return nil, diag.FromErr(err)
	}

	if !providerAPIConfig.IsKindValid() {
		return nil, diag.Errorf("invalid provider kind, got: %s, expected: %v", providerAPIConfig.Kind, api.ExpectedProviderKinds)
	}

	if providerAPIConfig.Kind == api.ProviderKindMock {
		return beclient.New(ctx, providerAPIConfig.Address), diag.Diagnostics{}
	}

	return beclient.New(ctx, providerAPIConfig.Address), diag.Diagnostics{}
}
