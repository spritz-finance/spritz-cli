package auth

import (
	"path/filepath"
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
		ExpiresAt:               time.Now().Add(time.Minute).Format(time.RFC3339),
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
		ExpiresAt:  time.Now().Add(-time.Minute).Format(time.RFC3339),
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
	if _, err := LoadDeviceState(path); err != nil {
		t.Fatalf("expected cleared file to read as nil, got: %v", err)
	}
}
