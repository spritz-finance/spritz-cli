package config

import "testing"

func TestResolveAccountAgentEndpoint(t *testing.T) {
	tests := []struct {
		url         string
		environment string
		baseURL     string
	}{
		{
			url:         "https://platform.spritz.finance/",
			environment: "production",
			baseURL:     "https://platform.spritz.finance",
		},
		{
			url:         "https://sandbox.spritz.finance",
			environment: "sandbox",
			baseURL:     "https://sandbox.spritz.finance",
		},
	}

	for _, tt := range tests {
		t.Run(tt.environment, func(t *testing.T) {
			endpoint, err := ResolveAccountAgentEndpoint(tt.url)
			if err != nil {
				t.Fatalf("ResolveAccountAgentEndpoint(%q): %v", tt.url, err)
			}
			if endpoint.Environment != tt.environment || endpoint.BaseURL != tt.baseURL {
				t.Fatalf("unexpected endpoint: %#v", endpoint)
			}
		})
	}
}

func TestResolveAccountAgentEndpointRejectsRedirectableOrigins(t *testing.T) {
	for _, rawURL := range []string{
		"http://platform.spritz.finance",
		"https://platform.spritz.finance.evil.example",
		"https://platform.spritz.finance/v1",
		"https://user@platform.spritz.finance",
		"https://platform.spritz.finance?next=https://evil.example",
	} {
		t.Run(rawURL, func(t *testing.T) {
			if _, err := ResolveAccountAgentEndpoint(rawURL); err == nil {
				t.Fatalf("expected %q to be rejected", rawURL)
			}
		})
	}
}
