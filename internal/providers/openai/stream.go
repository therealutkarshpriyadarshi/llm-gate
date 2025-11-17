package openai

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/therealutkarshpriyadarshi/llm-gate/internal/core/models"
)

// StreamReaderImpl implements the StreamReader interface for OpenAI
type StreamReaderImpl struct {
	reader io.ReadCloser
	scanner *bufio.Scanner
	model  string
	done   bool
}

// NewStreamReader creates a new stream reader
func NewStreamReader(reader io.ReadCloser, model string) *StreamReaderImpl {
	return &StreamReaderImpl{
		reader:  reader,
		scanner: bufio.NewScanner(reader),
		model:   model,
		done:    false,
	}
}

// Recv receives the next chunk from the stream
func (s *StreamReaderImpl) Recv() (*models.StreamChunk, error) {
	if s.done {
		return nil, io.EOF
	}

	if s.scanner == nil {
		s.scanner = bufio.NewScanner(s.reader)
	}

	// Read the next line
	for s.scanner.Scan() {
		line := s.scanner.Text()

		// Skip empty lines
		if line == "" {
			continue
		}

		// OpenAI uses SSE (Server-Sent Events) format with "data: " prefix
		if !strings.HasPrefix(line, "data: ") {
			continue
		}

		// Remove "data: " prefix
		data := strings.TrimPrefix(line, "data: ")

		// Check for stream end
		if data == "[DONE]" {
			s.done = true
			return nil, io.EOF
		}

		// Parse the JSON chunk
		var openAIChunk OpenAIStreamChunk
		if err := json.Unmarshal([]byte(data), &openAIChunk); err != nil {
			return nil, fmt.Errorf("failed to parse chunk: %w", err)
		}

		// Convert to unified format
		chunk := s.toUnifiedChunk(&openAIChunk)
		return chunk, nil
	}

	// Check for scanner errors
	if err := s.scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	// End of stream
	s.done = true
	return nil, io.EOF
}

// Close closes the stream
func (s *StreamReaderImpl) Close() error {
	if s.reader != nil {
		return s.reader.Close()
	}
	return nil
}

// toUnifiedChunk converts an OpenAI chunk to unified format
func (s *StreamReaderImpl) toUnifiedChunk(chunk *OpenAIStreamChunk) *models.StreamChunk {
	choices := make([]models.StreamChoice, len(chunk.Choices))
	for i, choice := range chunk.Choices {
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
		ID:       chunk.ID,
		Object:   chunk.Object,
		Created:  chunk.Created,
		Model:    chunk.Model,
		Choices:  choices,
		Provider: string(models.ProviderOpenAI),
	}
}

// StreamWriterImpl implements the StreamWriter interface
type StreamWriterImpl struct {
	writer io.Writer
	buf    *bytes.Buffer
}

// NewStreamWriter creates a new stream writer
func NewStreamWriter(writer io.Writer) *StreamWriterImpl {
	return &StreamWriterImpl{
		writer: writer,
		buf:    &bytes.Buffer{},
	}
}

// Write writes data to the buffer
func (w *StreamWriterImpl) Write(p []byte) (n int, err error) {
	return w.buf.Write(p)
}

// Flush writes buffered data to the underlying writer
func (w *StreamWriterImpl) Flush() error {
	_, err := w.writer.Write(w.buf.Bytes())
	if err != nil {
		return err
	}
	w.buf.Reset()

	// Flush if the writer supports it
	if flusher, ok := w.writer.(interface{ Flush() error }); ok {
		return flusher.Flush()
	}

	return nil
}
