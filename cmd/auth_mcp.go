package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spritz-finance/spritz-cli/internal/auth"
)

var authMCPCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Report the unavailable MCP credential-broker boundary",
	Long: `The MCP broker is fail-closed in this release. The CLI will not inject an
End User credential into a child resolved through an ambient Node/npm runtime.
That path cannot be treated as a secret boundary against same-user local code.

Until an integrity-verifiable packaged child is available, use only the
read-only MCP server in an operator-controlled disposable Sandbox/test
environment and revoke or rotate the credential after the session. Developer
workspace credentials remain a separate, unavailable grant.`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		accessValue, _ := cmd.Flags().GetString("access")
		if accessValue == "" {
			return fmt.Errorf("--access is required: use user for current End User account tooling")
		}
		access, err := auth.ParseAccessMode(accessValue)
		if err != nil {
			return err
		}
		return auth.RunMCPServer(cmd.Context(), cmd.InOrStdin(), cmd.OutOrStdout(), cmd.ErrOrStderr(), access)
	},
}

func init() {
	authMCPCmd.Flags().String("access", "", "Credential owner; current supported value: user (required)")
}
