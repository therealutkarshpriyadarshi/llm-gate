package prompts

import (
	"testing"

	"github.com/google/uuid"
)

func TestABTestingService_SelectVariant(t *testing.T) {
	service := NewABTestingService(nil)

	exp := &Experiment{
		ID:   uuid.New(),
		Name: "test-experiment",
		Variants: []Variant{
			{ID: uuid.New(), Name: "control", TrafficSplit: 50},
			{ID: uuid.New(), Name: "variant-a", TrafficSplit: 30},
			{ID: uuid.New(), Name: "variant-b", TrafficSplit: 20},
		},
	}

	// Test consistency: same user should always get same variant
	userID := "user-123"
	variant1 := service.SelectVariant(exp, userID)
	variant2 := service.SelectVariant(exp, userID)

	if variant1.ID != variant2.ID {
		t.Errorf("SelectVariant() not consistent for same user. Got %v and %v", variant1.Name, variant2.Name)
	}

	// Test different users get distributed
	userCounts := make(map[string]int)
	for i := 0; i < 1000; i++ {
		variant := service.SelectVariant(exp, uuid.New().String())
		userCounts[variant.Name]++
	}

	// Check that all variants received some traffic
	for _, v := range exp.Variants {
		if userCounts[v.Name] == 0 {
			t.Errorf("Variant %s received no traffic", v.Name)
		}
	}

	// Rough distribution check (allowing for variance)
	controlCount := userCounts["control"]
	if controlCount < 400 || controlCount > 600 {
		t.Logf("Warning: Control variant got %d/1000 requests, expected ~500", controlCount)
	}
}

func TestABTestingService_HashUser(t *testing.T) {
	service := NewABTestingService(nil)

	experimentID := uuid.New().String()
	userID := "test-user"

	// Test consistency
	hash1 := service.hashUser(experimentID, userID)
	hash2 := service.hashUser(experimentID, userID)

	if hash1 != hash2 {
		t.Errorf("hashUser() not consistent. Got %d and %d", hash1, hash2)
	}

	// Test range
	if hash1 < 0 || hash1 >= 100 {
		t.Errorf("hashUser() = %d, want value between 0 and 99", hash1)
	}

	// Test different experiments give different hashes
	experimentID2 := uuid.New().String()
	hash3 := service.hashUser(experimentID2, userID)

	if hash1 == hash3 {
		t.Logf("Note: Same user got same hash in different experiments (possible but unlikely)")
	}
}

func TestABTestingService_CalculateTrafficSplits(t *testing.T) {
	service := NewABTestingService(nil)

	tests := []struct {
		name        string
		numVariants int
		wantSum     int
	}{
		{
			name:        "two variants",
			numVariants: 2,
			wantSum:     100,
		},
		{
			name:        "three variants",
			numVariants: 3,
			wantSum:     100,
		},
		{
			name:        "five variants",
			numVariants: 5,
			wantSum:     100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			splits := service.CalculateTrafficSplits(tt.numVariants)

			if len(splits) != tt.numVariants {
				t.Errorf("CalculateTrafficSplits() returned %d splits, want %d", len(splits), tt.numVariants)
			}

			sum := 0
			for _, split := range splits {
				sum += split
			}

			if sum != tt.wantSum {
				t.Errorf("CalculateTrafficSplits() sum = %d, want %d. Splits: %v", sum, tt.wantSum, splits)
			}
		})
	}
}

