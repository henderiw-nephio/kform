package apply

import (
	"context"

	"github.com/spf13/cobra"
	docs "github.com/henderiw-nephio/kform/internal/docs/generated/applydocs"

)

// NewRunner returns a command runner.
func NewRunner(ctx context.Context, version string) *Runner {
	r := &Runner{}
	cmd := &cobra.Command{
		Use:   "apply [flags]",
		Args:  cobra.MaximumNArgs(0),
		Short:   docs.ApplyShort,
		Long:    docs.ApplyShort + "\n" + docs.ApplyLong,
		Example: docs.ApplyExamples,
		RunE: r.runE,
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
	AutoApprove bool
}

func (r *Runner) runE(c *cobra.Command, args []string) error {
	// initialize the providers -> provider factory
	// syntax check config -> build the dag
	// execute the dag
	// auto-apply -> depends on the flag if we approve the change or not.
	return nil
}
