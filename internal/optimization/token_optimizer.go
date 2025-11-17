package optimization

import (
	"regexp"
	"strings"

	"github.com/yourusername/llm-gate/internal/core/models"
)

// OptimizerConfig holds configuration for token optimization
type OptimizerConfig struct {
	// EnableCompression enables prompt compression
	EnableCompression bool

	// EnableTruncation enables response truncation
	EnableTruncation bool

	// MaxPromptTokens is the maximum number of tokens for prompts
	MaxPromptTokens int

	// MaxResponseTokens is the maximum number of tokens for responses
	MaxResponseTokens int

	// EnableSmartTruncation enables smart truncation (preserves important parts)
	EnableSmartTruncation bool
}

// TokenOptimizer handles token optimization
type TokenOptimizer struct {
	config *OptimizerConfig
}

// NewTokenOptimizer creates a new token optimizer
func NewTokenOptimizer(config *OptimizerConfig) *TokenOptimizer {
	if config == nil {
		config = &OptimizerConfig{
			EnableCompression:     true,
			EnableTruncation:      true,
			MaxPromptTokens:       4000,
			MaxResponseTokens:     2000,
			EnableSmartTruncation: true,
		}
	}
	return &TokenOptimizer{
		config: config,
	}
}

// OptimizeRequest optimizes a request to reduce token usage
func (o *TokenOptimizer) OptimizeRequest(req *models.UnifiedRequest) (*models.UnifiedRequest, error) {
	optimized := *req

	if o.config.EnableCompression {
		o.compressMessages(&optimized)
	}

	if o.config.EnableTruncation {
		o.truncateMessages(&optimized, o.config.MaxPromptTokens)
	}

	return &optimized, nil
}

// compressMessages compresses messages by removing unnecessary whitespace and content
func (o *TokenOptimizer) compressMessages(req *models.UnifiedRequest) {
	for i := range req.Messages {
		// Remove excessive whitespace
		req.Messages[i].Content = o.compressWhitespace(req.Messages[i].Content)

		// Remove markdown formatting if not essential
		if o.shouldRemoveFormatting(req.Messages[i].Content) {
			req.Messages[i].Content = o.removeMarkdownFormatting(req.Messages[i].Content)
		}
	}
}

// compressWhitespace removes excessive whitespace
func (o *TokenOptimizer) compressWhitespace(text string) string {
	// Replace multiple spaces with single space
	re := regexp.MustCompile(`\s+`)
	text = re.ReplaceAllString(text, " ")

	// Remove leading and trailing whitespace
	text = strings.TrimSpace(text)

	// Remove excessive newlines (more than 2)
	re = regexp.MustCompile(`\n{3,}`)
	text = re.ReplaceAllString(text, "\n\n")

	return text
}

// shouldRemoveFormatting determines if formatting should be removed
func (o *TokenOptimizer) shouldRemoveFormatting(text string) bool {
	// Don't remove formatting from code blocks
	if strings.Contains(text, "```") {
		return false
	}

	// Remove formatting if text is long and contains markdown
	return len(text) > 500 && (strings.Contains(text, "**") || strings.Contains(text, "__"))
}

// removeMarkdownFormatting removes markdown formatting
func (o *TokenOptimizer) removeMarkdownFormatting(text string) string {
	// Remove bold
	re := regexp.MustCompile(`\*\*([^\*]+)\*\*`)
	text = re.ReplaceAllString(text, "$1")

	re = regexp.MustCompile(`__([^_]+)__`)
	text = re.ReplaceAllString(text, "$1")

	// Remove italic
	re = regexp.MustCompile(`\*([^\*]+)\*`)
	text = re.ReplaceAllString(text, "$1")

	re = regexp.MustCompile(`_([^_]+)_`)
	text = re.ReplaceAllString(text, "$1")

	// Remove headers
	re = regexp.MustCompile(`^#+\s+`)
	text = re.ReplaceAllString(text, "")

	return text
}

// truncateMessages truncates messages to fit within token limit
func (o *TokenOptimizer) truncateMessages(req *models.UnifiedRequest, maxTokens int) {
	// Estimate tokens (rough approximation: 1 token ≈ 4 characters)
	estimatedTokens := o.estimateTokens(req)

	if estimatedTokens <= maxTokens {
		return
	}

	if o.config.EnableSmartTruncation {
		o.smartTruncate(req, maxTokens)
	} else {
		o.simpleTruncate(req, maxTokens)
	}
}

