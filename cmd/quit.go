package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"pb/util"
)

var quitCmd = &cobra.Command{
	Use:   "quit",
	Short: "Quits server",
	Long:  fmt.Sprintf(`Tell the remote %s server to quit.`, util.ProgramName),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := fmt.Sprintf("https://%s:%d%s", serverAddress, port, util.RequestQuit)
		_, err := doHTTPSRequest("POST", url, "")

		if err == nil {
			return err
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(quitCmd)
}
