package gemini

import (
	"context"
	"testing"
	"time"
)

func TestLogAnalysis(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	tests := []struct {
		name     string
		logs     string
		wantErr  bool
		validate func(t *testing.T, analysis string)
	}{
		{
			name: "connection errors",
			logs: `
ERROR Connection timeout
ERROR Connection refused
INFO Reconnected
ERROR Connection lost
INFO Connected successfully`,
			validate: func(t *testing.T, analysis string) {
				assertContains(t, analysis, "connection", "Should mention connection issues")
				assertContains(t, analysis, "pattern", "Should identify patterns")
				assertContains(t, analysis, "recommend", "Should include recommendations")
			},
		},
		{
			name: "cache performance issues",
			logs: `
WARN High cache miss rate detected
INFO Cache hit rate: 45%
WARN Memory usage above threshold
INFO Started cache cleanup
INFO Cache hit rate improved to 75%`,
			validate: func(t *testing.T, analysis string) {
				assertContains(t, analysis, "cache", "Should mention cache")
				assertContains(t, analysis, "performance", "Should discuss performance")
				assertContains(t, analysis, "improve", "Should mention improvements")
			},
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis, err := client.AnalyzeLogs(ctx, tt.logs)
			if (err != nil) != tt.wantErr {
				t.Errorf("AnalyzeLogs() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.validate != nil {
				tt.validate(t, analysis)
			}
		})
	}
}

func setupTestClient(t *testing.T) *Client {
	cfg := Config{
		APIKey:     getTestAPIKey(t),
		ModelName:  "gemini-1.5-flash",
		Timeout:    15 * time.Second,
		MaxRetries: 3,
	}

	client, err := NewClient(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to create test client: %v", err)
	}
	return client
}
