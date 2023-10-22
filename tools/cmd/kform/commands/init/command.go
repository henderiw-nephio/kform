package init

import (
	"context"
	"github.com/spf13/cobra"
	docs "github.com/henderiw-nephio/kform/internal/docs/generated/initdocs"

)

// NewRunner returns a command runner.
func NewRunner(ctx context.Context, version string) *Runner {
	r := &Runner{}
	cmd := &cobra.Command{
		Use:   "init [flags]",
		Args:  cobra.MaximumNArgs(0),
		Short:   docs.InitShort,
		Long:    docs.InitShort + "\n" + docs.InitLong,
		Example: docs.InitExamples,
		RunE: r.runE,
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
	Command *cobra.Command
}

func (r *Runner) runE(c *cobra.Command, args []string) error {

	// init and/or restore backend
	// syntax check config -> build the dag but dont use it
	// download module
	// download providers
	// -> lock files

	return nil
}
