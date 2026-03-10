package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/spritz-finance/spritz-cli/internal/config"
)

type UserInfo struct {
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
}

// ValidateAPIKey checks the key against GET /v1/users/me and returns user info.
// Uses a self-contained HTTP call to avoid circular dependency with the api package.
func ValidateAPIKey(ctx context.Context, apiKey string) (*UserInfo, error) {
	baseURL := strings.TrimRight(config.APIURL(), "/")
	req, err := http.NewRequestWithContext(ctx, "GET", baseURL+"/v1/users/me", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to reach Spritz API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized || resp.StatusCode == http.StatusForbidden {
		return nil, fmt.Errorf("invalid API key")
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected response from Spritz API (HTTP %d)", resp.StatusCode)
	}

	var user UserInfo
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user info: %w", err)
	}
	return &user, nil
}
