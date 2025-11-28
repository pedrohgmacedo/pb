package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"pb/clipboard"
	"pb/util"
)

var pasteCmd = &cobra.Command{
	Use:   "paste",
	Short: "Pastes text from the server's clipboard",
	Long:  fmt.Sprintf(`Retrieves text from the remote %s server's clipboard and prints it to standard output.`, util.ProgramName),
	RunE: func(cmd *cobra.Command, args []string) error {
		url := fmt.Sprintf("https://%s:%d%s", serverAddress, port, util.RequestPaste)
		pastedText, err := doHTTPSRequest("GET", url, "")
		
		// If server fails, try local clipboard
		if err != nil {
			if err := clipboard.Init(); err != nil {
				return fmt.Errorf("server unreachable and clipboard unavailable: %w", err)
			}
			data, err := clipboard.Paste()
			if err != nil {
				return fmt.Errorf("server unreachable and failed to read from local clipboard: %w", err)
			}
			fmt.Print(string(data))
			return nil
		}
		
		fmt.Print(pastedText)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pasteCmd)
}
