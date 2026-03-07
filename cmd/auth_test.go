package cmd

import (
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/denisbrodbeck/machineid"
	"github.com/spritz-finance/spritz-cli/internal/format"
)

func captureStdoutStderr(t *testing.T, fn func() error) (string, string, error) {
	t.Helper()

	oldStdout := os.Stdout
	oldStderr := os.Stderr
	stdoutR, stdoutW, err := os.Pipe()
	if err != nil {
		t.Fatalf("stdout pipe: %v", err)
	}
	stderrR, stderrW, err := os.Pipe()
	if err != nil {
		t.Fatalf("stderr pipe: %v", err)
	}

	os.Stdout = stdoutW
	os.Stderr = stderrW

	runErr := fn()

	_ = stdoutW.Close()
	_ = stderrW.Close()
	os.Stdout = oldStdout
	os.Stderr = oldStderr

	stdout, err := io.ReadAll(stdoutR)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	stderr, err := io.ReadAll(stderrR)
	if err != nil {
		t.Fatalf("read stderr: %v", err)
	}

	_ = stdoutR.Close()
	_ = stderrR.Close()

	return string(stdout), string(stderr), runErr
}

func assertJSONGolden(t *testing.T, actual string, goldenPath string) {
	t.Helper()

	expected, err := os.ReadFile(goldenPath)
	if err != nil {
		t.Fatalf("read golden file: %v", err)
	}

	normalizedActual := normalizeJSON(t, actual)
	if strings.TrimSpace(normalizedActual) != strings.TrimSpace(string(expected)) {
		t.Fatalf("json mismatch\nactual:\n%s\n\nexpected:\n%s", normalizedActual, expected)
	}
}

func normalizeJSON(t *testing.T, raw string) string {
	t.Helper()

	var value any
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		t.Fatalf("actual output was not json: %v\n%s", err, raw)
	}

	normalized, err := json.MarshalIndent(sortJSONValue(value), "", "  ")
	if err != nil {
		t.Fatalf("normalize actual json: %v", err)
	}

	return string(normalized)
}

func sortJSONValue(v any) any {
	switch x := v.(type) {
	case map[string]any:
		keys := make([]string, 0, len(x))
		for k := range x {
			keys = append(keys, k)
		}
		slices.Sort(keys)
		out := make(map[string]any, len(x))
		for _, k := range keys {
			out[k] = sortJSONValue(x[k])
		}
		return out
	case []any:
		out := make([]any, len(x))
		for i, item := range x {
			out[i] = sortJSONValue(item)
		}
		return out
	default:
		return v
	}
}

func resetAuthLoginFlags(t *testing.T) {
	t.Helper()
	_ = authLoginCmd.Flags().Set("api-key", "")
	_ = authLoginCmd.Flags().Set("json", "false")
	_ = authLoginCmd.Flags().Set("allow-file-storage", "false")
}

func resetAuthLogoutFlags(t *testing.T) {
	t.Helper()
	_ = authLogoutCmd.Flags().Set("json", "false")
}

func resetAuthDeviceFlags(t *testing.T) {
	t.Helper()
	_ = authDeviceStartCmd.Flags().Set("device-state-file", "")
	_ = authDeviceCompleteCmd.Flags().Set("device-state-file", "")
	_ = authDeviceCompleteCmd.Flags().Set("allow-file-storage", "false")
}

func writeEncryptedCredentialFile(t *testing.T, configDir, apiKey string) {
	t.Helper()

	id, err := machineid.ProtectedID("spritz")
	if err != nil {
		t.Fatalf("machine id: %v", err)
	}
	h := sha256.Sum256([]byte(id + "spritz-credential-v1"))
	block, err := aes.NewCipher(h[:])
	if err != nil {
		t.Fatalf("cipher: %v", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		t.Fatalf("gcm: %v", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		t.Fatalf("nonce: %v", err)
	}
	encrypted := gcm.Seal(nonce, nonce, []byte(apiKey), nil)

	if err := os.MkdirAll(configDir, 0700); err != nil {
		t.Fatalf("mkdir config: %v", err)
	}
	path := filepath.Join(configDir, "credentials")
	if err := os.WriteFile(path, encrypted, 0600); err != nil {
		t.Fatalf("write credentials: %v", err)
	}
}

func TestAuthDeviceStartWritesJSONToStdout(t *testing.T) {
	defer resetAuthDeviceFlags(t)
	configDir := t.TempDir()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/device/authorize" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"deviceCode":              "dc_secret",
			"userCode":                "ABCD1234",
			"verificationUri":         "https://app.spritz.finance/device",
			"verificationUriComplete": "https://app.spritz.finance/device?code=ABCD1234",
			"expiresIn":               600,
			"interval":                5,
		})
	}))
	defer server.Close()

	t.Setenv("SPRITZ_API_URL", server.URL)
	t.Setenv("SPRITZ_CONFIG_DIR", configDir)
	authDeviceStartCmd.SetContext(context.Background())

	stdout, stderr, err := captureStdoutStderr(t, func() error {
		return authDeviceStartCmd.RunE(authDeviceStartCmd, nil)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}

	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("stdout was not json: %v\n%s", err, stdout)
	}
	stateFile, _ := got["deviceStateFile"].(string)
	got["deviceStateFile"] = "__STATE_FILE__"
	got["expiresAt"] = "__EXPIRES_AT__"
	normalized, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal normalized device start json: %v", err)
	}
	assertJSONGolden(t, string(normalized), "testdata/login_device_start.golden.json")

	if stateFile == "" || !strings.HasPrefix(stateFile, configDir) {
		t.Fatalf("expected generated device state file in config dir, got %q", stateFile)
	}
}

