package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"sync/atomic"
	"testing"
	"time"
)

func setFastPolling(t *testing.T) {
	origInterval := minPollInterval
	origBackoff := slowDownBackoff
	minPollInterval = 100 * time.Millisecond
	slowDownBackoff = 100 * time.Millisecond
	t.Cleanup(func() {
		minPollInterval = origInterval
		slowDownBackoff = origBackoff
	})
}

func TestDeviceAuth_Success(t *testing.T) {
	setFastPolling(t)
	var pollCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/device/authorize":
			json.NewEncoder(w).Encode(deviceAuthResponse{
				DeviceCode:              "dc_secret",
				UserCode:                "ABCD1234",
				VerificationURI:         "https://app.spritz.finance/device",
				VerificationURIComplete: "https://app.spritz.finance/device?code=ABCD1234",
				ExpiresIn:               600,
				Interval:                1, // fast for testing
			})
		case "/v1/device/token":
			if pollCount.Add(1) < 3 {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(problemDetailResponse{Detail: "authorization_pending"})
				return
			}
			json.NewEncoder(w).Encode(DeviceTokenResponse{
				APIKey:      "ak_test_key",
				KeyID:       "key_123",
				Permissions: []string{"bank-accounts:read"},
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	t.Setenv("SPRITZ_API_URL", server.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	token, err := DeviceAuth(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.APIKey != "ak_test_key" {
		t.Fatalf("expected ak_test_key, got %s", token.APIKey)
	}
	if token.KeyID != "key_123" {
		t.Fatalf("expected key_123, got %s", token.KeyID)
	}
	if pollCount.Load() < 3 {
		t.Fatalf("expected at least 3 polls, got %d", pollCount.Load())
	}
}

func TestDeviceAuth_Expired(t *testing.T) {
	setFastPolling(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/device/authorize":
			json.NewEncoder(w).Encode(deviceAuthResponse{
				DeviceCode:              "dc_secret",
				UserCode:                "ABCD1234",
				VerificationURI:         "https://app.spritz.finance/device",
				VerificationURIComplete: "https://app.spritz.finance/device?code=ABCD1234",
				ExpiresIn:               600,
				Interval:                1,
			})
		case "/v1/device/token":
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(problemDetailResponse{Detail: "expired_token"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	t.Setenv("SPRITZ_API_URL", server.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := DeviceAuth(ctx)
	if err == nil {
		t.Fatal("expected error for expired token")
	}
	if err.Error() != "device code expired — please try again" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeviceAuth_SlowDown(t *testing.T) {
	setFastPolling(t)
	var pollCount atomic.Int32

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/device/authorize":
			json.NewEncoder(w).Encode(deviceAuthResponse{
				DeviceCode:              "dc_secret",
				UserCode:                "TEST5678",
				VerificationURI:         "https://app.spritz.finance/device",
				VerificationURIComplete: "https://app.spritz.finance/device?code=TEST5678",
				ExpiresIn:               600,
				Interval:                1,
			})
		case "/v1/device/token":
			n := pollCount.Add(1)
			if n == 1 {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(problemDetailResponse{Detail: "slow_down"})
				return
			}
			json.NewEncoder(w).Encode(DeviceTokenResponse{
				APIKey: "ak_after_slowdown",
				KeyID:  "key_456",
			})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	t.Setenv("SPRITZ_API_URL", server.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	token, err := DeviceAuth(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.APIKey != "ak_after_slowdown" {
		t.Fatalf("expected ak_after_slowdown, got %s", token.APIKey)
	}
}

func TestDeviceAuth_Timeout(t *testing.T) {
	setFastPolling(t)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/device/authorize":
			json.NewEncoder(w).Encode(deviceAuthResponse{
				DeviceCode:              "dc_secret",
				UserCode:                "TIMEOUT1",
				VerificationURI:         "https://app.spritz.finance/device",
				VerificationURIComplete: "https://app.spritz.finance/device?code=TIMEOUT1",
				ExpiresIn:               2, // 2 seconds
				Interval:                1,
			})
		case "/v1/device/token":
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(problemDetailResponse{Detail: "authorization_pending"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	t.Setenv("SPRITZ_API_URL", server.URL)

	ctx := context.Background()
	_, err := DeviceAuth(ctx)
	if err == nil {
		t.Fatal("expected timeout error")
	}
	if err.Error() != "authentication timed out" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeviceAuth_AuthorizeFails(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	t.Setenv("SPRITZ_API_URL", server.URL)

	_, err := DeviceAuth(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}

func TestDeviceComplete_MissingStateFile(t *testing.T) {
	path := t.TempDir() + "/missing-device-state.json"

	_, err := DeviceComplete(context.Background(), path)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "no pending device authorization at \""+path+"\" — run 'spritz auth device start --device-state-file "+path+"' first" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestDeviceComplete_SuccessClearsStateFile(t *testing.T) {
	setFastPolling(t)
	path := t.TempDir() + "/device-state.json"
	if err := SaveDeviceState(path, &DeviceState{
		DeviceCode:              "dc_secret",
		UserCode:                "ABCD1234",
		VerificationURI:         "https://app.spritz.finance/device",
		VerificationURIComplete: "https://app.spritz.finance/device?code=ABCD1234",
		Interval:                1,
		ExpiresAt:               time.Now().Add(time.Minute).Format(time.RFC3339),
	}); err != nil {
		t.Fatalf("save state: %v", err)
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/v1/device/token":
			json.NewEncoder(w).Encode(DeviceTokenResponse{APIKey: "ak_test_key", KeyID: "key_123"})
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	t.Setenv("SPRITZ_API_URL", server.URL)

	token, err := DeviceComplete(context.Background(), path)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if token.APIKey != "ak_test_key" {
		t.Fatalf("expected ak_test_key, got %q", token.APIKey)
	}
	if _, err := os.Stat(path); !os.IsNotExist(err) {
		t.Fatalf("expected state file to be removed, got err=%v", err)
	}
}

func TestRequestDeviceCode_MissingCodes(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		json.NewEncoder(w).Encode(deviceAuthResponse{
			DeviceCode: "",
			UserCode:   "",
		})
	}))
	defer server.Close()

	_, err := requestDeviceCode(context.Background(), server.URL)
	if err == nil {
		t.Fatal("expected error for missing codes")
	}
}
