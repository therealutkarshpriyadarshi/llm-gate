package cache

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
)

// RequestNormalizer normalizes requests for consistent cache key generation
type RequestNormalizer struct{}

// NewRequestNormalizer creates a new request normalizer
func NewRequestNormalizer() *RequestNormalizer {
	return &RequestNormalizer{}
}

// NormalizeRequest converts a request to a normalized string representation
// This ensures that semantically similar requests produce similar embeddings
func (n *RequestNormalizer) NormalizeRequest(req *models.ChatRequest) (string, error) {
	var parts []string

	// Add model (normalized to lowercase)
	if req.Model != "" {
		parts = append(parts, fmt.Sprintf("model:%s", strings.ToLower(req.Model)))
	}

	// Add messages (most important for semantic matching)
	if len(req.Messages) > 0 {
		messagesText := n.normalizeMessages(req.Messages)
		parts = append(parts, fmt.Sprintf("messages:%s", messagesText))
	}

	// Add temperature (rounded to 1 decimal place for fuzzy matching)
	if req.Temperature != nil {
		temp := fmt.Sprintf("%.1f", *req.Temperature)
		parts = append(parts, fmt.Sprintf("temp:%s", temp))
	}

	// Add max_tokens if specified
	if req.MaxTokens != nil {
		parts = append(parts, fmt.Sprintf("max_tokens:%d", *req.MaxTokens))
	}

	// Join all parts with delimiter
	normalized := strings.Join(parts, "|")
	return normalized, nil
}

// normalizeMessages converts messages to a normalized string
func (n *RequestNormalizer) normalizeMessages(messages []models.Message) string {
	var msgParts []string
	for _, msg := range messages {
		// Normalize role to lowercase
		role := strings.ToLower(msg.Role)
		// Trim whitespace from content
		content := strings.TrimSpace(msg.Content)
		msgParts = append(msgParts, fmt.Sprintf("%s:%s", role, content))
	}
	return strings.Join(msgParts, ";")
}

// GenerateCacheKey generates a deterministic cache key from a request
// This is used for exact matching (non-semantic)
func (n *RequestNormalizer) GenerateCacheKey(req *models.ChatRequest) (string, error) {
	// Create a map of relevant fields for key generation
	keyData := map[string]interface{}{
		"model":    req.Model,
		"messages": req.Messages,
	}

	// Add optional fields if present
	if req.Temperature != nil {
		keyData["temperature"] = *req.Temperature
	}
	if req.MaxTokens != nil {
		keyData["max_tokens"] = *req.MaxTokens
	}
	if req.TopP != nil {
		keyData["top_p"] = *req.TopP
	}

	// Convert to JSON for consistent ordering
	jsonData, err := json.Marshal(keyData)
	if err != nil {
		return "", fmt.Errorf("failed to marshal key data: %w", err)
	}

	// Generate SHA-256 hash
	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:]), nil
}

// ExtractSearchableText extracts the main text content for embedding
// This focuses on the most semantically relevant parts of the request
func (n *RequestNormalizer) ExtractSearchableText(req *models.ChatRequest) string {
	var parts []string

	// Extract message content (this is the primary semantic content)
	for _, msg := range req.Messages {
		if msg.Content != "" {
			parts = append(parts, msg.Content)
		}
	}

	return strings.Join(parts, "\n")
}

// SortedMapKeys returns sorted keys from a map for deterministic ordering
func SortedMapKeys(m map[string]interface{}) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
