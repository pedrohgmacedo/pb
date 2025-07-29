package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"io"
	"os"
	"path/filepath"
	"pb/util"
	"strings"
)

var addKeyCmd = &cobra.Command{
	Use:   "key-add [public key string]",
	Short: "Adds a public key to the server's authorized_keys",
	Long:  fmt.Sprintf(`Appends a given public key to the ~/.config/%s/authorized_keys file. The key can be provided as an argument or via standard input.`, util.ProgramName),
	Args:  cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		var keyToAdd string
		if len(args) == 1 {
			keyToAdd = args[0]
		} else {
			bytes, err := io.ReadAll(os.Stdin)
			if err != nil {
				return fmt.Errorf("failed to read public key from stdin: %w", err)
			}
			keyToAdd = string(bytes)
		}

		keyToAdd = strings.TrimSpace(keyToAdd)
		if _, _, _, _, err := ssh.ParseAuthorizedKey([]byte(keyToAdd)); err != nil {
			return fmt.Errorf("invalid public key provided: %w", err)
		}

		home, _ := os.UserHomeDir()
		configDir := filepath.Join(home, ".config", util.ProgramName)
		if err := os.MkdirAll(configDir, 0700); err != nil {
			return fmt.Errorf("could not create config directory: %w", err)
		}

		authKeysPath := filepath.Join(configDir, "authorized_keys")
		f, err := os.OpenFile(authKeysPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0600)
		if err != nil {
			return fmt.Errorf("could not open authorized_keys file: %w", err)
		}
		defer f.Close()

		if _, err := f.WriteString(keyToAdd + "\n"); err != nil {
			return fmt.Errorf("failed to write to authorized_keys file: %w", err)
		}

		fmt.Printf("Successfully added key to %s\n", authKeysPath)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(addKeyCmd)
}
