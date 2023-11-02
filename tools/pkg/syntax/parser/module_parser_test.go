package parser

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"testing"
	"testing/fstest"

	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/tools/pkg/fsys"
	"github.com/henderiw-nephio/kform/tools/pkg/recorder"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/types"
	"github.com/henderiw/logger/log"
)

var kformfile = `
apiVersion: meta.pkg.kform.io/v1alpha1
kind: KformFile
metadata:
  name: wirer-example
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

var kformMain = `
apiVersion: v1
kind: ConfigMap
metadata:
  name: main
data:
  spec:
  - provider:
      kubernetes:
        config:
  - data: # this should validate if the resource exists, if not an error will be thrown
      kubernetes_manifest:
        network:
          attributes:
            schema:
              apiVersion: infra.nephio.org/v1alpha1
              kind: Network
            for_each: $local.unique_networkinstances
          instances:
          - metadata:
              name: $each.value
              namespace: default
  - module:
      interface:
        attributes:
          source: ./interface
        inputParams:
          context: $input.context
`

var kformInput = `
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

var kformOutput = `
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
            apiVersion: k8s.cni.cncf.io/v1
            kind: NetworkAttachmentDefinition
          forEach: $module.interface.nad
        instances: [$each.value]
`

func buildFs(rootPath string, files map[string]string) fsys.FS {
	filemap := fstest.MapFS{}
	for path, data := range files {
		filemap[path] = &fstest.MapFile{Data: []byte(data)}
	}
	return fsys.NewMemFS(rootPath, filemap)
}

func TestGetKforms(t *testing.T) {
	cases := map[string]struct {
		path  string
		files map[string]string
	}{
		"Basic": {
			path: "./example",
			files: map[string]string{
				"KformFile.yaml": kformfile,
				"main.yaml":      kformMain,
				"input.yaml":     kformInput,
				"output.yaml":    kformOutput,
			},
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
				Level: slog.LevelDebug,
			}))

			recorder := recorder.New[diag.Diagnostic]()

			ctx := context.Background()
			ctx = context.WithValue(ctx, types.CtxKeyRecorder, recorder)

			log.IntoContext(ctx, logger)
			p := moduleparser{
				path:     tc.path,
				fsys:     buildFs(tc.path, tc.files),
				recorder: recorder,
			}

			//kf, kforms, err := p.getKforms(context.Background())
			//assert.Error(t, err)
			//fmt.Println(kf.Spec.RequiredProviders)
			//fmt.Println(kforms)

			m := p.Parse(ctx)
			fmt.Println(recorder.Get())
			fmt.Println(recorder.Get().Error())
			if !recorder.Get().HasError() {
				fmt.Println("provider req", m.ProviderRequirements.List())
				for name, provider := range m.ProviderConfigs.List() {
					fmt.Printf("provider: %s, data: %v\n", name.Name, *provider)
				}
			}
		})
	}
}
