package cmd

import (
	"fmt"
	"os"
	"os/exec"

	"github.com/spf13/cobra"
)

const installURL = "https://spritz.finance/install"

var updateCmd = &cobra.Command{
	Use:   "update",
	Short: "Update spritz to the latest version",
	Long: `Download and install the latest version of spritz.

Re-runs the official install script. Requires curl, bash, and cosign.

Examples:
  # Update to latest
  spritz update

  # Update to a specific version
  SPRITZ_VERSION=v1.2.3 spritz update`,
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Fprintln(os.Stderr, "Updating spritz...")
		c := exec.Command("bash", "-c", fmt.Sprintf("curl --proto '=https' --tlsv1.2 --fail --silent --show-error --location %s | bash", installURL))
		c.Stdout = os.Stderr // install output goes to stderr
		c.Stderr = os.Stderr
		if err := c.Run(); err != nil {
			return fmt.Errorf("update failed: %w", err)
		}
		return nil
	},
}
