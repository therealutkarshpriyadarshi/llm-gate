package bedrock

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sort"
	"strings"
	"time"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/interfaces"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
	"github.com/therealutkarshpriyadarshi/llm-gate/pkg/utils"
)

// Client implements the LLMProvider interface for AWS Bedrock
type Client struct {
	config     *Config
	httpClient *http.Client
}

// NewClient creates a new AWS Bedrock client
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
	return models.ProviderBedrock
}

// ChatCompletion sends a chat completion request
func (c *Client) ChatCompletion(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error) {
	// Validate the request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Convert to Bedrock format
	bedrockReq, err := c.toBedrockRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}

	// Make the API request with retries
	var bedrockResp BedrockResponse
	err = utils.RetryWithBackoff(ctx, c.config.MaxRetries, c.config.RetryDelay, func() error {
		resp, err := c.makeRequest(ctx, req.Model, bedrockReq, false)
		if err != nil {
			return err
		}
		defer resp.Body.Close()

		// Check for errors
		if resp.StatusCode != http.StatusOK {
			return c.handleErrorResponse(resp)
		}

		// Decode response
		if err := json.NewDecoder(resp.Body).Decode(&bedrockResp); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	// Convert to unified format
	return c.toUnifiedResponse(&bedrockResp, req), nil
}

// ChatCompletionStream sends a streaming chat completion request
func (c *Client) ChatCompletionStream(ctx context.Context, req *models.ChatRequest) (interfaces.StreamReader, error) {
	// Validate the request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("invalid request: %w", err)
	}

	// Convert to Bedrock format
	bedrockReq, err := c.toBedrockRequest(req)
	if err != nil {
		return nil, fmt.Errorf("failed to convert request: %w", err)
	}

	// Make the API request with streaming
	resp, err := c.makeRequest(ctx, req.Model, bedrockReq, true)
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
			SupportsFunctions: false,
		})
	}

	return models.ProviderCapabilities{
		Name:              models.ProviderBedrock,
		SupportsStreaming: true,
		SupportsFunctions: false,
		SupportsVision:    true,
		MaxTokens:         200000,
		Models:            modelInfos,
		DefaultTimeout:    c.config.Timeout,
	}
}

