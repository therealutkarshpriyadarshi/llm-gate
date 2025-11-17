package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/api/validators"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
	"github.com/therealutkarshpriyadarshi/llm-gate/internal/routing"
)

// ChatHandler handles chat completion requests
type ChatHandler struct {
	router *routing.Router
}

// NewChatHandler creates a new chat handler
func NewChatHandler(router *routing.Router) *ChatHandler {
	return &ChatHandler{
		router: router,
	}
}

// HandleChatCompletion handles POST /v1/chat/completions
func (h *ChatHandler) HandleChatCompletion(w http.ResponseWriter, r *http.Request) {
	// Parse request
	var req models.ChatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", fmt.Sprintf("Failed to parse request: %v", err))
		return
	}

	// Generate request ID if not provided
	if req.Metadata.RequestID == "" {
		req.Metadata.RequestID = uuid.New().String()
	}
	req.Metadata.Timestamp = time.Now()

	// Sanitize request
	validators.SanitizeChatRequest(&req)

	// Validate request
	if err := validators.ValidateChatRequest(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", err.Error())
		return
	}

	// Log request
	log.Info().
		Str("request_id", req.Metadata.RequestID).
		Str("model", req.Model).
		Int("messages", len(req.Messages)).
		Bool("stream", req.Stream).
		Msg("Chat completion request received")

	// Handle streaming vs non-streaming
	if req.Stream {
		h.handleStreamingRequest(w, r, &req)
	} else {
		h.handleNonStreamingRequest(w, r, &req)
	}
}

// handleNonStreamingRequest handles non-streaming chat completions
func (h *ChatHandler) handleNonStreamingRequest(w http.ResponseWriter, r *http.Request, req *models.ChatRequest) {
	ctx := r.Context()
	startTime := time.Now()

	// Route the request with fallback
	resp, err := h.router.RouteWithFallback(ctx, req)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", req.Metadata.RequestID).
			Msg("Failed to route request")
		h.writeError(w, http.StatusServiceUnavailable, "provider_error", fmt.Sprintf("Failed to complete request: %v", err))
		return
	}

	// Update metadata
	resp.Metadata.Latency = time.Since(startTime)

	// Log response
	log.Info().
		Str("request_id", req.Metadata.RequestID).
		Str("provider", resp.Provider).
		Dur("latency", resp.Metadata.Latency).
		Int("prompt_tokens", resp.Usage.PromptTokens).
		Int("completion_tokens", resp.Usage.CompletionTokens).
		Float64("cost", resp.Metadata.Cost).
		Msg("Chat completion successful")

	// Write response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Error().
			Err(err).
			Str("request_id", req.Metadata.RequestID).
			Msg("Failed to encode response")
	}
}

// handleStreamingRequest handles streaming chat completions
func (h *ChatHandler) handleStreamingRequest(w http.ResponseWriter, r *http.Request, req *models.ChatRequest) {
	ctx := r.Context()

	// Route the request to get a provider
	provider, err := h.router.Route(ctx, req)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", req.Metadata.RequestID).
			Msg("Failed to route streaming request")
		h.writeError(w, http.StatusServiceUnavailable, "provider_error", fmt.Sprintf("Failed to route request: %v", err))
		return
	}

	// Start streaming
	stream, err := provider.ChatCompletionStream(ctx, req)
	if err != nil {
		log.Error().
			Err(err).
			Str("request_id", req.Metadata.RequestID).
			Str("provider", string(provider.Name())).
			Msg("Failed to start stream")
		h.writeError(w, http.StatusServiceUnavailable, "stream_error", fmt.Sprintf("Failed to start stream: %v", err))
		return
	}
	defer stream.Close()

	// Set headers for SSE
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable buffering in nginx

	// Get flusher
	flusher, ok := w.(http.Flusher)
	if !ok {
		h.writeError(w, http.StatusInternalServerError, "internal_error", "Streaming not supported")
		return
	}

	log.Info().
		Str("request_id", req.Metadata.RequestID).
		Str("provider", string(provider.Name())).
		Msg("Starting streaming response")

	// Stream chunks
	for {
		chunk, err := stream.Recv()
		if err == io.EOF {
			// Send [DONE] message
			fmt.Fprintf(w, "data: [DONE]\n\n")
			flusher.Flush()
			break
		}

		if err != nil {
			log.Error().
				Err(err).
				Str("request_id", req.Metadata.RequestID).
				Msg("Error receiving stream chunk")
			// Send error and close
			errorResp := models.ErrorResponse{
				Error: models.ErrorDetail{
					Message: err.Error(),
					Type:    "stream_error",
				},
			}
			data, _ := json.Marshal(errorResp)
			fmt.Fprintf(w, "data: %s\n\n", data)
			flusher.Flush()
			break
		}

		// Write chunk
		data, err := json.Marshal(chunk)
		if err != nil {
			log.Error().
				Err(err).
				Str("request_id", req.Metadata.RequestID).
				Msg("Failed to marshal chunk")
			continue
		}

		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()
	}

	log.Info().
		Str("request_id", req.Metadata.RequestID).
		Msg("Streaming complete")
}

// writeError writes an error response
func (h *ChatHandler) writeError(w http.ResponseWriter, status int, errType, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	resp := models.ErrorResponse{
		Error: models.ErrorDetail{
			Message: message,
			Type:    errType,
		},
	}

	json.NewEncoder(w).Encode(resp)
}
