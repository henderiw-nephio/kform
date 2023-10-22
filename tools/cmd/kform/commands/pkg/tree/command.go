package treecmd

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
		Use:     "tree [DIR] [flags]",
		Args:    cobra.ExactArgs(1),
		Short:   docs.TreeShort,
		Long:    docs.TreeShort + "\n" + docs.TreeLong,
		Example: docs.TreeExamples,
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

	pkgrw := pkgio.NewPkgTreeReadWriter(r.rootPath)
	p := pkgio.Pipeline{
		Inputs:  []pkgio.Reader{pkgrw},
		Outputs: []pkgio.Writer{pkgrw},
	}
	return p.Execute()
}

/*
func (r *Runner) runE(c *cobra.Command, args []string) error {
	ctx := c.Context()
	var input kio.Reader
	var root = "."
	if len(args) == 0 {
		args = append(args, root)
	}
	r.pkgPath = args[0]

	input = kio.LocalPackageReader{
		PackagePath:       r.pkgPath,
		MatchFilesGlob:    r.getMatchFilesGlob(),
		PreserveSeqIndent: true,
		WrapBareSeqNode:   true,
	}
	fltrs := []kio.Filter{&filters.IsLocalConfig{
		IncludeLocalConfig: true,
	}}

	return HandleError(ctx, kio.Pipeline{
		Inputs:  []kio.Reader{input},
		Filters: fltrs,
		Outputs: []kio.Writer{TreeWriter{
			Root:   root,
			Writer: printer.FromContextOrDie(ctx).OutStream(),
		}},
	}.Execute())
}

func (r *Runner) getMatchFilesGlob() []string {
	return append([]string{}, kio.DefaultMatch...)
}

// ExitOnError if true, will cause commands to call os.Exit instead of returning an error.
// Used for skipping printing usage on failure.
var ExitOnError bool

// StackOnError if true, will print a stack trace on failure.
var StackOnError bool

func HandleError(ctx context.Context, err error) error {
	if err == nil {
		return nil
	}
	pr := printer.FromContextOrDie(ctx)
	if StackOnError {
		if err, ok := err.(*errors.Error); ok {
			pr.Printf("%s", err.Stack())
		}
	}

	if ExitOnError {
		pr.Printf("Error: %v\n", err)
		os.Exit(1)
	}
	return err
}
*/
