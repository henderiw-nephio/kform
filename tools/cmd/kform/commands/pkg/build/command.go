package buildcmd

import (
	"context"
	"fmt"

	docs "github.com/henderiw-nephio/kform/internal/docs/generated/pkgdocs"
	"github.com/henderiw-nephio/kform/tools/pkg/fsys"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio"
	"github.com/spf13/cobra"
)

// NewRunner returns a command runner.
func NewRunner(ctx context.Context, version string) *Runner {
	r := &Runner{}
	cmd := &cobra.Command{
		Use:     "build DIR [flags]",
		Args:    cobra.ExactArgs(1),
		Short:   docs.BuildShort,
		Long:    docs.BuildShort + "\n" + docs.BuildLong,
		Example: docs.BuildExamples,
		RunE:    r.runE,
	}

	r.Command = cmd

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
	r.rootPath = args[0]
	if err := fsys.ValidateDirPath(r.rootPath); err != nil {
		return err
	}
	fs := fsys.NewDiskFS(".")
	f, err := fs.Stat(r.rootPath)
	if err != nil {
		fs.MkdirAll(r.rootPath)
	} else if !f.IsDir() {
		return fmt.Errorf("cannot initialize a pkg on a file, please provide a directory instead, file: %s", r.rootPath)
	}

	pkgrw := pkgio.NewPkgBuildReadWriter(r.rootPath)
	p := pkgio.Pipeline{
		Inputs:  []pkgio.Reader{pkgrw},
		Outputs: []pkgio.Writer{pkgrw},
	}
	return p.Execute(c.Context())
}
