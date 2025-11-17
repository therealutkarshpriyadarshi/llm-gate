package cache

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/rs/zerolog/log"
)

// OpenAIEmbedder implements the Embedder interface using OpenAI's API
type OpenAIEmbedder struct {
	apiKey     string
	model      string
	httpClient *http.Client

	// Embedding cache to avoid redundant API calls
	cache      map[string][]float32
	cacheMutex sync.RWMutex

	// Rate limiting
	rateLimiter chan struct{}
}

// NewOpenAIEmbedder creates a new OpenAI embedder
func NewOpenAIEmbedder(apiKey, model string) *OpenAIEmbedder {
	if model == "" {
		model = "text-embedding-3-small" // Default model
	}

	return &OpenAIEmbedder{
		apiKey: apiKey,
		model:  model,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		cache:       make(map[string][]float32),
		rateLimiter: make(chan struct{}, 10), // Allow 10 concurrent requests
	}
}

// OpenAI API request/response structures
type embeddingRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
}

type embeddingResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float32 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Model string `json:"model"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

// Embed generates an embedding vector for the given text
func (e *OpenAIEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	// Check cache first
	e.cacheMutex.RLock()
	if cached, ok := e.cache[text]; ok {
		e.cacheMutex.RUnlock()
		log.Debug().Msg("embedding cache hit")
		return cached, nil
	}
	e.cacheMutex.RUnlock()

	// Rate limiting
	select {
	case e.rateLimiter <- struct{}{}:
		defer func() { <-e.rateLimiter }()
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Create request
	reqBody := embeddingRequest{
		Input: text,
		Model: e.model,
	}

	jsonData, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", "https://api.openai.com/v1/embeddings", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", e.apiKey))

	// Send request
	resp, err := e.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	// Read response
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	// Check status code
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("API request failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse response
	var embResp embeddingResponse
	if err := json.Unmarshal(body, &embResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if len(embResp.Data) == 0 {
		return nil, fmt.Errorf("no embeddings returned")
	}

	embedding := embResp.Data[0].Embedding

	// Cache the embedding
	e.cacheMutex.Lock()
	e.cache[text] = embedding
	e.cacheMutex.Unlock()

	log.Debug().
		Int("tokens", embResp.Usage.TotalTokens).
		Int("dimension", len(embedding)).
		Msg("generated embedding")

	return embedding, nil
}

// EmbedBatch generates embeddings for multiple texts
func (e *OpenAIEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	if len(texts) == 0 {
		return nil, fmt.Errorf("no texts provided")
	}

	// For simplicity, we'll embed them one by one
	// In production, you might want to use OpenAI's batch API
	embeddings := make([][]float32, len(texts))

	for i, text := range texts {
		embedding, err := e.Embed(ctx, text)
		if err != nil {
			return nil, fmt.Errorf("failed to embed text %d: %w", i, err)
		}
		embeddings[i] = embedding
	}

	return embeddings, nil
}

// ClearCache clears the embedding cache
func (e *OpenAIEmbedder) ClearCache() {
	e.cacheMutex.Lock()
	defer e.cacheMutex.Unlock()
	e.cache = make(map[string][]float32)
}

// GetCacheSize returns the number of cached embeddings
func (e *OpenAIEmbedder) GetCacheSize() int {
	e.cacheMutex.RLock()
	defer e.cacheMutex.RUnlock()
	return len(e.cache)
}

// MockEmbedder is a mock embedder for testing
type MockEmbedder struct {
	dimension int
}

// NewMockEmbedder creates a mock embedder that generates random embeddings
func NewMockEmbedder(dimension int) *MockEmbedder {
	return &MockEmbedder{
		dimension: dimension,
	}
}

// Embed generates a mock embedding (hash-based deterministic vector)
func (m *MockEmbedder) Embed(ctx context.Context, text string) ([]float32, error) {
	// Generate a deterministic embedding based on text
	// This is just for testing - not suitable for production
	embedding := make([]float32, m.dimension)

	// Use text hash to generate deterministic values
	for i := 0; i < m.dimension; i++ {
		// Simple hash function
		hash := 0
		for j, c := range text {
			hash += int(c) * (i + j + 1)
		}
		embedding[i] = float32(hash%1000) / 1000.0
	}

	return embedding, nil
}

// EmbedBatch generates mock embeddings for multiple texts
func (m *MockEmbedder) EmbedBatch(ctx context.Context, texts []string) ([][]float32, error) {
	embeddings := make([][]float32, len(texts))
	for i, text := range texts {
		embedding, err := m.Embed(ctx, text)
		if err != nil {
			return nil, err
		}
		embeddings[i] = embedding
	}
	return embeddings, nil
}
