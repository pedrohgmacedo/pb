// Package commands implements CLI commands
package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"pb/clipboard"
	"pb/util"
	"strconv"
)

var (
	// These variables are populated by the persistent flags and are available to all subcommands.
	serverAddress string
	port          int
	keyPath       string
	enableLogging bool
)

var rootCmd = &cobra.Command{
	Use:     util.ProgramName,
	Version: util.GitHead,
	Short:   "copies and pastes text between machines.",
	Long:    `A simple tool for sharing your clipboard over the network, using HTTPS and SSH key authentication.`,
	// This function runs before any subcommand executes.
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		// Enable logging if --log flag is set
		if enableLogging {
			clipboard.EnableLogging()
		}

		// This logic only applies to commands that have these flags.
		// The server command, for example, doesn't have a "server" flag.
		if cmd.Flags().Lookup("server") != nil {
			if !cmd.Flags().Changed("server") {
				if envServer := os.Getenv(util.EnvVarServer); envServer != "" {
					serverAddress = envServer
				}
			}
		}

		if cmd.Flags().Lookup("port") != nil {
			if !cmd.Flags().Changed("port") {
				if envPortStr := os.Getenv(util.EnvVarPort); envPortStr != "" {
					if envPort, err := strconv.Atoi(envPortStr); err == nil {
						port = envPort
					}
				}
			}
		}

		if cmd.Flags().Lookup("key") != nil {
			if !cmd.Flags().Changed("key") {
				if envKey := os.Getenv(util.EnvVarKey); envKey != "" {
					keyPath = envKey
				}
			}
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	// Hide the default completion command
	rootCmd.CompletionOptions.HiddenDefaultCmd = true

	// Define persistent flags available to all subcommands.
	// Individual commands can choose which of these to use.
	rootCmd.PersistentFlags().StringVarP(&serverAddress, "server", "s", "localhost", fmt.Sprintf("Server address (or %s)", util.EnvVarServer))
	rootCmd.PersistentFlags().IntVarP(&port, "port", "p", util.DefaultPort, fmt.Sprintf("Server port (or %s)", util.EnvVarPort))
	rootCmd.PersistentFlags().StringVar(&keyPath, "key", "", fmt.Sprintf("Path to private key (or %s)", util.EnvVarKey))
	rootCmd.PersistentFlags().BoolVar(&enableLogging, "log", false, "enable logging output for debugging.")
}
