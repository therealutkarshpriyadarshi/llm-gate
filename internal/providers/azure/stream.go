package azure

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
)

// StreamReaderImpl implements the StreamReader interface for Azure OpenAI
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

		// Check for stream end
		if data == "[DONE]" {
			return nil, io.EOF
		}

		// Parse the chunk
		var azureChunk AzureStreamChunk
		if err := json.Unmarshal([]byte(data), &azureChunk); err != nil {
			return nil, fmt.Errorf("failed to parse stream chunk: %w", err)
		}

		// Convert to unified format
		choices := make([]models.StreamChoice, len(azureChunk.Choices))
		for i, choice := range azureChunk.Choices {
			choices[i] = models.StreamChoice{
				Index: choice.Index,
				Delta: models.MessageDelta{
					Role:    choice.Delta.Role,
					Content: choice.Delta.Content,
				},
				FinishReason: choice.FinishReason,
			}
		}

		return &models.StreamChunk{
			ID:      azureChunk.ID,
			Object:  azureChunk.Object,
			Created: azureChunk.Created,
			Model:   azureChunk.Model,
			Choices: choices,
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
