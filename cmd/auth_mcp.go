package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/spritz-finance/spritz-cli/internal/auth"
)

var authMCPCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start the Spritz MCP server through the credential broker",
	Long: `Start the reviewed Spritz MCP server with the active credential injected
only into that child process's environment. The credential is not placed on
argv or written to a plaintext file. The broker does not accept an arbitrary
child command, so an agent cannot use it to print the credential.

Use --access user for the current End User account-agent tool surface. Developer
workspace-agent access is a separate future grant and fails closed.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		accessValue, _ := cmd.Flags().GetString("access")
		if accessValue == "" {
			return fmt.Errorf("--access is required: choose user or developer")
		}
		access, err := auth.ParseAccessMode(accessValue)
		if err != nil {
			return err
		}
		return auth.RunMCPServer(cmd.Context(), os.Stdin, os.Stdout, os.Stderr, access)
	},
}

func init() {
	authMCPCmd.Flags().String("access", "", "Credential owner: user or developer (required)")
}
