package auth

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

func TestSaveAndLoadDeviceState(t *testing.T) {
	path := filepath.Join(t.TempDir(), "device-state.json")
	state := &DeviceState{
		DeviceCode:              "dc_secret",
		UserCode:                "ABCD1234",
		VerificationURI:         "https://app.spritz.finance/device",
		VerificationURIComplete: "https://app.spritz.finance/device?code=ABCD1234",
		Interval:                5,
		CreatedAt:               time.Now().UTC().Format(time.RFC3339),
		ExpiresAt:               time.Now().Add(time.Minute).UTC().Format(time.RFC3339),
	}

	if err := SaveDeviceState(path, state); err != nil {
		t.Fatalf("unexpected save error: %v", err)
	}

	loaded, err := LoadDeviceState(path)
	if err != nil {
		t.Fatalf("unexpected load error: %v", err)
	}
	if loaded == nil {
		t.Fatal("expected device state")
	}
	if loaded.DeviceCode != state.DeviceCode {
		t.Fatalf("expected device code %q, got %q", state.DeviceCode, loaded.DeviceCode)
	}
}

func TestLoadDeviceStateExpiredClearsFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "device-state.json")
	state := &DeviceState{
		DeviceCode: "dc_secret",
		ExpiresAt:  time.Now().Add(-time.Minute).UTC().Format(time.RFC3339),
	}

	if err := SaveDeviceState(path, state); err != nil {
		t.Fatalf("unexpected save error: %v", err)
	}

	loaded, err := LoadDeviceState(path)
	if err == nil {
		t.Fatal("expected expiry error")
	}
	if loaded != nil {
		t.Fatal("expected nil state for expired device flow")
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected expired state file to be removed, got err=%v", err)
	}
}

func TestResolveDeviceStatePathUsesOnlyPendingSession(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("SPRITZ_CONFIG_DIR", configDir)

	path, err := NewDeviceStatePath()
	if err != nil {
		t.Fatalf("new device state path: %v", err)
	}
	if err := SaveDeviceState(path, &DeviceState{
		DeviceCode: "dc_secret",
		CreatedAt:  time.Now().UTC().Format(time.RFC3339),
		ExpiresAt:  time.Now().Add(time.Minute).UTC().Format(time.RFC3339),
	}); err != nil {
		t.Fatalf("save device state: %v", err)
	}

	resolved, err := ResolveDeviceStatePath("")
	if err != nil {
		t.Fatalf("resolve device state path: %v", err)
	}
	if resolved != path {
		t.Fatalf("expected %q, got %q", path, resolved)
	}
}

func TestResolveDeviceStatePathErrorsWhenMultiplePendingSessionsExist(t *testing.T) {
	configDir := t.TempDir()
	t.Setenv("SPRITZ_CONFIG_DIR", configDir)

	for i := 0; i < 2; i++ {
		path, err := NewDeviceStatePath()
		if err != nil {
			t.Fatalf("new device state path: %v", err)
		}
		if err := SaveDeviceState(path, &DeviceState{
			DeviceCode: "dc_secret",
			CreatedAt:  time.Now().Add(time.Duration(i) * time.Second).UTC().Format(time.RFC3339),
			ExpiresAt:  time.Now().Add(time.Minute).UTC().Format(time.RFC3339),
		}); err != nil {
			t.Fatalf("save device state: %v", err)
		}
	}

	_, err := ResolveDeviceStatePath("")
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "multiple pending device authorizations found") {
		t.Fatalf("unexpected error: %v", err)
	}
	if !strings.Contains(err.Error(), "--device-state-file") {
		t.Fatalf("expected explicit guidance in error, got: %v", err)
	}
}
