package gemini

import (
	"context"
	"testing"
)

func TestAPIDocGeneration(t *testing.T) {
	client := setupTestClient(t)
	defer client.Close()

	tests := []struct {
		name     string
		code     string
		wantErr  bool
		validate func(t *testing.T, doc string)
	}{
		{
			name: "cache set method",
			code: `
// Set stores a value in the cache with an optional TTL
func (c *Cache) Set(key string, value interface{}, ttl time.Duration) error {
	if key == "" {
		return ErrInvalidKey
	}
	if ttl < 0 {
		return ErrInvalidTTL
	}
	return c.store.Set(key, value, ttl)
}`,
			validate: func(t *testing.T, doc string) {
				assertContains(t, doc, "Parameters", "Should list parameters")
				assertContains(t, doc, "Returns", "Should describe return values")
				assertContains(t, doc, "error", "Should mention error handling")
			},
		},
		{
			name: "cache get method",
			code: `
// Get retrieves a value from the cache
func (c *Cache) Get(key string) (interface{}, error) {
	if key == "" {
		return nil, ErrInvalidKey
	}
	return c.store.Get(key)
}`,
			validate: func(t *testing.T, doc string) {
				assertContains(t, doc, "Parameters", "Should list parameters")
				assertContains(t, doc, "Returns", "Should describe return values")
				assertContains(t, doc, "interface{}", "Should mention return type")
			},
		},
	}

	ctx := context.Background()
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			doc, err := client.GenerateAPIDoc(ctx, tt.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("GenerateAPIDoc() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.validate != nil {
				tt.validate(t, doc)
			}
		})
	}
}
