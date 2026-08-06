package auth

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestRunMCPServerFailsClosedUntilWorkspaceGrantExists(t *testing.T) {
	t.Setenv("SPRITZ_DEVELOPER_CONSOLE_URL", "https://console.example.test")

	err := RunMCPServer(t.Context(), strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{})
	if !errors.Is(err, ErrDeveloperAccessNotEnabled) {
		t.Fatalf("expected ErrDeveloperAccessNotEnabled, got %v", err)
	}
	if !strings.Contains(err.Error(), "AI agent must not accept terms") {
		t.Fatalf("expected human-approval guidance, got %v", err)
	}
}
