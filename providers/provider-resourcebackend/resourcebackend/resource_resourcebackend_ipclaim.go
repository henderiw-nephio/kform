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

func resourceResourceBackendIPClaim() *schema.Resource {
	defaultTimout := 5 * time.Minute
	return &schema.Resource{
		CreateContext: resourceResourceBackendIPClaimCreate,
		ReadContext:   resourceResourceBackendIPClaimRead,
		Timeouts: &schema.ResourceTimeout{
			Create:  &defaultTimout,
			Read:    &defaultTimout,
			Default: &defaultTimout,
		},
	}
}

func resourceResourceBackendIPClaimCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]byte, diag.Diagnostics) {
	client := meta.(beclient.Client)

	u := &unstructured.Unstructured{}
	if err := json.Unmarshal(d.GetData(), u); err != nil {
		return nil, diag.FromErr(err)
	}

	newu, err := client.Claim(ctx, u, nil)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	b, err := json.Marshal(newu)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	return b, nil
}

func resourceResourceBackendIPClaimRead(ctx context.Context, d *schema.ResourceData, meta interface{}) ([]byte, diag.Diagnostics) {
	client := meta.(beclient.Client)

	u := &unstructured.Unstructured{}
	if err := json.Unmarshal(d.GetData(), u); err != nil {
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
