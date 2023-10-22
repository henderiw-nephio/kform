package login

import (
	"context"

	"github.com/google/go-containerregistry/pkg/name"
	"github.com/spf13/cobra"
)

// NewRunner returns a command runner.
func NewRunner(ctx context.Context, version string) *Runner {
	r := &Runner{}
	cmd := &cobra.Command{
		Use:   "login [OPTIONS] [SERVER]",
		Args:  cobra.ExactArgs(1),
		Short: "Log in to a registry",
		//Short:   docs.ConfigShort,
		//Long:    docs.ConfigShort + "\n" + docs.ConfigLong,
		//Example: docs.ConfigExamples,
		RunE: r.runE,
	}

	r.Command = cmd

	r.Command.Flags().StringVarP(&r.loginOpts.user, "username", "u", "", "Username")
	r.Command.Flags().StringVarP(&r.loginOpts.password, "password", "p", "", "Password")
	r.Command.Flags().BoolVarP(&r.loginOpts.passwordStdin, "password-stdin", "", false, "Take the password from stdin")
	return r
}

func NewCommand(ctx context.Context, version string) *cobra.Command {
	return NewRunner(ctx, version).Command
}

type Runner struct {
	Command   *cobra.Command
	loginOpts loginOptions
}

type loginOptions struct {
	serverAddress string
	user          string
	password      string
	passwordStdin bool
}

func (r *Runner) runE(c *cobra.Command, args []string) error {
	reg, err := name.NewRegistry(args[0])
	if err != nil {
		return err
	}
	r.loginOpts.serverAddress = reg.Name()

	// TODO login
	return nil
}
