package gemini

import (
	"context"
	"fmt"
	"os"
	"testing"
	"time"
)

func TestNewClient(t *testing.T) {
	// 直接设置 API 密钥
	apiKey := "AIzaSyCba1V_pxgMMIiG9Nxo-C9zKiXUNWGjM1k"
	if envKey := os.Getenv("GEMINI_API_KEY"); envKey != "" {
		apiKey = envKey
	}

	ctx := context.Background()
	cfg := Config{
		APIKey:     apiKey,
		ModelName:  "gemini-1.5-flash",
		Timeout:    15 * time.Second,
		MaxRetries: 3,
	}

	client, err := NewClient(ctx, cfg)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// 使用与官方示例相同的提示词
	resp, err := client.GenerateContent(ctx, "Explain how AI works")
	if err != nil {
		t.Logf("Error details: %v", err)
		t.Fatalf("Failed to generate content: %v", err)
	}

	if resp == "" {
		t.Error("Expected non-empty response")
	}

	// Print the response
	fmt.Printf("Gemini Response:\n%s\n", resp)

	t.Run("generate cache key name", func(t *testing.T) {
		keyName, err := client.GenerateCacheKeyName(ctx, "user session data for user_id abc123")
		if err != nil {
			t.Fatalf("Failed to generate key name: %v", err)
		}
		t.Logf("Generated key name: %s", keyName)
	})

	t.Run("analyze cache pattern", func(t *testing.T) {
		analysis, err := client.AnalyzeCachePattern(ctx, "hit, miss, hit, hit, miss")
		if err != nil {
			t.Fatalf("Failed to analyze pattern: %v", err)
		}
		t.Logf("Pattern analysis: %s", analysis)
	})

	t.Run("enhance error message", func(t *testing.T) {
		enhanced, err := client.EnhanceErrorMessage(ctx, "cache entry not found")
		if err != nil {
			t.Fatalf("Failed to enhance error message: %v", err)
		}
		t.Logf("Enhanced error message: %s", enhanced)
	})

	t.Run("suggest eviction strategy", func(t *testing.T) {
		suggestion, err := client.SuggestEvictionStrategy(ctx, "frequent repeated access to same items")
		if err != nil {
			t.Fatalf("Failed to suggest strategy: %v", err)
		}
		t.Logf("Strategy suggestion: %s", suggestion)
	})
}
