package config

import (
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/viper"
)

type AccountAgentEndpoint struct {
	Environment string
	BaseURL     string
}

var accountAgentEndpoints = map[string]AccountAgentEndpoint{
	"https://platform.spritz.finance": {
		Environment: "production",
		BaseURL:     "https://platform.spritz.finance",
	},
	"https://sandbox.spritz.finance": {
		Environment: "sandbox",
		BaseURL:     "https://sandbox.spritz.finance",
	},
}

func Dir() string {
	if d := os.Getenv("SPRITZ_CONFIG_DIR"); d != "" {
		return d
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "spritz")
}

func Init() {
	dir := Dir()
	os.MkdirAll(dir, 0700)

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(dir)

	viper.SetDefault("output", "csv")
	viper.SetDefault("api_url", "https://platform.spritz.finance")
	viper.SetDefault("developer_console_url", "https://console.spritz.finance")

	viper.SetEnvPrefix("SPRITZ")
	viper.AutomaticEnv()

	viper.ReadInConfig() // ignore error — file may not exist yet
}

func Set(key, value string) error {
	viper.Set(key, value)
	dir := Dir()
	os.MkdirAll(dir, 0700)
	return viper.WriteConfigAs(filepath.Join(dir, "config.yaml"))
}

func Get(key string) string {
	return viper.GetString(key)
}

func APIURL() string {
	if u := os.Getenv("SPRITZ_API_URL"); u != "" {
		return u
	}
	return viper.GetString("api_url")
}

func DeveloperConsoleURL() string {
	if u := os.Getenv("SPRITZ_DEVELOPER_CONSOLE_URL"); u != "" {
		return u
	}
	if u := viper.GetString("developer_console_url"); u != "" {
		return u
	}
	return "https://console.spritz.finance"
}

// ResolveAccountAgentEndpoint converts the configured API URL into the exact
// official endpoint and environment that the End User account-agent broker may
// pass to its fixed MCP child. Arbitrary URLs are deliberately rejected so a
// changed environment variable cannot redirect a keychain credential.
func ResolveAccountAgentEndpoint(rawURL string) (AccountAgentEndpoint, error) {
	trimmed := strings.TrimRight(strings.TrimSpace(rawURL), "/")
	parsed, err := url.Parse(trimmed)
	if err != nil {
		return AccountAgentEndpoint{}, fmt.Errorf("invalid Spritz API URL: %w", err)
	}
	if parsed.Scheme != "https" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" || (parsed.Path != "" && parsed.Path != "/") {
		return AccountAgentEndpoint{}, fmt.Errorf("Spritz account-agent broker requires an official HTTPS API origin")
	}
	endpoint, ok := accountAgentEndpoints[trimmed]
	if !ok {
		return AccountAgentEndpoint{}, fmt.Errorf(
			"Spritz account-agent broker does not allow API origin %q; expected https://platform.spritz.finance or https://sandbox.spritz.finance",
			rawURL,
		)
	}
	return endpoint, nil
}
