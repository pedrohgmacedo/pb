package commands

import (
	"fmt"
	"github.com/spf13/cobra"
	"golang.org/x/crypto/ssh"
	"pb/util"
)

var pubkeyCmd = &cobra.Command{
	Use:   "key-print",
	Short: "Prints the public key that will be used for authentication",
	Long:  fmt.Sprintf(`Finds the first available private key (checking ~/.config/%s/id_ed25519 first, then common ~/.ssh keys), derives the public key, and prints it in the authorized_keys format.`, util.ProgramName),
	RunE: func(cmd *cobra.Command, args []string) error {
		signer, err := getSigner()
		if err != nil {
			return err
		}

		pubKeyBytes := ssh.MarshalAuthorizedKey(signer.PublicKey())
		fmt.Print(string(pubKeyBytes))
		return nil
	},
}

func init() {
	rootCmd.AddCommand(pubkeyCmd)
}
