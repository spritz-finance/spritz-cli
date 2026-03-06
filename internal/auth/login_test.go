package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
)

func TestLogin_DeviceStartRequiresStateFile(t *testing.T) {
	_, err := Login(context.Background(), LoginOptions{DeviceStart: true})
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "--device-state-file is required with --device-start" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLogin_DeviceCompleteRequiresStateFile(t *testing.T) {
	_, err := Login(context.Background(), LoginOptions{DeviceComplete: true})
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "--device-state-file is required with --device-complete" {
		t.Fatalf("unexpected error: %v", err)
	}
}

func TestLogin_ReadsAPIKeyFromStdin(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/users/me" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if got := r.Header.Get("Authorization"); got != "Bearer ak_valid" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode(UserInfo{Email: "stdin@example.com", FirstName: "Stdin"})
	}))
	defer server.Close()

	t.Setenv("SPRITZ_API_URL", server.URL)
	t.Setenv("SPRITZ_CONFIG_DIR", t.TempDir())

	oldStdin := os.Stdin
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("pipe: %v", err)
	}
	if _, err := w.WriteString("  ak_valid\n"); err != nil {
		t.Fatalf("write stdin: %v", err)
	}
	if err := w.Close(); err != nil {
		t.Fatalf("close stdin writer: %v", err)
	}
	os.Stdin = r
	defer func() {
		os.Stdin = oldStdin
		r.Close()
	}()

	result, err := Login(context.Background(), LoginOptions{AllowFileStorage: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result == nil {
		t.Fatal("expected result")
	}
	if result.Mode != "stored_credentials" {
		t.Fatalf("expected stored_credentials, got %q", result.Mode)
	}
	if result.Email != "stdin@example.com" {
		t.Fatalf("expected stdin@example.com, got %q", result.Email)
	}
	if strings.TrimSpace(result.Storage) == "" {
		t.Fatal("expected storage method")
	}
}
