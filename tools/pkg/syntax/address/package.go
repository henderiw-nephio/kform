package address

import (
	"runtime"

	"github.com/henderiw-nephio/kform/tools/apis/kform/pkg/meta/v1alpha1"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
)

/*
https://github.com/henderiw-nephio/kform/releases/download/v0.0.1/provider-kubernetes_0.0.1_darwin_amd64
europe-docker.pkg.dev/srlinux/eu.gcr.io/provider-xxxx
github.com/henderiw-nephio/kform/provider-xxxx
*/

// address -> hostname, namespace, name
func GetPackage(nsn cache.NSN, reqs []v1alpha1.Provider) (*Package, error) {
	// TODO handle
	hostname, namespace, err := ParseSource(reqs[0].Source)
	if err != nil {
		return nil, err
	}
	return &Package{
		Type: PackageTypeProvider,
		Address: &Address{
			HostName:  hostname,
			Namespace: namespace,
			Name:      nsn.Name,
		},
		Version: reqs[0].Version,
		Platform: &Platform{
			OS:   runtime.GOOS,
			Arch: runtime.GOARCH,
		},
	}, nil
}
