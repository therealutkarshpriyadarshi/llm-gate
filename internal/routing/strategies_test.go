package routing

import (
	"testing"
	"time"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
)

func TestCostBasedStrategy(t *testing.T) {
	matrix := NewModelCapabilityMatrix()
	analyzer := NewQueryAnalyzer()
	strategy := NewCostBasedStrategy(matrix, analyzer)

	// This test would need proper mocking of the LLMProvider interface
	// For now, we'll test the strategy creation
	if strategy.Name() != "cost-based" {
		t.Errorf("Expected strategy name 'cost-based', got %s", strategy.Name())
	}
}

func TestLeastLatencyStrategy(t *testing.T) {
	matrix := NewModelCapabilityMatrix()
	strategy := NewLeastLatencyStrategy(matrix)

	if strategy.Name() != "least-latency" {
		t.Errorf("Expected strategy name 'least-latency', got %s", strategy.Name())
	}

	// Test latency updates
	strategy.UpdateLatency(models.ProviderOpenAI, 100*time.Millisecond)
	strategy.UpdateLatency(models.ProviderAnthropic, 200*time.Millisecond)

	latency := strategy.getLatency(models.ProviderOpenAI)
	if latency != 100*time.Millisecond {
		t.Errorf("Expected latency 100ms, got %v", latency)
	}
}

func TestWeightedStrategy(t *testing.T) {
	weights := map[models.ProviderType]int{
		models.ProviderOpenAI:    70,
		models.ProviderAnthropic: 30,
	}
	strategy := NewWeightedStrategy(weights)

	if strategy.Name() != "weighted" {
		t.Errorf("Expected strategy name 'weighted', got %s", strategy.Name())
	}

	// Test weight retrieval
	weight := strategy.getWeight(models.ProviderOpenAI)
	if weight != 70 {
		t.Errorf("Expected weight 70, got %d", weight)
	}

	// Test weight update
	strategy.SetWeight(models.ProviderOpenAI, 80)
	weight = strategy.getWeight(models.ProviderOpenAI)
	if weight != 80 {
		t.Errorf("Expected weight 80, got %d", weight)
	}
}

func TestStickySessionStrategy(t *testing.T) {
	fallback := NewRoundRobinStrategy()
	strategy := NewStickySessionStrategy(fallback)

	if strategy.Name() != "sticky-session" {
		t.Errorf("Expected strategy name 'sticky-session', got %s", strategy.Name())
	}

	// Test session key extraction
	req := &models.ChatRequest{
		Model: "gpt-4",
		Messages: []models.Message{
			{Role: "user", Content: "Hello"},
		},
		Metadata: models.RequestMetadata{
			UserID: "user123",
		},
	}

	sessionKey := strategy.getSessionKey(req)
	if sessionKey != "user123" {
		t.Errorf("Expected session key 'user123', got %s", sessionKey)
	}

	// Test clear session
	strategy.sessionMap["user123"] = models.ProviderOpenAI
	strategy.ClearSession("user123")
	if _, exists := strategy.sessionMap["user123"]; exists {
		t.Error("Session should have been cleared")
	}
}

func TestIntelligentStrategy(t *testing.T) {
	analyzer := NewQueryAnalyzer()
	matrix := NewModelCapabilityMatrix()
	strategy := NewIntelligentStrategy(analyzer, matrix)

	if strategy.Name() != "intelligent" {
		t.Errorf("Expected strategy name 'intelligent', got %s", strategy.Name())
	}

	// Test weight setting
	strategy.SetWeights(0.5, 0.3, 0.2)
	if strategy.costWeight != 0.5 {
		t.Errorf("Expected cost weight 0.5, got %f", strategy.costWeight)
	}
	if strategy.latencyWeight != 0.3 {
		t.Errorf("Expected latency weight 0.3, got %f", strategy.latencyWeight)
	}
	if strategy.capabilityWeight != 0.2 {
		t.Errorf("Expected capability weight 0.2, got %f", strategy.capabilityWeight)
	}
}

func TestHashBasedStrategy(t *testing.T) {
	strategy := NewHashBasedStrategy()

	if strategy.Name() != "hash-based" {
		t.Errorf("Expected strategy name 'hash-based', got %s", strategy.Name())
	}

	// Test that same user gets same provider
	// This would need proper provider mocking
}

