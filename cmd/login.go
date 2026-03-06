package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spritz-finance/spritz-cli/internal/auth"
)

var loginCmd = &cobra.Command{
	Use:   "login",
	Short: "Authenticate with the Spritz API",
	Long: `Authenticate with the Spritz API. Credentials are stored in the
system keychain (or encrypted file with --allow-file-storage).

With no flags, runs interactively: opens a browser for device approval
or prompts to paste an API key. Requires a TTY.

Examples:
  # Interactive (default — requires terminal)
  spritz login

  # Direct API key from stdin (no TTY needed, avoids shell history)
  printf '%s' "$SPRITZ_API_KEY" | spritz login
  op read "op://private/spritz/api key" | spritz login

  # Direct API key flag (works, but may be captured in shell history)
  spritz login --api-key ak_...

  # Two-step device flow (for AI agents and automation)
  spritz login --device-start --device-state-file /tmp/spritz-device.json
  spritz login --device-complete --device-state-file /tmp/spritz-device.json --json

  # Structured success output
  spritz login --json

  # Environment variable (no login needed, always takes precedence)
  export SPRITZ_API_KEY=ak_...`,
	RunE: func(cmd *cobra.Command, args []string) error {
		apiKey, _ := cmd.Flags().GetString("api-key")
		allowFile, _ := cmd.Flags().GetBool("allow-file-storage")
		deviceStart, _ := cmd.Flags().GetBool("device-start")
		deviceComplete, _ := cmd.Flags().GetBool("device-complete")
		deviceStateFile, _ := cmd.Flags().GetString("device-state-file")
		jsonOut, _ := cmd.Flags().GetBool("json")

		result, err := auth.Login(cmd.Context(), auth.LoginOptions{
			AllowFileStorage: allowFile,
			APIKey:           apiKey,
			DeviceStart:      deviceStart,
			DeviceComplete:   deviceComplete,
			DeviceStateFile:  deviceStateFile,
		})
		if err != nil {
			return err
		}

		if result == nil {
			return nil
		}

		if deviceStart || jsonOut {
			return json.NewEncoder(os.Stdout).Encode(result)
		}

		switch result.Mode {
		case "skipped":
			fmt.Fprintln(os.Stderr, "Login cancelled. Stored credentials unchanged.")
		case "stored_credentials":
			fmt.Fprintf(os.Stderr, "Logged in as %s. Key stored in %s.\n", result.Email, result.Storage)
			if result.EnvVarActive {
				fmt.Fprintln(os.Stderr, "Warning: SPRITZ_API_KEY is set and will continue to override stored credentials.")
			}
		}

		return nil
	},
}

func init() {
	loginCmd.Flags().String("api-key", "",
		"API key to store (non-interactive, works without TTY)")
	loginCmd.Flags().Bool("json", false,
		"Print structured JSON output for automation")
	loginCmd.Flags().Bool("allow-file-storage", false,
		"Allow falling back to encrypted file if system keychain is unavailable")
	loginCmd.Flags().Bool("device-start", false,
		"Initiate device authorization and print JSON to stdout")
	loginCmd.Flags().Bool("device-complete", false,
		"Complete a previously started device authorization")
	loginCmd.Flags().String("device-state-file", "",
		"Path to the device authorization state file used by --device-start/--device-complete")
	loginCmd.MarkFlagsMutuallyExclusive("api-key", "device-start", "device-complete")
}