// HealthCheck checks if the provider is healthy
func (c *Client) HealthCheck(ctx context.Context) models.ProviderHealth {
	start := time.Now()
	health := models.ProviderHealth{
		Name:      models.ProviderBedrock,
		LastCheck: start,
	}

	// Make a simple request to check connectivity
	req := &models.ChatRequest{
		Model: "anthropic.claude-3-haiku-20240307-v1:0",
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

// makeRequest makes an HTTP request to the Bedrock API
func (c *Client) makeRequest(ctx context.Context, modelID string, body interface{}, stream bool) (*http.Response, error) {
	// Marshal request body
	jsonBody, err := json.Marshal(body)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Build endpoint URL
	endpoint := "invoke"
	if stream {
		endpoint = "invoke-with-response-stream"
	}
	url := fmt.Sprintf("https://bedrock-runtime.%s.amazonaws.com/model/%s/%s",
		c.config.Region, modelID, endpoint)

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	// Sign the request with AWS Signature Version 4
	if err := c.signRequest(req, jsonBody); err != nil {
		return nil, fmt.Errorf("failed to sign request: %w", err)
	}

	// Make the request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

// signRequest signs the request using AWS Signature Version 4
func (c *Client) signRequest(req *http.Request, payload []byte) error {
	// Get current time
	t := time.Now().UTC()
	amzDate := t.Format("20060102T150405Z")
	dateStamp := t.Format("20060102")

	// Set required headers
	req.Header.Set("X-Amz-Date", amzDate)
	if c.config.SessionToken != "" {
		req.Header.Set("X-Amz-Security-Token", c.config.SessionToken)
	}

	// Create canonical request
	payloadHash := hashSHA256(payload)
	canonicalRequest := createCanonicalRequest(req, payloadHash)

	// Create string to sign
	credentialScope := fmt.Sprintf("%s/%s/bedrock/aws4_request", dateStamp, c.config.Region)
	stringToSign := createStringToSign(amzDate, credentialScope, canonicalRequest)

	// Calculate signature
	signature := c.calculateSignature(dateStamp, stringToSign)

	// Create authorization header
	authorization := fmt.Sprintf("AWS4-HMAC-SHA256 Credential=%s/%s, SignedHeaders=%s, Signature=%s",
		c.config.AccessKeyID,
		credentialScope,
		"content-type;host;x-amz-date",
		signature,
	)

	req.Header.Set("Authorization", authorization)

	return nil
}

// createCanonicalRequest creates the canonical request string
func createCanonicalRequest(req *http.Request, payloadHash string) string {
	// Canonical URI
	canonicalURI := req.URL.Path
	if canonicalURI == "" {
		canonicalURI = "/"
	}

	// Canonical query string
	canonicalQueryString := req.URL.Query().Encode()

	// Canonical headers
	var headers []string
	headers = append(headers, fmt.Sprintf("content-type:%s", req.Header.Get("Content-Type")))
	headers = append(headers, fmt.Sprintf("host:%s", req.Host))
	headers = append(headers, fmt.Sprintf("x-amz-date:%s", req.Header.Get("X-Amz-Date")))
	sort.Strings(headers)
	canonicalHeaders := strings.Join(headers, "\n") + "\n"

	// Signed headers
	signedHeaders := "content-type;host;x-amz-date"

	// Create canonical request
	return strings.Join([]string{
		req.Method,
		canonicalURI,
		canonicalQueryString,
		canonicalHeaders,
		signedHeaders,
		payloadHash,
	}, "\n")
}

// createStringToSign creates the string to sign
func createStringToSign(amzDate, credentialScope, canonicalRequest string) string {
	return strings.Join([]string{
		"AWS4-HMAC-SHA256",
		amzDate,
		credentialScope,
		hashSHA256([]byte(canonicalRequest)),
	}, "\n")
}

// calculateSignature calculates the AWS Signature V4 signature
func (c *Client) calculateSignature(dateStamp, stringToSign string) string {
	kDate := hmacSHA256([]byte("AWS4"+c.config.SecretAccessKey), []byte(dateStamp))
	kRegion := hmacSHA256(kDate, []byte(c.config.Region))
	kService := hmacSHA256(kRegion, []byte("bedrock"))
	kSigning := hmacSHA256(kService, []byte("aws4_request"))
	signature := hmacSHA256(kSigning, []byte(stringToSign))
	return hex.EncodeToString(signature)
}

// hashSHA256 creates a SHA256 hash
func hashSHA256(data []byte) string {
	hash := sha256.Sum256(data)
	return hex.EncodeToString(hash[:])
}

// hmacSHA256 creates an HMAC-SHA256 signature
func hmacSHA256(key, data []byte) []byte {
	h := hmac.New(sha256.New, key)
	h.Write(data)
	return h.Sum(nil)
}

// handleErrorResponse handles error responses from the API
func (c *Client) handleErrorResponse(resp *http.Response) error {
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read error response: %w", err)
	}

	var bedrockErr BedrockError
	if err := json.Unmarshal(body, &bedrockErr); err != nil {
		return fmt.Errorf("API error (status %d): %s", resp.StatusCode, string(body))
	}

	return fmt.Errorf("Bedrock API error: %s (type: %s)",
		bedrockErr.Message,
		bedrockErr.Type,
	)
}

// toBedrockRequest converts a unified request to Bedrock format
func (c *Client) toBedrockRequest(req *models.ChatRequest) (*BedrockRequest, error) {
	messages := make([]BedrockMessage, 0)
	var systemPrompt string

	// Extract system message and convert others
	for _, msg := range req.Messages {
		if msg.Role == "system" {
			if systemPrompt != "" {
				systemPrompt += "\n\n" + msg.Content
			} else {
				systemPrompt = msg.Content
			}
		} else {
			messages = append(messages, BedrockMessage{
				Role:    msg.Role,
				Content: msg.Content,
			})
		}
	}

	// Set max_tokens
	maxTokens := 4096
	if req.MaxTokens != nil {
		maxTokens = *req.MaxTokens
	}

	bedrockReq := &BedrockRequest{
		AnthropicVersion: "bedrock-2023-05-31",
		Messages:         messages,
		MaxTokens:        maxTokens,
		System:           systemPrompt,
	}

	// Map optional parameters
	if req.Temperature != nil {
		bedrockReq.Temperature = req.Temperature
	}
	if req.TopP != nil {
		bedrockReq.TopP = req.TopP
	}
	if req.Stop != nil {
		bedrockReq.StopSequences = req.Stop
	}

	return bedrockReq, nil
}

// toUnifiedResponse converts a Bedrock response to unified format
func (c *Client) toUnifiedResponse(resp *BedrockResponse, originalReq *models.ChatRequest) *models.ChatResponse {
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
	modelInfo, _ := c.GetModelInfo(originalReq.Model)
	costInfo := models.CalculateCost(usage, modelInfo, models.ProviderBedrock)

	return &models.ChatResponse{
		ID:       resp.ID,
		Object:   "chat.completion",
		Created:  time.Now().Unix(),
		Model:    originalReq.Model,
		Choices:  choices,
		Usage:    usage,
		Provider: string(models.ProviderBedrock),
		Metadata: models.ResponseMetadata{
			RequestID: originalReq.Metadata.RequestID,
			Cost:      costInfo.TotalCost,
			Timestamp: time.Now(),
		},
	}
}
