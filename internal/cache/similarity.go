package cache

import (
	"fmt"
	"math"
)

// SimilarityCalculator handles vector similarity calculations
type SimilarityCalculator struct{}

// NewSimilarityCalculator creates a new similarity calculator
func NewSimilarityCalculator() *SimilarityCalculator {
	return &SimilarityCalculator{}
}

// CosineSimilarity calculates the cosine similarity between two vectors
// Returns a value between -1 and 1, where 1 means identical, 0 means orthogonal, -1 means opposite
func (s *SimilarityCalculator) CosineSimilarity(a, b []float32) (float32, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("vector dimensions must match: %d != %d", len(a), len(b))
	}

	if len(a) == 0 {
		return 0, fmt.Errorf("vectors cannot be empty")
	}

	var dotProduct float64
	var magnitudeA float64
	var magnitudeB float64

	for i := 0; i < len(a); i++ {
		dotProduct += float64(a[i]) * float64(b[i])
		magnitudeA += float64(a[i]) * float64(a[i])
		magnitudeB += float64(b[i]) * float64(b[i])
	}

	magnitudeA = math.Sqrt(magnitudeA)
	magnitudeB = math.Sqrt(magnitudeB)

	if magnitudeA == 0 || magnitudeB == 0 {
		return 0, fmt.Errorf("vector magnitude cannot be zero")
	}

	similarity := dotProduct / (magnitudeA * magnitudeB)
	return float32(similarity), nil
}

// EuclideanDistance calculates the Euclidean distance between two vectors
// Returns a value >= 0, where 0 means identical vectors
func (s *SimilarityCalculator) EuclideanDistance(a, b []float32) (float32, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("vector dimensions must match: %d != %d", len(a), len(b))
	}

	if len(a) == 0 {
		return 0, fmt.Errorf("vectors cannot be empty")
	}

	var sum float64
	for i := 0; i < len(a); i++ {
		diff := float64(a[i] - b[i])
		sum += diff * diff
	}

	return float32(math.Sqrt(sum)), nil
}

// DotProduct calculates the dot product of two vectors
func (s *SimilarityCalculator) DotProduct(a, b []float32) (float32, error) {
	if len(a) != len(b) {
		return 0, fmt.Errorf("vector dimensions must match: %d != %d", len(a), len(b))
	}

	if len(a) == 0 {
		return 0, fmt.Errorf("vectors cannot be empty")
	}

	var product float64
	for i := 0; i < len(a); i++ {
		product += float64(a[i]) * float64(b[i])
	}

	return float32(product), nil
}

// Magnitude calculates the magnitude (L2 norm) of a vector
func (s *SimilarityCalculator) Magnitude(v []float32) (float32, error) {
	if len(v) == 0 {
		return 0, fmt.Errorf("vector cannot be empty")
	}

	var sum float64
	for _, val := range v {
		sum += float64(val) * float64(val)
	}

	return float32(math.Sqrt(sum)), nil
}

// Normalize normalizes a vector to unit length
func (s *SimilarityCalculator) Normalize(v []float32) ([]float32, error) {
	mag, err := s.Magnitude(v)
	if err != nil {
		return nil, err
	}

	if mag == 0 {
		return nil, fmt.Errorf("cannot normalize zero vector")
	}

	normalized := make([]float32, len(v))
	for i, val := range v {
		normalized[i] = val / mag
	}

	return normalized, nil
}

// FindMostSimilar finds the most similar vector from a list
// Returns the index of the most similar vector and its similarity score
func (s *SimilarityCalculator) FindMostSimilar(query []float32, candidates [][]float32) (int, float32, error) {
	if len(candidates) == 0 {
		return -1, 0, fmt.Errorf("no candidates provided")
	}

	maxSimilarity := float32(-2.0) // Lower than minimum possible similarity (-1)
	maxIndex := -1

	for i, candidate := range candidates {
		similarity, err := s.CosineSimilarity(query, candidate)
		if err != nil {
			continue // Skip invalid candidates
		}

		if similarity > maxSimilarity {
			maxSimilarity = similarity
			maxIndex = i
		}
	}

	if maxIndex == -1 {
		return -1, 0, fmt.Errorf("no valid candidates found")
	}

	return maxIndex, maxSimilarity, nil
}

// IsAboveThreshold checks if similarity score is above the threshold
func (s *SimilarityCalculator) IsAboveThreshold(similarity, threshold float32) bool {
	return similarity >= threshold
}
