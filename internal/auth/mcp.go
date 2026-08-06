package auth

import (
	"context"
	"io"
)

// RunMCPServer is the typed local entry point for Developer Access agent
// tooling. It remains fail-closed until the platform can issue and introspect a
// workspace-scoped device grant; the existing End User device flow must never
// be substituted.
func RunMCPServer(_ context.Context, _ io.Reader, _, _ io.Writer) error {
	return developerAccessNotEnabled()
}
