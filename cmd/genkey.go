package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"path/filepath"
	"pb/util"
)

var genkeyCmd = &cobra.Command{
	Use:   "key-gen",
	Short: fmt.Sprintf("Generates a new %s-specific SSH key", util.ProgramName),
	Long:  fmt.Sprintf(`Generates a new ed25519 SSH key pair specifically for %s in ~/.config/%s/`, util.ProgramName, util.ProgramName),
	RunE: func(cmd *cobra.Command, args []string) error {
		home, _ := os.UserHomeDir()
		keyDir := filepath.Join(home, ".config", util.ProgramName)
		keyPath := filepath.Join(keyDir, "id_ed25519")

		if _, err := os.Stat(keyPath); err == nil {
			return fmt.Errorf("%s key already exists at %s", util.ProgramName, keyPath)
		}

		if err := util.GenerateSSHKeys(keyDir); err != nil {
			return fmt.Errorf("failed to generate keys: %w", err)
		}

		fmt.Printf("New ed25519 key pair generated in %s/\n", keyDir)
		fmt.Println("You can now add this key to a server's authorized_keys file by running:")
		fmt.Printf("  %s key-add \"$(cat %s.pub)\" --server <server_address>\n", util.ProgramName, keyPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(genkeyCmd)
}
