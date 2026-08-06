package auth

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"

	"github.com/spritz-finance/spritz-cli/internal/config"
)

// Pin the reviewed broker target. A mutable dist-tag would let a newly published
// package receive an account credential before this CLI release reviewed it.
const mcpPackage = "@spritz-finance/mcp-server@0.3.2"

var mcpChildEnvironmentAllowlist = map[string]struct{}{
	"APPDATA":         {},
	"COLORTERM":       {},
	"COMSPEC":         {},
	"FORCE_COLOR":     {},
	"HOME":            {},
	"LANG":            {},
	"LC_ALL":          {},
	"LOCALAPPDATA":    {},
	"LOGNAME":         {},
	"NO_COLOR":        {},
	"PATH":            {},
	"PATHEXT":         {},
	"SHELL":           {},
	"SYSTEMROOT":      {},
	"TEMP":            {},
	"TERM":            {},
	"TMP":             {},
	"TMPDIR":          {},
	"USER":            {},
	"USERPROFILE":     {},
	"XDG_CACHE_HOME":  {},
	"XDG_RUNTIME_DIR": {},
}

type mcpDependencies struct {
	getAPIKey       func() (string, StorageMethod, error)
	validateAPIKey  func(context.Context, string) (*UserInfo, error)
	apiURL          func() string
	environ         func() []string
	runChildProcess func(context.Context, []string, []string, io.Reader, io.Writer, io.Writer) error
}

var defaultMCPDependencies = mcpDependencies{
	getAPIKey:      GetAPIKeyWithSource,
	validateAPIKey: ValidateAPIKey,
	apiURL:         config.APIURL,
	environ:        os.Environ,
	runChildProcess: func(ctx context.Context, args, env []string, stdin io.Reader, stdout, stderr io.Writer) error {
		cmd := exec.CommandContext(ctx, "npx", args...)
		cmd.Env = env
		cmd.Stdin = stdin
		cmd.Stdout = stdout
		cmd.Stderr = stderr
		if err := cmd.Run(); err != nil {
			return fmt.Errorf("Spritz MCP server exited: %w", err)
		}
		return nil
	},
}

// RunMCPServer starts the fixed End User account-agent MCP server when access
// is explicitly user-owned. Developer workspace-agent access is a different
// HMAC/scoped grant and remains fail-closed until its platform contract exists.
func RunMCPServer(ctx context.Context, stdin io.Reader, stdout, stderr io.Writer, access AccessMode) error {
	return runMCPServer(ctx, stdin, stdout, stderr, access, defaultMCPDependencies)
}

func runMCPServer(ctx context.Context, stdin io.Reader, stdout, stderr io.Writer, access AccessMode, deps mcpDependencies) error {
	switch access {
	case AccessDeveloper:
		return developerAccessNotEnabled()
	case AccessUser:
		// Continue below.
	default:
		return fmt.Errorf("unsupported access mode %q", access)
	}

	endpoint, err := config.ResolveAccountAgentEndpoint(deps.apiURL())
	if err != nil {
		return err
	}
	apiKey, storage, err := deps.getAPIKey()
	if err != nil {
		return err
	}
	if _, err := deps.validateAPIKey(ctx, apiKey); err != nil {
		return fmt.Errorf("active credential is not a valid End User account credential: %w", err)
	}

	env := accountAgentChildEnvironment(deps.environ(), apiKey, storage, endpoint)
	args := []string{"--yes", mcpPackage}
	return deps.runChildProcess(ctx, args, env, stdin, stdout, stderr)
}

func accountAgentChildEnvironment(parent []string, apiKey string, storage StorageMethod, endpoint config.AccountAgentEndpoint) []string {
	env := make([]string, 0, len(mcpChildEnvironmentAllowlist)+9)
	for _, entry := range parent {
		name, _, _ := strings.Cut(entry, "=")
		upper := strings.ToUpper(name)
		if _, ok := mcpChildEnvironmentAllowlist[upper]; ok || strings.HasPrefix(upper, "LC_") {
			env = append(env, entry)
		}
	}

	return append(env,
		"SPRITZ_API_KEY="+apiKey,
		"SPRITZ_CREDENTIAL_BROKER=spritz-cli",
		"SPRITZ_CREDENTIAL_ACCESS=user",
		"SPRITZ_CREDENTIAL_STORAGE="+storage.String(),
		"SPRITZ_CREDENTIAL_ENVIRONMENT="+endpoint.Environment,
		"SPRITZ_CREDENTIAL_API_BASE_URL="+endpoint.BaseURL,
		"NPM_CONFIG_USERCONFIG="+os.DevNull,
		"NPM_CONFIG_REGISTRY=https://registry.npmjs.org",
		"NPM_CONFIG_IGNORE_SCRIPTS=true",
	)
}
