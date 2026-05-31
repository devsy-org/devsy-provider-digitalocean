package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/devsy-org/devsy-provider-digitalocean/pkg/digitalocean"
	"github.com/devsy-org/devsy-provider-digitalocean/pkg/options"
	"github.com/devsy-org/log"
	"github.com/spf13/cobra"
)

// DescribeCmd holds the cmd flags.
type DescribeCmd struct{}

// NewDescribeCmd defines a command.
func NewDescribeCmd() *cobra.Command {
	cmd := &DescribeCmd{}
	return &cobra.Command{
		Use:   "describe",
		Short: "Retrieve description of the virtual machine",
		RunE: func(_ *cobra.Command, args []string) error {
			options, err := options.FromEnv(false)
			if err != nil {
				return err
			}

			return cmd.Run(context.Background(), options, log.Default)
		},
	}
}

// Run runs the command logic.
func (cmd *DescribeCmd) Run(ctx context.Context, options *options.Options, log log.Logger) error {
	description, err := digitalocean.NewDigitalOcean(options.Token).Describe(ctx, options.MachineID)
	if err != nil {
		return err
	}

	_, err = fmt.Fprint(os.Stdout, description)
	return err
}
