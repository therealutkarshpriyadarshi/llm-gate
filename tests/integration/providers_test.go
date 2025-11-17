// +build integration

package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/providers"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/providers/anthropic"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/providers/azure"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/providers/bedrock"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/providers/openai"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/providers/vertex"
)

// These tests require actual API keys and will only run when the integration build tag is used
// Example: go test -tags=integration ./tests/integration/...

func TestOpenAIProvider_Integration(t *testing.T) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		t.Skip("OPENAI_API_KEY not set, skipping integration test")
	}

	config := &openai.Config{
		APIKey:     apiKey,
		BaseURL:    "https://api.openai.com/v1",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
	}

	client, err := openai.NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create OpenAI client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	req := &models.ChatRequest{
		Model: "gpt-3.5-turbo",
		Messages: []models.Message{
			{Role: "user", Content: "Say 'Hello, World!' and nothing else."},
		},
	}

	resp, err := client.ChatCompletion(ctx, req)
	if err != nil {
		t.Fatalf("ChatCompletion failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Response is nil")
	}

	if len(resp.Choices) == 0 {
		t.Fatal("No choices in response")
	}

	t.Logf("OpenAI response: %s", resp.Choices[0].Message.Content)
	t.Logf("Tokens used: %d total (%d prompt + %d completion)",
		resp.Usage.TotalTokens, resp.Usage.PromptTokens, resp.Usage.CompletionTokens)
	t.Logf("Cost: $%.6f", resp.Metadata.Cost)
}

func TestAnthropicProvider_Integration(t *testing.T) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		t.Skip("ANTHROPIC_API_KEY not set, skipping integration test")
	}

	config := &anthropic.Config{
		APIKey:     apiKey,
		BaseURL:    "https://api.anthropic.com",
		Version:    "2023-06-01",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
	}

	client, err := anthropic.NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create Anthropic client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	maxTokens := 100
	req := &models.ChatRequest{
		Model: "claude-3-haiku-20240307",
		Messages: []models.Message{
			{Role: "user", Content: "Say 'Hello, World!' and nothing else."},
		},
		MaxTokens: &maxTokens,
	}

	resp, err := client.ChatCompletion(ctx, req)
	if err != nil {
		t.Fatalf("ChatCompletion failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Response is nil")
	}

	if len(resp.Choices) == 0 {
		t.Fatal("No choices in response")
	}

	t.Logf("Anthropic response: %s", resp.Choices[0].Message.Content)
	t.Logf("Tokens used: %d total (%d prompt + %d completion)",
		resp.Usage.TotalTokens, resp.Usage.PromptTokens, resp.Usage.CompletionTokens)
	t.Logf("Cost: $%.6f", resp.Metadata.Cost)
}

func TestAzureProvider_Integration(t *testing.T) {
	apiKey := os.Getenv("AZURE_OPENAI_API_KEY")
	endpoint := os.Getenv("AZURE_OPENAI_ENDPOINT")
	deploymentName := os.Getenv("AZURE_DEPLOYMENT_NAME")

	if apiKey == "" || endpoint == "" {
		t.Skip("AZURE_OPENAI_API_KEY or AZURE_OPENAI_ENDPOINT not set, skipping integration test")
	}

	if deploymentName == "" {
		deploymentName = "gpt-35-turbo"
	}

	config := &azure.Config{
		APIKey:         apiKey,
		Endpoint:       endpoint,
		APIVersion:     "2024-02-15-preview",
		DeploymentName: deploymentName,
		Timeout:        30 * time.Second,
		MaxRetries:     3,
		RetryDelay:     1 * time.Second,
	}

	client, err := azure.NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create Azure client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	req := &models.ChatRequest{
		Model: deploymentName,
		Messages: []models.Message{
			{Role: "user", Content: "Say 'Hello, World!' and nothing else."},
		},
	}

	resp, err := client.ChatCompletion(ctx, req)
	if err != nil {
		t.Fatalf("ChatCompletion failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Response is nil")
	}

	if len(resp.Choices) == 0 {
		t.Fatal("No choices in response")
	}

	t.Logf("Azure response: %s", resp.Choices[0].Message.Content)
	t.Logf("Tokens used: %d total (%d prompt + %d completion)",
		resp.Usage.TotalTokens, resp.Usage.PromptTokens, resp.Usage.CompletionTokens)
	t.Logf("Cost: $%.6f", resp.Metadata.Cost)
}

