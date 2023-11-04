package builder

import (
	"context"

	"github.com/henderiw-nephio/kform/tools/pkg/dag"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/fn/fns"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/record"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/vars"
	"github.com/henderiw-nephio/kform/tools/pkg/executor"
	"github.com/henderiw-nephio/kform/tools/pkg/recorder"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/types"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
)

// TBD this need to be optimized
func New(ctx context.Context, dag dag.DAG[*types.VertexContext]) (executor.DAGExecutor, error) {
	recorder := recorder.New[record.Record]()
	vars := cache.New[vars.Variable]()

	return executor.NewDAGExecutor[*types.VertexContext](ctx, dag, &executor.Config[*types.VertexContext]{
		Name: "tbd",
		Handler: fns.NewExecHandler(ctx, &fns.Config{
			RootModuleName: "tbd",
			ModuleName:     "tbd",
			Vars:           vars,
			Recorder:       recorder,
		}),
	})
}
