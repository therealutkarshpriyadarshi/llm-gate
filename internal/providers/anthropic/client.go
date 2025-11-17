package anthropic

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/interfaces"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
	"github.com/therealutkarshpriyadarshi/llm-gate/pkg/utils"
)

// Client implements the LLMProvider interface for Anthropic
type Client struct {
	config     *Config
	httpClient *http.Client
}

// NewClient creates a new Anthropic client
func NewClient(config *Config) (*Client, error) {
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return &Client{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}, nil
}

// Name returns the provider name
func (c *Client) Name() models.ProviderType {
	return models.ProviderAnthropic
}

// ChatCompletion sends a chat completion request
func (c *Client) ChatCompletion(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error) {
	// Validate the request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Convert to Anthropic format
	anthropicReq, err := c.toAnthropicRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}

	// Make the API request with retries
	var anthropicResp AnthropicResponse
	err = utils.RetryWithBackoff(ctx, c.config.MaxRetries, c.config.RetryDelay, func() error {
		resp, err := c.makeRequest(ctx, "/v1/messages", anthropicReq)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Check for errors
		if resp.StatusCode != http.StatusOK {
			return c.handleErrorResponse(resp)
		}

		// Decode response
		if err := json.NewDecoder(resp.Body).Decode(&anthropicResp); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Convert to unified format
	return c.toUnifiedResponse(&anthropicResp, req), nil
}

// ChatCompletionStream sends a streaming chat completion request
func (c *Client) ChatCompletionStream(ctx context.Context, req *models.ChatRequest) (interfaces.StreamReader, error) {
	// Validate the request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Convert to Anthropic format and enable streaming
	anthropicReq, err := c.toAnthropicRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}
	anthropicReq.Stream = true

	// Make the API request
	resp, err := c.makeRequest(ctx, "/v1/messages", anthropicReq)
	if err != nil {
		return nil, err
	}

	// Check for errors
	if resp.StatusCode != http.StatusOK {
		defer resp.Body.Close()
		return nil, c.handleErrorResponse(resp)
	}

	// Return stream reader
	return &StreamReaderImpl{
		reader: resp.Body,
		model:  req.Model,
	}, nil
}

// GetCapabilities returns the provider's capabilities
func (c *Client) GetCapabilities() models.ProviderCapabilities {
	modelInfos := make([]models.ModelInfo, 0, len(ModelPricing))
	for modelID, pricing := range ModelPricing {
		modelInfos = append(modelInfos, models.ModelInfo{
			ID:                modelID,
			Name:              modelID,
			MaxTokens:         pricing.MaxTokens,
			InputCostPer1K:    pricing.InputCostPer1K,
			OutputCostPer1K:   pricing.OutputCostPer1K,
			SupportsStreaming: true,
			SupportsFunctions: false, // Anthropic uses tools, not function calling
		})
	}

	return models.ProviderCapabilities{
		Name:              models.ProviderAnthropic,
		SupportsStreaming: true,
		SupportsFunctions: false,
		SupportsVision:    true, // Claude 3 supports vision
		MaxTokens:         200000,
		Models:            modelInfos,
		DefaultTimeout:    c.config.Timeout,
	}
}

// HealthCheck checks if the provider is healthy
func (c *Client) HealthCheck(ctx context.Context) models.ProviderHealth {
	start := time.Now()
	health := models.ProviderHealth{
		Name:      models.ProviderAnthropic,
		LastCheck: start,
	}

	// Make a simple request to check connectivity
	req := &models.ChatRequest{
		Model: "claude-3-haiku-20240307",
		Messages: []models.Message{
			{Role: "user", Content: "ping"},
		},
		MaxTokens: utils.IntPtr(1),
	}

	_, err := c.ChatCompletion(ctx, req)
	latency := time.Since(start)

	if err != nil {
		health.Healthy = false
		health.Message = err.Error()
	} else {
		health.Healthy = true
		health.Latency = latency
		health.Message = "OK"
	}

	return health
}

