package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
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
		case "/device/authorize":
			json.NewEncoder(w).Encode(deviceAuthResponse{
				DeviceCode:              "dc_secret",
				UserCode:                "ABCD1234",
				VerificationURI:         "https://app.spritz.finance/device",
				VerificationURIComplete: "https://app.spritz.finance/device?code=ABCD1234",
				ExpiresIn:               600,
				Interval:                1, // fast for testing
			})
		case "/device/token":
			if pollCount.Add(1) < 3 {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(deviceErrorResponse{Error: "authorization_pending"})
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
		case "/device/authorize":
			json.NewEncoder(w).Encode(deviceAuthResponse{
				DeviceCode:              "dc_secret",
				UserCode:                "ABCD1234",
				VerificationURI:         "https://app.spritz.finance/device",
				VerificationURIComplete: "https://app.spritz.finance/device?code=ABCD1234",
				ExpiresIn:               600,
				Interval:                1,
			})
		case "/device/token":
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(deviceErrorResponse{Error: "expired_token"})
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
		case "/device/authorize":
			json.NewEncoder(w).Encode(deviceAuthResponse{
				DeviceCode:              "dc_secret",
				UserCode:                "TEST5678",
				VerificationURI:         "https://app.spritz.finance/device",
				VerificationURIComplete: "https://app.spritz.finance/device?code=TEST5678",
				ExpiresIn:               600,
				Interval:                1,
			})
		case "/device/token":
			n := pollCount.Add(1)
			if n == 1 {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(deviceErrorResponse{Error: "slow_down"})
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
		case "/device/authorize":
			json.NewEncoder(w).Encode(deviceAuthResponse{
				DeviceCode:              "dc_secret",
				UserCode:                "TIMEOUT1",
				VerificationURI:         "https://app.spritz.finance/device",
				VerificationURIComplete: "https://app.spritz.finance/device?code=TIMEOUT1",
				ExpiresIn:               2, // 2 seconds
				Interval:                1,
			})
		case "/device/token":
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(deviceErrorResponse{Error: "authorization_pending"})
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
