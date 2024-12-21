package gemini

import (
	"strings"
	"testing"
)

func TestPromptBuilder(t *testing.T) {
	tests := []struct {
		name     string
		build    func() string
		contains []string
	}{
		{
			name: "key name prompt",
			build: func() string {
				return BuildKeyNamePrompt("user session data for session_id abc123")
			},
			contains: []string{
				"Generate a descriptive",
				"user session data for session_id abc123",
			},
		},
		{
			name: "access pattern prompt",
			build: func() string {
				return BuildAccessPatternPrompt("hit, miss, hit, hit, miss")
			},
			contains: []string{
				"Analyze the given cache access pattern",
				"hit, miss, hit, hit, miss",
			},
		},
		{
			name: "error message prompt",
			build: func() string {
				return BuildErrorMessagePrompt("cache entry not found")
			},
			contains: []string{
				"Convert the technical error message",
				"cache entry not found",
			},
		},
		{
			name: "custom prompt builder",
			build: func() string {
				return NewPromptBuilder().
					AddContext("Cache system with 1GB memory").
					AddInstruction("Suggest optimal cache size").
					AddExample("500MB usage", "Recommend 750MB cache size").
					AddInput("Current usage: 800MB").
					Build()
			},
			contains: []string{
				"Context: Cache system with 1GB memory",
				"Instruction: Suggest optimal cache size",
				"Example:",
				"Input: 500MB usage",
				"Output: Recommend 750MB cache size",
				"Input: Current usage: 800MB",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prompt := tt.build()

			// Verify all expected contents are present
			for _, expected := range tt.contains {
				if !strings.Contains(prompt, expected) {
					t.Errorf("prompt should contain %q but got:\n%s", expected, prompt)
				}
			}
		})
	}
}
