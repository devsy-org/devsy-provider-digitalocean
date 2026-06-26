package cmd

import (
	"context"
	"fmt"
	"os"

	"github.com/devsy-org/devsy-provider-digitalocean/pkg/digitalocean"
	"github.com/devsy-org/devsy-provider-digitalocean/pkg/options"
	"github.com/devsy-org/devsy/pkg/ssh"
	"github.com/digitalocean/godo"
	"github.com/spf13/cobra"
)

// CommandCmd holds the cmd flags
type CommandCmd struct{}

// NewCommandCmd defines a command
func NewCommandCmd() *cobra.Command {
	cmd := &CommandCmd{}
	commandCmd := &cobra.Command{
		Use:   "command",
		Short: "Run a command on the instance",
		RunE: func(_ *cobra.Command, args []string) error {
			options, err := options.FromEnv(false)
			if err != nil {
				return err
			}

			return cmd.Run(context.Background(), options)
		},
	}

	return commandCmd
}

// Run runs the command logic
func (cmd *CommandCmd) Run(ctx context.Context, options *options.Options) error {
	command := os.Getenv("COMMAND")
	if command == "" {
		return fmt.Errorf("command environment variable is missing")
	}

	// get private key
	privateKey, err := ssh.GetPrivateKeyRawBase(options.MachineFolder)
	if err != nil {
		return fmt.Errorf("load private key: %w", err)
	}

	// create client
	droplet, err := digitalocean.NewDigitalOcean(options.Token).GetByName(ctx, options.MachineID)
	if err != nil {
		return err
	} else if droplet == nil {
		return fmt.Errorf("droplet not found")
	}

	// get external ip
	externalIP, err := publicIPv4(droplet)
	if err != nil {
		return err
	}

	// dial external address
	sshClient, err := ssh.NewSSHClient("devsy", externalIP+":22", privateKey)
	if err != nil {
		return fmt.Errorf("create ssh client: %w", err)
	}
	defer sshClient.Close()

	// run command
	return ssh.Run(context.Background(), ssh.RunOptions{
		Client:  sshClient,
		Command: command,
		Stdin:   os.Stdin,
		Stdout:  os.Stdout,
		Stderr:  os.Stderr,
	})
}

// publicIPv4 returns the first public IPv4 address of the droplet.
func publicIPv4(droplet *godo.Droplet) (string, error) {
	if droplet.Networks == nil {
		return "", fmt.Errorf("couldn't find public ip address")
	}

	for _, network := range droplet.Networks.V4 {
		if network.Type == "public" && network.IPAddress != "" {
			return network.IPAddress, nil
		}
	}

	return "", fmt.Errorf("couldn't find a public ip address")
}
