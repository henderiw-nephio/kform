package pkg

import (
	"context"

	docs "github.com/henderiw-nephio/kform/internal/docs/generated/pkgdocs"
	buildcmd "github.com/henderiw-nephio/kform/tools/cmd/kform/commands/pkg/build"
	initcmd "github.com/henderiw-nephio/kform/tools/cmd/kform/commands/pkg/init"
	pullcmd "github.com/henderiw-nephio/kform/tools/cmd/kform/commands/pkg/pull"
	pushcmd "github.com/henderiw-nephio/kform/tools/cmd/kform/commands/pkg/push"
	"github.com/henderiw-nephio/kform/tools/cmd/kform/commands/pkg/tree"
	"github.com/spf13/cobra"
)

func NewCommand(ctx context.Context, version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "pkg",
		Short: docs.PkgShort,
		Long:  docs.PkgShort + "\n" + docs.PkgLong,
		//Example: docs.PkgExamples,
		RunE: func(cmd *cobra.Command, _ []string) error { return cmd.Usage() },
	}

	cmd.AddCommand(buildcmd.NewCommand(ctx, version))
	cmd.AddCommand(pushcmd.NewCommand(ctx, version))
	cmd.AddCommand(pullcmd.NewCommand(ctx, version))
	cmd.AddCommand(initcmd.NewCommand(ctx, version))
	cmd.AddCommand(treecmd.NewCommand(ctx, version))
	return cmd
}
