package resourcebackend

import (
	"context"
	"encoding/json"
	"time"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/schema"
	"github.com/nokia/k8s-ipam/pkg/proxy/beclient"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func dataSourceResourceBackendIPClaim() *schema.Resource {
	defaultTimout := 5 * time.Minute
	return &schema.Resource{
		ReadContext: dataSourceResourceBackendIPClaimRead,
		Timeouts: &schema.ResourceTimeout{
			Read:    &defaultTimout,
			Default: &defaultTimout,
		},
	}
}

func dataSourceResourceBackendIPClaimRead(ctx context.Context, d *schema.ResourceObject, meta interface{}) ([]byte, diag.Diagnostics) {
	client := meta.(beclient.Client)

	u := &unstructured.Unstructured{}
	if err := json.Unmarshal(d.GetObject(), u); err != nil {
		return nil, diag.FromErr(err)
	}
	newu, err := client.GetClaim(ctx, u, nil)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	b, err := json.Marshal(newu)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	return b, nil
}
