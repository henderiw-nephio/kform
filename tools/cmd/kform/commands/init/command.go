package init

import (
	"context"
	"fmt"
	"os"

	docs "github.com/henderiw-nephio/kform/internal/docs/generated/initdocs"
	"github.com/henderiw-nephio/kform/kform-sdk-go/pkg/diag"
	"github.com/henderiw-nephio/kform/tools/pkg/fsys"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/parser"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/types"
	"github.com/henderiw/logger/log"
	"github.com/spf13/cobra"
)

// NewRunner returns a command runner.
func NewRunner(ctx context.Context, version string) *Runner {
	r := &Runner{}
	cmd := &cobra.Command{
		Use:     "init DIR [flags]",
		Args:    cobra.MaximumNArgs(1),
		Short:   docs.InitShort,
		Long:    docs.InitShort + "\n" + docs.InitLong,
		Example: docs.InitExamples,
		RunE:    r.runE,
	}

	r.Command = cmd
	/*
		r.Command.Flags().StringVar(
			&r.FnConfigDir, "fn-config-dir", "", "dir where the function config files are located")
	*/
	return r
}

func NewCommand(ctx context.Context, version string) *cobra.Command {
	return NewRunner(ctx, version).Command
}

type Runner struct {
	Command  *cobra.Command
	rootPath string
}

func (r *Runner) runE(c *cobra.Command, args []string) error {
	ctx := c.Context()
	log := log.FromContext(ctx)
	//log := log.FromContext(ctx)
	r.rootPath = args[0]
	if err := fsys.ValidateDirPath(r.rootPath); err != nil {
		return err
	}
	// check if the root path exists
	_, err := os.Stat(r.rootPath)
	if err != nil {
		return fmt.Errorf("cannot init kform, path does not exist: %s", r.rootPath)
	}

	// initialize the recorder
	recorder := diag.NewRecorder()
	ctx = context.WithValue(ctx, types.CtxKeyRecorder, recorder)

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
	recorder.Print()

	provReqs := p.GetProviderRequirements(ctx)
	for nsn, reqs := range provReqs {
		fmt.Printf("provider: %s res: %v\n", nsn.Name, reqs)
	}

	// init and/or restore backend
	// syntax check config -> build the dag but dont use it (done)
	// download module
	// download providers
	// -> lock files

	return nil
}
