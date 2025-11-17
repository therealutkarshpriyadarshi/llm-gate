package prompts

import (
	"crypto/md5"
	"encoding/binary"
	"fmt"
	"math/rand"
	"sort"
)

// ABTestingService handles A/B testing logic for prompt experiments
type ABTestingService struct {
	expRepo *ExperimentRepository
}

// NewABTestingService creates a new A/B testing service
func NewABTestingService(expRepo *ExperimentRepository) *ABTestingService {
	return &ABTestingService{
		expRepo: expRepo,
	}
}

// SelectVariant selects a variant for a user based on traffic splitting
func (s *ABTestingService) SelectVariant(experiment *Experiment, userID string) Variant {
	if len(experiment.Variants) == 0 {
		panic("experiment has no variants")
	}

	// If only one variant, return it
	if len(experiment.Variants) == 1 {
		return experiment.Variants[0]
	}

	// Use consistent hashing to ensure the same user always gets the same variant
	hash := s.hashUser(experiment.ID.String(), userID)

	// Select variant based on traffic split
	cumulative := 0
	for _, variant := range experiment.Variants {
		cumulative += variant.TrafficSplit
		if hash < cumulative {
			return variant
		}
	}

	// Fallback to first variant (should never reach here if splits sum to 100)
	return experiment.Variants[0]
}

// SelectVariantRandom selects a variant randomly based on traffic splits
func (s *ABTestingService) SelectVariantRandom(experiment *Experiment) Variant {
	if len(experiment.Variants) == 0 {
		panic("experiment has no variants")
	}

	// Generate random number 0-99
	r := rand.Intn(100)

	// Select variant based on traffic split
	cumulative := 0
	for _, variant := range experiment.Variants {
		cumulative += variant.TrafficSplit
		if r < cumulative {
			return variant
		}
	}

	// Fallback to first variant
	return experiment.Variants[0]
}

// hashUser generates a consistent hash for a user within an experiment
// Returns a value between 0 and 99 for percentage-based selection
func (s *ABTestingService) hashUser(experimentID, userID string) int {
	// Create a deterministic hash based on experiment and user
	data := fmt.Sprintf("%s:%s", experimentID, userID)
	hash := md5.Sum([]byte(data))

	// Convert first 8 bytes to uint64
	num := binary.BigEndian.Uint64(hash[:8])

	// Return value between 0 and 99
	return int(num % 100)
}

// AnalyzeResults analyzes experiment results and determines statistical significance
func (s *ABTestingService) AnalyzeResults(metrics map[string]*ExperimentMetrics) *ExperimentAnalysis {
	analysis := &ExperimentAnalysis{
		VariantResults: make(map[string]*VariantResult),
	}

	// Calculate metrics for each variant
	for variantID, m := range metrics {
		result := &VariantResult{
			VariantID:    variantID,
			RequestCount: m.RequestCount,
			SuccessRate:  0,
			ErrorRate:    0,
			AvgLatency:   m.AvgLatencyMs,
			AvgCost:      0,
			AvgTokens:    0,
		}

		if m.RequestCount > 0 {
			result.SuccessRate = float64(m.SuccessCount) / float64(m.RequestCount)
			result.ErrorRate = float64(m.ErrorCount) / float64(m.RequestCount)
			result.AvgCost = m.TotalCost / float64(m.RequestCount)
			result.AvgTokens = float64(m.TotalTokens) / float64(m.RequestCount)
		}

		// Calculate user satisfaction
		totalFeedback := m.UserFeedbackPositive + m.UserFeedbackNegative
		if totalFeedback > 0 {
			result.UserSatisfaction = float64(m.UserFeedbackPositive) / float64(totalFeedback)
		}

		analysis.VariantResults[variantID] = result
	}

	// Determine winner (simplified - in production use proper statistical tests)
	analysis.Winner = s.determineWinner(analysis.VariantResults)
	analysis.IsSignificant = s.isStatisticallySignificant(analysis.VariantResults)

	return analysis
}

