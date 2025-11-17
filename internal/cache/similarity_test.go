package cache

import (
	"math"
	"testing"
)

func TestCosineSimilarity(t *testing.T) {
	calc := NewSimilarityCalculator()

	tests := []struct {
		name      string
		vectorA   []float32
		vectorB   []float32
		want      float32
		wantError bool
	}{
		{
			name:      "identical vectors",
			vectorA:   []float32{1.0, 0.0, 0.0},
			vectorB:   []float32{1.0, 0.0, 0.0},
			want:      1.0,
			wantError: false,
		},
		{
			name:      "orthogonal vectors",
			vectorA:   []float32{1.0, 0.0, 0.0},
			vectorB:   []float32{0.0, 1.0, 0.0},
			want:      0.0,
			wantError: false,
		},
		{
			name:      "opposite vectors",
			vectorA:   []float32{1.0, 0.0, 0.0},
			vectorB:   []float32{-1.0, 0.0, 0.0},
			want:      -1.0,
			wantError: false,
		},
		{
			name:      "similar vectors",
			vectorA:   []float32{1.0, 1.0, 0.0},
			vectorB:   []float32{1.0, 0.5, 0.0},
			want:      0.948683, // Approximate
			wantError: false,
		},
		{
			name:      "different dimensions",
			vectorA:   []float32{1.0, 0.0},
			vectorB:   []float32{1.0, 0.0, 0.0},
			wantError: true,
		},
		{
			name:      "empty vectors",
			vectorA:   []float32{},
			vectorB:   []float32{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calc.CosineSimilarity(tt.vectorA, tt.vectorB)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Use approximate comparison for floating point
			if math.Abs(float64(got-tt.want)) > 0.0001 {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestEuclideanDistance(t *testing.T) {
	calc := NewSimilarityCalculator()

	tests := []struct {
		name      string
		vectorA   []float32
		vectorB   []float32
		want      float32
		wantError bool
	}{
		{
			name:      "identical vectors",
			vectorA:   []float32{1.0, 2.0, 3.0},
			vectorB:   []float32{1.0, 2.0, 3.0},
			want:      0.0,
			wantError: false,
		},
		{
			name:      "unit distance",
			vectorA:   []float32{0.0, 0.0, 0.0},
			vectorB:   []float32{1.0, 0.0, 0.0},
			want:      1.0,
			wantError: false,
		},
		{
			name:      "3-4-5 triangle",
			vectorA:   []float32{0.0, 0.0},
			vectorB:   []float32{3.0, 4.0},
			want:      5.0,
			wantError: false,
		},
		{
			name:      "different dimensions",
			vectorA:   []float32{1.0, 2.0},
			vectorB:   []float32{1.0, 2.0, 3.0},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calc.EuclideanDistance(tt.vectorA, tt.vectorB)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if math.Abs(float64(got-tt.want)) > 0.0001 {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMagnitude(t *testing.T) {
	calc := NewSimilarityCalculator()

	tests := []struct {
		name      string
		vector    []float32
		want      float32
		wantError bool
	}{
		{
			name:      "unit vector",
			vector:    []float32{1.0, 0.0, 0.0},
			want:      1.0,
			wantError: false,
		},
		{
			name:      "3-4-5 vector",
			vector:    []float32{3.0, 4.0},
			want:      5.0,
			wantError: false,
		},
		{
			name:      "zero vector",
			vector:    []float32{0.0, 0.0, 0.0},
			want:      0.0,
			wantError: false,
		},
		{
			name:      "empty vector",
			vector:    []float32{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calc.Magnitude(tt.vector)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			if math.Abs(float64(got-tt.want)) > 0.0001 {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNormalize(t *testing.T) {
	calc := NewSimilarityCalculator()

	tests := []struct {
		name      string
		vector    []float32
		wantError bool
	}{
		{
			name:      "already normalized",
			vector:    []float32{1.0, 0.0, 0.0},
			wantError: false,
		},
		{
			name:      "needs normalization",
			vector:    []float32{3.0, 4.0},
			wantError: false,
		},
		{
			name:      "zero vector",
			vector:    []float32{0.0, 0.0, 0.0},
			wantError: true,
		},
		{
			name:      "empty vector",
			vector:    []float32{},
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := calc.Normalize(tt.vector)

			if tt.wantError {
				if err == nil {
					t.Errorf("expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("unexpected error: %v", err)
				return
			}

			// Check that the magnitude is 1
			mag, err := calc.Magnitude(got)
			if err != nil {
				t.Errorf("unexpected error calculating magnitude: %v", err)
				return
			}

			if math.Abs(float64(mag-1.0)) > 0.0001 {
				t.Errorf("normalized vector magnitude is %v, want 1.0", mag)
			}
		})
	}
}

func TestFindMostSimilar(t *testing.T) {
	calc := NewSimilarityCalculator()

	query := []float32{1.0, 0.0, 0.0}
	candidates := [][]float32{
		{0.5, 0.5, 0.0},  // Less similar
		{0.9, 0.1, 0.0},  // Most similar
		{0.0, 1.0, 0.0},  // Orthogonal
		{-1.0, 0.0, 0.0}, // Opposite
	}

	idx, similarity, err := calc.FindMostSimilar(query, candidates)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
		return
	}

	if idx != 1 {
		t.Errorf("got index %d, want 1", idx)
	}

	if similarity < 0.9 {
		t.Errorf("got similarity %v, expected > 0.9", similarity)
	}
}

func TestIsAboveThreshold(t *testing.T) {
	calc := NewSimilarityCalculator()

	tests := []struct {
		name       string
		similarity float32
		threshold  float32
		want       bool
	}{
		{
			name:       "above threshold",
			similarity: 0.95,
			threshold:  0.90,
			want:       true,
		},
		{
			name:       "equal to threshold",
			similarity: 0.90,
			threshold:  0.90,
			want:       true,
		},
		{
			name:       "below threshold",
			similarity: 0.85,
			threshold:  0.90,
			want:       false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := calc.IsAboveThreshold(tt.similarity, tt.threshold)
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}
