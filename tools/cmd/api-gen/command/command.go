package command

import (
	"context"
	"log/slog"

	"github.com/henderiw-nephio/kform/tools/pkg/loader"
	"github.com/henderiw-nephio/kform/tools/pkg/markers"
	"github.com/henderiw-nephio/kform/tools/pkg/openapi"
	"github.com/spf13/cobra"
)

func GetMain(ctx context.Context) *cobra.Command {
	showVersion := false
	paths := []string{}
	allowDangerousTypes := false
	ignoreUnexportedFields := false
	generateEmbeddedObjectMeta := false
	cmd := &cobra.Command{
		Use:          "api-gen",
		Short:        "api-gen is a tool to help generate api resources",
		Long:         "api-gen is a tool to help generate api resources",
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
			if len(paths) > 0 {
				l := loader.NewLoader()
				roots, err := l.LoadRoots(paths...)
				if err != nil {
					return err
				}
				if len(roots) == 0 {
					slog.Info("no roots")
					return nil
				}

				ctx = context.WithValue(ctx, "roots", roots)
				ctx = context.WithValue(ctx, "allowDangerousTypes", allowDangerousTypes)
				ctx = context.WithValue(ctx, "ignoreUnexportedFields", ignoreUnexportedFields)
				ctx = context.WithValue(ctx, "generateEmbeddedObjectMeta", generateEmbeddedObjectMeta)

				ctx = context.WithValue(ctx, "registry", markers.NewRegistry(ctx))
				ctx = context.WithValue(ctx, "collector", markers.NewCollector(ctx))
				ctx = context.WithValue(ctx, "checker", loader.NewTypeChecker())

				g := openapi.NewGenerator()
				g.RegisterMarkers(ctx)
				g.Generate(ctx)

				return nil
			}
			return cmd.Usage()
		},
	}
	cmd.Flags().BoolVarP(&showVersion, "version", "v", false, "show version")
	cmd.Flags().StringSliceVarP(&paths, "paths", "p", []string{}, "defines the paths for which we generate the apis")
	cmd.Flags().BoolVarP(&allowDangerousTypes, "allowDangerousTypes", "", false, "allowDangerousTypes")
	cmd.Flags().BoolVarP(&ignoreUnexportedFields, "ignoreUnexportedFields", "", false, "ignoreUnexportedFields")
	cmd.Flags().BoolVarP(&generateEmbeddedObjectMeta, "generateEmbeddedObjectMeta", "", false, "generateEmbeddedObjectMeta")

	return cmd
}

type Runner struct {
	Command *cobra.Command
	Ctx     context.Context
}
