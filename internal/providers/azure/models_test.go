package azure

import (
	"testing"
)

func TestGetModelPricing(t *testing.T) {
	tests := []struct {
		name           string
		modelID        string
		wantInputCost  float64
		wantOutputCost float64
		wantMaxTokens  int
	}{
		{
			name:           "GPT-4",
			modelID:        "gpt-4",
			wantInputCost:  0.03,
			wantOutputCost: 0.06,
			wantMaxTokens:  8192,
		},
		{
			name:           "GPT-4 32K",
			modelID:        "gpt-4-32k",
			wantInputCost:  0.06,
			wantOutputCost: 0.12,
			wantMaxTokens:  32768,
		},
		{
			name:           "GPT-4 Turbo",
			modelID:        "gpt-4-turbo",
			wantInputCost:  0.01,
			wantOutputCost: 0.03,
			wantMaxTokens:  128000,
		},
		{
			name:           "GPT-3.5 Turbo",
			modelID:        "gpt-35-turbo",
			wantInputCost:  0.0015,
			wantOutputCost: 0.002,
			wantMaxTokens:  4096,
		},
		{
			name:           "GPT-3.5 Turbo 16K",
			modelID:        "gpt-35-turbo-16k",
			wantInputCost:  0.003,
			wantOutputCost: 0.004,
			wantMaxTokens:  16384,
		},
		{
			name:           "Unknown model defaults to GPT-3.5 Turbo pricing",
			modelID:        "unknown-model",
			wantInputCost:  0.0015,
			wantOutputCost: 0.002,
			wantMaxTokens:  4096,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			inputCost, outputCost, maxTokens := GetModelPricing(tt.modelID)

			if inputCost != tt.wantInputCost {
				t.Errorf("GetModelPricing(%s) inputCost = %v, want %v", tt.modelID, inputCost, tt.wantInputCost)
			}

			if outputCost != tt.wantOutputCost {
				t.Errorf("GetModelPricing(%s) outputCost = %v, want %v", tt.modelID, outputCost, tt.wantOutputCost)
			}

			if maxTokens != tt.wantMaxTokens {
				t.Errorf("GetModelPricing(%s) maxTokens = %v, want %v", tt.modelID, maxTokens, tt.wantMaxTokens)
			}
		})
	}
}

func TestModelPricing(t *testing.T) {
	expectedModels := []string{
		"gpt-4",
		"gpt-4-32k",
		"gpt-4-turbo",
		"gpt-35-turbo",
		"gpt-35-turbo-16k",
	}

	for _, modelID := range expectedModels {
		if _, ok := ModelPricing[modelID]; !ok {
			t.Errorf("Model %s not found in ModelPricing map", modelID)
		}
	}

	// Verify all models have positive pricing
	for modelID, pricing := range ModelPricing {
		if pricing.InputCostPer1K <= 0 {
			t.Errorf("Model %s has invalid InputCostPer1K: %v", modelID, pricing.InputCostPer1K)
		}
		if pricing.OutputCostPer1K <= 0 {
			t.Errorf("Model %s has invalid OutputCostPer1K: %v", modelID, pricing.OutputCostPer1K)
		}
		if pricing.MaxTokens <= 0 {
			t.Errorf("Model %s has invalid MaxTokens: %v", modelID, pricing.MaxTokens)
		}
	}
}
