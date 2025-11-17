package vertex

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

// Client implements the LLMProvider interface for Google Vertex AI
type Client struct {
	config     *Config
	httpClient *http.Client
	accessToken string
}

// NewClient creates a new Google Vertex AI client
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
	return models.ProviderVertex
}

// ChatCompletion sends a chat completion request
func (c *Client) ChatCompletion(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error) {
	// Validate the request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Convert to Vertex format
	vertexReq, err := c.toVertexRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}

	// Make the API request with retries
	var vertexResp VertexResponse
	err = utils.RetryWithBackoff(ctx, c.config.MaxRetries, c.config.RetryDelay, func() error {
		resp, err := c.makeRequest(ctx, req.Model, vertexReq, false)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Check for errors
		if resp.StatusCode != http.StatusOK {
			return c.handleErrorResponse(resp)
		}

		// Decode response
		if err := json.NewDecoder(resp.Body).Decode(&vertexResp); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Convert to unified format
	return c.toUnifiedResponse(&vertexResp, req), nil
}

// ChatCompletionStream sends a streaming chat completion request
func (c *Client) ChatCompletionStream(ctx context.Context, req *models.ChatRequest) (interfaces.StreamReader, error) {
	// Validate the request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Convert to Vertex format
	vertexReq, err := c.toVertexRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}

	// Make the API request with streaming
	resp, err := c.makeRequest(ctx, req.Model, vertexReq, true)
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
		Name:              models.ProviderVertex,
		SupportsStreaming: true,
		SupportsFunctions: true,
		SupportsVision:    true,
		MaxTokens:         1048576, // Gemini 1.5 Pro supports 1M tokens
		Models:            modelInfos,
		DefaultTimeout:    c.config.Timeout,
	}
}

// HealthCheck checks if the provider is healthy
func (c *Client) HealthCheck(ctx context.Context) models.ProviderHealth {
	start := time.Now()
	health := models.ProviderHealth{
		Name:      models.ProviderVertex,
		LastCheck: start,
	}

	// Make a simple request to check connectivity
	req := &models.ChatRequest{
		Model: "gemini-1.5-flash",
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

// makeRequest makes an HTTP request to the Vertex AI API
func (c *Client) makeRequest(ctx context.Context, modelID string, body interface{}, stream bool) (*http.Response, error) {
	// Marshal request body
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build URL
	endpoint := "generateContent"
	if stream {
		endpoint = "streamGenerateContent?alt=sse"
	}
	url := fmt.Sprintf("https://%s-aiplatform.googleapis.com/v1/projects/%s/locations/%s/publishers/google/models/%s:%s",
		c.config.Location,
		c.config.ProjectID,
		c.config.Location,
		modelID,
		endpoint,
	)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")

	// Add authentication
	if c.config.APIKey != "" {
		req.Header.Set("X-Goog-Api-Key", c.config.APIKey)
	} else {
		// For service account authentication, we would need to get an access token
		// This is a simplified version - in production, use Google Cloud SDK
		req.Header.Set("Authorization", "Bearer "+c.accessToken)
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

	var vertexErr VertexError
	if err := json.Unmarshal(body, &vertexErr); err != nil {
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return fmt.Errorf("Vertex AI API error: %s (code: %d, status: %s)",
		vertexErr.Error.Message,
		vertexErr.Error.Code,
		vertexErr.Error.Status,
	)
}

// toVertexRequest converts a unified request to Vertex AI format
func (c *Client) toVertexRequest(req *models.ChatRequest) (*VertexRequest, error) {
	contents := make([]VertexMessage, len(req.Messages))
	for i, msg := range req.Messages {
		role := msg.Role
		// Vertex uses "user" and "model" roles
		if role == "assistant" {
			role = "model"
		}
		if role == "system" {
			// Vertex doesn't have a system role, prepend to first user message
			role = "user"
		}

		contents[i] = VertexMessage{
			Role: role,
			Parts: []VertexPart{
				{Text: msg.Content},
			},
		}
	}

	vertexReq := &VertexRequest{
		Contents: contents,
	}

	// Add generation config if parameters are set
	if req.Temperature != nil || req.MaxTokens != nil || req.TopP != nil || req.Stop != nil {
		vertexReq.GenerationConfig = &VertexGenConfig{
			Temperature:     req.Temperature,
			MaxOutputTokens: req.MaxTokens,
			TopP:            req.TopP,
			StopSequences:   req.Stop,
		}
	}

	return vertexReq, nil
}

// toUnifiedResponse converts a Vertex AI response to unified format
func (c *Client) toUnifiedResponse(resp *VertexResponse, originalReq *models.ChatRequest) *models.ChatResponse {
	choices := make([]models.Choice, len(resp.Candidates))
	for i, candidate := range resp.Candidates {
		// Combine all parts into a single content string
		var content string
		for _, part := range candidate.Content.Parts {
			content += part.Text
		}

		role := candidate.Content.Role
		if role == "model" {
			role = "assistant"
		}

		choices[i] = models.Choice{
			Index: candidate.Index,
			Message: models.Message{
				Role:    role,
				Content: content,
			},
			FinishReason: candidate.FinishReason,
		}
	}

	usage := models.Usage{
		PromptTokens:     resp.UsageMetadata.PromptTokenCount,
		CompletionTokens: resp.UsageMetadata.CandidatesTokenCount,
		TotalTokens:      resp.UsageMetadata.TotalTokenCount,
	}

	// Calculate cost
	modelInfo, _ := c.GetModelInfo(originalReq.Model)
	costInfo := models.CalculateCost(usage, modelInfo, models.ProviderVertex)

	return &models.ChatResponse{
		ID:       fmt.Sprintf("vertex-%d", time.Now().Unix()),
		Object:   "chat.completion",
		Created:  time.Now().Unix(),
		Model:    originalReq.Model,
		Choices:  choices,
		Usage:    usage,
		Provider: string(models.ProviderVertex),
		Metadata: models.ResponseMetadata{
			RequestID: originalReq.Metadata.RequestID,
			Cost:      costInfo.TotalCost,
			Timestamp: time.Now(),
		},
	}
}