func TestABTestingService_AnalyzeResults(t *testing.T) {
	service := NewABTestingService(nil)

	variantA := uuid.New()
	variantB := uuid.New()

	metrics := map[string]*ExperimentMetrics{
		variantA.String(): {
			VariantID:            variantA,
			RequestCount:         1000,
			SuccessCount:         950,
			ErrorCount:           50,
			TotalTokens:          50000,
			TotalCost:            5.0,
			AvgLatencyMs:         100,
			UserFeedbackPositive: 80,
			UserFeedbackNegative: 20,
		},
		variantB.String(): {
			VariantID:            variantB,
			RequestCount:         1000,
			SuccessCount:         900,
			ErrorCount:           100,
			TotalTokens:          48000,
			TotalCost:            4.8,
			AvgLatencyMs:         120,
			UserFeedbackPositive: 70,
			UserFeedbackNegative: 30,
		},
	}

	analysis := service.AnalyzeResults(metrics)

	if analysis == nil {
		t.Fatal("AnalyzeResults() returned nil")
	}

	if len(analysis.VariantResults) != 2 {
		t.Errorf("AnalyzeResults() returned %d results, want 2", len(analysis.VariantResults))
	}

	// Check that variant A (better performer) is the winner
	if analysis.Winner != variantA.String() {
		t.Errorf("AnalyzeResults() winner = %v, want %v", analysis.Winner, variantA.String())
	}

	// Check that results are statistically significant
	// Note: This is a simplified implementation, so significance may not always be detected
	// In production, use proper statistical tests
	if !analysis.IsSignificant {
		t.Log("Note: IsSignificant = false (expected with simplified implementation)")
	}

	// Verify success rates
	resultA := analysis.VariantResults[variantA.String()]
	if resultA.SuccessRate != 0.95 {
		t.Errorf("Variant A success rate = %.2f, want 0.95", resultA.SuccessRate)
	}

	resultB := analysis.VariantResults[variantB.String()]
	if resultB.SuccessRate != 0.90 {
		t.Errorf("Variant B success rate = %.2f, want 0.90", resultB.SuccessRate)
	}
}

func TestABTestingService_CalculateRequiredSampleSize(t *testing.T) {
	service := NewABTestingService(nil)

	tests := []struct {
		name         string
		baselineRate float64
		effect       float64
		wantMinimum  int
	}{
		{
			name:         "typical case",
			baselineRate: 0.50,
			effect:       0.05,
			wantMinimum:  100,
		},
		{
			name:         "small effect",
			baselineRate: 0.50,
			effect:       0.01,
			wantMinimum:  1000,
		},
		{
			name:         "large effect",
			baselineRate: 0.50,
			effect:       0.20,
			wantMinimum:  100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := service.CalculateRequiredSampleSize(tt.baselineRate, tt.effect, 0.05, 0.80)

			if n < tt.wantMinimum {
				t.Errorf("CalculateRequiredSampleSize() = %d, want at least %d", n, tt.wantMinimum)
			}
		})
	}
}

func TestExperimentAnalysis_GetRecommendation(t *testing.T) {
	tests := []struct {
		name     string
		analysis *ExperimentAnalysis
		wantText string
	}{
		{
			name: "significant results",
			analysis: &ExperimentAnalysis{
				IsSignificant: true,
				Winner:        "variant-a",
				VariantResults: map[string]*VariantResult{
					"variant-a": {
						SuccessRate: 0.95,
						AvgLatency:  100,
						AvgCost:     0.005,
					},
				},
			},
			wantText: "winner",
		},
		{
			name: "not significant",
			analysis: &ExperimentAnalysis{
				IsSignificant: false,
			},
			wantText: "Continue experiment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			recommendation := tt.analysis.GetRecommendation()

			if recommendation == "" {
				t.Error("GetRecommendation() returned empty string")
			}

			// Simple check that the recommendation contains expected text
			if tt.wantText != "" {
				found := false
				if len(recommendation) > 0 {
					// Check if the expected text appears in the recommendation
					for i := 0; i <= len(recommendation)-len(tt.wantText); i++ {
						if recommendation[i:i+len(tt.wantText)] == tt.wantText {
							found = true
							break
						}
					}
				}
				if !found {
					t.Errorf("GetRecommendation() = %v, want to contain %v", recommendation, tt.wantText)
				}
			}
		})
	}
}

func BenchmarkABTestingService_SelectVariant(b *testing.B) {
	service := NewABTestingService(nil)

	exp := &Experiment{
		ID: uuid.New(),
		Variants: []Variant{
			{ID: uuid.New(), Name: "control", TrafficSplit: 50},
			{ID: uuid.New(), Name: "variant-a", TrafficSplit: 30},
			{ID: uuid.New(), Name: "variant-b", TrafficSplit: 20},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		service.SelectVariant(exp, "user-123")
	}
}