func TestBedrockProvider_Integration(t *testing.T) {
	accessKeyID := os.Getenv("AWS_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("AWS_SECRET_ACCESS_KEY")
	region := os.Getenv("AWS_REGION")

	if accessKeyID == "" || secretAccessKey == "" {
		t.Skip("AWS credentials not set, skipping integration test")
	}

	if region == "" {
		region = "us-east-1"
	}

	config := &bedrock.Config{
		AccessKeyID:     accessKeyID,
		SecretAccessKey: secretAccessKey,
		Region:          region,
		Timeout:         60 * time.Second,
		MaxRetries:      3,
		RetryDelay:      1 * time.Second,
	}

	client, err := bedrock.NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create Bedrock client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	maxTokens := 100
	req := &models.ChatRequest{
		Model: "anthropic.claude-3-haiku-20240307-v1:0",
		Messages: []models.Message{
			{Role: "user", Content: "Say 'Hello, World!' and nothing else."},
		},
		MaxTokens: &maxTokens,
	}

	resp, err := client.ChatCompletion(ctx, req)
	if err != nil {
		t.Fatalf("ChatCompletion failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Response is nil")
	}

	if len(resp.Choices) == 0 {
		t.Fatal("No choices in response")
	}

	t.Logf("Bedrock response: %s", resp.Choices[0].Message.Content)
	t.Logf("Tokens used: %d total (%d prompt + %d completion)",
		resp.Usage.TotalTokens, resp.Usage.PromptTokens, resp.Usage.CompletionTokens)
	t.Logf("Cost: $%.6f", resp.Metadata.Cost)
}

func TestVertexProvider_Integration(t *testing.T) {
	projectID := os.Getenv("GCP_PROJECT_ID")
	apiKey := os.Getenv("VERTEX_API_KEY")
	location := os.Getenv("VERTEX_LOCATION")

	if projectID == "" || apiKey == "" {
		t.Skip("GCP_PROJECT_ID or VERTEX_API_KEY not set, skipping integration test")
	}

	if location == "" {
		location = "us-central1"
	}

	config := &vertex.Config{
		ProjectID:  projectID,
		Location:   location,
		APIKey:     apiKey,
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
	}

	client, err := vertex.NewClient(config)
	if err != nil {
		t.Fatalf("Failed to create Vertex client: %v", err)
	}
	defer client.Close()

	ctx := context.Background()
	req := &models.ChatRequest{
		Model: "gemini-1.5-flash",
		Messages: []models.Message{
			{Role: "user", Content: "Say 'Hello, World!' and nothing else."},
		},
	}

	resp, err := client.ChatCompletion(ctx, req)
	if err != nil {
		t.Fatalf("ChatCompletion failed: %v", err)
	}

	if resp == nil {
		t.Fatal("Response is nil")
	}

	if len(resp.Choices) == 0 {
		t.Fatal("No choices in response")
	}

	t.Logf("Vertex response: %s", resp.Choices[0].Message.Content)
	t.Logf("Tokens used: %d total (%d prompt + %d completion)",
		resp.Usage.TotalTokens, resp.Usage.PromptTokens, resp.Usage.CompletionTokens)
	t.Logf("Cost: $%.6f", resp.Metadata.Cost)
}

func TestProviderFactory_Integration(t *testing.T) {
	factory := providers.NewFactory()

	// Test factory with all available providers
	tests := []struct {
		name         string
		providerType models.ProviderType
		configMap    map[string]interface{}
		model        string
		envCheck     func() bool
	}{
		{
			name:         "OpenAI via factory",
			providerType: models.ProviderOpenAI,
			configMap: map[string]interface{}{
				"api_key":  os.Getenv("OPENAI_API_KEY"),
				"base_url": "https://api.openai.com/v1",
			},
			model: "gpt-3.5-turbo",
			envCheck: func() bool {
				return os.Getenv("OPENAI_API_KEY") != ""
			},
		},
		{
			name:         "Anthropic via factory",
			providerType: models.ProviderAnthropic,
			configMap: map[string]interface{}{
				"api_key":  os.Getenv("ANTHROPIC_API_KEY"),
				"base_url": "https://api.anthropic.com",
				"version":  "2023-06-01",
			},
			model: "claude-3-haiku-20240307",
			envCheck: func() bool {
				return os.Getenv("ANTHROPIC_API_KEY") != ""
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if !tt.envCheck() {
				t.Skipf("Environment variables not set for %s", tt.name)
			}

			provider, err := factory.CreateProvider(tt.providerType, tt.configMap)
			if err != nil {
				t.Fatalf("Failed to create provider: %v", err)
			}
			defer provider.Close()

			ctx := context.Background()
			maxTokens := 50
			req := &models.ChatRequest{
				Model: tt.model,
				Messages: []models.Message{
					{Role: "user", Content: "Say hi"},
				},
				MaxTokens: &maxTokens,
			}

			resp, err := provider.ChatCompletion(ctx, req)
			if err != nil {
				t.Fatalf("ChatCompletion failed: %v", err)
			}

			if resp == nil {
				t.Fatal("Response is nil")
			}

			t.Logf("Provider %s responded successfully", tt.providerType)
		})
	}
}
