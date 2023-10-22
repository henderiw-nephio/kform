package auth

import (
	"context"

	"github.com/henderiw-nephio/kform/tools/cmd/kform/commands/auth/get"
	"github.com/henderiw-nephio/kform/tools/cmd/kform/commands/auth/login"
	"github.com/henderiw-nephio/kform/tools/cmd/kform/commands/auth/logout"
	"github.com/spf13/cobra"
)

func NewCommand(ctx context.Context, version string) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "auth",
		Short: "Login or access registry credentials",
		//Short:   docs.ConfigShort,
		//Long:    docs.ConfigShort + "\n" + docs.ConfigLong,
		//Example: docs.ConfigExamples,
		RunE: func(cmd *cobra.Command, _ []string) error { return cmd.Usage() },
	}

	cmd.AddCommand(get.NewCommand(ctx, version))
	cmd.AddCommand(login.NewCommand(ctx, version))
	cmd.AddCommand(logout.NewCommand(ctx, version))
	return cmd
}
