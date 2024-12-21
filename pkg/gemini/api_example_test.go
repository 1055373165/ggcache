package gemini

import (
	"context"
	"testing"
)

func TestAPIExampleGeneration(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	tests := []struct {
		name     string
		useCase  string
		wantErr  bool
		validate func(t *testing.T, example string)
	}{
		{
			name:    "rate limiter example",
			useCase: "Implementing a rate limiter using the cache with a sliding window",
			validate: func(t *testing.T, example string) {
				assertContains(t, example, "rate limit", "Should mention rate limiting")
				assertContains(t, example, "window", "Should mention time window")
				assertContains(t, example, "error", "Should include error handling")
			},
		},
		{
			name:    "session storage example",
			useCase: "Using cache for temporary session storage with automatic expiration",
			validate: func(t *testing.T, example string) {
				assertContains(t, example, "session", "Should mention session")
				assertContains(t, example, "TTL", "Should mention TTL/expiration")
				assertContains(t, example, "error", "Should include error handling")
			},
		},
		{
			name:    "cache warming example",
			useCase: "Implementing cache warming/preloading for frequently accessed data",
			validate: func(t *testing.T, example string) {
				assertContains(t, example, "warm", "Should mention warming")
				assertContains(t, example, "load", "Should mention loading")
				assertContains(t, example, "error", "Should include error handling")
			},
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			example, err := client.GenerateAPIExample(ctx, tt.useCase)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateAPIExample() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.validate != nil {
				tt.validate(t, example)
			}
		})
	}
}
