package initcmd

import (
	"context"
	"fmt"

	docs "github.com/henderiw-nephio/kform/internal/docs/generated/pkgdocs"
	kformpkgmetav1alpha1 "github.com/henderiw-nephio/kform/tools/apis/kform/pkg/meta/v1alpha1"
	"github.com/henderiw-nephio/kform/tools/pkg/fsys"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio"
	"github.com/spf13/cobra"
)

// NewRunner returns a command runner.
func NewRunner(ctx context.Context, version string) *Runner {
	r := &Runner{}
	cmd := &cobra.Command{
		Use:     "init PACKAGE-TYPE DIR [flags]",
		Args:    cobra.ExactArgs(2),
		Short:   docs.InitShort,
		Long:    docs.InitShort + "\n" + docs.InitLong,
		Example: docs.InitExamples,
		RunE:    r.runE,
	}

	r.Command = cmd

	r.Command.Flags().StringVar(&r.description, "description", "sample description", "short description of the package.")

	//r.Command.Usage()
	return r
}

func NewCommand(ctx context.Context, version string) *cobra.Command {
	return NewRunner(ctx, version).Command
}

type Runner struct {
	Command     *cobra.Command
	rootPath    string
	pkgType     string
	description string
}

func (r *Runner) runE(c *cobra.Command, args []string) error {
	r.rootPath = args[1]
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

	var pkgrw pkgio.PkgInitReadWriter
	r.pkgType = args[0]
	switch r.pkgType {
	case string(kformpkgmetav1alpha1.PkgKindProvider):
		pkgrw = pkgio.NewPkgInitReadWriter(r.rootPath, kformpkgmetav1alpha1.PkgKindProvider, []string{"image", "schemas/provider", "schemas/resources"})
	case string(kformpkgmetav1alpha1.PkgKindModule):
		pkgrw = pkgio.NewPkgInitReadWriter(r.rootPath, kformpkgmetav1alpha1.PkgKindModule, []string{})
	default:
		return fmt.Errorf("unsupported packageType, expecting %s or %s, got %s", kformpkgmetav1alpha1.PkgKindProvider, kformpkgmetav1alpha1.PkgKindModule, r.pkgType)
	}
	p := pkgio.Pipeline{
		Inputs:  []pkgio.Reader{pkgrw},
		Outputs: []pkgio.Writer{pkgrw},
	}
	return p.Execute()
}
