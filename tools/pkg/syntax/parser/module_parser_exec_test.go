package parser

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"testing"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/fn/fns"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/record"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/vars"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio"
	"github.com/henderiw-nephio/kform/tools/pkg/recorder"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/types"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
	"github.com/henderiw/logger/log"
	"github.com/stretchr/testify/assert"
)

var kformfileTest = `
apiVersion: meta.pkg.kform.io/v1alpha1
kind: KformFile
metadata:
  name: test
spec:
  kind: module
  requiredProviders:
    aws:
      source: .terraform/providers/aws
    gke:
      source: .terraform/providers/gke
    kubernetes:
      source: .terraform/providers/kubernetes
`

var kformInputTest = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: input
data:
  spec:
  - input:
      context:
        attributes:
          schema:
            apiVersion: v1
            kind: ConfigMap
        default:
        - metadata:
            name: context
          data:
            clusterName: dummy
`

var kformOutputTest = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: outputs
data:
  spec:
  - output:
      nads:
        attributes:
          schema:
            apiVersion: v1
            kind: ConfigMap
        value: $input.context
`

func TestKformsExec(t *testing.T) {
	cases := map[string]struct {
		path  string
		files map[string]string
	}{
		"InputOutput": {
			path: "./test",
			files: map[string]string{
				"KformFile.yaml": kformfileTest,
				"input.yaml":     kformInputTest,
				"output.yaml":    kformOutputTest,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}))

			precorder := recorder.New[diag.Diagnostic]()

			ctx := context.Background()
			ctx = context.WithValue(ctx, types.CtxKeyRecorder, precorder)

			log.IntoContext(ctx, logger)
			p := moduleparser{
				path:     tc.path,
				fsys:     buildFs(tc.path, tc.files),
				recorder: precorder,
			}
			ctx = context.WithValue(ctx, types.CtxKeyModuleName, cache.NSN{Name: "test"})
			m := p.Parse(ctx)
			fmt.Println(precorder.Get())
			fmt.Println(precorder.Get().Error())
			if precorder.Get().HasError() {
				assert.Error(t, precorder.Get().Error())
			}
			m.GenerateDAG(ctx)

			rrecorder := recorder.New[record.Record]()
			varsCache := cache.New[vars.Variable]()
			rmfn := fns.NewModuleFn(&fns.Config{RootModuleName: m.NSN.Name, Vars: varsCache, Recorder: rrecorder})

			if err := rmfn.Run(ctx, &types.VertexContext{
				FileName:     filepath.Join("test", pkgio.PkgFileMatch[0]),
				ModuleName:   m.NSN.Name,
				BlockType:    types.BlockTypeModule,
				BlockName:    m.NSN.Name,
				DAG:          m.DAG,
				BlockContext: types.KformBlockContext{},
			}, map[string]any{}); err != nil {
				assert.Error(t, err)
			}

			for nsn, v := range varsCache.List() {
				fmt.Println("nsn", nsn.Name, "vars", v)
			}
		})
	}
}
