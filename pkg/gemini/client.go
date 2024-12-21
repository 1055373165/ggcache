package gemini

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	genai "github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
)

// Client wraps the Gemini client with additional functionality
type Client struct {
	client *genai.Client
	model  *genai.GenerativeModel
	mu     sync.RWMutex
	cfg    Config
}

// Config holds the configuration for the Gemini client
type Config struct {
	APIKey     string
	ModelName  string // e.g., "gemini-1.5-flash"
	Timeout    time.Duration
	MaxRetries int
}

// NewClient creates a new Gemini client
func NewClient(ctx context.Context, cfg Config) (*Client, error) {
	if cfg.APIKey == "" {
		return nil, fmt.Errorf("API key is required")
	}

	if cfg.ModelName == "" {
		cfg.ModelName = "gemini-1.5-flash" // default model
	}

	if cfg.Timeout == 0 {
		cfg.Timeout = 10 * time.Second // default timeout
	}

	if cfg.MaxRetries == 0 {
		cfg.MaxRetries = 3 // default retries
	}

	client, err := genai.NewClient(ctx, option.WithAPIKey(cfg.APIKey))
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %v", err)
	}

	model := client.GenerativeModel(cfg.ModelName)
	model.SetTemperature(0.7) // 设置生成文本的创造性程度

	return &Client{
		client: client,
		model:  model,
		cfg:    cfg,
	}, nil
}

// GenerateContent generates content using the Gemini model with retries
func (c *Client) GenerateContent(ctx context.Context, prompt string) (string, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var lastErr error
	for attempt := 0; attempt < c.cfg.MaxRetries; attempt++ {
		// Create a timeout context for this attempt
		timeoutCtx, cancel := context.WithTimeout(ctx, c.cfg.Timeout)
		defer cancel()

		// 创建请求内容
		resp, err := c.model.GenerateContent(timeoutCtx, genai.Text(prompt))
		if err == nil {
			if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
				return "", fmt.Errorf("no content generated")
			}
			return fmt.Sprintf("%v", resp.Candidates[0].Content.Parts[0]), nil
		}

		lastErr = err
		log.Printf("Attempt %d failed: %v", attempt+1, err)

		// Wait before retrying, with exponential backoff
		if attempt < c.cfg.MaxRetries-1 { // Don't sleep after the last attempt
			time.Sleep(time.Duration(attempt+1) * 2 * time.Second)
		}
	}

	return "", fmt.Errorf("failed to generate content after %d attempts: %v", c.cfg.MaxRetries, lastErr)
}

// Close closes the Gemini client
func (c *Client) Close() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.client != nil {
		c.client.Close()
	}
}

// GenerateCacheKeyName generates a descriptive cache key name
func (c *Client) GenerateCacheKeyName(ctx context.Context, description string) (string, error) {
	prompt := BuildKeyNamePrompt(description)
	return c.GenerateContent(ctx, prompt)
}

// AnalyzeCachePattern analyzes cache access patterns
func (c *Client) AnalyzeCachePattern(ctx context.Context, pattern string) (string, error) {
	prompt := BuildAccessPatternPrompt(pattern)
	return c.GenerateContent(ctx, prompt)
}

// EnhanceErrorMessage converts technical error messages to user-friendly ones
func (c *Client) EnhanceErrorMessage(ctx context.Context, errorMsg string) (string, error) {
	prompt := BuildErrorMessagePrompt(errorMsg)
	return c.GenerateContent(ctx, prompt)
}

// SuggestEvictionStrategy suggests a cache eviction strategy
func (c *Client) SuggestEvictionStrategy(ctx context.Context, requirements string) (string, error) {
	prompt := BuildEvictionStrategyPrompt(requirements)
	return c.GenerateContent(ctx, prompt)
}

// AnalyzeLogs analyzes system logs to identify anomalies and patterns
func (c *Client) AnalyzeLogs(ctx context.Context, logs string) (string, error) {
	prompt := BuildLogAnalysisPrompt(logs)
	return c.GenerateContent(ctx, prompt)
}

// GenerateStatusReport generates a natural language summary of system status
func (c *Client) GenerateStatusReport(ctx context.Context, metrics string) (string, error) {
	prompt := BuildStatusReportPrompt(metrics)
	return c.GenerateContent(ctx, prompt)
}

// GenerateAPIDoc generates API documentation for the given code
func (c *Client) GenerateAPIDoc(ctx context.Context, code string) (string, error) {
	prompt := BuildAPIDocPrompt(code)
	return c.GenerateContent(ctx, prompt)
}

// GenerateAPIExample generates usage examples for the API
func (c *Client) GenerateAPIExample(ctx context.Context, useCase string) (string, error) {
	prompt := BuildAPIExamplePrompt(useCase)
	return c.GenerateContent(ctx, prompt)
}
