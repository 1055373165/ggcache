package gemini

import (
	"context"
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

	t.Run("generate cache key name", func(t *testing.T) {
		_, err := client.GenerateCacheKeyName(ctx, "user session data for user_id abc123")
		if err != nil {
			t.Fatalf("Failed to generate key name: %v", err)
		}
	})

	t.Run("analyze cache pattern", func(t *testing.T) {
		_, err := client.AnalyzeCachePattern(ctx, "hit, miss, hit, hit, miss")
		if err != nil {
			t.Fatalf("Failed to analyze pattern: %v", err)
		}
	})

	t.Run("enhance error message", func(t *testing.T) {
		_, err := client.EnhanceErrorMessage(ctx, "cache entry not found")
		if err != nil {
			t.Fatalf("Failed to enhance error message: %v", err)
		}
	})

	t.Run("suggest eviction strategy", func(t *testing.T) {
		_, err := client.SuggestEvictionStrategy(ctx, "frequent repeated access to same items")
		if err != nil {
			t.Fatalf("Failed to suggest strategy: %v", err)
		}
	})

	t.Run("analyze logs", func(t *testing.T) {
		logs := `2024-01-01 10:00:01 ERROR Connection timeout
2024-01-01 10:00:05 ERROR Connection timeout
2024-01-01 10:00:10 INFO Connected
2024-01-01 10:00:15 ERROR Cache miss rate high`

		_, err := client.AnalyzeLogs(ctx, logs)
		if err != nil {
			t.Fatalf("Failed to analyze logs: %v", err)
		}
	})

	t.Run("generate status report", func(t *testing.T) {
		metrics := `Cache size: 1.2GB
Hit rate: 85%
Avg latency: 50ms
Active connections: 100
Memory usage: 75%`

		_, err := client.GenerateStatusReport(ctx, metrics)
		if err != nil {
			t.Fatalf("Failed to generate status report: %v", err)
		}
	})

	t.Run("generate api doc", func(t *testing.T) {
		code := `// Set stores a value in the cache with the given key and TTL
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) error {
	if key == "" {
		return ErrInvalidKey
	}
	if ttl < 0 {
		return ErrInvalidTTL
	}
	return c.store.Set(key, value, ttl)
}`

		_, err := client.GenerateAPIDoc(ctx, code)
		if err != nil {
			t.Fatalf("Failed to generate API doc: %v", err)
		}
	})

	t.Run("generate api example", func(t *testing.T) {
		useCase := "Implementing a rate limiter using the cache"
		_, err := client.GenerateAPIExample(ctx, useCase)
		if err != nil {
			t.Fatalf("Failed to generate API example: %v", err)
		}
	})
}
