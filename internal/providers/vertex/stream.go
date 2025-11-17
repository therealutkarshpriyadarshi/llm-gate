package vertex

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
)

// StreamReaderImpl implements the StreamReader interface for Vertex AI
type StreamReaderImpl struct {
	reader  io.ReadCloser
	scanner *bufio.Scanner
	model   string
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

		// Parse the chunk
		var vertexChunk VertexStreamChunk
		if err := json.Unmarshal([]byte(data), &vertexChunk); err != nil {
			return nil, fmt.Errorf("failed to parse stream chunk: %w", err)
		}

		// Convert to unified format
		if len(vertexChunk.Candidates) == 0 {
			continue
		}

		candidate := vertexChunk.Candidates[0]

		// Combine all parts into content
		var content string
		for _, part := range candidate.Content.Parts {
			content += part.Text
		}

		role := candidate.Content.Role
		if role == "model" {
			role = "assistant"
		}

		return &models.StreamChunk{
			ID:      fmt.Sprintf("vertex-stream-%d", candidate.Index),
			Object:  "chat.completion.chunk",
			Created: 0,
			Model:   s.model,
			Choices: []models.StreamChoice{
				{
					Index: candidate.Index,
					Delta: models.MessageDelta{
						Role:    role,
						Content: content,
					},
					FinishReason: candidate.FinishReason,
				},
			},
		}, nil
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
