package auth

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

func TestValidateAPIKey_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/users/me" {
			t.Fatalf("unexpected path: %s", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer ak_valid" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		json.NewEncoder(w).Encode(UserInfo{Email: "test@example.com", FirstName: "Test"})
	}))
	defer server.Close()

	os.Setenv("SPRITZ_API_URL", server.URL)
	defer os.Unsetenv("SPRITZ_API_URL")

	user, err := ValidateAPIKey(context.Background(), "ak_valid")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if user.Email != "test@example.com" {
		t.Fatalf("expected test@example.com, got %s", user.Email)
	}
}

func TestValidateAPIKey_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
	}))
	defer server.Close()

	os.Setenv("SPRITZ_API_URL", server.URL)
	defer os.Unsetenv("SPRITZ_API_URL")

	_, err := ValidateAPIKey(context.Background(), "ak_bad")
	if err == nil {
		t.Fatal("expected error for invalid key")
	}
	if err.Error() != "invalid API key" {
		t.Fatalf("expected 'invalid API key', got: %v", err)
	}
}

func TestValidateAPIKey_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	os.Setenv("SPRITZ_API_URL", server.URL)
	defer os.Unsetenv("SPRITZ_API_URL")

	_, err := ValidateAPIKey(context.Background(), "ak_test")
	if err == nil {
		t.Fatal("expected error for server error")
	}
}
