package routing

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
)

// QueryComplexity represents the complexity level of a query
type QueryComplexity string

const (
	ComplexitySimple  QueryComplexity = "simple"
	ComplexityMedium  QueryComplexity = "medium"
	ComplexityComplex QueryComplexity = "complex"
)

// QueryAnalysis contains the results of query analysis
type QueryAnalysis struct {
	// Complexity is the determined complexity level
	Complexity QueryComplexity

	// EstimatedTokens is the estimated number of input tokens
	EstimatedTokens int

	// Language is the detected language of the query
	Language string

	// RequiresReasoning indicates if the query requires multi-step reasoning
	RequiresReasoning bool

	// RequiresCodeGeneration indicates if code generation is needed
	RequiresCodeGeneration bool

	// RequiresLongContext indicates if long context is needed
	RequiresLongContext bool

	// Categories are the query categories (e.g., "math", "code", "creative")
	Categories []string
}

// QueryAnalyzer analyzes queries to determine routing decisions
type QueryAnalyzer struct {
	// Configuration
	simpleThreshold  int
	complexThreshold int

	// Regular expressions for pattern matching
	codePattern      *regexp.Regexp
	mathPattern      *regexp.Regexp
	reasoningPattern *regexp.Regexp
}

// NewQueryAnalyzer creates a new query analyzer
func NewQueryAnalyzer() *QueryAnalyzer {
	return &QueryAnalyzer{
		simpleThreshold:  100,  // tokens
		complexThreshold: 1000, // tokens
		codePattern:      regexp.MustCompile(`(?i)(write|create|generate|implement|code|function|class|program|script|debug|fix)`),
		mathPattern:      regexp.MustCompile(`(?i)(calculate|solve|equation|formula|math|compute|sum|average|integrate|derivative)`),
		reasoningPattern: regexp.MustCompile(`(?i)(why|how|explain|analyze|compare|evaluate|reason|think|understand|because)`),
	}
}

// Analyze analyzes a chat request and returns the query analysis
func (qa *QueryAnalyzer) Analyze(req *models.ChatRequest) *QueryAnalysis {
	analysis := &QueryAnalysis{
		Categories: []string{},
	}

	// Combine all messages for analysis
	fullText := qa.combineMessages(req.Messages)

	// Estimate tokens (rough approximation: 1 token ≈ 4 characters)
	analysis.EstimatedTokens = qa.estimateTokens(fullText)

	// Detect language
	analysis.Language = qa.detectLanguage(fullText)

	// Determine complexity
	analysis.Complexity = qa.determineComplexity(fullText, analysis.EstimatedTokens)

	// Check for code generation
	analysis.RequiresCodeGeneration = qa.codePattern.MatchString(fullText)
	if analysis.RequiresCodeGeneration {
		analysis.Categories = append(analysis.Categories, "code")
	}

	// Check for mathematical operations
	if qa.mathPattern.MatchString(fullText) {
		analysis.Categories = append(analysis.Categories, "math")
	}

	// Check for reasoning requirements
	analysis.RequiresReasoning = qa.reasoningPattern.MatchString(fullText) ||
		qa.hasMultipleSteps(fullText)
	if analysis.RequiresReasoning {
		analysis.Categories = append(analysis.Categories, "reasoning")
	}

	// Check for long context
	analysis.RequiresLongContext = len(req.Messages) > 10 ||
		analysis.EstimatedTokens > qa.complexThreshold

	// Determine if creative writing
	if qa.isCreativeWriting(fullText) {
		analysis.Categories = append(analysis.Categories, "creative")
	}

	return analysis
}

// combineMessages combines all messages into a single text
func (qa *QueryAnalyzer) combineMessages(messages []models.Message) string {
	var parts []string
	for _, msg := range messages {
		parts = append(parts, msg.Content)
	}
	return strings.Join(parts, " ")
}

// estimateTokens estimates the number of tokens in the text
// This is a rough approximation. For production, use a proper tokenizer
func (qa *QueryAnalyzer) estimateTokens(text string) int {
	// Rough estimate: 1 token ≈ 4 characters for English
	// Count words and adjust
	words := strings.Fields(text)
	return len(words) + len(text)/4
}

