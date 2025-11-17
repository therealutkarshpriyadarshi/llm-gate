package bedrock

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
			name:           "Claude 3 Opus",
			modelID:        "anthropic.claude-3-opus-20240229-v1:0",
			wantInputCost:  0.015,
			wantOutputCost: 0.075,
			wantMaxTokens:  200000,
		},
		{
			name:           "Claude 3 Sonnet",
			modelID:        "anthropic.claude-3-sonnet-20240229-v1:0",
			wantInputCost:  0.003,
			wantOutputCost: 0.015,
			wantMaxTokens:  200000,
		},
		{
			name:           "Claude 3 Haiku",
			modelID:        "anthropic.claude-3-haiku-20240307-v1:0",
			wantInputCost:  0.00025,
			wantOutputCost: 0.00125,
			wantMaxTokens:  200000,
		},
		{
			name:           "Claude 2.1",
			modelID:        "anthropic.claude-v2:1",
			wantInputCost:  0.008,
			wantOutputCost: 0.024,
			wantMaxTokens:  100000,
		},
		{
			name:           "Llama 2 70B",
			modelID:        "meta.llama2-70b-chat-v1",
			wantInputCost:  0.00195,
			wantOutputCost: 0.00256,
			wantMaxTokens:  4096,
		},
		{
			name:           "Titan Text Express",
			modelID:        "amazon.titan-text-express-v1",
			wantInputCost:  0.0002,
			wantOutputCost: 0.0006,
			wantMaxTokens:  8000,
		},
		{
			name:           "Unknown model defaults to Haiku pricing",
			modelID:        "unknown-model",
			wantInputCost:  0.00025,
			wantOutputCost: 0.00125,
			wantMaxTokens:  200000,
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
		"anthropic.claude-3-opus-20240229-v1:0",
		"anthropic.claude-3-sonnet-20240229-v1:0",
		"anthropic.claude-3-haiku-20240307-v1:0",
		"anthropic.claude-v2:1",
		"anthropic.claude-v2",
		"anthropic.claude-instant-v1",
		"meta.llama2-70b-chat-v1",
		"meta.llama2-13b-chat-v1",
		"amazon.titan-text-express-v1",
		"amazon.titan-text-lite-v1",
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
