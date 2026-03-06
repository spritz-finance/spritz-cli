package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/spritz-finance/spritz-cli/internal/config"
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

func deviceStatePath() string {
	return filepath.Join(config.Dir(), "device-auth-pending.json")
}

// SaveDeviceState writes pending device auth state to disk.
func SaveDeviceState(state *DeviceState) error {
	data, err := json.Marshal(state)
	if err != nil {
		return err
	}
	path := deviceStatePath()
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

// LoadDeviceState reads pending device auth state from disk.
// Returns nil if no pending state exists.
func LoadDeviceState() (*DeviceState, error) {
	data, err := os.ReadFile(deviceStatePath())
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
		ClearDeviceState()
		return nil, fmt.Errorf("corrupt device auth state — cleared")
	}
	if time.Now().After(expires) {
		ClearDeviceState()
		return nil, fmt.Errorf("device code expired — run 'spritz login --device-start' again")
	}
	return &state, nil
}

// ClearDeviceState removes any pending device auth state.
func ClearDeviceState() {
	os.Remove(deviceStatePath())
}
