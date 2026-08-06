package cmd

import (
	"github.com/spf13/cobra"
	"github.com/spritz-finance/spritz-cli/internal/auth"
)

var authDeviceStartCmd = &cobra.Command{
	Use:   "start",
	Short: "Start a headless device authorization session",
	Long: `Start a headless device authorization session.

If --device-state-file is omitted, spritz creates a unique pending state file
automatically and returns its path in the JSON response.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		deviceStateFile, _ := cmd.Flags().GetString("device-state-file")
		accessValue, _ := cmd.Flags().GetString("access")
		access, err := auth.ParseAccessMode(accessValue)
		if err != nil {
			return err
		}
		result, err := auth.StartDeviceLoginForAccess(cmd.Context(), deviceStateFile, access)
		if err != nil {
			return err
		}
		return writeJSON(result)
	},
}

func init() {
	authDeviceStartCmd.Flags().String("device-state-file", "", "Path to the device authorization state file; defaults to an auto-generated unique path")
	authDeviceStartCmd.Flags().String("access", string(auth.AccessUser), "Grant owner: user or developer (Developer Access fails closed until platform support is enabled)")
}
