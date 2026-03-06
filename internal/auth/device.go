package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// DeviceState is persisted between --device-start and --device-complete.
type DeviceState struct {
	DeviceCode              string `json:"deviceCode"`
	UserCode                string `json:"userCode"`
	VerificationURI         string `json:"verificationUri"`
	VerificationURIComplete string `json:"verificationUriComplete"`
	Interval                int    `json:"interval"`
	ExpiresAt               string `json:"expiresAt"`
}
// SaveDeviceState writes pending device auth state to disk.
func SaveDeviceState(path string, state *DeviceState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	if path == "" {
		return fmt.Errorf("device state file is required")
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// LoadDeviceState reads pending device auth state from disk.
// Returns nil if no pending state exists.
func LoadDeviceState(path string) (*DeviceState, error) {
	if path == "" {
		return nil, fmt.Errorf("device state file is required")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	var state DeviceState
	if err := json.Unmarshal(data, &state); err != nil {
		return nil, err
	}
	expires, err := time.Parse(time.RFC3339, state.ExpiresAt)
	if err != nil {
		ClearDeviceState(path)
		return nil, fmt.Errorf("corrupt device auth state — cleared")
	}
	if time.Now().After(expires) {
		ClearDeviceState(path)
		return nil, fmt.Errorf("device code expired — run 'spritz login --device-start' again")
	}
	return &state, nil
}

// ClearDeviceState removes any pending device auth state.
func ClearDeviceState(path string) {
	if path == "" {
		return
	}
	os.Remove(path)
}
