package anthropic

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
)

// StreamReaderImpl implements the StreamReader interface for Anthropic
type StreamReaderImpl struct {
	reader        io.ReadCloser
	scanner       *bufio.Scanner
	model         string
	currentIndex  int
	currentText   string
	totalTokens   int
	finishReason  string
}

// Recv receives the next chunk from the stream
func (s *StreamReaderImpl) Recv() (*models.StreamChunk, error) {
	if s.scanner == nil {
		s.scanner = bufio.NewScanner(s.reader)
		// Increase buffer size for large responses
		buf := make([]byte, 0, 64*1024)
		s.scanner.Buffer(buf, 1024*1024)
	}

	for s.scanner.Scan() {
		line := s.scanner.Text()

		// Skip empty lines
		if line == "" {
			continue
		}

		// Parse SSE format (data: {...})
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		// Check for stream end
		if data == "[DONE]" {
			return nil, io.EOF
		}

		// Parse the event
		var event AnthropicStreamEvent
		if err := json.Unmarshal([]byte(data), &event); err != nil {
			return nil, fmt.Errorf("failed to parse stream event: %w", err)
		}

		// Handle different event types
		switch event.Type {
		case "message_start":
			// Initial message metadata
			if event.Message != nil {
				s.model = event.Message.Model
			}
			continue

		case "content_block_start":
			// Start of a new content block
			continue

		case "content_block_delta":
			// Incremental content update
			if event.Delta != nil && event.Delta.Text != "" {
				s.currentText += event.Delta.Text
				return &models.StreamChunk{
					ID:      fmt.Sprintf("chunk-%d", s.currentIndex),
					Object:  "chat.completion.chunk",
					Created: 0,
					Model:   s.model,
					Choices: []models.StreamChoice{
						{
							Index: 0,
							Delta: models.MessageDelta{
								Role:    "assistant",
								Content: event.Delta.Text,
							},
						},
					},
				}, nil
			}

		case "content_block_stop":
			// End of content block
			s.currentIndex++
			continue

		case "message_delta":
			// Message metadata update (includes stop reason)
			if event.Delta != nil && event.Delta.StopReason != "" {
				s.finishReason = event.Delta.StopReason
			}
			continue

		case "message_stop":
			// End of message
			return nil, io.EOF

		case "error":
			return nil, fmt.Errorf("stream error: %v", event)
		}
	}

	if err := s.scanner.Err(); err != nil {
		return nil, err
	}

	return nil, io.EOF
}

// Close closes the stream
func (s *StreamReaderImpl) Close() error {
	return s.reader.Close()
}

// parseSSE parses a Server-Sent Event line
func parseSSE(line string) (eventType, data string, err error) {
	line = strings.TrimSpace(line)

	if strings.HasPrefix(line, "event: ") {
		eventType = strings.TrimPrefix(line, "event: ")
		return eventType, "", nil
	}

	if strings.HasPrefix(line, "data: ") {
		data = strings.TrimPrefix(line, "data: ")
		return "", data, nil
	}

	return "", "", nil
}

// combineStreamChunks combines multiple stream chunks into a full response
func combineStreamChunks(chunks []*models.StreamChunk) *models.ChatResponse {
	if len(chunks) == 0 {
		return nil
	}

	var content bytes.Buffer
	var lastChunk *models.StreamChunk

	for _, chunk := range chunks {
		lastChunk = chunk
		if len(chunk.Choices) > 0 && chunk.Choices[0].Delta.Content != "" {
			content.WriteString(chunk.Choices[0].Delta.Content)
		}
	}

	finishReason := ""
	if lastChunk != nil && len(lastChunk.Choices) > 0 {
		finishReason = lastChunk.Choices[0].FinishReason
	}

	return &models.ChatResponse{
		ID:      lastChunk.ID,
		Object:  "chat.completion",
		Created: lastChunk.Created,
		Model:   lastChunk.Model,
		Choices: []models.Choice{
			{
				Index: 0,
				Message: models.Message{
					Role:    "assistant",
					Content: content.String(),
				},
				FinishReason: finishReason,
			},
		},
		Provider: string(models.ProviderAnthropic),
	}
}
