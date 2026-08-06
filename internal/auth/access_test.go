package auth

import (
	"errors"
	"strings"
	"testing"
)

func TestParseAccessMode(t *testing.T) {
	for _, value := range []string{string(AccessUser), string(AccessDeveloper)} {
		mode, err := ParseAccessMode(value)
		if err != nil {
			t.Fatalf("ParseAccessMode(%q): %v", value, err)
		}
		if string(mode) != value {
			t.Fatalf("ParseAccessMode(%q) = %q", value, mode)
		}
	}
}

func TestParseAccessModeRejectsUnknownMode(t *testing.T) {
	_, err := ParseAccessMode("production")
	if err == nil || !strings.Contains(err.Error(), "expected user or developer") {
		t.Fatalf("expected a typed access-mode error, got %v", err)
	}
}

func TestDeveloperAccessFailsClosed(t *testing.T) {
	t.Setenv("SPRITZ_DEVELOPER_CONSOLE_URL", "https://console.example.test")

	_, err := StartDeviceLoginForAccess(t.Context(), "", AccessDeveloper)
	if !errors.Is(err, ErrDeveloperAccessNotEnabled) {
		t.Fatalf("expected ErrDeveloperAccessNotEnabled, got %v", err)
	}
	if !strings.Contains(err.Error(), "https://console.example.test") {
		t.Fatalf("expected console guidance, got %v", err)
	}
	if !strings.Contains(err.Error(), "AI agent must not accept terms") {
		t.Fatalf("expected human-approval guidance, got %v", err)
	}
}
