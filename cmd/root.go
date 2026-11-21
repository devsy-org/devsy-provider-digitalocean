package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/devsy-org/devsy/logging"
	"github.com/devsy-org/devsy/pkg/log"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
)

// NewRootCmd returns a new root command
func NewRootCmd() *cobra.Command {
	doCmd := &cobra.Command{
		Use:           "devsy-provider-digitalocean",
		Short:         "digitalocean provider commands",
		SilenceErrors: true,
		SilenceUsage:  true,

		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			return initProviderLogging()
		},
	}

	return doCmd
}

func initProviderLogging() error {
	log.Init(log.Config{DefaultLevel: log.LevelInfoName, Format: "json"})

	cfg, invalid := logging.ResolveProvider(
		os.Getenv("DEVSY_LOG_LEVEL"),
		os.Getenv("DEVSY_LOG_FORMAT"),
	)
	if invalid.Level {
		return fmt.Errorf("invalid DEVSY_LOG_LEVEL")
	}
	if invalid.Format {
		return fmt.Errorf("invalid DEVSY_LOG_FORMAT")
	}

	log.Init(
		log.Config{
			Level:        logging.LevelName(cfg.Level),
			DefaultLevel: log.LevelInfoName,
			Format:       string(cfg.Format),
		},
	)
	return nil
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	// build the root command
	rootCmd := BuildRoot()

	// execute command
	err := rootCmd.Execute()
	if err != nil {
		if exitErr, ok := err.(*ssh.ExitError); ok {
			os.Exit(exitErr.ExitStatus())
		}
		if exitErr, ok := err.(*exec.ExitError); ok {
			if len(exitErr.Stderr) > 0 {
				log.Error(string(exitErr.Stderr))
			}
			os.Exit(exitErr.ExitCode())
		}

		log.Fatal(err)
	}
}

// BuildRoot creates a new root command from the
func BuildRoot() *cobra.Command {
	rootCmd := NewRootCmd()

	rootCmd.AddCommand(NewCreateCmd())
	rootCmd.AddCommand(NewStatusCmd())
	rootCmd.AddCommand(NewDeleteCmd())
	rootCmd.AddCommand(NewStartCmd())
	rootCmd.AddCommand(NewStopCmd())
	rootCmd.AddCommand(NewCommandCmd())
	rootCmd.AddCommand(NewInitCmd())
	rootCmd.AddCommand(NewDescribeCmd())
	return rootCmd
}
