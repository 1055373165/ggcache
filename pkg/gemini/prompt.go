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
				AddExample("hit, hit, miss, hit, miss, miss", "Analysis:\n- Calculate hit rate\n- Identify access patterns\n- Suggest potential optimizations")

	// For enhancing error messages
	ErrorMessagePrompt = NewPromptBuilder().
				AddInstruction("Convert the technical error message into a user-friendly explanation.").
				AddExample("connection refused", "Unable to connect to the service. Please check your internet connection and try again.")

	// For suggesting cache eviction strategies
	EvictionStrategyPrompt = NewPromptBuilder().
				AddInstruction("Suggest an appropriate cache eviction strategy based on the access pattern and requirements.").
				AddExample("frequent repeated access to same items", "Recommendation: Consider strategies optimized for frequently accessed items")

	// For analyzing system logs
	LogAnalysisPrompt = NewPromptBuilder().
				AddInstruction("Analyze the system logs and provide a comprehensive analysis that includes:\n"+
			"1. Identify and describe error patterns and their frequency\n"+
			"2. Assess the impact on system stability and performance\n"+
			"3. Analyze trends and correlations between events\n"+
			"4. Provide specific recommendations for addressing issues\n"+
			"Focus on actionable insights and potential preventive measures.").
		AddExample("ERROR Connection timeout\nERROR Connection timeout\nINFO Connected",
			"Analysis:\n"+
				"Pattern Identification:\n"+
				"- Detected repeated connection timeouts followed by recovery\n"+
				"- Pattern suggests intermittent connectivity issues\n\n"+
				"Impact Assessment:\n"+
				"- System stability affected by connection interruptions\n"+
				"- Potential impact on service availability\n\n"+
				"Recommendations:\n"+
				"- Implement connection retry mechanism\n"+
				"- Set up monitoring for connection health\n"+
				"- Consider fallback mechanisms for critical operations")

	// For generating system status reports
	StatusReportPrompt = NewPromptBuilder().
				AddInstruction("Generate a comprehensive system status report that includes:\n"+
			"1. Performance analysis of current metrics\n"+
			"2. System health evaluation\n"+
			"3. Trend analysis (improvements or degradation)\n"+
			"4. Specific recommendations for optimization or action items\n"+
			"Focus on trends and relative changes rather than absolute values.").
		AddExample("Cache utilization: moderate\nHit rate: improved\nLatency: stable",
			"Performance Analysis:\n"+
				"- System performance is stable with moderate resource utilization\n"+
				"- Cache efficiency shows positive trends with improved hit rates\n"+
				"- Response times remain within acceptable ranges\n\n"+
				"Recommendations:\n"+
				"- Continue monitoring hit rate trends\n"+
				"- Consider proactive cache warming if utilization increases\n"+
				"- Set up alerts for significant latency changes")

	// For generating API documentation
	APIDocPrompt = NewPromptBuilder().
			AddInstruction("Generate clear and comprehensive API documentation based on the provided code. Include description, parameters, return values, and any important notes.").
			AddExample("func (c *Cache) Get(key string) (interface{}, error)",
			"### Get\nRetrieves a value from the cache using the specified key.\n\n**Parameters:**\n- key (string): The unique identifier for the cached item\n\n**Returns:**\n- interface{}: The cached value if found\n- error: Error if the key doesn't exist or if retrieval fails")

	// For generating API usage examples
	APIExamplePrompt = NewPromptBuilder().
				AddInstruction("Generate practical and easy-to-understand examples demonstrating how to use the API. Include common use cases and best practices.").
				AddExample("Basic cache operations",
			"Demonstrate initialization, setting values with TTL, retrieving values, and error handling")
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

// BuildLogAnalysisPrompt creates a prompt for analyzing system logs
func BuildLogAnalysisPrompt(logs string) string {
	return LogAnalysisPrompt.AddInput(logs).Build()
}

// BuildStatusReportPrompt creates a prompt for generating system status reports
func BuildStatusReportPrompt(metrics string) string {
	return StatusReportPrompt.AddInput(metrics).Build()
}

// BuildAPIDocPrompt creates a prompt for generating API documentation
func BuildAPIDocPrompt(code string) string {
	return APIDocPrompt.AddInput(code).Build()
}

// BuildAPIExamplePrompt creates a prompt for generating API usage examples
func BuildAPIExamplePrompt(useCase string) string {
	return APIExamplePrompt.AddInput(useCase).Build()
}
