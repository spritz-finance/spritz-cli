package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spritz-finance/spritz-cli/internal/auth"
)

var logoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove stored Spritz credentials and revoke the API key",
	RunE: func(cmd *cobra.Command, args []string) error {
		if os.Getenv("SPRITZ_API_KEY") != "" {
			fmt.Fprintln(os.Stderr, "Warning: SPRITZ_API_KEY environment variable is set.")
			fmt.Fprintln(os.Stderr, "Commands will continue to authenticate via the env var.")
			fmt.Fprintln(os.Stderr, "Unset it to fully log out.")
		}

		if !auth.HasStoredCredentials() {
			fmt.Println("No stored credentials found.")
			return nil
		}

		// Attempt server-side revocation before deleting local credentials
		keyID := auth.LoadKeyMetadata()
		if keyID != "" {
			apiKey, err := auth.GetAPIKey()
			if err == nil {
				if err := auth.RevokeAPIKey(apiKey, keyID); err != nil {
					fmt.Fprintf(os.Stderr, "Warning: could not revoke key server-side: %v\n", err)
					fmt.Fprintln(os.Stderr, "You can revoke it manually at https://app.spritz.finance/settings/api-keys")
				} else {
					fmt.Fprintln(os.Stderr, "API key revoked on server.")
				}
			}
		}

		if err := auth.DeleteAPIKey(); err != nil {
			return fmt.Errorf("failed to remove credentials: %w", err)
		}

		fmt.Println("Stored credentials removed.")
		return nil
	},
}
