package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spritz-finance/spritz-cli/internal/auth"
)

var authCmd = &cobra.Command{
	Use:   "auth",
	Short: "Manage Spritz authentication",
	Long: `Manage Spritz authentication for humans, agents, and CI.

Use 'spritz auth login' for interactive login, 'spritz auth device' for
headless agent flows, and 'spritz auth status' to inspect the active
credential source.`,
}

func init() {
	authCmd.AddCommand(authLoginCmd)
	authCmd.AddCommand(authLogoutCmd)
	authCmd.AddCommand(authStatusCmd)
	authCmd.AddCommand(authDeviceCmd)
	authDeviceCmd.AddCommand(authDeviceStartCmd)
	authDeviceCmd.AddCommand(authDeviceCompleteCmd)
}

func writeJSON(value any) error {
	return json.NewEncoder(os.Stdout).Encode(value)
}

func printStoredCredentialResult(result *auth.LoginResult) {
	switch result.Mode {
	case "skipped":
		fmt.Fprintln(os.Stderr, "Login cancelled. Stored credentials unchanged.")
	case "stored_credentials":
		fmt.Fprintf(os.Stderr, "Logged in as %s. Key stored in %s.\n", result.Email, result.Storage)
		if result.EnvVarActive {
			fmt.Fprintln(os.Stderr, "Warning: SPRITZ_API_KEY is set and will continue to override stored credentials.")
		}
	}
}
