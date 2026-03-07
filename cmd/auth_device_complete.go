package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spritz-finance/spritz-cli/internal/auth"
)

var authDeviceCompleteCmd = &cobra.Command{
	Use:   "complete",
	Short: "Complete a headless device authorization session",
	Long: `Complete a headless device authorization session.

If --device-state-file is omitted, spritz uses the only pending device session.
If multiple pending sessions exist, spritz requires an explicit state file.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		allowFile, _ := cmd.Flags().GetBool("allow-file-storage")
		deviceStateFile, _ := cmd.Flags().GetString("device-state-file")
		result, err := auth.CompleteDeviceLogin(cmd.Context(), deviceStateFile, allowFile)
		if err != nil {
			return err
		}
		return writeJSON(result)
	},
}

func init() {
	authDeviceCompleteCmd.Flags().Bool("allow-file-storage", false, "Allow falling back to encrypted file if system keychain is unavailable")
	authDeviceCompleteCmd.Flags().String("device-state-file", "", "Path to the device authorization state file; defaults to the only pending session")
}
