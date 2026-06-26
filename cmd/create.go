package cmd

import (
	"context"
	"encoding/base64"
	"fmt"
	"strconv"

	"github.com/devsy-org/devsy-provider-digitalocean/pkg/digitalocean"
	"github.com/devsy-org/devsy-provider-digitalocean/pkg/options"
	"github.com/devsy-org/devsy/pkg/ssh"
	"github.com/digitalocean/godo"
	"github.com/spf13/cobra"
)

// CreateCmd holds the cmd flags
type CreateCmd struct{}

// NewCreateCmd defines a command
func NewCreateCmd() *cobra.Command {
	cmd := &CreateCmd{}
	createCmd := &cobra.Command{
		Use:   "create",
		Short: "Create an instance",
		RunE: func(_ *cobra.Command, args []string) error {
			options, err := options.FromEnv(false)
			if err != nil {
				return err
			}

			return cmd.Run(context.Background(), options)
		},
	}

	return createCmd
}

// Run runs the command logic
func (cmd *CreateCmd) Run(ctx context.Context, options *options.Options) error {
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

func GetInjectKeypairScript(machineFolder, machineID string) (string, error) {
	publicKeyBase, err := ssh.GetPublicKeyBase(machineFolder)
	if err != nil {
		return "", err
	}

	publicKey, err := base64.StdEncoding.DecodeString(publicKeyBase)
	if err != nil {
		return "", err
	}

	resultScript := `#!/bin/sh

# Mount volume to home
mkdir -p /home/devsy
mount -o discard,defaults,noatime /dev/disk/by-id/scsi-0DO_Volume_` + machineID + ` /home/devsy

# Move docker data dir
service docker stop
cat > /etc/docker/daemon.json << EOF
{
  "data-root": "/home/devsy/.docker-daemon",
  "live-restore": true
}
EOF
# Make sure we only copy if volumes isn't initialized
if [ ! -d "/home/devsy/.docker-daemon" ]; then
  mkdir -p /home/devsy/.docker-daemon
  rsync -aP /var/lib/docker/ /home/devsy/.docker-daemon
fi
service docker start

# Create Devsy user and configure ssh
useradd devsy -d /home/devsy
if grep -q sudo /etc/groups; then
	usermod -aG sudo devsy
elif grep -q wheel /etc/groups; then
	usermod -aG wheel devsy
fi
echo "devsy ALL=(ALL) NOPASSWD:ALL" > /etc/sudoers.d/91-devsy
mkdir -p /home/devsy/.ssh
echo '` + string(publicKey) + `' > /home/devsy/.ssh/authorized_keys
chmod 0700 /home/devsy/.ssh
chmod 0600 /home/devsy/.ssh/authorized_keys
chown devsy:devsy /home/devsy
chown -R devsy:devsy /home/devsy/.ssh

# Make sure we don't get limited
ufw allow 22/tcp || true
`

	return resultScript, nil
}

func buildInstance(options *options.Options) (*godo.DropletCreateRequest, error) {
	// generate ssh keys
	userData, err := GetInjectKeypairScript(options.MachineFolder, options.MachineID)
	if err != nil {
		return nil, err
	}

	// generate instance object
	instance := &godo.DropletCreateRequest{
		Name:   options.MachineID,
		Region: options.Region,
		Size:   options.MachineType,
		Image: godo.DropletCreateImage{
			Slug: options.DiskImage,
		},
		UserData: userData,
		Tags:     []string{"devsy"},
	}

	return instance, nil
}
