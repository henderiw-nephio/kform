package resourcebackend

import (
	"context"
	"encoding/json"
	"time"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/schema"
	ipamv1alpha1 "github.com/nokia/k8s-ipam/apis/resource/ipam/v1alpha1"
	"github.com/nokia/k8s-ipam/pkg/proxy/beclient"
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

	u := &ipamv1alpha1.IPClaim{}
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

	u := &ipamv1alpha1.IPClaim{}
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
