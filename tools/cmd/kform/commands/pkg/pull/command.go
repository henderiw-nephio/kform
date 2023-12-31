package pullcmd

import (
	"context"
	"fmt"
	"os"

	docs "github.com/henderiw-nephio/kform/internal/docs/generated/pkgdocs"
	kformpkgmetav1alpha1 "github.com/henderiw-nephio/kform/tools/apis/kform/pkg/meta/v1alpha1"
	"github.com/henderiw-nephio/kform/tools/pkg/pkgio"
	"github.com/henderiw-nephio/kform/tools/pkg/syntax/address"
	"github.com/pkg/errors"
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
	r.Command.Flags().StringVarP(&r.kind, "kind", "", "module", "package kind (module or provider, default: module)")
	return r
}

func NewCommand(ctx context.Context, version string) *cobra.Command {
	return NewRunner(ctx, version).Command
}

type Runner struct {
	Command *cobra.Command
	kind    string
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

	if err := kformpkgmetav1alpha1.ValidatePackageType(r.kind); err != nil {
		return errors.Wrap(err, "invalid packageType")
	}

	pkg, err := address.GetPackageFromRef(args[0])
	if err != nil {
		return errors.Wrap(err, "cannot get package from ref")
	}

	pkgrw := pkgio.NewPkgPullReadWriter(rootPath, pkg, kformpkgmetav1alpha1.PkgKind(r.kind))
	p := pkgio.Pipeline{
		Inputs:  []pkgio.Reader{pkgrw},
		Outputs: []pkgio.Writer{pkgrw},
	}
	return p.Execute(c.Context())

}
