package pkgclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	provclient "github.com/henderiw-nephio/kform/providers/provider-kubernetes/kubernetes/client"
	"github.com/henderiw-nephio/kform/providers/provider-kubernetes/kubernetes/client/pkgclient/pkgutil"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/yaml"
)

type Config struct {
	Dir               string
	IgnoreAnnotations []string
	IgnoreLabels      []string
}

func New(cfg Config) (provclient.Client, error) {
	pb, err := pkgutil.GetPackage(cfg.Dir, "*.yaml")
	if err != nil {
		return nil, err
	}

	res := map[v1.ObjectReference]client.Object{}
	for _, n := range pb.Nodes {
		u := unstructured.Unstructured{}
		if err := yaml.Unmarshal([]byte(n.MustString()), &u); err != nil {
			return nil, err
		}
		//fmt.Println(u.GetKind())
		res[v1.ObjectReference{
			APIVersion: u.GetAPIVersion(),
			Kind:       u.GetKind(),
			Name:       u.GetName(),
			Namespace:  u.GetNamespace(),
		}] = &u
	}

	return &pkgclient{
		cfg:       cfg,
		resources: res,
	}, nil
}

type pkgclient struct {
	cfg       Config
	resources map[v1.ObjectReference]client.Object
}

func (r *pkgclient) Get(ctx context.Context, key client.ObjectKey, o client.Object, opts ...client.GetOption) error {

	objRef := v1.ObjectReference{
		APIVersion: o.GetObjectKind().GroupVersionKind().GroupVersion().Identifier(),
		Kind:       o.GetObjectKind().GroupVersionKind().Kind,
		Namespace:  key.Namespace,
		Name:       key.Name,
	}
	u, ok := r.resources[objRef]
	if !ok {
		return fmt.Errorf("resource not found: %v", objRef.String())
	}
	if o == nil {
		return fmt.Errorf("resource not found: %v", objRef.String())
	}
	b, err := json.Marshal(u)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, o)
}

func (r *pkgclient) List(ctx context.Context, list client.ObjectList, opts ...client.ListOption) error {
	objref := v1.ObjectReference{
		APIVersion: list.GetObjectKind().GroupVersionKind().GroupVersion().Identifier(),
		Kind:       list.GetObjectKind().GroupVersionKind().Kind,
	}
	ul := unstructured.UnstructuredList{}
	ul.SetAPIVersion(list.GetObjectKind().GroupVersionKind().GroupVersion().Identifier())
	ul.SetKind(list.GetObjectKind().GroupVersionKind().Kind)
	for ref, u := range r.resources {
		if ref.APIVersion == objref.APIVersion && ref.Kind == objref.Kind {
			ul.Items = append(ul.Items, *u.(*unstructured.Unstructured))
		}
	}
	b, err := json.Marshal(ul)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, list)
}

func (r *pkgclient) Create(ctx context.Context, o client.Object, ao ...client.CreateOption) error {
	objRef := v1.ObjectReference{
		APIVersion: o.GetObjectKind().GroupVersionKind().GroupVersion().Identifier(),
		Kind:       o.GetObjectKind().GroupVersionKind().Kind,
		Namespace:  o.GetNamespace(),
		Name:       o.GetName(),
	}
	r.resources[objRef] = o
	return nil
}

func (r *pkgclient) Delete(ctx context.Context, o client.Object, ao ...client.DeleteOption) error {
	objRef := v1.ObjectReference{
		APIVersion: o.GetObjectKind().GroupVersionKind().GroupVersion().Identifier(),
		Kind:       o.GetObjectKind().GroupVersionKind().Kind,
		Namespace:  o.GetNamespace(),
		Name:       o.GetName(),
	}
	delete(r.resources, objRef)
	return nil
}

func (r *pkgclient) Update(ctx context.Context, o client.Object, ao ...client.UpdateOption) error {
	return r.Create(ctx, o)
}

func (r *pkgclient) Patch(ctx context.Context, o client.Object, patch client.Patch, ao ...client.PatchOption) error {
	return r.Create(ctx, o)
}

func (r *pkgclient) DeleteAllOf(ctx context.Context, obj client.Object, opts ...client.DeleteAllOfOption) error {
	return errors.New("not implemented")
}

func (r *pkgclient) Apply(ctx context.Context, o client.Object, ao ...provclient.ApplyOption) error {
	return r.Create(ctx, o)
}
