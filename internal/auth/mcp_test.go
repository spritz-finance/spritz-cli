package auth

import (
	"bytes"
	"context"
	"errors"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/spritz-finance/spritz-cli/internal/config"
)

func TestRunMCPServerFailsClosedForDeveloperAccess(t *testing.T) {
	t.Setenv("SPRITZ_DEVELOPER_CONSOLE_URL", "https://console.example.test")

	err := RunMCPServer(t.Context(), strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}, AccessDeveloper)
	if !errors.Is(err, ErrDeveloperAccessNotEnabled) {
		t.Fatalf("expected ErrDeveloperAccessNotEnabled, got %v", err)
	}
	if !strings.Contains(err.Error(), "AI agent must not accept terms") {
		t.Fatalf("expected human-approval guidance, got %v", err)
	}
}

func TestRunMCPServerBrokersValidatedEndUserCredential(t *testing.T) {
	const apiKey = "sk_live_test-secret"
	var gotArgs, gotEnv []string
	var validatedKey string

	deps := mcpDependencies{
		getAPIKey: func() (string, StorageMethod, error) {
			return apiKey, StorageKeychain, nil
		},
		validateAPIKey: func(_ context.Context, key string) (*UserInfo, error) {
			validatedKey = key
			return &UserInfo{Email: "owner@example.com"}, nil
		},
		apiURL: func() string { return "https://platform.spritz.finance" },
		environ: func() []string {
			return []string{
				"PATH=/usr/bin:/bin",
				"HOME=/home/tester",
				"SPRITZ_API_KEY=stale",
				"SPRITZ_API_BASE_URL=https://evil.example",
				"SPRITZ_CREDENTIAL_ACCESS=developer",
				"NODE_OPTIONS=--require=/tmp/key-printer.js",
				"NPM_CONFIG_REGISTRY=https://evil.example",
				"AWS_SECRET_ACCESS_KEY=ambient-cloud-secret",
				"GITHUB_TOKEN=ambient-github-secret",
			}
		},
		runChildProcess: func(_ context.Context, args, env []string, _ io.Reader, _, _ io.Writer) error {
			gotArgs = slices.Clone(args)
			gotEnv = slices.Clone(env)
			return nil
		},
	}

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	if err := runMCPServer(t.Context(), strings.NewReader(""), stdout, stderr, AccessUser, deps); err != nil {
		t.Fatalf("runMCPServer: %v", err)
	}
	if validatedKey != apiKey {
		t.Fatalf("validated key %q, want brokered key", validatedKey)
	}
	if strings.Contains(strings.Join(gotArgs, " "), apiKey) {
		t.Fatal("credential appeared in child argv")
	}
	if !slices.Equal(gotArgs, []string{"--yes", mcpPackage}) {
		t.Fatalf("unexpected fixed child args: %#v", gotArgs)
	}

	assertEnvValue(t, gotEnv, "SPRITZ_API_KEY", apiKey)
	assertEnvValue(t, gotEnv, "SPRITZ_CREDENTIAL_BROKER", "spritz-cli")
	assertEnvValue(t, gotEnv, "SPRITZ_CREDENTIAL_ACCESS", "user")
	assertEnvValue(t, gotEnv, "SPRITZ_CREDENTIAL_ENVIRONMENT", "production")
	assertEnvValue(t, gotEnv, "SPRITZ_CREDENTIAL_API_BASE_URL", "https://platform.spritz.finance")
	assertEnvValue(t, gotEnv, "NPM_CONFIG_REGISTRY", "https://registry.npmjs.org")
	assertEnvMissing(t, gotEnv, "SPRITZ_API_BASE_URL")
	assertEnvMissing(t, gotEnv, "NODE_OPTIONS")
	assertEnvMissing(t, gotEnv, "AWS_SECRET_ACCESS_KEY")
	assertEnvMissing(t, gotEnv, "GITHUB_TOKEN")
	if strings.Contains(stdout.String()+stderr.String(), apiKey) {
		t.Fatal("credential appeared in broker output")
	}
}

func TestRunMCPServerRejectsUnvalidatedCredential(t *testing.T) {
	calledChild := false
	deps := mcpDependencies{
		getAPIKey: func() (string, StorageMethod, error) {
			return "int_not-a-user-key", StorageEnv, nil
		},
		validateAPIKey: func(context.Context, string) (*UserInfo, error) {
			return nil, errors.New("invalid API key")
		},
		apiURL:  func() string { return "https://platform.spritz.finance" },
		environ: func() []string { return nil },
		runChildProcess: func(context.Context, []string, []string, io.Reader, io.Writer, io.Writer) error {
			calledChild = true
			return nil
		},
	}

	err := runMCPServer(t.Context(), strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}, AccessUser, deps)
	if err == nil || !strings.Contains(err.Error(), "not a valid End User account credential") {
		t.Fatalf("expected End User validation error, got %v", err)
	}
	if calledChild {
		t.Fatal("child process started with an invalid credential")
	}
}

func TestRunMCPServerRejectsNonOfficialEndpointBeforeReadingKey(t *testing.T) {
	readKey := false
	deps := mcpDependencies{
		getAPIKey: func() (string, StorageMethod, error) {
			readKey = true
			return "secret", StorageEnv, nil
		},
		validateAPIKey: func(context.Context, string) (*UserInfo, error) {
			return &UserInfo{}, nil
		},
		apiURL:  func() string { return "https://evil.example" },
		environ: func() []string { return nil },
		runChildProcess: func(context.Context, []string, []string, io.Reader, io.Writer, io.Writer) error {
			return nil
		},
	}

	err := runMCPServer(t.Context(), strings.NewReader(""), &bytes.Buffer{}, &bytes.Buffer{}, AccessUser, deps)
	if err == nil || !strings.Contains(err.Error(), "does not allow API origin") {
		t.Fatalf("expected endpoint allowlist error, got %v", err)
	}
	if readKey {
		t.Fatal("credential was read before endpoint validation")
	}
}

func TestAccountAgentChildEnvironmentUsesSandboxAttestation(t *testing.T) {
	env := accountAgentChildEnvironment(
		[]string{"PATH=/usr/bin"},
		"sk_test_secret",
		StorageFile,
		config.AccountAgentEndpoint{Environment: "sandbox", BaseURL: "https://sandbox.spritz.finance"},
	)
	assertEnvValue(t, env, "SPRITZ_CREDENTIAL_ENVIRONMENT", "sandbox")
	assertEnvValue(t, env, "SPRITZ_CREDENTIAL_API_BASE_URL", "https://sandbox.spritz.finance")
}

func assertEnvValue(t *testing.T, env []string, name, want string) {
	t.Helper()
	prefix := name + "="
	for _, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			if got := strings.TrimPrefix(entry, prefix); got != want {
				t.Fatalf("%s=%q, want %q", name, got, want)
			}
			return
		}
	}
	t.Fatalf("%s missing from environment", name)
}

func assertEnvMissing(t *testing.T, env []string, name string) {
	t.Helper()
	prefix := name + "="
	for _, entry := range env {
		if strings.HasPrefix(entry, prefix) {
			t.Fatalf("%s unexpectedly survived environment sanitization", name)
		}
	}
}
