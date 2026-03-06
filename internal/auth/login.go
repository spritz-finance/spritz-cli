package auth

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"

	"golang.org/x/term"
)

// LoginOptions controls login behavior.
type LoginOptions struct {
	AllowFileStorage bool
}

// Login runs the interactive login flow. It must be called from a TTY.
func Login(ctx context.Context, opts LoginOptions) error {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return fmt.Errorf("login requires an interactive terminal")
	}

	if envKey := os.Getenv("SPRITZ_API_KEY"); envKey != "" {
		fmt.Fprintln(os.Stderr, "SPRITZ_API_KEY environment variable is set.")
		fmt.Fprintln(os.Stderr, "The env var always takes precedence over stored credentials.")
		fmt.Fprintln(os.Stderr, "Unset it first, then re-run 'spritz login'.")
		return fmt.Errorf("SPRITZ_API_KEY is set — unset it to use stored credentials")
	}

	// Check if already logged in
	if HasStoredCredentials() {
		fmt.Print("You are already logged in. Overwrite? [y/N] ")
		reader := bufio.NewReader(os.Stdin)
		answer, _ := reader.ReadString('\n')
		if strings.TrimSpace(strings.ToLower(answer)) != "y" {
			return nil
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
			return err
		}
		apiKey = token.APIKey
		keyID = token.KeyID
	case "2":
		fmt.Print("API key: ")
		keyBytes, err := term.ReadPassword(int(os.Stdin.Fd()))
		fmt.Println() // newline after hidden input
		if err != nil {
			return fmt.Errorf("failed to read API key: %w", err)
		}
		apiKey = strings.TrimSpace(string(keyBytes))
	default:
		return fmt.Errorf("invalid choice: %s", choice)
	}

	if apiKey == "" {
		return fmt.Errorf("no API key provided")
	}

	user, err := ValidateAPIKey(apiKey)
	if err != nil {
		return err
	}

	method, err := StoreAPIKey(apiKey, opts.AllowFileStorage)
	if err != nil {
		return fmt.Errorf("failed to store credentials: %w", err)
	}

	StoreKeyMetadata(keyID)

	fmt.Printf("Logged in as %s. Key stored in %s.\n", user.Email, method)
	return nil
}
