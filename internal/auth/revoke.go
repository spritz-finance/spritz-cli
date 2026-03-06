package auth

import (
	"fmt"
	"net/http"
	"time"

	"github.com/spritz-finance/spritz-cli/internal/config"
)

// RevokeAPIKey attempts to revoke the key server-side via DELETE /v1/api-keys/{keyId}.
// Requires the current API key for authentication. Best-effort: returns error but
// callers may choose to proceed with local cleanup regardless.
func RevokeAPIKey(apiKey, keyID string) error {
	if apiKey == "" || keyID == "" {
		return fmt.Errorf("missing credentials for revocation")
	}

	url := config.APIURL() + "/v1/api-keys/" + keyID
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create revocation request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{Timeout: 15 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to reach Spritz API: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil // already revoked
	}
	if resp.StatusCode >= 400 {
		return fmt.Errorf("revocation failed (HTTP %d)", resp.StatusCode)
	}
	return nil
}
