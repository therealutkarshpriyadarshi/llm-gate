package openai

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

// Client implements the LLMProvider interface for OpenAI
type Client struct {
	config     *Config
	httpClient *http.Client
}

// NewClient creates a new OpenAI client
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
	return models.ProviderOpenAI
}

// ChatCompletion sends a chat completion request
func (c *Client) ChatCompletion(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error) {
	// Validate the request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Convert to OpenAI format
	openAIReq := c.toOpenAIRequest(req)

	// Make the API request with retries
	var openAIResp OpenAIChatResponse
	err := utils.RetryWithBackoff(ctx, c.config.MaxRetries, c.config.RetryDelay, func() error {
		resp, err := c.makeRequest(ctx, "/chat/completions", openAIReq)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Check for errors
		if resp.StatusCode != http.StatusOK {
			return c.handleErrorResponse(resp)
		}

		// Decode response
		if err := json.NewDecoder(resp.Body).Decode(&openAIResp); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Convert to unified format
	return c.toUnifiedResponse(&openAIResp, req), nil
}

// ChatCompletionStream sends a streaming chat completion request
func (c *Client) ChatCompletionStream(ctx context.Context, req *models.ChatRequest) (interfaces.StreamReader, error) {
	// Validate the request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Convert to OpenAI format and enable streaming
	openAIReq := c.toOpenAIRequest(req)
	openAIReq.Stream = true

	// Make the API request
	resp, err := c.makeRequest(ctx, "/chat/completions", openAIReq)
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
			SupportsFunctions: true,
		})
	}

	return models.ProviderCapabilities{
		Name:              models.ProviderOpenAI,
		SupportsStreaming: true,
		SupportsFunctions: true,
		SupportsVision:    true,
		MaxTokens:         128000,
		Models:            modelInfos,
		DefaultTimeout:    c.config.Timeout,
	}
}

// HealthCheck checks if the provider is healthy
func (c *Client) HealthCheck(ctx context.Context) models.ProviderHealth {
	start := time.Now()
	health := models.ProviderHealth{
		Name:      models.ProviderOpenAI,
		LastCheck: start,
	}

	// Make a simple request to check connectivity
	req := &models.ChatRequest{
		Model: "gpt-3.5-turbo",
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
		SupportsFunctions: true,
	}, nil
}

// Close closes the client
func (c *Client) Close() error {
	c.httpClient.CloseIdleConnections()
	return nil
}

// makeRequest makes an HTTP request to the OpenAI API
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
	req.Header.Set("Authorization", "Bearer "+c.config.APIKey)
	if c.config.Organization != "" {
		req.Header.Set("OpenAI-Organization", c.config.Organization)
	}

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

	var openAIErr OpenAIError
	if err := json.Unmarshal(body, &openAIErr); err != nil {
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return fmt.Errorf("OpenAI API error: %s (type: %s, code: %s)",
		openAIErr.Error.Message,
		openAIErr.Error.Type,
		openAIErr.Error.Code,
	)
}

// toOpenAIRequest converts a unified request to OpenAI format
func (c *Client) toOpenAIRequest(req *models.ChatRequest) *OpenAIChatRequest {
	messages := make([]OpenAIMessage, len(req.Messages))
	for i, msg := range req.Messages {
		messages[i] = OpenAIMessage{
			Role:    msg.Role,
			Content: msg.Content,
			Name:    msg.Name,
		}
	}

	return &OpenAIChatRequest{
		Model:            req.Model,
		Messages:         messages,
		Temperature:      req.Temperature,
		MaxTokens:        req.MaxTokens,
		TopP:             req.TopP,
		N:                req.N,
		Stream:           req.Stream,
		Stop:             req.Stop,
		PresencePenalty:  req.PresencePenalty,
		FrequencyPenalty: req.FrequencyPenalty,
		User:             req.User,
	}
}

// toUnifiedResponse converts an OpenAI response to unified format
func (c *Client) toUnifiedResponse(resp *OpenAIChatResponse, originalReq *models.ChatRequest) *models.ChatResponse {
	choices := make([]models.Choice, len(resp.Choices))
	for i, choice := range resp.Choices {
		choices[i] = models.Choice{
			Index: choice.Index,
			Message: models.Message{
				Role:    choice.Message.Role,
				Content: choice.Message.Content,
				Name:    choice.Message.Name,
			},
			FinishReason: choice.FinishReason,
		}
	}

	usage := models.Usage{
		PromptTokens:     resp.Usage.PromptTokens,
		CompletionTokens: resp.Usage.CompletionTokens,
		TotalTokens:      resp.Usage.TotalTokens,
	}

	// Calculate cost
	modelInfo, _ := c.GetModelInfo(resp.Model)
	costInfo := models.CalculateCost(usage, modelInfo, models.ProviderOpenAI)

	return &models.ChatResponse{
		ID:       resp.ID,
		Object:   resp.Object,
		Created:  resp.Created,
		Model:    resp.Model,
		Choices:  choices,
		Usage:    usage,
		Provider: string(models.ProviderOpenAI),
		Metadata: models.ResponseMetadata{
			RequestID: originalReq.Metadata.RequestID,
			Cost:      costInfo.TotalCost,
			Timestamp: time.Now(),
		},
	}
}
