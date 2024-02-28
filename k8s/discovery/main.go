package main

import (
	"flag"
	"fmt"
	"path/filepath"

	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/discovery"
	"k8s.io/client-go/tools/clientcmd"
)

func main() {
	// Path to kubeconfig file
	kubeconfig := flag.String("kubeconfig", filepath.Join(
		// Change this path according to your kubeconfig location
		"/Users/henderiw", ".kube", "config"),
		"(optional) absolute path to the kubeconfig file")

	flag.Parse()

	// Load kubeconfig file
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	// Create a Kubernetes clientset
	/*
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			panic(err.Error())
		}
	*/

	// Create a Discovery client
	discoveryClient := discovery.NewDiscoveryClientForConfigOrDie(config)

	// Retrieve the list of supported API groups
	apiGroups, err := discoveryClient.ServerGroups()
	if err != nil {
		panic(err.Error())
	}
	for _, group := range apiGroups.Groups {
		fmt.Printf("  - Name: %s\n", group.Name)
		for _, version := range group.Versions {
			fmt.Printf("    Versions: %v\n", version.Version)
		}
		fmt.Printf("    Kind: %v\n", group.Kind)
		fmt.Println()
	}

	// Print the API groups and versions
	fmt.Println("API Groups:")
	apiResources, err := discoveryClient.ServerPreferredResources()
	if err != nil {
		panic(err.Error())
	}

	// Print the resources
	fmt.Println("Supported Resources:")
	for _, apiResourceList := range apiResources {
		gv, err := schema.ParseGroupVersion(apiResourceList.GroupVersion)
		if err != nil {
			continue
		}
		for _, apiResource := range apiResourceList.APIResources {
			fmt.Printf("  - Kind: %s\n", apiResource.Kind)
			fmt.Printf("    Group: %s\n", gv.Group)
			fmt.Printf("    Version: %s\n", gv.Version)
			fmt.Printf("    Namespaced: %t\n", apiResource.Namespaced)
			fmt.Printf("    Singular: %s\n", apiResource.SingularName)
			fmt.Printf("    Verbs: %s\n", apiResource.Verbs)
			fmt.Println()
		}
	}

}
