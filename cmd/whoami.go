package cmd

import (
	"strconv"

	"github.com/spf13/cobra"
	"github.com/spritz-finance/spritz-cli/internal/auth"
	"github.com/spritz-finance/spritz-cli/internal/format"
)

var whoamiCmd = &cobra.Command{
	Use:   "whoami",
	Short: "Show the active authenticated user",
	Long: `Show the active authenticated user and where credentials are coming from.

This is useful for debugging automation, especially when SPRITZ_API_KEY is
set and overrides stored credentials.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		apiKey, source, err := auth.GetAPIKeyWithSource()
		if err != nil {
			return err
		}

		user, err := auth.ValidateAPIKey(apiKey)
		if err != nil {
			return err
		}

		headers := []string{"email", "firstName", "source", "envOverride", "storedCredentials"}
		rows := [][]string{{
			user.Email,
			user.FirstName,
			source.String(),
			strconv.FormatBool(source == auth.StorageEnv && auth.HasStoredCredentials()),
			strconv.FormatBool(auth.HasStoredCredentials()),
		}}

		return format.Global.Write(headers, rows)
	},
}
