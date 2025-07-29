package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"net/url"
	"pb/util"
)

var openCmd = &cobra.Command{
	Use:   "open [url]",
	Short: "Opens a URL on the server",
	Long:  `Sends a URL to the remote %s server to be opened in the default browser.`,
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		urlToOpen := args[0]
		if _, err := url.ParseRequestURI(urlToOpen); err != nil {
			return fmt.Errorf("invalid URL provided: %w", err)
		}

		requestURL := fmt.Sprintf("https://%s:%d%s", serverAddress, port, util.RequestOpen)
		_, err := doHTTPSRequest("POST", requestURL, urlToOpen)
		if err == nil {
			fmt.Printf("Successfully requested server to open URL: %s\n", urlToOpen)
		}
		return err
	},
}

func init() {
	rootCmd.AddCommand(openCmd)
}