// GetModelInfo returns information about a specific model
func (c *Client) GetModelInfo(modelID string) (models.ModelInfo, error) {
	inputCost, outputCost, maxTokens := GetModelPricing(modelID)

	return models.ModelInfo{
		ID:                modelID,
		Name:              modelID,
		MaxTokens:         maxTokens,
		InputCostPer1K:    inputCost,
		OutputCostPer1K:   outputCost,
		SupportsStreaming: true,
		SupportsFunctions: false,
	}, nil
}

// Close closes the client
func (c *Client) Close() error {
	c.httpClient.CloseIdleConnections()
	return nil
}

// makeRequest makes an HTTP request to the Anthropic API
func (c *Client) makeRequest(ctx context.Context, path string, body interface{}) (*http.Response, error) {
	// Marshal request body
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	url := c.config.BaseURL + path
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", c.config.APIKey)
	req.Header.Set("anthropic-version", c.config.Version)

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// handleErrorResponse handles error responses from the API
func (c *Client) handleErrorResponse(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read error response: %w", err)
	}

	var anthropicErr AnthropicError
	if err := json.Unmarshal(body, &anthropicErr); err != nil {
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return fmt.Errorf("Anthropic API error: %s (type: %s)",
		anthropicErr.Error.Message,
		anthropicErr.Error.Type,
	)
}

// toAnthropicRequest converts a unified request to Anthropic format
func (c *Client) toAnthropicRequest(req *models.ChatRequest) (*AnthropicRequest, error) {
	messages := make([]AnthropicMessage, 0)
	var systemPrompt string

	// Extract system message and convert others
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			// Anthropic uses a separate system parameter
			if systemPrompt != "" {
				systemPrompt += "\n\n" + msg.Content
			} else {
				systemPrompt = msg.Content
			}
		} else {
			messages = append(messages, AnthropicMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	// Ensure we have at least one message
	if len(messages) == 0 {
		return nil, fmt.Errorf("at least one non-system message is required")
	}

	// Set max_tokens (required by Anthropic)
	maxTokens := 4096
	if req.MaxTokens != nil {
		maxTokens = *req.MaxTokens
	}

	anthropicReq := &AnthropicRequest{
		Model:     req.Model,
		Messages:  messages,
		MaxTokens: maxTokens,
		System:    systemPrompt,
	}

	// Map optional parameters
	if req.Temperature != nil {
		anthropicReq.Temperature = req.Temperature
	}
	if req.TopP != nil {
		anthropicReq.TopP = req.TopP
	}
	if req.Stop != nil {
		anthropicReq.StopSeqs = req.Stop
	}

	return anthropicReq, nil
}

// toUnifiedResponse converts an Anthropic response to unified format
func (c *Client) toUnifiedResponse(resp *AnthropicResponse, originalReq *models.ChatRequest) *models.ChatResponse {
	// Combine all content blocks
	var content string
	for _, block := range resp.Content {
		if block.Type == "text" {
			content += block.Text
		}
	}

	choices := []models.Choice{
		{
			Index: 0,
			Message: models.Message{
				Role:    resp.Role,
				Content: content,
			},
			FinishReason: resp.StopReason,
		},
	}

	usage := models.Usage{
		PromptTokens:     resp.Usage.InputTokens,
		CompletionTokens: resp.Usage.OutputTokens,
		TotalTokens:      resp.Usage.InputTokens + resp.Usage.OutputTokens,
	}

	// Calculate cost
	modelInfo, _ := c.GetModelInfo(resp.Model)
	costInfo := models.CalculateCost(usage, modelInfo, models.ProviderAnthropic)

	return &models.ChatResponse{
		ID:       resp.ID,
		Object:   "chat.completion",
		Created:  time.Now().Unix(),
		Model:    resp.Model,
		Choices:  choices,
		Usage:    usage,
		Provider: string(models.ProviderAnthropic),
		Metadata: models.ResponseMetadata{
			RequestID: originalReq.Metadata.RequestID,
			Cost:      costInfo.TotalCost,
			Timestamp: time.Now(),
		},
	}
}
