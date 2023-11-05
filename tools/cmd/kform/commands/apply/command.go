package apply

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	docs "github.com/henderiw-nephio/kform/internal/docs/generated/applydocs"
	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/fn/fns"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/record"
	"github.com/henderiw-nephio/kform/tools/pkg/exec/vars"
	"github.com/henderiw-nephio/kform/tools/pkg/fsys"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio"
	"github.com/henderiw-nephio/kform/tools/pkg/recorder"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/parser"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/types"
	"github.com/henderiw-nephio/kform/tools/pkg/util/cache"
	"github.com/henderiw/logger/log"
	"github.com/spf13/cobra"
	"sigs.k8s.io/yaml"
)

// NewRunner returns a command runner.
func NewRunner(ctx context.Context, version string) *Runner {
	r := &Runner{}
	cmd := &cobra.Command{
		Use:     "apply [flags]",
		Args:    cobra.ExactArgs(1),
		Short:   docs.ApplyShort,
		Long:    docs.ApplyShort + "\n" + docs.ApplyLong,
		Example: docs.ApplyExamples,
		RunE:    r.runE,
	}

	r.Command = cmd

	r.Command.Flags().BoolVar(
		&r.AutoApprove, "auto-approve", false, "skip interactive approval of plan before applying")

	return r
}

func NewCommand(ctx context.Context, version string) *cobra.Command {
	return NewRunner(ctx, version).Command
}

type Runner struct {
	Command     *cobra.Command
	rootPath    string
	AutoApprove bool
}

func (r *Runner) runE(c *cobra.Command, args []string) error {
	ctx := c.Context()
	log := log.FromContext(ctx)

	r.rootPath = args[0]
	// validate the rootpath, so far we assume we run a directory calling the main function
	// but not within the main fn
	if err := fsys.ValidateDirPath(r.rootPath); err != nil {
		return err
	}
	// check if the root path exists
	_, err := os.Stat(r.rootPath)
	if err != nil {
		return fmt.Errorf("cannot init kform, path does not exist: %s", r.rootPath)
	}

	// initialize the recorder
	parserecorder := recorder.New[diag.Diagnostic]()
	ctx = context.WithValue(ctx, types.CtxKeyRecorder, parserecorder)

	// syntax check config -> build the dag
	log.Info("parsing modules")
	p, err := parser.NewKformParser(ctx, r.rootPath)
	if err != nil {
		return err
	}
	p.Parse(ctx)
	if parserecorder.Get().HasError() {
		parserecorder.Print()
		log.Error("failed parsing modules", "error", parserecorder.Get().Error())
		return parserecorder.Get().Error()
	}
	parserecorder.Print()
	fmt.Println("provider requirements", p.GetProviderRequirements(ctx))
	fmt.Println("provider configs", p.GetProviderConfigs(ctx))
	providerInventory, err := p.InitProviderInventory(ctx)
	if err != nil {
		log.Error("failed initializing provider inventory", "error", err)
		return err
	}
	providerInstances := p.InitProviderInstances(ctx)

	rm, err := p.GetRootModule(ctx)
	if err != nil {
		log.Error("failed parsing no root module found")
		return fmt.Errorf("failed parsing no root module found")
	}

	for nsn := range providerInstances.List() {
		fmt.Println("provider instance", nsn.Name)
	}

	runrecorder := recorder.New[record.Record]()
	varsCache := cache.New[vars.Variable]()

	// run the provider DAG
	log.Info("create provider runner")
	rmfn := fns.NewModuleFn(&fns.Config{
		Provider:          true,
		RootModuleName:    rm.NSN.Name,
		Vars:              varsCache,
		Recorder:          runrecorder,
		ProviderInstances: providerInstances,
		ProviderInventory: providerInventory,
	})
	log.Info("executing provider runner DAG")
	if err := rmfn.Run(ctx, &types.VertexContext{
		FileName:     filepath.Join(r.rootPath, pkgio.PkgFileMatch[0]),
		ModuleName:   rm.NSN.Name,
		BlockType:    types.BlockTypeModule,
		BlockName:    rm.NSN.Name,
		DAG:          rm.ProviderDAG, // we supply the provider DAG here
		BlockContext: types.KformBlockContext{},
	}, map[string]any{}); err != nil {
		log.Error("failed running provider DAG", "err", err)
		return err
	}
	log.Info("success executing provider DAG")

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	go func() {
		<-ctx.Done()
		providerInstances := providerInstances.List()
		fmt.Println("context Done", len(providerInstances))
		for nsn, provider := range providerInstances {
			if provider != nil {
				provider.Close(ctx)
				log.Info("closing provider", "nsn", nsn)
				continue
			}
			log.Info("closing provider nil", "nsn", nsn)
		}
	}()

	runrecorder.Print()

	for nsn := range providerInstances.List() {
		fmt.Println("provider instance", nsn.Name)
	}

	runrecorder = recorder.New[record.Record]()
	varsCache = cache.New[vars.Variable]()

	rmfn = fns.NewModuleFn(&fns.Config{
		RootModuleName:    rm.NSN.Name,
		Vars:              varsCache,
		Recorder:          runrecorder,
		ProviderInstances: providerInstances,
		ProviderInventory: providerInventory,
	})

	log.Info("executing module")
	if err := rmfn.Run(ctx, &types.VertexContext{
		FileName:     filepath.Join(r.rootPath, pkgio.PkgFileMatch[0]),
		ModuleName:   rm.NSN.Name,
		BlockType:    types.BlockTypeModule,
		BlockName:    rm.NSN.Name,
		DAG:          rm.DAG,
		BlockContext: types.KformBlockContext{},
	}, map[string]any{}); err != nil {
		log.Error("failed executing module", "err", err)
		return err
	}
	log.Info("success executing module")

	fsys := fsys.NewDiskFS(r.rootPath)
	if err := fsys.MkdirAll("out"); err != nil {
		return err
	}
	for nsn, v := range varsCache.List() {
		fmt.Println("nsn", nsn, "value", v.Data)
		for outputVarName, instances := range v.Data {
			for idx, instance := range instances {
				b, err := yaml.Marshal(instance)
				if err != nil {
					return err
				}
				fsys.WriteFile(filepath.Join("out", fmt.Sprintf("%s%d.yaml", outputVarName, idx)), b)
			}
		}

	}

	runrecorder.Print()
	// auto-apply -> depends on the flag if we approve the change or not.
	cancel()
	return nil
}
