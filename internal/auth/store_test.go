package auth

import "testing"

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
