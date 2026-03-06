package auth

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/pkg/browser"
	"github.com/spritz-finance/spritz-cli/internal/config"
)

const deviceFlowTimeout = 10 * time.Minute

// minPollInterval is the floor for the polling interval.
// slowDownBackoff is the amount added on slow_down responses (RFC 8628 §3.5).
// Both are vars to allow test overrides.
var (
	minPollInterval = 5 * time.Second
	slowDownBackoff = 5 * time.Second
)

type deviceAuthResponse struct {
	DeviceCode              string `json:"deviceCode"`
	UserCode                string `json:"userCode"`
	VerificationURI         string `json:"verificationUri"`
	VerificationURIComplete string `json:"verificationUriComplete"`
	ExpiresIn               int    `json:"expiresIn"`
	Interval                int    `json:"interval"`
}

type DeviceTokenResponse struct {
	APIKey      string   `json:"apiKey"`
	KeyID       string   `json:"keyId"`
	Permissions []string `json:"permissions"`
	ExpiresAt   *string  `json:"expiresAt"`
	KeyName     *string  `json:"keyName"`
}

type deviceErrorResponse struct {
	Error string `json:"error"`
}

// DeviceAuth runs the device authorization flow (RFC 8628-style).
// It requests a device code, opens the browser for user approval,
// and polls until the user approves or the code expires.
// Returns the full token response so callers can store metadata.
func DeviceAuth(ctx context.Context) (*DeviceTokenResponse, error) {
	baseURL := config.APIURL()

	// Step 1: Request device code
	auth, err := requestDeviceCode(ctx, baseURL)
	if err != nil {
		return nil, err
	}

	interval := time.Duration(auth.Interval) * time.Second
	if interval < minPollInterval {
		interval = minPollInterval
	}

	timeout := time.Duration(auth.ExpiresIn) * time.Second
	if timeout <= 0 || timeout > deviceFlowTimeout {
		timeout = deviceFlowTimeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Always show the code and URL — helps SSH, tmux, headless, wrong browser profile
	fmt.Fprintf(os.Stderr, "Your code: %s\n", auth.UserCode)
	fmt.Fprintf(os.Stderr, "Verification URL: %s\n", auth.VerificationURIComplete)
	if err := browser.OpenURL(auth.VerificationURIComplete); err != nil {
		fmt.Fprintln(os.Stderr, "Could not open browser automatically.")
	}
	fmt.Fprintln(os.Stderr, "Waiting for approval...")

	// Step 4: Poll for token
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("authentication timed out")
		case <-ticker.C:
			token, err := pollDeviceToken(ctx, baseURL, auth.DeviceCode)
			if err == errAuthPending {
				continue
			}
			if err == errSlowDown {
				interval += slowDownBackoff
				ticker.Reset(interval)
				continue
			}
			if err != nil {
				return nil, err
			}
			return token, nil
		}
	}
}

var (
	errAuthPending = fmt.Errorf("authorization_pending")
	errSlowDown    = fmt.Errorf("slow_down")
)

var httpClient = &http.Client{Timeout: 15 * time.Second}

func requestDeviceCode(ctx context.Context, baseURL string) (*deviceAuthResponse, error) {
	body, _ := json.Marshal(map[string]string{"client_id": "spritz-cli"})
	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/device/authorize", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to request device code: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device authorization failed (HTTP %d)", resp.StatusCode)
	}

	var auth deviceAuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&auth); err != nil {
		return nil, fmt.Errorf("failed to decode device auth response: %w", err)
	}
	if auth.DeviceCode == "" || auth.UserCode == "" {
		return nil, fmt.Errorf("invalid device auth response: missing codes")
	}
	return &auth, nil
}

func pollDeviceToken(ctx context.Context, baseURL, deviceCode string) (*DeviceTokenResponse, error) {
	body, _ := json.Marshal(map[string]string{
		"device_code": deviceCode,
		"grant_type":  "urn:ietf:params:oauth:grant-type:device_code",
	})
	req, err := http.NewRequestWithContext(ctx, "POST", baseURL+"/device/token", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to poll for token: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusOK {
		var token DeviceTokenResponse
		if err := json.NewDecoder(resp.Body).Decode(&token); err != nil {
			return nil, fmt.Errorf("failed to decode token response: %w", err)
		}
		if token.APIKey == "" {
			return nil, fmt.Errorf("empty API key in token response")
		}
		return &token, nil
	}

	var errResp deviceErrorResponse
	if err := json.NewDecoder(resp.Body).Decode(&errResp); err != nil {
		return nil, fmt.Errorf("device token request failed (HTTP %d)", resp.StatusCode)
	}

	switch errResp.Error {
	case "authorization_pending":
		return nil, errAuthPending
	case "slow_down":
		return nil, errSlowDown
	case "expired_token":
		return nil, fmt.Errorf("device code expired — please try again")
	default:
		return nil, fmt.Errorf("device token error: %s", errResp.Error)
	}
}