// detectLanguage detects the primary language of the text
// This is a simple implementation. For production, consider using a proper language detection library
func (qa *QueryAnalyzer) detectLanguage(text string) string {
	// Count character types
	var latin, cyrillic, cjk, other int

	for _, r := range text {
		switch {
		case unicode.In(r, unicode.Latin):
			latin++
		case unicode.In(r, unicode.Cyrillic):
			cyrillic++
		case unicode.In(r, unicode.Han, unicode.Hiragana, unicode.Katakana, unicode.Hangul):
			cjk++
		default:
			other++
		}
	}

	total := latin + cyrillic + cjk + other
	if total == 0 {
		return "unknown"
	}

	// Determine primary script
	if float64(latin)/float64(total) > 0.5 {
		return "en" // Assume English for Latin script
	} else if float64(cyrillic)/float64(total) > 0.5 {
		return "ru" // Russian for Cyrillic
	} else if float64(cjk)/float64(total) > 0.5 {
		return "zh" // Chinese/Japanese/Korean
	}

	return "unknown"
}

// determineComplexity determines the complexity of the query
func (qa *QueryAnalyzer) determineComplexity(text string, tokens int) QueryComplexity {
	// Token-based classification
	if tokens < qa.simpleThreshold {
		return ComplexitySimple
	} else if tokens < qa.complexThreshold {
		return ComplexityMedium
	}

	// Check for complexity indicators
	complexityScore := 0

	// Long sentences indicate complexity
	sentences := strings.Split(text, ".")
	for _, sentence := range sentences {
		words := strings.Fields(sentence)
		if len(words) > 20 {
			complexityScore++
		}
	}

	// Multiple questions
	questionCount := strings.Count(text, "?")
	if questionCount > 2 {
		complexityScore++
	}

	// Technical terms
	technicalPatterns := []string{
		"algorithm", "implementation", "architecture", "optimization",
		"integration", "deployment", "configuration", "infrastructure",
	}
	for _, term := range technicalPatterns {
		if strings.Contains(strings.ToLower(text), term) {
			complexityScore++
			break
		}
	}

	if complexityScore >= 2 {
		return ComplexityComplex
	}

	return ComplexityMedium
}

// hasMultipleSteps checks if the query requires multiple steps
func (qa *QueryAnalyzer) hasMultipleSteps(text string) bool {
	// Look for step indicators
	stepPatterns := []string{
		"first", "second", "third", "then", "next", "after that",
		"step 1", "step 2", "finally", "lastly",
	}

	count := 0
	lowerText := strings.ToLower(text)
	for _, pattern := range stepPatterns {
		if strings.Contains(lowerText, pattern) {
			count++
		}
	}

	return count >= 2
}

// isCreativeWriting checks if the query is for creative writing
func (qa *QueryAnalyzer) isCreativeWriting(text string) bool {
	creativePatterns := []string{
		"write a story", "create a poem", "compose", "creative",
		"imagine", "fictional", "narrative", "character", "plot",
	}

	lowerText := strings.ToLower(text)
	for _, pattern := range creativePatterns {
		if strings.Contains(lowerText, pattern) {
			return true
		}
	}

	return false
}

// ShouldUseAdvancedModel determines if an advanced model should be used
func (qa *QueryAnalyzer) ShouldUseAdvancedModel(analysis *QueryAnalysis) bool {
	return analysis.Complexity == ComplexityComplex ||
		analysis.RequiresCodeGeneration ||
		analysis.RequiresReasoning ||
		analysis.RequiresLongContext
}

// GetRecommendedModels returns recommended models based on analysis
func (qa *QueryAnalyzer) GetRecommendedModels(analysis *QueryAnalysis) []string {
	var models []string

	switch analysis.Complexity {
	case ComplexitySimple:
		// Use cheaper, faster models for simple queries
		models = []string{
			"gpt-3.5-turbo",
			"claude-3-haiku",
			"gpt-4o-mini",
		}
	case ComplexityMedium:
		// Use mid-tier models
		models = []string{
			"gpt-4o-mini",
			"claude-3-sonnet",
			"gpt-4-turbo",
		}
	case ComplexityComplex:
		// Use most capable models
		models = []string{
			"gpt-4",
			"claude-3-opus",
			"gpt-4-turbo",
			"claude-3-sonnet",
		}
	}

	// Adjust for code generation
	if analysis.RequiresCodeGeneration {
		models = []string{
			"gpt-4",
			"claude-3-opus",
			"gpt-4-turbo",
		}
	}

	return models
}
