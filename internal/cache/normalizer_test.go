package cache

import (
	"strings"
	"testing"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
)

func TestNormalizeRequest(t *testing.T) {
	normalizer := NewRequestNormalizer()

	tests := []struct {
		name    string
		request *models.ChatRequest
		want    []string // Expected substrings in normalized output
	}{
		{
			name: "basic request",
			request: &models.ChatRequest{
				Model: "gpt-4",
				Messages: []models.Message{
					{Role: "user", Content: "Hello, world!"},
				},
			},
			want: []string{"model:gpt-4", "messages:", "user:Hello, world!"},
		},
		{
			name: "request with temperature",
			request: &models.ChatRequest{
				Model: "gpt-3.5-turbo",
				Messages: []models.Message{
					{Role: "user", Content: "Test"},
				},
				Temperature: floatPtr(0.7),
			},
			want: []string{"model:gpt-3.5-turbo", "temp:0.7"},
		},
		{
			name: "request with max_tokens",
			request: &models.ChatRequest{
				Model: "gpt-4",
				Messages: []models.Message{
					{Role: "user", Content: "Test"},
				},
				MaxTokens: intPtr(100),
			},
			want: []string{"max_tokens:100"},
		},
		{
			name: "multi-message conversation",
			request: &models.ChatRequest{
				Model: "gpt-4",
				Messages: []models.Message{
					{Role: "system", Content: "You are a helpful assistant"},
					{Role: "user", Content: "What is the weather?"},
					{Role: "assistant", Content: "I don't have real-time data"},
					{Role: "user", Content: "OK, thanks"},
				},
			},
			want: []string{"system:You are a helpful assistant", "user:What is the weather?"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizer.NormalizeRequest(tt.request)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Errorf("normalized request %q does not contain %q", got, want)
				}
			}
		})
	}
}

func TestGenerateCacheKey(t *testing.T) {
	normalizer := NewRequestNormalizer()

	req1 := &models.ChatRequest{
		Model: "gpt-4",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	req2 := &models.ChatRequest{
		Model: "gpt-4",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
	}

	req3 := &models.ChatRequest{
		Model: "gpt-4",
		Messages: []models.Message{
			{Role: "user", Content: "Goodbye"},
		},
	}

	key1, err := normalizer.GenerateCacheKey(req1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	key2, err := normalizer.GenerateCacheKey(req2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	key3, err := normalizer.GenerateCacheKey(req3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Same requests should produce same key
	if key1 != key2 {
		t.Errorf("identical requests produced different keys: %s != %s", key1, key2)
	}

	// Different requests should produce different keys
	if key1 == key3 {
		t.Errorf("different requests produced same key: %s", key1)
	}

	// Keys should be hex-encoded SHA256 (64 characters)
	if len(key1) != 64 {
		t.Errorf("key length is %d, want 64", len(key1))
	}
}

func TestExtractSearchableText(t *testing.T) {
	normalizer := NewRequestNormalizer()

	tests := []struct {
		name    string
		request *models.ChatRequest
		want    []string // Expected substrings
	}{
		{
			name: "single message",
			request: &models.ChatRequest{
				Messages: []models.Message{
					{Role: "user", Content: "What is the capital of France?"},
				},
			},
			want: []string{"What is the capital of France?"},
		},
		{
			name: "multiple messages",
			request: &models.ChatRequest{
				Messages: []models.Message{
					{Role: "system", Content: "You are a helpful assistant"},
					{Role: "user", Content: "Hello"},
					{Role: "assistant", Content: "Hi there!"},
					{Role: "user", Content: "How are you?"},
				},
			},
			want: []string{"You are a helpful assistant", "Hello", "Hi there!", "How are you?"},
		},
		{
			name: "empty content filtered",
			request: &models.ChatRequest{
				Messages: []models.Message{
					{Role: "user", Content: "Test"},
					{Role: "assistant", Content: ""},
					{Role: "user", Content: "Another test"},
				},
			},
			want: []string{"Test", "Another test"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := normalizer.ExtractSearchableText(tt.request)

			for _, want := range tt.want {
				if !strings.Contains(got, want) {
					t.Errorf("extracted text %q does not contain %q", got, want)
				}
			}
		})
	}
}

func TestNormalizeMessages(t *testing.T) {
	normalizer := NewRequestNormalizer()

	messages := []models.Message{
		{Role: "User", Content: "  Hello  "},  // Mixed case, extra whitespace
		{Role: "ASSISTANT", Content: "Hi!"},
		{Role: "user", Content: "How are you?"},
	}

	result := normalizer.normalizeMessages(messages)

	// Should normalize to lowercase roles
	if !strings.Contains(result, "user:Hello") {
		t.Errorf("result does not contain 'user:Hello': %s", result)
	}

	// Should normalize to lowercase and trim
	if !strings.Contains(result, "assistant:Hi!") {
		t.Errorf("result does not contain 'assistant:Hi!': %s", result)
	}
}

// Helper functions
func floatPtr(f float64) *float64 {
	return &f
}

func intPtr(i int) *int {
	return &i
}
