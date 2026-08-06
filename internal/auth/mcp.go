package auth

import (
	"context"
	"errors"
	"fmt"
	"io"
)

var ErrMCPBrokerNotEnabled = errors.New("MCP credential broker is not enabled")

// RunMCPServer remains fail-closed until the CLI can launch a bundled,
// integrity-verifiable child without ambient executables or mutable caches.
func RunMCPServer(_ context.Context, _ io.Reader, _, _ io.Writer, access AccessMode) error {
	switch access {
	case AccessDeveloper:
		return developerAccessNotEnabled()
	case AccessUser:
		return fmt.Errorf(
			"%w: this CLI will not inject an End User credential into a dynamically resolved MCP child; use only the read-only MCP server in an operator-controlled disposable Sandbox/test environment, then revoke or rotate that credential",
			ErrMCPBrokerNotEnabled,
		)
	default:
		return fmt.Errorf("unsupported access mode %q", access)
	}
}
