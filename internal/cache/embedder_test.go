package cache

import (
	"context"
	"testing"
)

func TestMockEmbedder(t *testing.T) {
	embedder := NewMockEmbedder(128)

	t.Run("embed single text", func(t *testing.T) {
		text := "Hello, world!"
		embedding, err := embedder.Embed(context.Background(), text)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(embedding) != 128 {
			t.Errorf("expected embedding dimension 128, got %d", len(embedding))
		}

		// Mock embedder should be deterministic
		embedding2, err := embedder.Embed(context.Background(), text)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(embedding) != len(embedding2) {
			t.Errorf("embeddings have different lengths")
		}

		for i := range embedding {
			if embedding[i] != embedding2[i] {
				t.Errorf("embeddings differ at index %d: %f != %f", i, embedding[i], embedding2[i])
			}
		}
	})

	t.Run("embed batch", func(t *testing.T) {
		texts := []string{"Hello", "World", "Test"}
		embeddings, err := embedder.EmbedBatch(context.Background(), texts)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(embeddings) != len(texts) {
			t.Errorf("expected %d embeddings, got %d", len(texts), len(embeddings))
		}

		for i, emb := range embeddings {
			if len(emb) != 128 {
				t.Errorf("embedding %d has dimension %d, want 128", i, len(emb))
			}
		}
	})

	t.Run("different texts produce different embeddings", func(t *testing.T) {
		text1 := "Hello"
		text2 := "World"

		emb1, _ := embedder.Embed(context.Background(), text1)
		emb2, _ := embedder.Embed(context.Background(), text2)

		// They should be different
		same := true
		for i := range emb1 {
			if emb1[i] != emb2[i] {
				same = false
				break
			}
		}

		if same {
			t.Error("different texts produced identical embeddings")
		}
	})
}

func TestOpenAIEmbedder(t *testing.T) {
	// This is a basic test that doesn't require an actual API key
	// Real integration tests would go in a separate test suite

	t.Run("create embedder", func(t *testing.T) {
		embedder := NewOpenAIEmbedder("test-api-key", "text-embedding-3-small")

		if embedder == nil {
			t.Error("expected non-nil embedder")
		}

		if embedder.apiKey != "test-api-key" {
			t.Errorf("api key is %s, want test-api-key", embedder.apiKey)
		}

		if embedder.model != "text-embedding-3-small" {
			t.Errorf("model is %s, want text-embedding-3-small", embedder.model)
		}
	})

	t.Run("create embedder with default model", func(t *testing.T) {
		embedder := NewOpenAIEmbedder("test-api-key", "")

		if embedder.model != "text-embedding-3-small" {
			t.Errorf("model is %s, want text-embedding-3-small (default)", embedder.model)
		}
	})

	t.Run("cache operations", func(t *testing.T) {
		embedder := NewOpenAIEmbedder("test-api-key", "text-embedding-3-small")

		// Cache should be empty initially
		if embedder.GetCacheSize() != 0 {
			t.Errorf("cache size is %d, want 0", embedder.GetCacheSize())
		}

		// Clear cache
		embedder.ClearCache()

		if embedder.GetCacheSize() != 0 {
			t.Errorf("cache size after clear is %d, want 0", embedder.GetCacheSize())
		}
	})
}

func BenchmarkMockEmbedder(b *testing.B) {
	embedder := NewMockEmbedder(1536) // Same dimension as OpenAI
	text := "This is a test sentence for benchmarking embedding generation."
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = embedder.Embed(ctx, text)
	}
}

func BenchmarkMockEmbedderBatch(b *testing.B) {
	embedder := NewMockEmbedder(1536)
	texts := []string{
		"First sentence",
		"Second sentence",
		"Third sentence",
		"Fourth sentence",
		"Fifth sentence",
	}
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = embedder.EmbedBatch(ctx, texts)
	}
}
