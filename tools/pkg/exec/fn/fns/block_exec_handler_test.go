package fns

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/google/go-cmp/cmp"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/record"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/vars"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/vctx"
	"github.com/henderiw-nephio/kform/tools/pkg/recorder"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/types"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
	"github.com/henderiw/logger/log"
	"github.com/stretchr/testify/assert"
	"k8s.io/utils/pointer"
)

func TestExecHandlerLocal(t *testing.T) {
	cases := map[string]struct {
		vars map[string]vars.Variable
		vCtx *vctx.VertexContext
		want vars.Variable
	}{
		"Single": {
			vars: map[string]vars.Variable{
				"input.a": {
					Data: map[string][]any{
						vars.DummyKey: {"a", "b"},
					},
				},
			},
			vCtx: &vctx.VertexContext{
				FileName:   "a.yaml",
				ModuleName: "a",
				BlockType:  types.BlockTypeLocal,
				BlockName:  "local.a",
				BlockContext: types.KformBlockContext{
					Attributes: &types.KformBlockAttributes{},
					Value:      "$input.a",
				},
			},
			want: vars.Variable{
				Data: map[string][]any{
					vars.DummyKey: {[]any{"a", "b"}},
				},
			},
		},
		"Count": {
			vars: map[string]vars.Variable{
				"input.a": {
					Data: map[string][]any{
						vars.DummyKey: {"a", "b"},
					},
				},
			},
			vCtx: &vctx.VertexContext{
				FileName:   "a.yaml",
				ModuleName: "a",
				BlockType:  types.BlockTypeLocal,
				BlockName:  "local.a",
				BlockContext: types.KformBlockContext{
					Attributes: &types.KformBlockAttributes{
						Count: pointer.String("5"),
					},
					Value: "$input.a",
				},
			},
			want: vars.Variable{
				Data: map[string][]any{
					vars.DummyKey: {
						[]any{"a", "b"},
						[]any{"a", "b"},
						[]any{"a", "b"},
						[]any{"a", "b"},
						[]any{"a", "b"},
					},
				},
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}))
			ctx := context.Background()
			log.IntoContext(ctx, logger)

			recorder := recorder.New[record.Record]()

			varsCache := cache.New[vars.Variable]()
			for name, v := range tc.vars {
				varsCache.Add(ctx, cache.NSN{Name: name}, v)
			}

			h := &ExecHandler{
				RootModuleName: "dummy",
				ModuleName:     tc.vCtx.BlockName,
				FnsMap: NewMap(ctx, &Config{
					Vars:     varsCache,
					Recorder: recorder,
				}),
				Vars:     varsCache,
				Recorder: recorder,
			}
			success := h.BlockRun(ctx, tc.vCtx.BlockName, tc.vCtx)
			if !success {
				t.Errorf("want success, but failed\n")
			}
			got, err := varsCache.Get(cache.NSN{Name: tc.vCtx.BlockName})
			if err != nil {
				assert.Error(t, err)
				return
			}
			if diff := cmp.Diff(tc.want, got); diff != "" {
				t.Errorf("-want, +got:\n%s", diff)
			}
		})
	}
}
