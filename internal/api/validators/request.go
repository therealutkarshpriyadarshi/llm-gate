package validators

import (
	"fmt"
	"strings"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
)

// ValidateChatRequest validates a chat completion request
func ValidateChatRequest(req *models.ChatRequest) error {
	if req == nil {
		return fmt.Errorf("request cannot be nil")
	}

	// Use the built-in validation
	if err := req.Validate(); err != nil {
		return err
	}

	// Additional validation
	if len(req.Messages) > 100 {
		return fmt.Errorf("too many messages: maximum 100 allowed, got %d", len(req.Messages))
	}

	// Validate message content length
	for i, msg := range req.Messages {
		if len(msg.Content) > 100000 {
			return fmt.Errorf("message %d content too long: maximum 100,000 characters allowed", i)
		}
	}

	return nil
}

// SanitizeChatRequest sanitizes a chat completion request
func SanitizeChatRequest(req *models.ChatRequest) {
	if req == nil {
		return
	}

	// Trim whitespace from messages
	for i := range req.Messages {
		req.Messages[i].Content = strings.TrimSpace(req.Messages[i].Content)
		req.Messages[i].Role = strings.ToLower(strings.TrimSpace(req.Messages[i].Role))
		req.Messages[i].Name = strings.TrimSpace(req.Messages[i].Name)
	}

	// Trim model name
	req.Model = strings.TrimSpace(req.Model)

	// Trim user
	req.User = strings.TrimSpace(req.User)
}
