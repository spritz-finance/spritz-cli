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

func TestCompleteDeviceLogin_NoPendingState(t *testing.T) {
	t.Setenv("SPRITZ_CONFIG_DIR", t.TempDir())

	_, err := CompleteDeviceLogin(context.Background(), "", false)
	if err == nil {
		t.Fatal("expected error")
	}
	if err.Error() != "no pending device authorization — run 'spritz auth device start' first" {
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
		_ = json.NewEncoder(w).Encode(UserInfo{Email: "stdin@example.com", FirstName: "Stdin"})
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
		_ = r.Close()
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
