package fns

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/vars"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/types"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
	"github.com/henderiw/logger/log"
	"github.com/stretchr/testify/assert"
)

func TestRunInput(t *testing.T) {
	cases := map[string]struct {
		vars        map[string]vars.Variable
		localVars   map[string]any
		vCtx        *types.VertexContext
		want        vars.Variable
		expectedErr bool
	}{
		"NoVarWithDefault": {
			vars:      map[string]vars.Variable{},
			localVars: map[string]any{},
			vCtx: &types.VertexContext{
				FileName:   "a.yaml",
				ModuleName: "a",
				BlockType:  types.BlockTypeInput,
				BlockName:  "input.a",
				BlockContext: types.KformBlockContext{
					Default: []any{"a", "b"},
				},
			},
			want: vars.Variable{
				Data: map[string][]any{
					vars.DummyKey: {"a", "b"},
				},
			},
			expectedErr: false,
		},
		"NoVarNoDefault": {
			vars:      map[string]vars.Variable{},
			localVars: map[string]any{},
			vCtx: &types.VertexContext{
				FileName:     "a.yaml",
				ModuleName:   "a",
				BlockType:    types.BlockTypeInput,
				BlockName:    "input.a",
				BlockContext: types.KformBlockContext{},
			},
			expectedErr: true,
		},
		"VarWithDefault": {
			vars: map[string]vars.Variable{
				"input.a": {
					Data: map[string][]any{
						vars.DummyKey: {"a", "b"},
					},
				},
			},
			localVars: map[string]any{},
			vCtx: &types.VertexContext{
				FileName:   "a.yaml",
				ModuleName: "a",
				BlockType:  types.BlockTypeInput,
				BlockName:  "input.a",
				BlockContext: types.KformBlockContext{
					Default: []any{"c", "d"},
				},
			},
			want: vars.Variable{
				Data: map[string][]any{
					vars.DummyKey: {"a", "b"},
				},
			},
			expectedErr: false,
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}))
			ctx := context.Background()
			log.IntoContext(ctx, logger)

			varsCache := cache.New[vars.Variable]()
			for name, v := range tc.vars {
				varsCache.Add(ctx, cache.NSN{Name: name}, v)
			}

			input := &input{vars: varsCache}
			err := input.Run(ctx, tc.vCtx, tc.localVars)
			if err != nil {
				assert.Error(t, err)
				return
			}
			got, err := varsCache.Get(cache.NSN{Name: tc.vCtx.BlockName})
			if err != nil {
				if !tc.expectedErr {
					assert.Error(t, err)
				}
				return
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("-want, +got:\n%s", diff)
			}
		})
	}
}