func TestAuthLoginJSONDoesNotPrintHumanSuccessToStderr(t *testing.T) {
	defer resetAuthLoginFlags(t)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/users/me" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Header.Get("Authorization") != "Bearer ak_valid" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"email": "cmd@example.com", "firstName": "Cmd"})
	}))
	defer server.Close()

	t.Setenv("SPRITZ_API_URL", server.URL)
	t.Setenv("SPRITZ_CONFIG_DIR", t.TempDir())
	authLoginCmd.SetContext(context.Background())
	_ = authLoginCmd.Flags().Set("api-key", "ak_valid")
	_ = authLoginCmd.Flags().Set("json", "true")
	_ = authLoginCmd.Flags().Set("allow-file-storage", "true")

	stdout, stderr, err := captureStdoutStderr(t, func() error {
		return authLoginCmd.RunE(authLoginCmd, nil)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.Contains(stderr, "Logged in as") {
		t.Fatalf("expected no human success text on stderr, got %q", stderr)
	}

	var got map[string]any
	if err := json.Unmarshal([]byte(stdout), &got); err != nil {
		t.Fatalf("stdout was not json: %v\n%s", err, stdout)
	}
	got["storage"] = "__STORAGE__"
	normalized, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("marshal normalized login json: %v", err)
	}
	assertJSONGolden(t, string(normalized), "testdata/login_json_success.golden.json")
}

func TestAuthLogoutJSONNoStoredCredentials(t *testing.T) {
	defer resetAuthLogoutFlags(t)
	t.Setenv("SPRITZ_CONFIG_DIR", t.TempDir())
	_ = authLogoutCmd.Flags().Set("json", "true")

	stdout, stderr, err := captureStdoutStderr(t, func() error {
		return authLogoutCmd.RunE(authLogoutCmd, nil)
	})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("expected empty stderr, got %q", stderr)
	}
	assertJSONGolden(t, stdout, "testdata/logout_json_no_credentials.golden.json")
}

func TestAuthStatusJSONShowsEnvSource(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/users/me" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Header.Get("Authorization") != "Bearer ak_env" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"email": "env@example.com", "firstName": "Env"})
	}))
	defer server.Close()

	t.Setenv("SPRITZ_API_URL", server.URL)
	t.Setenv("SPRITZ_API_KEY", "ak_env")
	t.Setenv("SPRITZ_CONFIG_DIR", t.TempDir())

	buf := &bytes.Buffer{}
	format.Global = format.New("json", false, buf)

	if err := authStatusCmd.RunE(authStatusCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertJSONGolden(t, buf.String(), "testdata/whoami_env_source.golden.json")
}

func TestAuthStatusJSONShowsEnvOverrideWhenStoredCredentialsExist(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/users/me" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if r.Header.Get("Authorization") != "Bearer ak_env" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		_ = json.NewEncoder(w).Encode(map[string]string{"email": "env@example.com", "firstName": "Env"})
	}))
	defer server.Close()

	configDir := t.TempDir()
	writeEncryptedCredentialFile(t, configDir, "ak_stored")

	t.Setenv("SPRITZ_API_URL", server.URL)
	t.Setenv("SPRITZ_API_KEY", "ak_env")
	t.Setenv("SPRITZ_CONFIG_DIR", configDir)

	buf := &bytes.Buffer{}
	format.Global = format.New("json", false, buf)

	if err := authStatusCmd.RunE(authStatusCmd, nil); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	assertJSONGolden(t, buf.String(), "testdata/whoami_env_override.golden.json")
}
