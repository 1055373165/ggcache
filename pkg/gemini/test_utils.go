package gemini

import (
	"os"
	"strings"
	"testing"
)

// getTestAPIKey returns the API key for testing
func getTestAPIKey(t *testing.T) string {
	t.Helper()

	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		apiKey = "AIzaSyCba1V_pxgMMIiG9Nxo-C9zKiXUNWGjM1k" // default test key
	}
	return apiKey
}

// assertContains checks if a string contains a substring (case-insensitive)
func assertContains(t *testing.T, got, want, message string) {
	t.Helper()
	if !containsIgnoreCase(got, want) {
		t.Errorf("%s: response should contain '%s', got: %s", message, want, got)
	}
}

// containsIgnoreCase checks if a string contains a substring (case-insensitive)
func containsIgnoreCase(s, substr string) bool {
	s, substr = strings.ToLower(s), strings.ToLower(substr)
	return strings.Contains(s, substr)
}
