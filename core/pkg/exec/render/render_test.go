package render

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/henderiw-nephio/kform/syntax/pkg/dag"
	kformtypes "github.com/henderiw-nephio/kform/syntax/pkg/dag/types"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
	//"sigs.k8s.io/yaml"
)

var templ1 string = `
apiVersion: inv.nephio.org/v1alpha1
kind: node
metadata:
  name: "'server-', $count.index"
  namespace: default
  labels:
    x: $t1.id1.spec.interfaceName
spec:
  labels:
    topo.nephio.org/position: server
    topo.nephio.org/rack: rack1
  provider: server.nephio.com
  count: size($t2.id2)
`

var templ1Rendered string = `
apiVersion: inv.nephio.org/v1alpha1
kind: node
metadata:
  name: server-4
  namespace: default
  labels:
    x: e1-1
spec:
  labels:
    topo.nephio.org/position: server
    topo.nephio.org/rack: rack1
  provider: server.nephio.com
  count: 4
`

func TestRender(t *testing.T) {
	cases := map[string]struct {
		vars           map[string]kformtypes.Variable
		initVars       map[string]any
		templateIn     string
		templateWanted string
	}{
		"Normal": {

			vars: map[string]kformtypes.Variable{
				"t1.id1": {
					RenderedObject: map[string]any{
						"spec": map[string]any{
							"interfaceName": "e1-1",
						},
					},
				},
				"t2.id2": {
					RenderedObject: []string{"a", "b", "c", "d"},
				},
			},
			initVars: map[string]any{
				"count.index": 4,
			},
			templateIn:     templ1,
			templateWanted: templ1Rendered,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// init vars in the var store
			ctx := context.Background()
			varStore := dag.New[kformtypes.Variable]()
			for varName, v := range tc.vars {
				if err := varStore.AddVertex(ctx, varName, v); err != nil {
					assert.NoError(t, err)
					return
				}
			}
			// init the rendered
			rdr, _ := New(varStore, tc.initVars)

			// init the marshalled data
			in, err := parseYamlObject([]byte(tc.templateIn))
			if err != nil {
				assert.NoError(t, err)
				return
			}
			want, err := parseYamlObject([]byte(tc.templateWanted))
			if err != nil {
				assert.NoError(t, err)
				return
			}

			// test the render fn
			out, err := rdr.Render(ctx, in)
			if err != nil {
				assert.NoError(t, err)
				return
			}

			if diff := cmp.Diff(out, want); diff != "" {
				t.Errorf("-want, +got:\n%s", diff)
			}
		})
	}
}

func parseYamlObject(b []byte) (map[string]any, error) {

	fmt.Println(string(b))
	x := map[string]any{}
	if err := yaml.Unmarshal(b, &x); err != nil {
		return nil, err
	}
	return x, nil
}