// estimateTokens estimates the number of tokens in a request
func (o *TokenOptimizer) estimateTokens(req *models.UnifiedRequest) int {
	totalChars := 0
	for _, msg := range req.Messages {
		totalChars += len(msg.Content)
	}

	// Rough approximation: 1 token ≈ 4 characters
	return totalChars / 4
}

// smartTruncate truncates messages intelligently, preserving important content
func (o *TokenOptimizer) smartTruncate(req *models.UnifiedRequest, maxTokens int) {
	if len(req.Messages) == 0 {
		return
	}

	// Always keep the system message and the latest user message
	systemMsg := ""
	latestUserMsg := ""
	var middleMessages []models.Message

	for i, msg := range req.Messages {
		if msg.Role == "system" && i == 0 {
			systemMsg = msg.Content
		} else if msg.Role == "user" && i == len(req.Messages)-1 {
			latestUserMsg = msg.Content
		} else {
			middleMessages = append(middleMessages, msg)
		}
	}

	// Estimate tokens for system and latest user message
	preservedTokens := (len(systemMsg) + len(latestUserMsg)) / 4
	remainingTokens := maxTokens - preservedTokens

	if remainingTokens <= 0 {
		// If system and latest message exceed limit, truncate latest message
		targetChars := maxTokens * 4
		if len(systemMsg) > 0 {
			targetChars -= len(systemMsg)
		}
		latestUserMsg = o.truncateString(latestUserMsg, targetChars)

		req.Messages = []models.Message{}
		if systemMsg != "" {
			req.Messages = append(req.Messages, models.Message{Role: "system", Content: systemMsg})
		}
		req.Messages = append(req.Messages, models.Message{Role: "user", Content: latestUserMsg})
		return
	}

	// Truncate middle messages to fit remaining tokens
	targetChars := remainingTokens * 4
	currentChars := 0
	for _, msg := range middleMessages {
		currentChars += len(msg.Content)
	}

	if currentChars > targetChars {
		// Remove oldest messages until we fit
		for currentChars > targetChars && len(middleMessages) > 0 {
			currentChars -= len(middleMessages[0].Content)
			middleMessages = middleMessages[1:]
		}
	}

	// Reconstruct messages
	req.Messages = []models.Message{}
	if systemMsg != "" {
		req.Messages = append(req.Messages, models.Message{Role: "system", Content: systemMsg})
	}
	req.Messages = append(req.Messages, middleMessages...)
	req.Messages = append(req.Messages, models.Message{Role: "user", Content: latestUserMsg})
}

// simpleTruncate truncates messages by removing oldest messages
func (o *TokenOptimizer) simpleTruncate(req *models.UnifiedRequest, maxTokens int) {
	targetChars := maxTokens * 4

	// Remove oldest messages until we fit
	for o.estimateTokens(req)*4 > targetChars && len(req.Messages) > 1 {
		req.Messages = req.Messages[1:]
	}

	// If still too large, truncate the content
	if o.estimateTokens(req)*4 > targetChars && len(req.Messages) > 0 {
		lastIdx := len(req.Messages) - 1
		req.Messages[lastIdx].Content = o.truncateString(req.Messages[lastIdx].Content, targetChars)
	}
}

// truncateString truncates a string to a target length
func (o *TokenOptimizer) truncateString(s string, targetLength int) string {
	if len(s) <= targetLength {
		return s
	}

	// Truncate and add ellipsis
	return s[:targetLength-3] + "..."
}

// OptimizeContextWindow optimizes a conversation context window
func (o *TokenOptimizer) OptimizeContextWindow(messages []models.Message, maxTokens int) []models.Message {
	req := &models.UnifiedRequest{Messages: messages}
	o.truncateMessages(req, maxTokens)
	return req.Messages
}

// CalculateTokenSavings calculates the token savings from optimization
func (o *TokenOptimizer) CalculateTokenSavings(originalReq, optimizedReq *models.UnifiedRequest) int {
	originalTokens := o.estimateTokens(originalReq)
	optimizedTokens := o.estimateTokens(optimizedReq)
	return originalTokens - optimizedTokens
}
