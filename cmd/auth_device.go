package cmd

import "github.com/spf13/cobra"

var authDeviceCmd = &cobra.Command{
	Use:   "device",
	Short: "Run headless device authorization flows",
	Long: `Run headless device authorization flows for agents and automation.

Start creates a pending device session and returns JSON to stdout. Complete
finishes the pending session and stores validated credentials locally.`,
}
