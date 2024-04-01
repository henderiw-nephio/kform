package resourcebackend

import (
	"context"
	"encoding/json"
	"time"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/schema"
	vlanv1alpha1 "github.com/nokia/k8s-ipam/apis/resource/vlan/v1alpha1"
	"github.com/nokia/k8s-ipam/pkg/proxy/beclient"
)

func resourceResourceBackendVLANClaim() *schema.Resource {
	defaultTimout := 5 * time.Minute
	return &schema.Resource{
		CreateContext: resourceResourceBackendVLANClaimCreate,
		ReadContext:   resourceResourceBackendVLANClaimRead,
		Timeouts: &schema.ResourceTimeout{
			Create:  &defaultTimout,
			Read:    &defaultTimout,
			Default: &defaultTimout,
		},
	}
}

func resourceResourceBackendVLANClaimCreate(ctx context.Context, d *schema.ResourceObject, meta interface{}) ([]byte, diag.Diagnostics) {
	client := meta.(beclient.Client)

	u := &vlanv1alpha1.VLANClaim{}
	if err := json.Unmarshal(d.GetObject(), u); err != nil {
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

func resourceResourceBackendVLANClaimRead(ctx context.Context, d *schema.ResourceObject, meta interface{}) ([]byte, diag.Diagnostics) {
	client := meta.(beclient.Client)

	u := &vlanv1alpha1.VLANClaim{}
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
