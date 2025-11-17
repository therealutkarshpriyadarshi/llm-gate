package bedrock

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
)

// StreamReaderImpl implements the StreamReader interface for AWS Bedrock
type StreamReaderImpl struct {
	reader       io.ReadCloser
	scanner      *bufio.Scanner
	model        string
	currentText  string
	finishReason string
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

		// Parse SSE format
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		data := strings.TrimPrefix(line, "data: ")

		// Check for stream end
		if data == "[DONE]" {
			return nil, io.EOF
		}

		// Parse the chunk
		var chunk BedrockStreamChunk
		if err := json.Unmarshal([]byte(data), &chunk); err != nil {
			return nil, fmt.Errorf("failed to parse stream chunk: %w", err)
		}

		// Handle different event types
		switch chunk.Type {
		case "message_start":
			continue

		case "content_block_start":
			continue

		case "content_block_delta":
			if chunk.Delta != nil && chunk.Delta.Text != "" {
				s.currentText += chunk.Delta.Text
				return &models.StreamChunk{
					ID:      "bedrock-stream",
					Object:  "chat.completion.chunk",
					Created: 0,
					Model:   s.model,
					Choices: []models.StreamChoice{
						{
							Index: 0,
							Delta: models.MessageDelta{
								Role:    "assistant",
								Content: chunk.Delta.Text,
							},
						},
					},
				}, nil
			}

		case "content_block_stop":
			continue

		case "message_delta":
			if chunk.Delta != nil && chunk.Delta.StopReason != "" {
				s.finishReason = chunk.Delta.StopReason
			}
			continue

		case "message_stop":
			return nil, io.EOF
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
