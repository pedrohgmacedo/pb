package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"os"
	"pb/clipboard"
	"pb/util"
)

var (
	rosebudFlag bool
)

const maxClipboardSize = 200 * 1024 * 1024 // 200MB

var copyCmd = &cobra.Command{
	Use:   "copy [data to copy]",
	Short: "Copies data to the server's clipboard",
	Long:  fmt.Sprintf(`Copies the provided data argument or standard input to the remote %s server's clipboard.`, util.ProgramName),
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var dataToCopy []byte
		if len(args) == 1 {
			dataToCopy = []byte(args[0])
		} else {
			bytes, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read from stdin: %w", err)
			}
			dataToCopy = bytes
		}

		// Check size limit
		if len(dataToCopy) > maxClipboardSize && !rosebudFlag {
			return fmt.Errorf("data too large: %d bytes (max %d bytes, use --rosebud to bypass)", len(dataToCopy), maxClipboardSize)
		}

		url := fmt.Sprintf("https://%s:%d%s", serverAddress, port, util.RequestCopy)
		_, err := doHTTPSRequest("POST", url, string(dataToCopy))
		
		// If server fails, try local clipboard
		if err != nil {
			if err := clipboard.Init(); err != nil {
				return fmt.Errorf("server unreachable and clipboard unavailable: %w", err)
			}
			if err := clipboard.Copy(dataToCopy); err != nil {
				return fmt.Errorf("server unreachable and failed to write to local clipboard: %w", err)
			}
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(copyCmd)
	copyCmd.Flags().BoolVar(&rosebudFlag, "rosebud", false, "bypass clipboard size limit")
}
