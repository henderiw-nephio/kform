package kubernetes

import (
	"context"
	"encoding/json"
	"time"

	"github.com/henderiw-nephio/kform/kform-plugin/kfprotov1/kfplugin1"
	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/schema"
	"github.com/henderiw-nephio/kform/providers/provider-kubernetes/kubernetes/client"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
)

func dataSourceKubernetesManifest() *schema.Resource {
	defaultTimout := 5 * time.Minute
	return &schema.Resource{
		ReadContext: dataSourceKubernetesManifestRead,
		Timeouts: &schema.ResourceTimeout{
			Read:    &defaultTimout,
			Default: &defaultTimout,
		},
	}
}

func dataSourceKubernetesManifestRead(ctx context.Context, d *schema.ResourceObject, meta interface{}) ([]byte, diag.Diagnostics) {
	client := meta.(client.Client)

	u := &unstructured.Unstructured{}
	if err := json.Unmarshal(d.GetObject(), u); err != nil {
		return nil, diag.FromErr(err)
	}
	nsn := types.NamespacedName{Namespace: u.GetNamespace(), Name: u.GetName()}
	if d.GetScope() != kfplugin1.Scope_CLUSTER {
		nsn = types.NamespacedName{Name: u.GetName()}
	}
	newu := &unstructured.Unstructured{}
	newu.SetAPIVersion(u.GetAPIVersion())
	newu.SetKind(u.GetKind())
	if err := client.Get(ctx, nsn, newu); err != nil {
		return nil, diag.FromErr(err)
	}
	b, err := json.Marshal(newu)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	return b, nil
}
