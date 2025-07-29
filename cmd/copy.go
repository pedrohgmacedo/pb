package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"os"
	"pb/util"
)

var copyCmd = &cobra.Command{
	Use:   "copy [text to copy]",
	Short: "Copies text to the server's clipboard",
	Long:  fmt.Sprintf(`Copies the provided text argument or standard input to the remote %s server's clipboard.`, util.ProgramName),
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var dataToCopy string
		if len(args) == 1 {
			dataToCopy = args[0]
		} else {
			bytes, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read from stdin: %w", err)
			}
			dataToCopy = string(bytes)
		}

		url := fmt.Sprintf("https://%s:%d%s", serverAddress, port, util.RequestCopy)
		_, err := doHTTPSRequest("POST", url, dataToCopy)
		return err
	},
}

func init() {
	rootCmd.AddCommand(copyCmd)
}
