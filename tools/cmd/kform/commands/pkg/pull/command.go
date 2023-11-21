package pullcmd

import (
	"context"

	docs "github.com/henderiw-nephio/kform/internal/docs/generated/pkgdocs"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio/oras"
	"github.com/spf13/cobra"
)

// NewRunner returns a command runner.
func NewRunner(ctx context.Context, version string) *Runner {
	r := &Runner{}
	cmd := &cobra.Command{
		Use:     "pull REF DIR [flags]",
		Args:    cobra.ExactArgs(2),
		Short:   docs.PullShort,
		Long:    docs.PullShort + "\n" + docs.PullLong,
		Example: docs.PullExamples,
		RunE:    r.runE,
	}

	r.Command = cmd

	return r
}

func NewCommand(ctx context.Context, version string) *cobra.Command {
	return NewRunner(ctx, version).Command
}

type Runner struct {
	Command *cobra.Command
}

func (r *Runner) runE(c *cobra.Command, args []string) error {

	/*
		client, err := registry.NewClient()
		if err != nil {
			return err
		}
	*/
	if err := oras.Pull(c.Context(), args[0], nil); err != nil {
		return err
	}

	//fmt.Println(result)

	return nil

}
