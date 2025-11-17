package models

import (
	"testing"

	"github.com/therealutkarshpriyadarshi/llm-gate/pkg/utils"
)

func TestChatRequest_Validate(t *testing.T) {
	tests := []struct {
		name    string
		req     *ChatRequest
		wantErr bool
	}{
		{
			name: "valid request",
			req: &ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			wantErr: false,
		},
		{
			name: "empty model",
			req: &ChatRequest{
				Model: "",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
			},
			wantErr: true,
		},
		{
			name: "no messages",
			req: &ChatRequest{
				Model:    "gpt-4",
				Messages: []Message{},
			},
			wantErr: true,
		},
		{
			name: "invalid role",
			req: &ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "invalid", Content: "Hello"},
				},
			},
			wantErr: true,
		},
		{
			name: "empty message content",
			req: &ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: ""},
				},
			},
			wantErr: true,
		},
		{
			name: "invalid temperature",
			req: &ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
				Temperature: utils.Float64Ptr(3.0),
			},
			wantErr: true,
		},
		{
			name: "invalid top_p",
			req: &ChatRequest{
				Model: "gpt-4",
				Messages: []Message{
					{Role: "user", Content: "Hello"},
				},
				TopP: utils.Float64Ptr(1.5),
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.req.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("ChatRequest.Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestChatRequest_Getters(t *testing.T) {
	t.Run("GetTemperature with value", func(t *testing.T) {
		req := &ChatRequest{
			Temperature: utils.Float64Ptr(0.7),
		}
		if got := req.GetTemperature(); got != 0.7 {
			t.Errorf("GetTemperature() = %v, want %v", got, 0.7)
		}
	})

	t.Run("GetTemperature default", func(t *testing.T) {
		req := &ChatRequest{}
		if got := req.GetTemperature(); got != 1.0 {
			t.Errorf("GetTemperature() = %v, want %v", got, 1.0)
		}
	})

	t.Run("GetMaxTokens with value", func(t *testing.T) {
		req := &ChatRequest{
			MaxTokens: utils.IntPtr(100),
		}
		if got := req.GetMaxTokens(); got != 100 {
			t.Errorf("GetMaxTokens() = %v, want %v", got, 100)
		}
	})

	t.Run("GetMaxTokens default", func(t *testing.T) {
		req := &ChatRequest{}
		if got := req.GetMaxTokens(); got != 0 {
			t.Errorf("GetMaxTokens() = %v, want %v", got, 0)
		}
	})
}
