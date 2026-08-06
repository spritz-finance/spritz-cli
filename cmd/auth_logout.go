package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spritz-finance/spritz-cli/internal/auth"
)

type logoutDependencies struct {
	hasStoredCredentials func() bool
	deleteAPIKey         func() error
}

var defaultLogoutDependencies = logoutDependencies{
	hasStoredCredentials: auth.HasStoredCredentials,
	deleteAPIKey:         auth.DeleteAPIKey,
}

var authLogoutCmd = &cobra.Command{
	Use:   "logout",
	Short: "Remove locally stored Spritz credentials",
	Long: `Remove credentials from keychain and encrypted file.

This does not revoke the API key on the server. To revoke it, visit
https://app.spritz.finance/api-keys.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		return runAuthLogout(cmd, defaultLogoutDependencies)
	},
}

func runAuthLogout(cmd *cobra.Command, deps logoutDependencies) error {
	jsonOut, _ := cmd.Flags().GetBool("json")
	envSet := os.Getenv("SPRITZ_API_KEY") != ""
	hadStored := deps.hasStoredCredentials()

	if envSet {
		fmt.Fprintln(os.Stderr, "Warning: SPRITZ_API_KEY environment variable is set.")
		fmt.Fprintln(os.Stderr, "Commands will continue to authenticate via the env var.")
		fmt.Fprintln(os.Stderr, "Unset it to fully log out.")
	}

	if !hadStored {
		if jsonOut {
			return json.NewEncoder(os.Stdout).Encode(map[string]any{
				"removedStoredCredentials": false,
				"envVarActive":             envSet,
				"serverRevocationURL":      "https://app.spritz.finance/api-keys",
			})
		}
		fmt.Println("No stored credentials found.")
		return nil
	}

	if err := deps.deleteAPIKey(); err != nil {
		return fmt.Errorf("failed to remove credentials: %w", err)
	}

	if jsonOut {
		return json.NewEncoder(os.Stdout).Encode(map[string]any{
			"removedStoredCredentials": true,
			"envVarActive":             envSet,
			"serverRevocationURL":      "https://app.spritz.finance/api-keys",
		})
	}

	fmt.Println("Stored credentials removed.")
	fmt.Println("To revoke the API key server-side, visit https://app.spritz.finance/api-keys")
	return nil
}

func init() {
	authLogoutCmd.Flags().Bool("json", false, "Print structured JSON output for automation")
}
