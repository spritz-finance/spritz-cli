package auth

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// LoginOptions controls login behavior.
type LoginOptions struct {
	AllowFileStorage bool
	APIKey           string // If set, skip interactive prompts (works without TTY)
	DeviceStart      bool   // Just initiate device flow and exit
	DeviceComplete   bool   // Complete a previously started device flow
}

type LoginResult struct {
	Mode                    string `json:"mode"`
	Email                   string `json:"email,omitempty"`
	FirstName               string `json:"firstName,omitempty"`
	Storage                 string `json:"storage,omitempty"`
	EnvVarActive            bool   `json:"envVarActive"`
	UserCode                string `json:"userCode,omitempty"`
	VerificationURI         string `json:"verificationUri,omitempty"`
	VerificationURIComplete string `json:"verificationUriComplete,omitempty"`
	ExpiresAt               string `json:"expiresAt,omitempty"`
}

// Login runs the login flow. Interactive prompts require a TTY;
// pass APIKey in opts to authenticate non-interactively (e.g. piped envs).
func Login(ctx context.Context, opts LoginOptions) (*LoginResult, error) {
	// Two-step device flow: start
	if opts.DeviceStart {
		state, err := DeviceStart(ctx)
		if err != nil {
			return nil, err
		}
		return &LoginResult{
			Mode:                    "device_start",
			EnvVarActive:            os.Getenv("SPRITZ_API_KEY") != "",
			UserCode:                state.UserCode,
			VerificationURI:         state.VerificationURI,
			VerificationURIComplete: state.VerificationURIComplete,
			ExpiresAt:               state.ExpiresAt,
		}, nil
	}

	// Two-step device flow: complete
	if opts.DeviceComplete {
		token, err := DeviceComplete(ctx)
		if err != nil {
			return nil, err
		}
		return storeValidatedKey(token.APIKey, token.KeyID, opts.AllowFileStorage)
	}

	// Non-interactive: API key provided directly
	if opts.APIKey != "" {
		return loginWithKey(opts.APIKey, opts.AllowFileStorage)
	}

	// API key from stdin pipe (non-TTY, no --api-key flag)
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		key, err := io.ReadAll(os.Stdin)
		if err != nil {
			return nil, fmt.Errorf("failed to read from stdin: %w", err)
		}
		k := strings.TrimSpace(string(key))
		if k == "" {
			return nil, fmt.Errorf("login requires an interactive terminal, stdin, --api-key, or --device-start/--device-complete")
		}
		return loginWithKey(k, opts.AllowFileStorage)
	}

	// Interactive flow
	if HasStoredCredentials() {
		fmt.Print("You are already logged in. Overwrite? [y/N] ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(answer)) != "y" {
			return &LoginResult{Mode: "skipped", EnvVarActive: os.Getenv("SPRITZ_API_KEY") != ""}, nil
		}
	}

	fmt.Println("How would you like to authenticate?")
	fmt.Println("  1. Browser (recommended)")
	fmt.Println("  2. Paste API key")
	fmt.Print("Choice [1]: ")

	reader := bufio.NewReader(os.Stdin)
	choice, _ := reader.ReadString('\n')
	choice = strings.TrimSpace(choice)

	var apiKey string
	var keyID string

	switch choice {
	case "", "1":
		token, err := DeviceAuth(ctx)
		if err != nil {
			return nil, err
		}
		apiKey = token.APIKey
		keyID = token.KeyID
	case "2":
		fmt.Print("API key: ")
		keyBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println()
		if err != nil {
			return nil, fmt.Errorf("failed to read API key: %w", err)
		}
		apiKey = strings.TrimSpace(string(keyBytes))
	default:
		return nil, fmt.Errorf("invalid choice: %s", choice)
	}

	if apiKey == "" {
		return nil, fmt.Errorf("no API key provided")
	}

	return storeValidatedKey(apiKey, keyID, opts.AllowFileStorage)
}

func loginWithKey(apiKey string, allowFile bool) (*LoginResult, error) {
	return storeValidatedKey(apiKey, "", allowFile)
}

func storeValidatedKey(apiKey, keyID string, allowFile bool) (*LoginResult, error) {
	user, err := ValidateAPIKey(apiKey)
	if err != nil {
		return nil, err
	}

	method, err := StoreAPIKey(apiKey, allowFile)
	if err != nil {
		return nil, fmt.Errorf("failed to store credentials: %w", err)
	}

	StoreKeyMetadata(keyID)

	return &LoginResult{
		Mode:         "stored_credentials",
		Email:        user.Email,
		FirstName:    user.FirstName,
		Storage:      method.String(),
		EnvVarActive: os.Getenv("SPRITZ_API_KEY") != "",
	}, nil
}
