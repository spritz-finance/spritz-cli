package auth

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/spritz-finance/spritz-cli/internal/config"
)

// DeviceState is persisted between device auth start and complete.
type DeviceState struct {
	DeviceCode              string `json:"deviceCode"`
	UserCode                string `json:"userCode"`
	VerificationURI         string `json:"verificationUri"`
	VerificationURIComplete string `json:"verificationUriComplete"`
	Interval                int    `json:"interval"`
	CreatedAt               string `json:"createdAt"`
	ExpiresAt               string `json:"expiresAt"`
}

type PendingDeviceState struct {
	Path  string
	State *DeviceState
}

var errNoDeviceState = errors.New("no pending device authorization")

func DeviceStateDir() string {
	return filepath.Join(config.Dir(), "device")
}

func NewDeviceStatePath() (string, error) {
	suffix := make([]byte, 4)
	if _, err := rand.Read(suffix); err != nil {
		return "", fmt.Errorf("generate device state suffix: %w", err)
	}
	name := fmt.Sprintf("device-%s-%s.json", time.Now().UTC().Format("20060102T150405Z"), hex.EncodeToString(suffix))
	return filepath.Join(DeviceStateDir(), name), nil
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

// LoadDeviceState reads a pending device auth state from disk.
func LoadDeviceState(path string) (*DeviceState, error) {
	if path == "" {
		return nil, fmt.Errorf("device state file is required")
	}
	state, err := readDeviceState(path, true)
	if errors.Is(err, errNoDeviceState) {
		return nil, nil
	}
	return state, err
}

func readDeviceState(path string, clearInvalid bool) (*DeviceState, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, errNoDeviceState
		}
		return nil, err
	}

	var state DeviceState
	if err := json.Unmarshal(data, &state); err != nil {
		if clearInvalid {
			ClearDeviceState(path)
		}
		return nil, fmt.Errorf("corrupt device auth state at %q — cleared", path)
	}

	expires, err := time.Parse(time.RFC3339, state.ExpiresAt)
	if err != nil {
		if clearInvalid {
			ClearDeviceState(path)
		}
		return nil, fmt.Errorf("corrupt device auth state at %q — cleared", path)
	}
	if time.Now().After(expires) {
		if clearInvalid {
			ClearDeviceState(path)
		}
		return nil, fmt.Errorf("device code expired for %q — cleared", path)
	}

	if state.CreatedAt == "" {
		state.CreatedAt = time.Now().UTC().Format(time.RFC3339)
	}

	return &state, nil
}

func ListPendingDeviceStates() ([]PendingDeviceState, error) {
	entries, err := os.ReadDir(DeviceStateDir())
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	pending := make([]PendingDeviceState, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".json") {
			continue
		}

		path := filepath.Join(DeviceStateDir(), entry.Name())
		state, err := readDeviceState(path, true)
		if err != nil {
			if errors.Is(err, errNoDeviceState) || strings.Contains(err.Error(), "— cleared") {
				continue
			}
			return nil, err
		}
		pending = append(pending, PendingDeviceState{Path: path, State: state})
	}

	sort.Slice(pending, func(i, j int) bool {
		left, leftErr := time.Parse(time.RFC3339, pending[i].State.CreatedAt)
		right, rightErr := time.Parse(time.RFC3339, pending[j].State.CreatedAt)
		if leftErr != nil || rightErr != nil {
			return pending[i].Path < pending[j].Path
		}
		return left.After(right)
	})

	return pending, nil
}

func ResolveDeviceStatePath(path string) (string, error) {
	if path != "" {
		return path, nil
	}

	pending, err := ListPendingDeviceStates()
	if err != nil {
		return "", err
	}

	switch len(pending) {
	case 0:
		return "", fmt.Errorf("no pending device authorization — run 'spritz auth device start' first")
	case 1:
		return pending[0].Path, nil
	default:
		paths := make([]string, len(pending))
		for i, candidate := range pending {
			paths[i] = candidate.Path
		}
		return "", fmt.Errorf("multiple pending device authorizations found — rerun with --device-state-file and choose one of: %s", strings.Join(paths, ", "))
	}
}

// ClearDeviceState removes any pending device auth state.
func ClearDeviceState(path string) {
	if path == "" {
		return
	}
	_ = os.Remove(path)
}
