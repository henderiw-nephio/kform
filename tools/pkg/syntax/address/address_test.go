package address

import (
	"fmt"
	"runtime"
	"testing"
)

/*
https://github.com/henderiw-nephio/kform/releases/download/v0.0.1/provider-kubernetes_0.0.1_darwin_amd64
europe-docker.pkg.dev/srlinux/eu.gcr.io/provider-xxxx
github.com/henderiw-nephio/kform/provider-xxxx
*/

// .kform/providers/github.com/henderiw-nephio_kform/kubernetes/0.0.1/darwin_arm64/<binary>
// .kform/providers/kubernetes/<binary>
// .kform/providers/<hostname>/<namespace>/<name>/<version>/<platform>/<binary>
//

func TestPackage(t *testing.T) {
	getPlatform := func() *Platform {
		return &Platform{
			OS:   runtime.GOOS,
			Arch: runtime.GOARCH,
		}
	}
	cases := map[string]struct {
		Type      string
		hostName  string
		namespace string
		name      string
		version   string
		// result
		local bool
		url   string
		csurl string
		path  string
	}{
		"Remote": {
			Type:      "provider",
			hostName:  "github.com",
			namespace: "henderiw-nephio/kform",
			name:      "kubernetes",
			version:   "v0.0.1",
			local:     false,
			url:       fmt.Sprintf("https://github.com/henderiw-nephio/kform/releases/download/v0.0.1/provider-kubernetes_0.0.1_%s", getPlatform().String()),
			csurl:     "https://github.com/henderiw-nephio/kform/releases/download/v0.0.1/kform_0.0.1_checksums.txt",
			path:      fmt.Sprintf("github.com/henderiw-nephio_kform/kubernetes/v0.0.1/%s/provider-kubernetes_0.0.1_%s", getPlatform().String(), getPlatform().String()),
		},
		"Local": {
			Type:      "provider",
			hostName:  "",
			namespace: "",
			name:      "kubernetes",
			version:   "",
			local:     true,
			path:      "kubernetes/provider-kubernetes",
			// url is not relevant since the path is local
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			p := &Package{
				Type: PackageTypeProvider,
				Address: &Address{
					HostName:  tc.hostName,
					Namespace: tc.namespace,
					Name:      tc.name,
				},
				Version:  tc.version,
				Platform: getPlatform(),
			}

			if tc.local != p.IsLocal() {
				t.Errorf("want: %v, got: %v", tc.local, p.IsLocal())
			}
			fmt.Println("path", p.Path())
			if tc.path != p.Path() {
				t.Errorf("want: %v, got: %v", tc.local, p.IsLocal())
			}
			if !p.IsLocal() {
				fmt.Println("URL", p.URL())
				if tc.url != p.URL() {
					t.Errorf("want: %v, got: %v", tc.url, p.URL())
				}
				fmt.Println("checksumURL", p.ChecksumURL())
				if tc.csurl != p.ChecksumURL() {
					t.Errorf("want: %v, got: %v", tc.url, p.URL())
				}
			}
		})
	}
}
