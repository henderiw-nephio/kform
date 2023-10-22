package get

import (
	"context"

	"github.com/spf13/cobra"
)

// NewRunner returns a command runner.
func NewRunner(ctx context.Context, version string) *Runner {
	r := &Runner{}
	cmd := &cobra.Command{
		Use:   "get [REGISTRY]",
		Args:  cobra.MaximumNArgs(1),
		Short: "Implements a credential helper",
		//Short:   docs.ConfigShort,
		//Long:    docs.ConfigShort + "\n" + docs.ConfigLong,
		//Example: docs.ConfigExamples,
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
