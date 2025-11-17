package integration

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/interfaces"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/providers"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/routing"
)

func TestIntelligentRouting_EndToEnd(t *testing.T) {
	// Setup
	analyzer := routing.NewQueryAnalyzer()
	matrix := routing.NewModelCapabilityMatrix()

	// Create mock providers
	openAIProvider := providers.NewMockProvider(
		models.ProviderOpenAI,
		"gpt-4",
		func(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error) {
			return &models.ChatResponse{
				ID:    "openai-response",
				Model: "gpt-4",
				Choices: []models.Choice{
					{
						Message: models.Message{
							Role:    "assistant",
							Content: "Response from OpenAI",
						},
					},
				},
			}, nil
		},
	)

	anthropicProvider := providers.NewMockProvider(
		models.ProviderAnthropic,
		"claude-3-sonnet",
		func(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error) {
			return &models.ChatResponse{
				ID:    "anthropic-response",
				Model: "claude-3-sonnet",
				Choices: []models.Choice{
					{
						Message: models.Message{
							Role:    "assistant",
							Content: "Response from Anthropic",
						},
					},
				},
			}, nil
		},
	)

	providerList := []interfaces.LLMProvider{openAIProvider, anthropicProvider}

	t.Run("Query Analysis", func(t *testing.T) {
		req := &models.ChatRequest{
			Model: "gpt-4",
			Messages: []models.Message{
				{Role: "user", Content: "Write a Python function to calculate fibonacci numbers"},
			},
		}

		analysis := analyzer.Analyze(req)

		if analysis.Complexity != routing.ComplexityMedium {
			t.Errorf("Expected medium complexity, got %s", analysis.Complexity)
		}

		if !analysis.RequiresCodeGeneration {
			t.Error("Expected code generation requirement")
		}

		found := false
		for _, cat := range analysis.Categories {
			if cat == "code" {
				found = true
				break
			}
		}
		if !found {
			t.Error("Expected 'code' category")
		}
	})

	t.Run("Cost-Based Routing", func(t *testing.T) {
		strategy := routing.NewCostBasedStrategy(matrix, analyzer)

		req := &models.ChatRequest{
			Model: "gpt-3.5-turbo",
			Messages: []models.Message{
				{Role: "user", Content: "Hello"},
			},
		}

		provider, err := strategy.SelectProvider(providerList, req)
		if err != nil {
			t.Fatalf("Failed to select provider: %v", err)
		}

		if provider == nil {
			t.Fatal("Provider should not be nil")
		}
	})

	t.Run("Intelligent Strategy", func(t *testing.T) {
		strategy := routing.NewIntelligentStrategy(analyzer, matrix)
		strategy.SetWeights(0.4, 0.3, 0.3)

		req := &models.ChatRequest{
			Model: "gpt-4",
			Messages: []models.Message{
				{Role: "user", Content: "Explain quantum computing"},
			},
		}

		provider, err := strategy.SelectProvider(providerList, req)
		if err != nil {
			t.Fatalf("Failed to select provider: %v", err)
		}

		if provider == nil {
			t.Fatal("Provider should not be nil")
		}
	})

	t.Run("Circuit Breaker", func(t *testing.T) {
		config := routing.CircuitBreakerConfig{
			MaxFailures:      2,
			Timeout:          100 * time.Millisecond,
			MaxConcurrent:    10,
			SuccessThreshold: 2,
			FailureRatio:     0.5,
			MinSamples:       5,
		}

		manager := routing.NewCircuitBreakerManager(config)
		ctx := context.Background()

		// Trigger failures
		testErr := errors.New("test error")
		for i := 0; i < 2; i++ {
			_ = manager.Execute(ctx, models.ProviderOpenAI, func(ctx context.Context) error {
				return testErr
			})
		}

		// Circuit should be open
		err := manager.Execute(ctx, models.ProviderOpenAI, func(ctx context.Context) error {
			return nil
		})

		if !errors.Is(err, routing.ErrCircuitOpen) {
			t.Errorf("Expected circuit to be open, got error: %v", err)
		}

		// Wait for timeout
		time.Sleep(150 * time.Millisecond)

		// Should allow request now
		err = manager.Execute(ctx, models.ProviderOpenAI, func(ctx context.Context) error {
			return nil
		})

		if err != nil && errors.Is(err, routing.ErrCircuitOpen) {
			t.Error("Circuit should not be open after timeout")
		}
	})

	t.Run("Fallback Chain", func(t *testing.T) {
		// Create a provider that fails
		failingProvider := providers.NewMockProvider(
			models.ProviderOpenAI,
			"gpt-4",
			func(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error) {
				return nil, errors.New("provider error")
			},
		)

		// Create a provider that succeeds
		successProvider := providers.NewMockProvider(
			models.ProviderAnthropic,
			"claude-3-sonnet",
			func(ctx context.Context, req *models.ChatRequest) (*models.ChatResponse, error) {
				return &models.ChatResponse{
					ID: "success-response",
					Choices: []models.Choice{
						{
							Message: models.Message{
								Role:    "assistant",
								Content: "Success!",
							},
						},
					},
				}, nil
			},
		)

		fallbackProviders := []interfaces.LLMProvider{failingProvider, successProvider}

		config := routing.DefaultFallbackChainConfig()
		config.MaxAttempts = 2
		config.EnableRetry = false

		chain := routing.NewFallbackChain(fallbackProviders, config)

		req := &models.ChatRequest{
			Model: "gpt-4",
			Messages: []models.Message{
				{Role: "user", Content: "Hello"},
			},
		}

		ctx := context.Background()
		response, err := chain.Execute(ctx, req)

		if err != nil {
			t.Fatalf("Fallback chain should have succeeded: %v", err)
		}

		if response == nil {
			t.Fatal("Response should not be nil")
		}

		if response.ID != "success-response" {
			t.Errorf("Expected success-response, got %s", response.ID)
		}
	})

	t.Run("Rules Engine", func(t *testing.T) {
		engine := routing.NewRulesEngine(analyzer, routing.NewRoundRobinStrategy())

		// Add rule: code queries go to OpenAI
		rule := routing.NewRuleBuilder("code-to-openai").
			WithPriority(100).
			WithCategory("code").
			SelectProvider(models.ProviderOpenAI).
			Build()

		engine.AddRule(rule)

		req := &models.ChatRequest{
			Model: "gpt-4",
			Messages: []models.Message{
				{Role: "user", Content: "Write a function to sort an array"},
			},
		}

		ctx := context.Background()
		provider, err := engine.Evaluate(ctx, req, providerList)

		if err != nil {
			t.Fatalf("Failed to evaluate rules: %v", err)
		}

		if provider.Name() != models.ProviderOpenAI {
			t.Errorf("Expected OpenAI provider, got %s", provider.Name())
		}
	})

	t.Run("Sticky Sessions", func(t *testing.T) {
		strategy := routing.NewStickySessionStrategy(routing.NewRoundRobinStrategy())

		req1 := &models.ChatRequest{
			Model: "gpt-4",
			Messages: []models.Message{
				{Role: "user", Content: "Hello"},
			},
			Metadata: models.RequestMetadata{
				UserID: "user123",
			},
		}

		// First request
		provider1, err := strategy.SelectProvider(providerList, req1)
		if err != nil {
			t.Fatalf("Failed to select provider: %v", err)
		}

		// Second request from same user
		req2 := &models.ChatRequest{
			Model: "gpt-4",
			Messages: []models.Message{
				{Role: "user", Content: "How are you?"},
			},
			Metadata: models.RequestMetadata{
				UserID: "user123",
			},
		}

		provider2, err := strategy.SelectProvider(providerList, req2)
		if err != nil {
			t.Fatalf("Failed to select provider: %v", err)
		}

		// Should be the same provider
		if provider1.Name() != provider2.Name() {
			t.Errorf("Sticky session failed: got %s and %s", provider1.Name(), provider2.Name())
		}
	})
}

