package apply

import (
	"context"
	"fmt"
	"os"

	docs "github.com/henderiw-nephio/kform/internal/docs/generated/applydocs"
	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/tools/pkg/fsys"
	"github.com/henderiw-nephio/kform/tools/pkg/recorder"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/parser"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/types"
	"github.com/henderiw/logger/log"
	"github.com/spf13/cobra"
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
	recorder := recorder.New[diag.Diagnostic]()
	ctx = context.WithValue(ctx, types.CtxKeyRecorder, recorder)

	// syntax check config -> build the dag
	log.Info("parsing modules")
	p, err := parser.NewKformParser(ctx, r.rootPath)
	if err != nil {
		return err
	}
	p.Parse(ctx)
	if recorder.Get().HasError() {
		recorder.Print()
		log.Error("failed parsing modules", "error", recorder.Get().Error())
		return recorder.Get().Error()
	}
	log.Info("generate dag(s)")

	rm := p.GetRootModule(ctx)
	if rm == nil {
		log.Error("failed parsing no root module found")
		return fmt.Errorf("failed parsing no root module found")
	}
	fmt.Println(rm.NSN.Name, rm.Kind)
	rm.DAG.Print(rm.NSN.Name)

	recorder.Print()

	// initialize the providers -> provider factory
	// execute the dag
	// auto-apply -> depends on the flag if we approve the change or not.
	return nil
}
