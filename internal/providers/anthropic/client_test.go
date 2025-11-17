package anthropic

import (
	"testing"
	"time"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
)

func TestNewClient(t *testing.T) {
	tests := []struct {
		name    string
		config  *Config
		wantErr bool
	}{
		{
			name: "valid config",
			config: &Config{
				APIKey:     "sk-ant-test-key",
				BaseURL:    "https://api.anthropic.com",
				Version:    "2023-06-01",
				Timeout:    30 * time.Second,
				MaxRetries: 3,
				RetryDelay: 1 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "invalid config - missing API key",
			config: &Config{
				BaseURL:    "https://api.anthropic.com",
				Version:    "2023-06-01",
				Timeout:    30 * time.Second,
				MaxRetries: 3,
				RetryDelay: 1 * time.Second,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client, err := NewClient(tt.config)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewClient() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && client == nil {
				t.Error("NewClient() returned nil client")
			}
		})
	}
}

func TestClient_Name(t *testing.T) {
	config := &Config{
		APIKey:     "sk-ant-test-key",
		BaseURL:    "https://api.anthropic.com",
		Version:    "2023-06-01",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	if client.Name() != models.ProviderAnthropic {
		t.Errorf("Client.Name() = %v, want %v", client.Name(), models.ProviderAnthropic)
	}
}

func TestClient_GetCapabilities(t *testing.T) {
	config := &Config{
		APIKey:     "sk-ant-test-key",
		BaseURL:    "https://api.anthropic.com",
		Version:    "2023-06-01",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	capabilities := client.GetCapabilities()

	if capabilities.Name != models.ProviderAnthropic {
		t.Errorf("GetCapabilities() Name = %v, want %v", capabilities.Name, models.ProviderAnthropic)
	}

	if !capabilities.SupportsStreaming {
		t.Error("GetCapabilities() SupportsStreaming = false, want true")
	}

	if capabilities.SupportsFunctions {
		t.Error("GetCapabilities() SupportsFunctions = true, want false")
	}

	if !capabilities.SupportsVision {
		t.Error("GetCapabilities() SupportsVision = false, want true")
	}

	if capabilities.MaxTokens != 200000 {
		t.Errorf("GetCapabilities() MaxTokens = %v, want 200000", capabilities.MaxTokens)
	}

	if len(capabilities.Models) != len(ModelPricing) {
		t.Errorf("GetCapabilities() Models count = %v, want %v", len(capabilities.Models), len(ModelPricing))
	}

	// Verify all models are streaming-capable
	for _, model := range capabilities.Models {
		if !model.SupportsStreaming {
			t.Errorf("Model %s should support streaming", model.ID)
		}
	}
}

func TestClient_GetModelInfo(t *testing.T) {
	config := &Config{
		APIKey:     "sk-ant-test-key",
		BaseURL:    "https://api.anthropic.com",
		Version:    "2023-06-01",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	tests := []struct {
		name      string
		modelID   string
		wantError bool
	}{
		{
			name:      "Claude 3 Opus",
			modelID:   "claude-3-opus-20240229",
			wantError: false,
		},
		{
			name:      "Claude 3 Sonnet",
			modelID:   "claude-3-sonnet-20240229",
			wantError: false,
		},
		{
			name:      "Unknown model",
			modelID:   "unknown-model",
			wantError: false, // Should return default pricing
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modelInfo, err := client.GetModelInfo(tt.modelID)
			if (err != nil) != tt.wantError {
				t.Errorf("GetModelInfo() error = %v, wantError %v", err, tt.wantError)
				return
			}

			if !tt.wantError {
				if modelInfo.ID != tt.modelID {
					t.Errorf("GetModelInfo() ID = %v, want %v", modelInfo.ID, tt.modelID)
				}
				if modelInfo.InputCostPer1K <= 0 {
					t.Errorf("GetModelInfo() InputCostPer1K = %v, want > 0", modelInfo.InputCostPer1K)
				}
				if modelInfo.OutputCostPer1K <= 0 {
					t.Errorf("GetModelInfo() OutputCostPer1K = %v, want > 0", modelInfo.OutputCostPer1K)
				}
			}
		})
	}
}

func TestClient_ToAnthropicRequest(t *testing.T) {
	config := &Config{
		APIKey:     "sk-ant-test-key",
		BaseURL:    "https://api.anthropic.com",
		Version:    "2023-06-01",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	tests := []struct {
		name           string
		request        *models.ChatRequest
		wantErr        bool
		wantSystemMsg  string
		wantMsgCount   int
	}{
		{
			name: "simple user message",
			request: &models.ChatRequest{
				Model: "claude-3-sonnet-20240229",
				Messages: []models.Message{
					{Role: "user", Content: "Hello"},
				},
			},
			wantErr:       false,
			wantSystemMsg: "",
			wantMsgCount:  1,
		},
		{
			name: "system and user messages",
			request: &models.ChatRequest{
				Model: "claude-3-sonnet-20240229",
				Messages: []models.Message{
					{Role: "system", Content: "You are a helpful assistant"},
					{Role: "user", Content: "Hello"},
				},
			},
			wantErr:       false,
			wantSystemMsg: "You are a helpful assistant",
			wantMsgCount:  1,
		},
		{
			name: "multiple system messages",
			request: &models.ChatRequest{
				Model: "claude-3-sonnet-20240229",
				Messages: []models.Message{
					{Role: "system", Content: "First instruction"},
					{Role: "system", Content: "Second instruction"},
					{Role: "user", Content: "Hello"},
				},
			},
			wantErr:       false,
			wantSystemMsg: "First instruction\n\nSecond instruction",
			wantMsgCount:  1,
		},
		{
			name: "only system message",
			request: &models.ChatRequest{
				Model: "claude-3-sonnet-20240229",
				Messages: []models.Message{
					{Role: "system", Content: "You are a helpful assistant"},
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			anthropicReq, err := client.toAnthropicRequest(tt.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("toAnthropicRequest() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				if anthropicReq.System != tt.wantSystemMsg {
					t.Errorf("toAnthropicRequest() System = %v, want %v", anthropicReq.System, tt.wantSystemMsg)
				}
				if len(anthropicReq.Messages) != tt.wantMsgCount {
					t.Errorf("toAnthropicRequest() Messages count = %v, want %v", len(anthropicReq.Messages), tt.wantMsgCount)
				}
				// MaxTokens should be set
				if anthropicReq.MaxTokens <= 0 {
					t.Errorf("toAnthropicRequest() MaxTokens = %v, want > 0", anthropicReq.MaxTokens)
				}
			}
		})
	}
}

func TestClient_Close(t *testing.T) {
	config := &Config{
		APIKey:     "sk-ant-test-key",
		BaseURL:    "https://api.anthropic.com",
		Version:    "2023-06-01",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	err = client.Close()
	if err != nil {
		t.Errorf("Close() error = %v, want nil", err)
	}
}

func TestClient_ToUnifiedResponse(t *testing.T) {
	config := &Config{
		APIKey:     "sk-ant-test-key",
		BaseURL:    "https://api.anthropic.com",
		Version:    "2023-06-01",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		RetryDelay: 1 * time.Second,
	}

	client, err := NewClient(config)
	if err != nil {
		t.Fatalf("NewClient() error = %v", err)
	}

	anthropicResp := &AnthropicResponse{
		ID:   "msg_123",
		Type: "message",
		Role: "assistant",
		Content: []AnthropicContentBlock{
			{Type: "text", Text: "Hello! How can I help you today?"},
		},
		Model:      "claude-3-sonnet-20240229",
		StopReason: "end_turn",
		Usage: AnthropicUsage{
			InputTokens:  10,
			OutputTokens: 20,
		},
	}

	originalReq := &models.ChatRequest{
		Model: "claude-3-sonnet-20240229",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
		Metadata: models.RequestMetadata{
			RequestID: "req_123",
		},
	}

	unifiedResp := client.toUnifiedResponse(anthropicResp, originalReq)

	if unifiedResp.ID != "msg_123" {
		t.Errorf("toUnifiedResponse() ID = %v, want msg_123", unifiedResp.ID)
	}

	if unifiedResp.Model != "claude-3-sonnet-20240229" {
		t.Errorf("toUnifiedResponse() Model = %v, want claude-3-sonnet-20240229", unifiedResp.Model)
	}

	if unifiedResp.Provider != string(models.ProviderAnthropic) {
		t.Errorf("toUnifiedResponse() Provider = %v, want %v", unifiedResp.Provider, models.ProviderAnthropic)
	}

	if len(unifiedResp.Choices) != 1 {
		t.Errorf("toUnifiedResponse() Choices count = %v, want 1", len(unifiedResp.Choices))
	}

	if unifiedResp.Choices[0].Message.Content != "Hello! How can I help you today?" {
		t.Errorf("toUnifiedResponse() Content = %v, want 'Hello! How can I help you today?'", unifiedResp.Choices[0].Message.Content)
	}

	if unifiedResp.Usage.PromptTokens != 10 {
		t.Errorf("toUnifiedResponse() PromptTokens = %v, want 10", unifiedResp.Usage.PromptTokens)
	}

	if unifiedResp.Usage.CompletionTokens != 20 {
		t.Errorf("toUnifiedResponse() CompletionTokens = %v, want 20", unifiedResp.Usage.CompletionTokens)
	}

	if unifiedResp.Usage.TotalTokens != 30 {
		t.Errorf("toUnifiedResponse() TotalTokens = %v, want 30", unifiedResp.Usage.TotalTokens)
	}

	if unifiedResp.Metadata.Cost <= 0 {
		t.Errorf("toUnifiedResponse() Cost = %v, want > 0", unifiedResp.Metadata.Cost)
	}

	if unifiedResp.Metadata.RequestID != "req_123" {
		t.Errorf("toUnifiedResponse() RequestID = %v, want req_123", unifiedResp.Metadata.RequestID)
	}
}
