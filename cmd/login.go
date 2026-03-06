package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spritz-finance/spritz-cli/internal/auth"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the Spritz API",
	Long:  "Authenticate interactively via browser or by pasting an API key.",
	RunE: func(cmd *cobra.Command, args []string) error {
		allowFile, _ := cmd.Flags().GetBool("allow-file-storage")
		return auth.Login(cmd.Context(), auth.LoginOptions{
			AllowFileStorage: allowFile,
		})
	},
}

func init() {
	loginCmd.Flags().Bool("allow-file-storage", false,
		"Allow falling back to encrypted file if system keychain is unavailable")
}
