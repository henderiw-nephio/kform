package pkgio

import (
	"context"
	"testing"

	kformpkgmetav1alpha1 "github.com/henderiw-nephio/kform/tools/apis/kform/pkg/meta/v1alpha1"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio/data"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/address"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
	"github.com/stretchr/testify/assert"
)

func TestPkgProviderRead(t *testing.T) {
	cases := map[string]struct {
		rootPath string
		reqs     map[string]kformpkgmetav1alpha1.Provider
	}{
		"dnn": {
			rootPath: "../../../examples/nf-test-dnn-example",
			reqs: map[string]kformpkgmetav1alpha1.Provider{
				"kubernetes": {
					Source:  "github.com/henderiw-nephio/kform",
					Version: "0.0.1",
				},
				"resourcebackend": {
					Source:  "github.com/henderiw-nephio/kform",
					Version: "0.0.1",
				},
			},
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			providers := cache.New[*address.Package]()
			for name, req := range tc.reqs {
				pkg, err := address.GetPackage(cache.NSN{Name: name}, req.Source)
				assert.NoError(t, err)
				providers.Add(ctx, cache.NSN{Name: name}, pkg)
			}

			rw := NewPkgProviderReadWriter(tc.rootPath, providers)
			data := data.New()
			data, err := rw.Read(ctx, data)
			assert.NoError(t, err)
			//data.Print()
			data, err = rw.Process(ctx, data)
			assert.NoError(t, err)
			data.Print()

			err = rw.Write(ctx, data)
			assert.NoError(t, err)
		})
	}
}
