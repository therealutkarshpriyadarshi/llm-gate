package vertex

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
			name:           "Gemini Pro",
			modelID:        "gemini-pro",
			wantInputCost:  0.000125,
			wantOutputCost: 0.000375,
			wantMaxTokens:  32760,
		},
		{
			name:           "Gemini Pro Vision",
			modelID:        "gemini-pro-vision",
			wantInputCost:  0.000125,
			wantOutputCost: 0.000375,
			wantMaxTokens:  16384,
		},
		{
			name:           "Gemini 1.5 Pro",
			modelID:        "gemini-1.5-pro",
			wantInputCost:  0.00125,
			wantOutputCost: 0.00375,
			wantMaxTokens:  1048576,
		},
		{
			name:           "Gemini 1.5 Flash",
			modelID:        "gemini-1.5-flash",
			wantInputCost:  0.000075,
			wantOutputCost: 0.0003,
			wantMaxTokens:  1048576,
		},
		{
			name:           "Text Bison",
			modelID:        "text-bison",
			wantInputCost:  0.000125,
			wantOutputCost: 0.000125,
			wantMaxTokens:  8196,
		},
		{
			name:           "Text Bison 32K",
			modelID:        "text-bison-32k",
			wantInputCost:  0.000125,
			wantOutputCost: 0.000125,
			wantMaxTokens:  32000,
		},
		{
			name:           "Chat Bison",
			modelID:        "chat-bison",
			wantInputCost:  0.000125,
			wantOutputCost: 0.000125,
			wantMaxTokens:  8196,
		},
		{
			name:           "Chat Bison 32K",
			modelID:        "chat-bison-32k",
			wantInputCost:  0.000125,
			wantOutputCost: 0.000125,
			wantMaxTokens:  32000,
		},
		{
			name:           "Unknown model defaults to Gemini Pro pricing",
			modelID:        "unknown-model",
			wantInputCost:  0.000125,
			wantOutputCost: 0.000375,
			wantMaxTokens:  32760,
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
		"gemini-pro",
		"gemini-pro-vision",
		"gemini-1.5-pro",
		"gemini-1.5-flash",
		"text-bison",
		"text-bison-32k",
		"chat-bison",
		"chat-bison-32k",
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