func TestModelCapabilityMatrix_Integration(t *testing.T) {
	matrix := routing.NewModelCapabilityMatrix()

	t.Run("Cost Estimation Accuracy", func(t *testing.T) {
		// Test GPT-4
		cost := matrix.EstimateCost("gpt-4", 1000, 500)
		expectedCost := (1000.0/1000)*0.03 + (500.0/1000)*0.06
		if cost != expectedCost {
			t.Errorf("Expected cost %.6f, got %.6f", expectedCost, cost)
		}

		// Test GPT-3.5-turbo (cheaper)
		cost35 := matrix.EstimateCost("gpt-3.5-turbo", 1000, 500)
		if cost35 >= cost {
			t.Error("GPT-3.5-turbo should be cheaper than GPT-4")
		}
	})

	t.Run("Model Recommendations", func(t *testing.T) {
		analyzer := routing.NewQueryAnalyzer()

		// Simple query should recommend cheaper models
		simpleReq := &models.ChatRequest{
			Model: "",
			Messages: []models.Message{
				{Role: "user", Content: "Hello"},
			},
		}

		analysis := analyzer.Analyze(simpleReq)
		models := analyzer.GetRecommendedModels(analysis)

		if len(models) == 0 {
			t.Error("Should recommend models for simple query")
		}

		// Should include a cheap model
		hasCheapModel := false
		for _, model := range models {
			if model == "gpt-3.5-turbo" || model == "claude-3-haiku" {
				hasCheapModel = true
				break
			}
		}
		if !hasCheapModel {
			t.Error("Simple query should recommend cheap models")
		}
	})
}

func BenchmarkIntelligentRouting(b *testing.B) {
	analyzer := routing.NewQueryAnalyzer()
	matrix := routing.NewModelCapabilityMatrix()
	strategy := routing.NewIntelligentStrategy(analyzer, matrix)

	openAIProvider := providers.NewMockProvider(models.ProviderOpenAI, "gpt-4", nil)
	providerList := []interfaces.LLMProvider{openAIProvider}

	req := &models.ChatRequest{
		Model: "gpt-4",
		Messages: []models.Message{
			{Role: "user", Content: "Hello, how are you?"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = strategy.SelectProvider(providerList, req)
	}
}
