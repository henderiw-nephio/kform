package render

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/vars"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/yaml"
)

var itfceDefault string = `
apiVersion: req.nephio.org/v1alpha1
kind: Interface
metadata:
  name: n3
spec:
  networkInstance:
    name: default
status: {}
`

var itfceVpc string = `
apiVersion: req.nephio.org/v1alpha1
kind: Interface
metadata:
  name: default
spec:
  networkInstance:
    name: vpc-ran
  cniType: sriov
  attachmentType: vlan
  ipFamilyPolicy: dualstack
status: {}
`

func buildInterfaceDefault() []any {
	out := []any{}
	out = append(out, unmarshal(itfceDefault))
	return out
}

func buildInterfaceVpc() []any {
	out := []any{}
	out = append(out, unmarshal(itfceVpc))
	return out
}

func buildInterfaces() []any {
	out := []any{}
	out = append(out, buildInterfaceDefault()[0])
	out = append(out, buildInterfaceVpc()[0])
	return out
}

func unmarshal(v string) map[string]any {
	out := map[string]any{}
	if err := yaml.Unmarshal([]byte(v), &out); err != nil {
		panic(err)
	}
	return out
}

func TestRender(t *testing.T) {

	cases := map[string]struct {
		vars       map[string][]any
		localVars  map[string]any
		expression string
		want       any
	}{
		"Count": {
			vars:       map[string][]any{}, // Empty vars
			expression: "5",
			want:       "5",
		},
		"CountSize": {
			vars: map[string][]any{
				"input.interface": {"a", "b", "c", "d"},
			},
			expression: "size($input.interface)",
			want:       int64(4),
		},

		"CountConditionalFalse": {
			vars: map[string][]any{
				"input.interface": buildInterfaceDefault(),
			},
			expression: `$input.interface[0].spec.networkInstance.name != 'default' && (($input.interface[0].spec.ipFamilyPolicy == 'ipv4Only') || ($input.interface[0].spec.ipFamilyPolicy == 'dualstack')) ? 1 : 0`,
			want:       int64(0),
		},

		"CountConditionalTrue": {
			vars: map[string][]any{
				"input.interface": buildInterfaceVpc(),
			},
			expression: `$input.interface[0].spec.networkInstance.name != 'default' && (($input.interface[0].spec.ipFamilyPolicy == 'ipv4Only') || ($input.interface[0].spec.ipFamilyPolicy == 'dualstack')) ? 1 : 0`,
			want:       int64(1),
		},
		"ForEach": {
			vars: map[string][]any{
				"input.interface": buildInterfaces(),
			},
			expression: `$input.interface.all(i, i.metadata.name == "default")`,
			want:       false,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			// initialize the vars
			varCache := cache.New[vars.Variable]()
			for k, v := range tc.vars {
				varCache.Add(ctx, cache.NSN{Name: k}, vars.Variable{
					Data: map[string][]any{
						"local": v,
					},
				})
			}
			r := &Renderer{
				Vars:      varCache,
				LocalVars: tc.localVars,
			}

			v, err := r.Render(ctx, tc.expression)
			if err != nil {
				assert.NoError(t, err)
				return
			}
			fmt.Println("value", v)
			fmt.Println("reflect type", reflect.TypeOf(v))
			if diff := cmp.Diff(v, tc.want); diff != "" {
				t.Errorf("-want, +got:\n%s", diff)
			}

		})
	}
}
