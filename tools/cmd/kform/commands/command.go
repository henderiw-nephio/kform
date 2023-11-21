package commands

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/adrg/xdg"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/henderiw-nephio/kform/tools/cmd/kform/commands/apply"
	"github.com/henderiw-nephio/kform/tools/cmd/kform/commands/auth"
	initcmd "github.com/henderiw-nephio/kform/tools/cmd/kform/commands/init"
	"github.com/henderiw-nephio/kform/tools/cmd/kform/commands/pkg"
	"github.com/henderiw-nephio/kform/tools/pkg/fsys"
	"github.com/henderiw-nephio/kform/tools/pkg/store"
)

const (
	defaultConfigFileSubDir  = "kform"
	defaultConfigFileName    = "kform"
	defaultConfigFileNameExt = "yaml"
	defaultConfigEnvPrefix   = "KFORM"
	defaultDBPath            = "plugins_db"
)

var (
	configFile string
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
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			// create plugin store if it does not exist
			_, err := store.New(cmd.Context(),
				filepath.Join(xdg.ConfigHome, defaultConfigFileSubDir, defaultDBPath))
			return err
		},
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
	// ensure the viper config directory exists

	cobra.CheckErr(fsys.EnsureDir(ctx, xdg.ConfigHome, defaultConfigFileSubDir))
	// initialize viper settings
	cobra.OnInitialize(initConfig)

	cmd.AddCommand(initcmd.NewCommand(ctx, version))
	cmd.AddCommand(apply.NewCommand(ctx, version))
	cmd.AddCommand(auth.NewCommand(ctx, version))
	cmd.AddCommand(pkg.NewCommand(ctx, version))
	cmd.PersistentFlags().StringVar(&configFile, "config", "c", fmt.Sprintf("Default config file (%s/%s/%s.%s)", xdg.ConfigHome, defaultConfigFileSubDir, defaultConfigFileName, defaultConfigFileNameExt))

	return cmd
}

type Runner struct {
	Command *cobra.Command
	Ctx     context.Context
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	if configFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(configFile)
	} else {

		viper.AddConfigPath(filepath.Join(xdg.ConfigHome, defaultConfigFileName, defaultConfigFileNameExt))
		viper.SetConfigType(defaultConfigFileNameExt)
		viper.SetConfigName(defaultConfigFileSubDir)

		_ = viper.SafeWriteConfig()
	}

	//viper.Set("kubecontext", kubecontext)
	//viper.Set("kubeconfig", kubeconfig)

	viper.SetEnvPrefix(defaultConfigEnvPrefix)
	viper.AutomaticEnv() // read in environment variables that match

	// If a config file is found, read it in.
	if err := viper.ReadInConfig(); err == nil {
		_ = 1
	}
}
