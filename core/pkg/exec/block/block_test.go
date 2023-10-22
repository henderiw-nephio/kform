package block

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/tools/pkg/dag"
	blockv1alpha1 "github.com/henderiw-nephio/kform/tools/apis/kform/block/v1alpha1"
	"github.com/stretchr/testify/assert"
)

func TestRun(t *testing.T) {
	cases := map[string]struct {
		blockName   string
		vars        map[string]blockv1alpha1.Variable // initializes the dag
		want        any
		expectedErr bool
	}{
		"NoLoopAttr": {
			blockName: "a.b",
			vars: map[string]blockv1alpha1.Variable{
				"a.b": {
					Object: "x",
					Attributes: map[string]any{
						"a": "b",
						"c": "d",
					},
				},
			},
			want:        "x",
			expectedErr: false,
		},
		"Count": {
			blockName: "a.b",
			vars: map[string]blockv1alpha1.Variable{
				"a.b": {
					Object: "x",
					Attributes: map[string]any{
						"count": 4,
						"c":     "d",
					},
				},
			},
			want:        []any{"x", "x", "x", "x"},
			expectedErr: false,
		},
		"CountVariable": {
			blockName: "a.b",
			vars: map[string]blockv1alpha1.Variable{
				"a.b": {
					Object: "$count.index",
					Attributes: map[string]any{
						"count": 4,
						"c":     "d",
					},
				},
			},
			want:        []any{int64(0), int64(1), int64(2), int64(3)},
			expectedErr: false,
		},
		"ForEach": {
			blockName: "a.b",
			vars: map[string]blockv1alpha1.Variable{
				"a.b": {
					Object: "x",
					Attributes: map[string]any{
						"for_each": map[string]any{
							"a": "b",
							"c": "d",
							"e": "f",
							"g": "h",
						},
					},
				},
			},
			want:        []any{"x", "x", "x", "x"},
			expectedErr: false,
		},
		"ForEachVariable": {
			blockName: "a.b",
			vars: map[string]blockv1alpha1.Variable{
				"a.b": {
					Object: map[string]any{
						"key": "$each.value",
					},
					Attributes: map[string]any{
						"for_each": map[string]any{
							"a": "b",
							"c": "d",
							"e": "f",
							"g": "h",
						},
					},
				},
			},
			want: []map[string]any{
				{"key": "b"},
				{"key": "d"},
				{"key": "f"},
				{"key": "h"},
			},
			expectedErr: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			varStore := dag.New[blockv1alpha1.Variable]()
			for varName, v := range tc.vars {
				if err := varStore.AddVertex(ctx, varName, v); err != nil {
					assert.NoError(t, err)
					return
				}
			}
			b := &block{
				recorder: diag.NewRecorder(),
				varStore: varStore,
			}
			diags := b.Run(ctx, tc.blockName)
			if diags.HasError() && tc.expectedErr {
				assert.Error(t, diags.Error())
				return
			}

			got, err := varStore.GetVertex(tc.blockName)
			if err != nil {
				assert.NoError(t, err)
				return
			}
			fmt.Println("test name:", name)
			if !reflect.DeepEqual(tc.want, got.RenderedObject) {
				t.Errorf("\n-want %v\n+got: %v\n", tc.want, got.RenderedObject)
			}
		})
	}
}

func TestEvalAttr(t *testing.T) {
	cases := map[string]struct {
		vars        map[string]blockv1alpha1.Variable // initializes the dag
		attrs       Attrs
		want        evalAttrCtx
		expectedErr bool
	}{
		"NonePresent": {
			attrs: map[string]any{
				"a": "b",
				"c": "d",
			},
			want: evalAttrCtx{
				Total:     0,
				Count:     0,
				ForEaches: []ForEach{},
			},
			expectedErr: false,
		},
		"CountPresent": {
			attrs: map[string]any{
				"a":     "b",
				"count": 4,
			},
			want: evalAttrCtx{
				Total:     4,
				Count:     4,
				ForEaches: []ForEach{},
			},
			expectedErr: false,
		},
		"ForEachPresent": {
			attrs: map[string]any{
				"for_each": map[string]any{"x": "y"},
				"a":        "b",
			},
			want: evalAttrCtx{
				Total: 1,
				Count: 0,
				ForEaches: []ForEach{
					{Key: "x", Value: "y"},
				},
			},
			expectedErr: false,
		},
		"BothPresentForEach": {
			attrs: map[string]any{
				"for_each": map[string]any{"x": "y"},
				"a":        "b",
				"count":    4,
			},
			want: evalAttrCtx{
				Total: 1,
				Count: 4,
				ForEaches: []ForEach{
					{Key: "x", Value: "y"},
				},
			},
			expectedErr: false,
		},
		"BothPresentCount": {
			attrs: map[string]any{
				"for_each": map[string]any{
					"a": "b",
					"c": "d",
					"e": "f",
					"g": "h",
				},
				"a":     "b",
				"count": 2,
			},
			want: evalAttrCtx{
				Total: 2,
				Count: 2,
				ForEaches: []ForEach{
					{Key: "a", Value: "b"},
					{Key: "c", Value: "d"},
					{Key: "e", Value: "f"},
					{Key: "g", Value: "h"},
				},
			},
			expectedErr: false,
		},
	}

	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			ctx := context.Background()
			varStore := dag.New[blockv1alpha1.Variable]()
			for varName, v := range tc.vars {
				if err := varStore.AddVertex(ctx, varName, v); err != nil {
					assert.NoError(t, err)
					return
				}
			}
			b := &block{
				recorder: diag.NewRecorder(),
				varStore: varStore,
			}
			got, diags := b.evalAttrs(ctx, tc.attrs)
			if diags.HasError() && !tc.expectedErr {
				assert.NoError(t, diags.Error())
				return
			}
			if diags.HasError() && tc.expectedErr {
				assert.Error(t, diags.Error())
				return
			}
			if diff := cmp.Diff(tc.want.Total, got.Total); diff != "" {
				t.Errorf("-want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(tc.want.Count, got.Count); diff != "" {
				t.Errorf("-want, +got:\n%s", diff)
			}
			if diff := cmp.Diff(len(tc.want.ForEaches), len(got.ForEaches)); diff != "" {
				t.Errorf("-want, +got:\n%s", diff)
			}
		})
	}
}
