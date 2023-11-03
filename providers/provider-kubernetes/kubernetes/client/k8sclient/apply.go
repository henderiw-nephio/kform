package k8sclient

import (
	"context"
	"encoding/json"

	provclient "github.com/henderiw-nephio/kform/providers/provider-kubernetes/kubernetes/client"
	"github.com/pkg/errors"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// An APIPatchingApplicator applies changes to an object by either creating or
// patching it in a Kubernetes API server.
type APIPatchingApplicator struct {
	client.Client
}

// NewAPIPatchingApplicator returns an Applicator that applies changes to an
// object by either creating or patching it in a Kubernetes API server.
func NewAPIPatchingApplicator(c client.Client) APIPatchingApplicator {
	return APIPatchingApplicator{c}
}

// Apply changes to the supplied object. The object will be created if it does
// not exist, or patched if it does. If the object does exist, it will only be
// patched if the passed object has the same or an empty resource version.
func (a APIPatchingApplicator) Apply(ctx context.Context, o client.Object, ao ...provclient.ApplyOption) error {
	//if o.GetNamespace() == "" {
	//	o.SetNamespace("default")
	//}

	m, ok := o.(metav1.Object)
	if !ok {
		return errors.New("cannot access object metadata")
	}

	if m.GetName() == "" && m.GetGenerateName() != "" {
		return errors.Wrap(a.Create(ctx, o), "cannot create object")
	}

	desired := o.DeepCopyObject()

	err := a.Get(ctx, types.NamespacedName{Name: m.GetName(), Namespace: m.GetNamespace()}, o)
	if kerrors.IsNotFound(err) {
		// TODO: Apply ApplyOptions here too?
		return errors.Wrap(a.Create(ctx, o), "cannot create object")
	}
	if err != nil {
		return errors.Wrap(err, "cannot get object")
	}

	for _, fn := range ao {
		if err := fn(ctx, o, desired); err != nil {
			return err
		}
	}

	// TODO: Allow callers to override the kind of patch used.
	return errors.Wrap(a.Patch(ctx, o, &patch{desired.(client.Object)}), "cannot patch object")
}

type patch struct{ from client.Object }

func (p *patch) Type() types.PatchType                { return types.MergePatchType }
func (p *patch) Data(_ client.Object) ([]byte, error) { return json.Marshal(p.from) }
