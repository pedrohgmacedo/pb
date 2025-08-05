package commands

import (
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"pb/server"
	"pb/util"
)

var fallback bool

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts the listener server",
	Long:  fmt.Sprintf(`Starts the listener server. It will use the --port flag if provided, otherwise the %s environment variable, otherwise the default port.`, util.EnvVarPort),
	RunE: func(cmd *cobra.Command, args []string) error {
		// The 'port' variable is populated by the root command's persistent flag and PersistentPreRun logic.

		return server.Serve(context.Background(), port, "", fallback)
	},
}

func init() {
	rootCmd.AddCommand(serverCmd)
	serverCmd.PersistentFlags().BoolVar(&fallback, "fallback", false, "uses the fallback clipboard implementation.")
}
