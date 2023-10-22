package render

import (
	"context"
	"testing"

	"github.com/google/go-cmp/cmp"
	blockv1alpha1 "github.com/henderiw-nephio/kform/tools/apis/kform/block/v1alpha1"
	"github.com/henderiw-nephio/kform/tools/pkg/dag"
	"github.com/stretchr/testify/assert"
)

func TestGetVarsFromExpression(t *testing.T) {
	cases := map[string]struct {
		vars        map[string]blockv1alpha1.Variable
		initVars    map[string]any
		expr        string
		want        map[string]any
		expectedErr bool
	}{
		"NoAttr": {
			vars:        map[string]blockv1alpha1.Variable{},
			expr:        "a",
			want:        map[string]any{},
			expectedErr: false,
		},
		"VarNotFound": {
			vars:        map[string]blockv1alpha1.Variable{},
			expr:        "$a.b.c",
			expectedErr: true,
		},
		"SingleVar": {
			vars: map[string]blockv1alpha1.Variable{
				"a.b": {RenderedObject: "v1"},
			},
			expr: "$a.b.c",
			want: map[string]any{
				"a.b": "v1",
			},
			expectedErr: true,
		},
		"MultipleVar": {
			vars: map[string]blockv1alpha1.Variable{
				"a.b": {RenderedObject: "render"},
				"v.w": {RenderedObject: nil},
			},
			expr: "$a.b.c, $v.w",
			want: map[string]any{
				"a.b": "render",
				"v.w": nil,
			},
			expectedErr: true,
		},
		"OverlappingVar": {
			vars: map[string]blockv1alpha1.Variable{
				"a.b": {RenderedObject: "render"},
				"v.w": {RenderedObject: nil},
			},
			expr: "$a.b.c, $a.b",
			want: map[string]any{
				"a.b": "render",
			},
			expectedErr: true,
		},
		"InitVars": {
			vars: map[string]blockv1alpha1.Variable{
				"a.b": {RenderedObject: "render"},
				"v.w": {RenderedObject: nil},
			},
			initVars: map[string]any{
				"count.index": 4,
			},
			expr: "$a.b.c, $count.index",
			want: map[string]any{
				"a.b":         "render",
				"count.index": 4,
			},
			expectedErr: true,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			// init vars in the var store
			ctx := context.Background()
			varStore := dag.New[blockv1alpha1.Variable]()
			for varName, v := range tc.vars {
				if err := varStore.AddVertex(ctx, varName, v); err != nil {
					assert.NoError(t, err)
					return
				}
			}
			// init vars with the test data
			vars, err := newVars(varStore, tc.initVars)
			if err != nil {
				assert.NoError(t, err)
				return
			}
			// actual test
			varsFromExpr, err := vars.GetVarsFromExpression(tc.expr)
			if tc.expectedErr && err != nil {
				assert.Error(t, err)
				return
			}
			if err != nil {
				assert.NoError(t, err)
			}
			if diff := cmp.Diff(varsFromExpr, tc.want); diff != "" {
				t.Errorf("-want, +got:\n%s", diff)
			}
		})
	}
}
