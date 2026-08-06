package auth

import (
	"bytes"
	"errors"
	"strings"
	"testing"
)

func TestRunMCPServerFailsClosedForDeveloperAccess(t *testing.T) {
	t.Setenv("SPRITZ_DEVELOPER_CONSOLE_URL", "https://console.example.test")

	err := RunMCPServer(t.Context(), strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}, AccessDeveloper)
	if !errors.Is(err, ErrDeveloperAccessNotEnabled) {
		t.Fatalf("expected ErrDeveloperAccessNotEnabled, got %v", err)
	}
	if !strings.Contains(err.Error(), "AI agent must not accept terms") {
		t.Fatalf("expected human-approval guidance, got %v", err)
	}
}

func TestRunMCPServerFailsClosedForEndUserAccess(t *testing.T) {
	t.Setenv("SPRITZ_API_KEY", "must-not-be-brokered")

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	err := RunMCPServer(t.Context(), strings.NewReader(""), stdout, stderr, AccessUser)
	if !errors.Is(err, ErrMCPBrokerNotEnabled) {
		t.Fatalf("expected ErrMCPBrokerNotEnabled, got %v", err)
	}
	if !strings.Contains(err.Error(), "dynamically resolved MCP child") {
		t.Fatalf("expected integrity-boundary guidance, got %v", err)
	}
	if strings.Contains(stdout.String()+stderr.String(), "must-not-be-brokered") {
		t.Fatal("credential appeared in broker output")
	}
}
