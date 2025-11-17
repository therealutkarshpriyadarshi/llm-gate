package routing

import (
	"testing"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
)

func TestQueryAnalyzer_Analyze(t *testing.T) {
	analyzer := NewQueryAnalyzer()

	tests := []struct {
		name     string
		request  *models.ChatRequest
		wantComplexity QueryComplexity
		wantCategories []string
		wantReasoning  bool
		wantCode       bool
	}{
		{
			name: "simple greeting",
			request: &models.ChatRequest{
				Model: "gpt-3.5-turbo",
				Messages: []models.Message{
					{Role: "user", Content: "Hello, how are you?"},
				},
			},
			wantComplexity: ComplexitySimple,
			wantCategories: []string{},
			wantReasoning:  false,
			wantCode:       false,
		},
		{
			name: "code generation request",
			request: &models.ChatRequest{
				Model: "gpt-4",
				Messages: []models.Message{
					{Role: "user", Content: "Write a function to calculate fibonacci numbers in Python"},
				},
			},
			wantComplexity: ComplexityMedium,
			wantCategories: []string{"code"},
			wantReasoning:  false,
			wantCode:       true,
		},
		{
			name: "complex reasoning task",
			request: &models.ChatRequest{
				Model: "gpt-4",
				Messages: []models.Message{
					{Role: "user", Content: "Explain why neural networks work well for image recognition. First, describe how convolutional layers extract features. Then, explain the role of pooling layers. Finally, discuss how fully connected layers make predictions."},
				},
			},
			wantComplexity: ComplexityComplex,
			wantCategories: []string{"reasoning"},
			wantReasoning:  true,
			wantCode:       false,
		},
		{
			name: "math problem",
			request: &models.ChatRequest{
				Model: "gpt-4",
				Messages: []models.Message{
					{Role: "user", Content: "Calculate the integral of x^2 from 0 to 5"},
				},
			},
			wantComplexity: ComplexitySimple,
			wantCategories: []string{"math"},
			wantReasoning:  false,
			wantCode:       false,
		},
		{
			name: "creative writing",
			request: &models.ChatRequest{
				Model: "gpt-4",
				Messages: []models.Message{
					{Role: "user", Content: "Write a story about a robot that learns to paint"},
				},
			},
			wantComplexity: ComplexitySimple,
			wantCategories: []string{"creative"},
			wantReasoning:  false,
			wantCode:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			analysis := analyzer.Analyze(tt.request)

			if analysis.Complexity != tt.wantComplexity {
				t.Errorf("Complexity = %v, want %v", analysis.Complexity, tt.wantComplexity)
			}

			if analysis.RequiresReasoning != tt.wantReasoning {
				t.Errorf("RequiresReasoning = %v, want %v", analysis.RequiresReasoning, tt.wantReasoning)
			}

			if analysis.RequiresCodeGeneration != tt.wantCode {
				t.Errorf("RequiresCodeGeneration = %v, want %v", analysis.RequiresCodeGeneration, tt.wantCode)
			}

			// Check categories
			for _, wantCat := range tt.wantCategories {
				found := false
				for _, cat := range analysis.Categories {
					if cat == wantCat {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected category %s not found in %v", wantCat, analysis.Categories)
				}
			}
		})
	}
}

func TestQueryAnalyzer_EstimateTokens(t *testing.T) {
	analyzer := NewQueryAnalyzer()

	tests := []struct {
		name       string
		text       string
		wantMinTokens int
		wantMaxTokens int
	}{
		{
			name: "short text",
			text: "Hello world",
			wantMinTokens: 1,
			wantMaxTokens: 10,
		},
		{
			name: "medium text",
			text: "This is a longer piece of text that should have more tokens estimated for it because it contains more words and characters.",
			wantMinTokens: 20,
			wantMaxTokens: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tokens := analyzer.estimateTokens(tt.text)
			if tokens < tt.wantMinTokens || tokens > tt.wantMaxTokens {
				t.Errorf("estimateTokens() = %v, want between %v and %v", tokens, tt.wantMinTokens, tt.wantMaxTokens)
			}
		})
	}
}

func TestQueryAnalyzer_DetectLanguage(t *testing.T) {
	analyzer := NewQueryAnalyzer()

	tests := []struct {
		name     string
		text     string
		wantLang string
	}{
		{
			name:     "english text",
			text:     "Hello, how are you today?",
			wantLang: "en",
		},
		{
			name:     "empty text",
			text:     "",
			wantLang: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lang := analyzer.detectLanguage(tt.text)
			if lang != tt.wantLang {
				t.Errorf("detectLanguage() = %v, want %v", lang, tt.wantLang)
			}
		})
	}
}

func TestQueryAnalyzer_ShouldUseAdvancedModel(t *testing.T) {
	analyzer := NewQueryAnalyzer()

	tests := []struct {
		name     string
		analysis *QueryAnalysis
		want     bool
	}{
		{
			name: "simple query",
			analysis: &QueryAnalysis{
				Complexity:             ComplexitySimple,
				RequiresCodeGeneration: false,
				RequiresReasoning:      false,
				RequiresLongContext:    false,
			},
			want: false,
		},
		{
			name: "complex query",
			analysis: &QueryAnalysis{
				Complexity:             ComplexityComplex,
				RequiresCodeGeneration: false,
				RequiresReasoning:      false,
				RequiresLongContext:    false,
			},
			want: true,
		},
		{
			name: "code generation",
			analysis: &QueryAnalysis{
				Complexity:             ComplexityMedium,
				RequiresCodeGeneration: true,
				RequiresReasoning:      false,
				RequiresLongContext:    false,
			},
			want: true,
		},
		{
			name: "reasoning required",
			analysis: &QueryAnalysis{
				Complexity:             ComplexityMedium,
				RequiresCodeGeneration: false,
				RequiresReasoning:      true,
				RequiresLongContext:    false,
			},
			want: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := analyzer.ShouldUseAdvancedModel(tt.analysis); got != tt.want {
				t.Errorf("ShouldUseAdvancedModel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestQueryAnalyzer_GetRecommendedModels(t *testing.T) {
	analyzer := NewQueryAnalyzer()

	tests := []struct {
		name     string
		analysis *QueryAnalysis
		wantContains string
	}{
		{
			name: "simple query",
			analysis: &QueryAnalysis{
				Complexity: ComplexitySimple,
			},
			wantContains: "gpt-3.5-turbo",
		},
		{
			name: "complex query",
			analysis: &QueryAnalysis{
				Complexity: ComplexityComplex,
			},
			wantContains: "gpt-4",
		},
		{
			name: "code generation",
			analysis: &QueryAnalysis{
				Complexity:             ComplexityMedium,
				RequiresCodeGeneration: true,
			},
			wantContains: "gpt-4",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			models := analyzer.GetRecommendedModels(tt.analysis)
			if len(models) == 0 {
				t.Error("GetRecommendedModels() returned empty list")
				return
			}

			found := false
			for _, model := range models {
				if model == tt.wantContains {
					found = true
					break
				}
			}

			if !found {
				t.Errorf("GetRecommendedModels() = %v, want to contain %v", models, tt.wantContains)
			}
		})
	}
}

func BenchmarkQueryAnalyzer_Analyze(b *testing.B) {
	analyzer := NewQueryAnalyzer()
	req := &models.ChatRequest{
		Model: "gpt-4",
		Messages: []models.Message{
			{Role: "user", Content: "Write a function to calculate fibonacci numbers in Python"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		analyzer.Analyze(req)
	}
}
