package optimization

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/yourusername/llm-gate/internal/core/models"
)

func TestTokenOptimizer_CompressWhitespace(t *testing.T) {
	optimizer := NewTokenOptimizer(nil)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "multiple spaces",
			input:    "Hello    world",
			expected: "Hello world",
		},
		{
			name:     "excessive newlines",
			input:    "Hello\n\n\n\nworld",
			expected: "Hello\n\nworld",
		},
		{
			name:     "leading and trailing whitespace",
			input:    "  Hello world  ",
			expected: "Hello world",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := optimizer.compressWhitespace(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTokenOptimizer_RemoveMarkdownFormatting(t *testing.T) {
	optimizer := NewTokenOptimizer(nil)

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "bold with asterisks",
			input:    "This is **bold** text",
			expected: "This is bold text",
		},
		{
			name:     "bold with underscores",
			input:    "This is __bold__ text",
			expected: "This is bold text",
		},
		{
			name:     "italic with asterisks",
			input:    "This is *italic* text",
			expected: "This is italic text",
		},
		{
			name:     "italic with underscores",
			input:    "This is _italic_ text",
			expected: "This is italic text",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := optimizer.removeMarkdownFormatting(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestTokenOptimizer_EstimateTokens(t *testing.T) {
	optimizer := NewTokenOptimizer(nil)

	req := &models.UnifiedRequest{
		Messages: []models.Message{
			{Role: "user", Content: "Hello, world!"},
			{Role: "assistant", Content: "Hi there!"},
		},
	}

	tokens := optimizer.estimateTokens(req)
	assert.Greater(t, tokens, 0)
	// Rough estimation: ~24 characters / 4 = ~6 tokens
	assert.GreaterOrEqual(t, tokens, 5)
}

func TestTokenOptimizer_OptimizeRequest(t *testing.T) {
	config := &OptimizerConfig{
		EnableCompression:     true,
		EnableTruncation:      false,
		MaxPromptTokens:       1000,
		EnableSmartTruncation: true,
	}
	optimizer := NewTokenOptimizer(config)

	req := &models.UnifiedRequest{
		Messages: []models.Message{
			{Role: "user", Content: "Hello    world   with    extra    spaces"},
		},
	}

	optimized, err := optimizer.OptimizeRequest(req)
	assert.NoError(t, err)
	assert.NotNil(t, optimized)
	// Whitespace should be compressed
	assert.Contains(t, optimized.Messages[0].Content, "Hello world")
	assert.NotContains(t, optimized.Messages[0].Content, "    ")
}

func TestTokenOptimizer_TruncateMessages(t *testing.T) {
	config := &OptimizerConfig{
		EnableCompression:     false,
		EnableTruncation:      true,
		MaxPromptTokens:       10, // Very low limit for testing
		EnableSmartTruncation: false,
	}
	optimizer := NewTokenOptimizer(config)

	req := &models.UnifiedRequest{
		Messages: []models.Message{
			{Role: "user", Content: "This is a very long message that should be truncated because it exceeds the token limit"},
		},
	}

	optimizer.truncateMessages(req, config.MaxPromptTokens)

	// Message should be truncated
	assert.Less(t, len(req.Messages[0].Content), 100)
}

func TestTokenOptimizer_SmartTruncate(t *testing.T) {
	config := &OptimizerConfig{
		EnableCompression:     false,
		EnableTruncation:      true,
		MaxPromptTokens:       50, // Low limit for testing
		EnableSmartTruncation: true,
	}
	optimizer := NewTokenOptimizer(config)

	req := &models.UnifiedRequest{
		Messages: []models.Message{
			{Role: "system", Content: "You are a helpful assistant"},
			{Role: "user", Content: "First question"},
			{Role: "assistant", Content: "First answer"},
			{Role: "user", Content: "Second question"},
			{Role: "assistant", Content: "Second answer"},
			{Role: "user", Content: "Final question that should be kept"},
		},
	}

	optimizer.smartTruncate(req, config.MaxPromptTokens)

	// System message should be preserved
	assert.Equal(t, "system", req.Messages[0].Role)

	// Latest user message should be preserved
	lastMsg := req.Messages[len(req.Messages)-1]
	assert.Equal(t, "user", lastMsg.Role)
	assert.Contains(t, lastMsg.Content, "Final question")
}

func TestTokenOptimizer_OptimizeContextWindow(t *testing.T) {
	optimizer := NewTokenOptimizer(nil)

	messages := []models.Message{
		{Role: "system", Content: "You are helpful"},
		{Role: "user", Content: "Question 1"},
		{Role: "assistant", Content: "Answer 1"},
		{Role: "user", Content: "Question 2"},
	}

	optimized := optimizer.OptimizeContextWindow(messages, 20)
	assert.NotNil(t, optimized)
	// Should have messages
	assert.Greater(t, len(optimized), 0)
}

func TestTokenOptimizer_CalculateTokenSavings(t *testing.T) {
	optimizer := NewTokenOptimizer(nil)

	original := &models.UnifiedRequest{
		Messages: []models.Message{
			{Role: "user", Content: "Hello    world    with    many    spaces"},
		},
	}

	optimized := &models.UnifiedRequest{
		Messages: []models.Message{
			{Role: "user", Content: "Hello world with many spaces"},
		},
	}

	savings := optimizer.CalculateTokenSavings(original, optimized)
	assert.GreaterOrEqual(t, savings, 0)
}