func TestStrategyFactory(t *testing.T) {
	analyzer := NewQueryAnalyzer()
	matrix := NewModelCapabilityMatrix()
	factory := NewStrategyFactory(analyzer, matrix)

	tests := []struct {
		name         string
		strategyName string
		wantErr      bool
	}{
		{
			name:         "round-robin",
			strategyName: "round-robin",
			wantErr:      false,
		},
		{
			name:         "random",
			strategyName: "random",
			wantErr:      false,
		},
		{
			name:         "cost-based",
			strategyName: "cost-based",
			wantErr:      false,
		},
		{
			name:         "least-latency",
			strategyName: "least-latency",
			wantErr:      false,
		},
		{
			name:         "weighted",
			strategyName: "weighted",
			wantErr:      false,
		},
		{
			name:         "sticky-session",
			strategyName: "sticky-session",
			wantErr:      false,
		},
		{
			name:         "intelligent",
			strategyName: "intelligent",
			wantErr:      false,
		},
		{
			name:         "hash-based",
			strategyName: "hash-based",
			wantErr:      false,
		},
		{
			name:         "unknown",
			strategyName: "unknown-strategy",
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			strategy, err := factory.CreateStrategy(tt.strategyName, nil)
			if (err != nil) != tt.wantErr {
				t.Errorf("CreateStrategy() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && strategy == nil {
				t.Error("CreateStrategy() returned nil strategy")
			}
		})
	}
}

func TestModelCapabilityMatrix(t *testing.T) {
	matrix := NewModelCapabilityMatrix()

	// Test getting capabilities
	cap := matrix.GetCapability("gpt-4")
	if cap == nil {
		t.Error("Expected to find gpt-4 capability")
	} else {
		if cap.Model != "gpt-4" {
			t.Errorf("Expected model 'gpt-4', got %s", cap.Model)
		}
		if cap.Provider != models.ProviderOpenAI {
			t.Errorf("Expected provider OpenAI, got %s", cap.Provider)
		}
	}

	// Test unknown model
	cap = matrix.GetCapability("unknown-model")
	if cap != nil {
		t.Error("Expected nil for unknown model")
	}

	// Test getting models by complexity
	simpleModels := matrix.GetModelsByComplexity(ComplexitySimple)
	if len(simpleModels) == 0 {
		t.Error("Expected to find simple models")
	}

	complexModels := matrix.GetModelsByComplexity(ComplexityComplex)
	if len(complexModels) == 0 {
		t.Error("Expected to find complex models")
	}

	// Test getting cheapest model
	cheapest := matrix.GetCheapestModel(1000, 500)
	if cheapest == nil {
		t.Error("Expected to find cheapest model")
	}

	// Test getting fastest model
	fastest := matrix.GetFastestModel()
	if fastest == nil {
		t.Error("Expected to find fastest model")
	}

	// Test cost estimation
	cost := matrix.EstimateCost("gpt-4", 1000, 500)
	if cost <= 0 {
		t.Error("Expected positive cost estimate")
	}

	// Test context length support
	supports := matrix.SupportsContextLength("gpt-4", 5000)
	if !supports {
		t.Error("Expected gpt-4 to support 5000 tokens")
	}

	supports = matrix.SupportsContextLength("gpt-4", 10000)
	if supports {
		t.Error("Expected gpt-4 to not support 10000 tokens (max is 8192)")
	}
}

func TestModelCapabilityMatrix_GetModelsByCategory(t *testing.T) {
	matrix := NewModelCapabilityMatrix()

	codeModels := matrix.GetModelsByCategory("code")
	if len(codeModels) == 0 {
		t.Error("Expected to find models good for code")
	}

	speedModels := matrix.GetModelsByCategory("speed")
	if len(speedModels) == 0 {
		t.Error("Expected to find models good for speed")
	}
}

func BenchmarkRoundRobinStrategy(b *testing.B) {
	strategy := NewRoundRobinStrategy()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Note: This would need proper provider mocking for full benchmark
		_ = strategy.Name()
	}
}
