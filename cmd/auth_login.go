package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spritz-finance/spritz-cli/internal/auth"
)

var authLoginCmd = &cobra.Command{
	Use:   "login",
	Short: "Log in to your Spritz account",
	Long: `Log in to your Spritz account.

This command is the human-friendly login path. For headless agents, prefer:

  spritz auth device start
  spritz auth device complete`,
	RunE: func(cmd *cobra.Command, args []string) error {
		apiKey, _ := cmd.Flags().GetString("api-key")
		allowFile, _ := cmd.Flags().GetBool("allow-file-storage")
		jsonOut, _ := cmd.Flags().GetBool("json")

		result, err := auth.Login(cmd.Context(), auth.LoginOptions{
			AllowFileStorage: allowFile,
			APIKey:           apiKey,
		})
		if err != nil {
			return err
		}

		if jsonOut {
			return writeJSON(result)
		}

		printStoredCredentialResult(result)
		return nil
	},
}

func init() {
	authLoginCmd.Flags().String("api-key", "", "API key to store (non-interactive, works without TTY)")
	authLoginCmd.Flags().Bool("json", false, "Print structured JSON output for automation")
	authLoginCmd.Flags().Bool("allow-file-storage", false, "Allow falling back to encrypted file if system keychain is unavailable")
}
