package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"pb/util"
)

var pasteCmd = &cobra.Command{
	Use:   "paste",
	Short: "Pastes text from the server's clipboard",
	Long:  fmt.Sprintf(`Retrieves text from the remote %s server's clipboard and prints it to standard output.`, util.ProgramName),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := fmt.Sprintf("https://%s:%d%s", serverAddress, port, util.RequestPaste)
		pastedText, err := doHTTPSRequest("GET", url, "")
		if err != nil {
			return err
		}
		fmt.Print(pastedText)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pasteCmd)
}
