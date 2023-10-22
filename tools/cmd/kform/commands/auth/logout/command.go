package logout

import (
	"context"
	"log"
	"os"

	"github.com/docker/cli/cli/config"
	"github.com/google/go-containerregistry/pkg/authn"
	"github.com/google/go-containerregistry/pkg/name"
	"github.com/spf13/cobra"
)

// NewRunner returns a command runner.
func NewRunner(ctx context.Context, version string) *Runner {
	r := &Runner{}
	cmd := &cobra.Command{
		Use:   "logout [SERVER]",
		Args:  cobra.ExactArgs(1),
		Short: "Log out of a registry",
		//Short:   docs.ConfigShort,
		//Long:    docs.ConfigShort + "\n" + docs.ConfigLong,
		//Example: docs.ConfigExamples,
		RunE: r.runE,
	}

	r.Command = cmd

	return r
}

func NewCommand(ctx context.Context, version string) *cobra.Command {
	return NewRunner(ctx, version).Command
}

type Runner struct {
	Command *cobra.Command
}

func (r *Runner) runE(c *cobra.Command, args []string) error {
	reg, err := name.NewRegistry(args[0])
	if err != nil {
		return err
	}
	serverAddress := reg.Name()

	cf, err := config.Load(os.Getenv("DOCKER_CONFIG"))
	if err != nil {
		return err
	}
	creds := cf.GetCredentialsStore(serverAddress)
	if serverAddress == name.DefaultRegistry {
		serverAddress = authn.DefaultAuthKey
	}
	if err := creds.Erase(serverAddress); err != nil {
		return err
	}

	if err := cf.Save(); err != nil {
		return err
	}
	log.Printf("logged out via %s", cf.Filename)
	return nil
}
