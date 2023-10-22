package commands

import (
	"context"

	"github.com/henderiw-nephio/kform/tools/cmd/kform/commands/apply"
	"github.com/henderiw-nephio/kform/tools/cmd/kform/commands/auth"
	initcmd "github.com/henderiw-nephio/kform/tools/cmd/kform/commands/init"
	"github.com/henderiw-nephio/kform/tools/cmd/kform/commands/pkg"
	"github.com/spf13/cobra"
)

func GetMain(ctx context.Context) *cobra.Command {
	//showVersion := false
	cmd := &cobra.Command{
		Use:          "kform",
		Short:        "kform is a KRM orchestration tool",
		Long:         "kform is a KRM orchestration tool",
		SilenceUsage: true,
		// We handle all errors in main after return from cobra so we can
		// adjust the error message coming from libraries
		SilenceErrors: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			h, err := cmd.Flags().GetBool("help")
			if err != nil {
				return err
			}
			if h {
				return cmd.Help()
			}

			return cmd.Usage()
		},
	}

	cmd.AddCommand(initcmd.NewCommand(ctx, version))
	cmd.AddCommand(apply.NewCommand(ctx, version))
	cmd.AddCommand(auth.NewCommand(ctx, version))
	cmd.AddCommand(pkg.NewCommand(ctx, version))
	//cmd.Flags().BoolVarP(&showVersion, "version", "v", false, "show version")

	return cmd
}

type Runner struct {
	Command *cobra.Command
	Ctx     context.Context
}
