package builder

import (
	"context"

	"github.com/henderiw-nephio/kform/tools/pkg/dag"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/fn/fns"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/record"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/vars"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/vctx"
	"github.com/henderiw-nephio/kform/tools/pkg/executor"
	"github.com/henderiw-nephio/kform/tools/pkg/recorder"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
)


// TBD this need to be optimized
func New(ctx context.Context, dag dag.DAG[*vctx.VertexContext]) executor.Executor {
	recorder := recorder.New[record.Record]()
	vars := cache.New[vars.Variable]()

	e := executor.New[*vctx.VertexContext](ctx, dag, &executor.Config[*vctx.VertexContext]{
		Name: "tbd",
		Handler: &fns.ExecHandler{
			RootModuleName: "tbd",
			ModuleName:     "tbd",
			FnsMap: fns.NewMap(ctx, &fns.Config{
				Recorder: recorder,
				Vars:     vars,
			}),
			Vars:     vars,
			Recorder: recorder,
		},
	})
	return e
}
