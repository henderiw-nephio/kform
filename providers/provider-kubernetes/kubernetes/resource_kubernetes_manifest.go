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

func resourceKubernetesManifest() *schema.Resource {
	defaultTimout := 5 * time.Minute
	return &schema.Resource{
		CreateContext: resourceKubernetesManifestCreate,
		ReadContext:   resourceKubernetesManifestRead,
		Timeouts: &schema.ResourceTimeout{
			Create:  &defaultTimout,
			Read:    &defaultTimout,
			Default: &defaultTimout,
		},
	}
}

func resourceKubernetesManifestCreate(ctx context.Context, newObj *schema.ResourceObject, meta interface{}) ([]byte, diag.Diagnostics) {
	// TODO distinguish between offline and api
	client := meta.(client.Client)

	u := &unstructured.Unstructured{}
	if err := json.Unmarshal(newObj.GetObject(), u); err != nil {
		return nil, diag.FromErr(err)
	}

	// TODO
	/*
		stateConf := &resource.StateChangeConf{
			Target:  expandPodTargetState(d.Get("target_state").([]interface{})),
			Pending: []string{string(corev1.PodPending)},
			Timeout: d.Timeout(schema.TimeoutCreate),
			Refresh: func() (interface{}, string, error) {
				out, err := conn.CoreV1().Pods(metadata.Namespace).Get(ctx, metadata.Name, metav1.GetOptions{})
				if err != nil {
					log.Printf("[ERROR] Received error: %#v", err)
					return out, "Error", err
				}

				statusPhase := fmt.Sprintf("%v", out.Status.Phase)
				log.Printf("[DEBUG] Pods %s status received: %#v", out.Name, statusPhase)
				return out, statusPhase, nil
			},
		}
		_, err = stateConf.WaitForStateContext(ctx)
		if err != nil {
			lastWarnings, wErr := getLastWarningsForObject(ctx, conn, out.ObjectMeta, "Pod", 3)
			if wErr != nil {
				return diag.FromErr(wErr)
			}
			return diag.Errorf("%s%s", err, stringifyEvents(lastWarnings))
		}
		log.Printf("[INFO] Pod %s created", out.Name)
	*/

	if err := client.Apply(ctx, u); err != nil {
		return nil, diag.FromErr(err)
	}

	b, err := json.Marshal(u)
	if err != nil {
		return nil, diag.FromErr(err)
	}

	return resourceKubernetesManifestRead(ctx, &schema.ResourceObject{Scope: newObj.GetScope(), Obj: b}, meta)
}

func resourceKubernetesManifestRead(ctx context.Context, d *schema.ResourceObject, meta interface{}) ([]byte, diag.Diagnostics) {
	client := meta.(client.Client)

	u := &unstructured.Unstructured{}
	if err := json.Unmarshal(d.GetObject(), u); err != nil {
		return nil, diag.FromErr(err)
	}
	nsn := types.NamespacedName{Namespace: u.GetNamespace(), Name: u.GetName()}
	if d.GetScope() == kfplugin1.Scope_CLUSTER {
		nsn = types.NamespacedName{Name: u.GetName()}
	}
	newu := &unstructured.Unstructured{}
	if err := client.Get(ctx, nsn, newu); err != nil {
		return nil, diag.FromErr(err)
	}
	b, err := json.Marshal(newu)
	if err != nil {
		return nil, diag.FromErr(err)
	}
	return b, nil
}


