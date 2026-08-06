package cmd

import (
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
child command, so an agent cannot use it to print the credential.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		return auth.RunMCPServer(cmd.Context(), os.Stdin, os.Stdout, os.Stderr)
	},
}