// determineWinner determines the best performing variant
func (s *ABTestingService) determineWinner(results map[string]*VariantResult) string {
	if len(results) == 0 {
		return ""
	}

	type scoredVariant struct {
		variantID string
		score     float64
	}

	scores := make([]scoredVariant, 0, len(results))

	// Calculate composite score for each variant
	for variantID, result := range results {
		// Weighted score: success rate (40%), latency (30%), cost (20%), satisfaction (10%)
		score := result.SuccessRate*0.4 +
			(1.0-normalizeLatency(result.AvgLatency))*0.3 +
			(1.0-normalizeCost(result.AvgCost))*0.2 +
			result.UserSatisfaction*0.1

		scores = append(scores, scoredVariant{variantID, score})
	}

	// Sort by score descending
	sort.Slice(scores, func(i, j int) bool {
		return scores[i].score > scores[j].score
	})

	return scores[0].variantID
}

// normalizeLatency normalizes latency to 0-1 range (lower is better)
func normalizeLatency(latency float64) float64 {
	// Assume 0ms is best (0.0) and 5000ms is worst (1.0)
	if latency <= 0 {
		return 0
	}
	if latency >= 5000 {
		return 1
	}
	return latency / 5000
}

// normalizeCost normalizes cost to 0-1 range (lower is better)
func normalizeCost(cost float64) float64 {
	// Assume $0 is best (0.0) and $0.10 per request is worst (1.0)
	if cost <= 0 {
		return 0
	}
	if cost >= 0.10 {
		return 1
	}
	return cost / 0.10
}

// isStatisticallySignificant determines if results are statistically significant
// This is a simplified version - in production use proper statistical tests (chi-square, t-test)
func (s *ABTestingService) isStatisticallySignificant(results map[string]*VariantResult) bool {
	// Require at least 2 variants and 100 requests per variant
	if len(results) < 2 {
		return false
	}

	for _, result := range results {
		if result.RequestCount < 100 {
			return false
		}
	}

	// Check if there's a meaningful difference in success rates
	var successRates []float64
	for _, result := range results {
		successRates = append(successRates, result.SuccessRate)
	}

	sort.Float64s(successRates)
	if len(successRates) >= 2 {
		// Require at least 5% difference between best and worst
		diff := successRates[len(successRates)-1] - successRates[0]
		return diff >= 0.05
	}

	return false
}

// ExperimentAnalysis contains the analysis of an experiment
type ExperimentAnalysis struct {
	VariantResults map[string]*VariantResult
	Winner         string
	IsSignificant  bool
}

// VariantResult contains metrics for a single variant
type VariantResult struct {
	VariantID        string
	RequestCount     int
	SuccessRate      float64
	ErrorRate        float64
	AvgLatency       float64
	AvgCost          float64
	AvgTokens        float64
	UserSatisfaction float64
}

// GetRecommendation provides a recommendation based on the analysis
func (a *ExperimentAnalysis) GetRecommendation() string {
	if !a.IsSignificant {
		return "Continue experiment - not enough data for statistical significance"
	}

	winner := a.VariantResults[a.Winner]
	if winner == nil {
		return "No clear winner"
	}

	return fmt.Sprintf(
		"Variant %s is the clear winner with %.2f%% success rate, avg latency of %.2fms, and avg cost of $%.4f",
		a.Winner, winner.SuccessRate*100, winner.AvgLatency, winner.AvgCost,
	)
}

// CalculateTrafficSplits calculates optimal traffic splits for an experiment
func (s *ABTestingService) CalculateTrafficSplits(numVariants int) []int {
	if numVariants <= 0 {
		return []int{}
	}

	// Equal split
	baseShare := 100 / numVariants
	splits := make([]int, numVariants)

	for i := 0; i < numVariants; i++ {
		splits[i] = baseShare
	}

	// Distribute remainder
	remainder := 100 - (baseShare * numVariants)
	for i := 0; i < remainder; i++ {
		splits[i]++
	}

	return splits
}

// CalculateRequiredSampleSize calculates the required sample size for statistical significance
func (s *ABTestingService) CalculateRequiredSampleSize(baselineRate, minimumDetectableEffect, alpha, power float64) int {
	// Simplified calculation - in production use proper power analysis
	// This uses a rough approximation of the formula

	if baselineRate <= 0 || baselineRate >= 1 {
		return 1000 // Default minimum
	}

	// Rough approximation: n ≈ 16 * p * (1-p) / (effect^2)
	p := baselineRate
	effect := minimumDetectableEffect

	n := 16 * p * (1 - p) / (effect * effect)

	// Apply corrections for alpha and power
	alphaCorrection := 2.0 // For alpha = 0.05
	powerCorrection := 1.5 // For power = 0.80

	n = n * alphaCorrection * powerCorrection

	// Ensure minimum sample size
	if n < 100 {
		n = 100
	}

	return int(n)
}
