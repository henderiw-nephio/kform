package main

import (
	"context"
	"fmt"

	"github.com/henderiw-nephio/kform/providers/provider-kubernetes/kubernetes/provclient/pkgclient"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
)

const (
	srcDir = "./examples/crd"
)

func main() {
	/*
		pb, err := pkgutil.GetPackage(srcDir, "*.yaml")
		if err != nil {
			panic(err)
		}
		res := map[v1.ObjectReference]client.Object{}
		for _, n := range pb.Nodes {
			fmt.Println(n.MustString())
			u := unstructured.Unstructured{}
			fmt.Println(n.GetApiVersion())
			if err := yaml.Unmarshal([]byte(n.MustString()), &u); err != nil {
				panic(err)
			}
			fmt.Println(u.GetKind())
			res[v1.ObjectReference{
				APIVersion: u.GetAPIVersion(),
				Kind:       u.GetKind(),
				Name:       u.GetName(),
				Namespace:  u.GetNamespace(),
			}] = &u
		}

		for ref := range res {
			fmt.Println(ref.String())
		}
	*/
	c, err := pkgclient.New(pkgclient.Config{Dir: srcDir})
	if err != nil {
		panic(err)
	}
	u := unstructured.Unstructured{}
	u.SetAPIVersion("apiextensions.k8s.io/v1")
	u.SetKind("CustomResourceDefinition")
	if err := c.Get(context.Background(), types.NamespacedName{Name: "nodepools.inv.nephio.org"}, &u); err != nil {
		panic(err)
	}
	fmt.Println(u.GetAPIVersion())
	fmt.Println(u.GetKind())
	fmt.Println(u.GetName())

	fmt.Println("+++++++++++++++++")
	//fmt.Println(unstructured.NestedMap(u.Object, "spec"))

	ul := unstructured.UnstructuredList{}
	ul.SetAPIVersion("apiextensions.k8s.io/v1")
	ul.SetKind("CustomResourceDefinition")
	if err := c.List(context.Background(), &ul); err != nil {
		panic(err)
	}
	fmt.Println(ul.GetAPIVersion())
	fmt.Println(ul.GetKind())
	fmt.Println("+++++++++++++++++")
	for _, u := range ul.Items {
		fmt.Printf("  %s %s %s\n", u.GetAPIVersion(), u.GetKind(), u.GetName())
	}
}
