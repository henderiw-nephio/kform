package pushcmd

import (
	"context"
	"fmt"
	"os"

	docs "github.com/henderiw-nephio/kform/internal/docs/generated/pkgdocs"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/address"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

// NewRunner returns a command runner.
func NewRunner(ctx context.Context, version string) *Runner {
	r := &Runner{}
	cmd := &cobra.Command{
		Use:     "push REF DIR [flags]",
		Args:    cobra.ExactArgs(2),
		Short:   docs.PushShort,
		Long:    docs.PushShort + "\n" + docs.PushLong,
		Example: docs.PushExamples,
		RunE:    r.runE,
	}

	r.Command = cmd

	r.Command.Flags().BoolVarP(&r.releaser, "releaser", "", false, "push command is used as a releaser e.g. as a github action")
	return r
}

func NewCommand(ctx context.Context, version string) *cobra.Command {
	return NewRunner(ctx, version).Command
}

type Runner struct {
	Command  *cobra.Command
	releaser bool
}

func (r *Runner) runE(c *cobra.Command, args []string) error {
	rootPath := args[1]
	f, err := os.Stat(rootPath)
	if err != nil {
		return fmt.Errorf("cannot create a pkg, rootpath %s does not exist", rootPath)
	}
	if !f.IsDir() {
		return fmt.Errorf("cannot initialize a pkg on a file, please provide a directory instead, file: %s", rootPath)
	}

	pkg, err := address.GetPackageFromRef(args[0])
	if err != nil {
		return errors.Wrap(err, "cannot get package from ref")
	}

	pkgrw := pkgio.NewPkgPushReadWriter(rootPath, pkg, r.releaser)
	p := pkgio.Pipeline{
		Inputs:  []pkgio.Reader{pkgrw},
		Outputs: []pkgio.Writer{pkgrw},
	}
	return p.Execute(c.Context())

}
