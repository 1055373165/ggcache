package gemini

import (
	"fmt"
	"strings"
)

// PromptBuilder helps build structured prompts for different use cases
type PromptBuilder struct {
	parts []string
}

// NewPromptBuilder creates a new PromptBuilder
func NewPromptBuilder() *PromptBuilder {
	return &PromptBuilder{
		parts: make([]string, 0),
	}
}

// AddContext adds context information to the prompt
func (b *PromptBuilder) AddContext(context string) *PromptBuilder {
	if context != "" {
		b.parts = append(b.parts, fmt.Sprintf("Context: %s", context))
	}
	return b
}

// AddInstruction adds an instruction to the prompt
func (b *PromptBuilder) AddInstruction(instruction string) *PromptBuilder {
	if instruction != "" {
		b.parts = append(b.parts, fmt.Sprintf("Instruction: %s", instruction))
	}
	return b
}

// AddExample adds an example to the prompt
func (b *PromptBuilder) AddExample(input, output string) *PromptBuilder {
	if input != "" && output != "" {
		b.parts = append(b.parts, fmt.Sprintf("Example:\nInput: %s\nOutput: %s", input, output))
	}
	return b
}

// AddInput adds the actual input to process
func (b *PromptBuilder) AddInput(input string) *PromptBuilder {
	if input != "" {
		b.parts = append(b.parts, fmt.Sprintf("Input: %s", input))
	}
	return b
}

// Build creates the final prompt string
func (b *PromptBuilder) Build() string {
	return strings.Join(b.parts, "\n\n")
}

// Common prompt templates for cache-related tasks
var (
	// For generating descriptive cache key names
	KeyNamePrompt = NewPromptBuilder().
			AddInstruction("Generate a descriptive and meaningful cache key name based on the given information. The key should be concise but informative.").
			AddExample("user profile data for user_id 123", "user_profile:123").
			AddExample("product inventory count for category electronics", "inventory_count:electronics")

	// For analyzing cache access patterns
	AccessPatternPrompt = NewPromptBuilder().
				AddInstruction("Analyze the given cache access pattern and provide insights about usage patterns and potential optimizations.").
				AddExample("hit, hit, miss, hit, miss, miss", "Analysis:\n- Hit rate: 50%\n- Shows clustering of misses\n- Consider increasing cache size")

	// For enhancing error messages
	ErrorMessagePrompt = NewPromptBuilder().
				AddInstruction("Convert the technical error message into a user-friendly explanation.").
				AddExample("connection refused", "Unable to connect to the service. Please check your internet connection and try again.")

	// For suggesting cache eviction strategies
	EvictionStrategyPrompt = NewPromptBuilder().
				AddInstruction("Suggest an appropriate cache eviction strategy based on the access pattern and requirements.").
				AddExample("frequent repeated access to same items", "Recommendation: LRU (Least Recently Used) strategy would be optimal for this pattern")
)

// Helper functions for common cache-related prompts

// BuildKeyNamePrompt creates a prompt for generating cache key names
func BuildKeyNamePrompt(description string) string {
	return KeyNamePrompt.AddInput(description).Build()
}

// BuildAccessPatternPrompt creates a prompt for analyzing cache access patterns
func BuildAccessPatternPrompt(pattern string) string {
	return AccessPatternPrompt.AddInput(pattern).Build()
}

// BuildErrorMessagePrompt creates a prompt for enhancing error messages
func BuildErrorMessagePrompt(errorMsg string) string {
	return ErrorMessagePrompt.AddInput(errorMsg).Build()
}

// BuildEvictionStrategyPrompt creates a prompt for suggesting eviction strategies
func BuildEvictionStrategyPrompt(requirements string) string {
	return EvictionStrategyPrompt.AddInput(requirements).Build()
}
