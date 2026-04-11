package cmd

import (
	"context"
	"fmt"
	"strconv"

	"github.com/skevetter/devpod-provider-digitalocean/pkg/digitalocean"
	"github.com/skevetter/devpod-provider-digitalocean/pkg/options"
	"github.com/skevetter/log"
	"github.com/spf13/cobra"
)

// StartCmd holds the cmd flags
type StartCmd struct{}

// NewStartCmd defines a command
func NewStartCmd() *cobra.Command {
	cmd := &StartCmd{}
	startCmd := &cobra.Command{
		Use:   "start",
		Short: "Start an instance",
		RunE: func(_ *cobra.Command, args []string) error {
			options, err := options.FromEnv(false)
			if err != nil {
				return err
			}

			return cmd.Run(context.Background(), options, log.Default)
		},
	}

	return startCmd
}

// Run runs the command logic
func (cmd *StartCmd) Run(ctx context.Context, options *options.Options, log log.Logger) error {
	req, err := buildInstance(options)
	if err != nil {
		return err
	}

	diskSize, err := strconv.Atoi(options.DiskSize)
	if err != nil {
		return fmt.Errorf("parse disk size: %w", err)
	}

	return digitalocean.NewDigitalOcean(options.Token).Create(ctx, req, diskSize)
}
