package auth

import (
	"testing"
)

func TestStorageMethod_String(t *testing.T) {
	tests := []struct {
		method StorageMethod
		want   string
	}{
		{StorageEnv, "environment variable"},
		{StorageKeychain, "system keychain"},
		{StorageFile, "encrypted file"},
		{StorageMethod(99), "unknown"},
	}
	for _, tt := range tests {
		if got := tt.method.String(); got != tt.want {
			t.Errorf("StorageMethod(%d).String() = %q, want %q", tt.method, got, tt.want)
		}
	}
}

func TestGetAPIKeyWithSource_EnvWins(t *testing.T) {
	t.Setenv("SPRITZ_API_KEY", "ak_env")

	key, source, err := GetAPIKeyWithSource()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if key != "ak_env" {
		t.Fatalf("expected env key, got %q", key)
	}
	if source != StorageEnv {
		t.Fatalf("expected StorageEnv, got %v", source)
	}
}
