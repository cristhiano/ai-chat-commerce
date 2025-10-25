package search

import (
	"math"
)

// LevenshteinService handles Levenshtein distance calculations
type LevenshteinService struct {
}

// NewLevenshteinService creates a new Levenshtein service
func NewLevenshteinService() *LevenshteinService {
	return &LevenshteinService{}
}

// CalculateDistance calculates Levenshtein distance between two strings
func (ls *LevenshteinService) CalculateDistance(s1, s2 string) int {
	r1, r2 := []rune(s1), []rune(s2)
	rows := len(r1) + 1
	cols := len(r2) + 1

	d := make([][]int, rows)
	for i := range d {
		d[i] = make([]int, cols)
	}

	// Initialize first row and column
	for i := 0; i < rows; i++ {
		d[i][0] = i
	}
	for j := 0; j < cols; j++ {
		d[0][j] = j
	}

	// Fill the matrix
	for i := 1; i < rows; i++ {
		for j := 1; j < cols; j++ {
			cost := 0
			if r1[i-1] != r2[j-1] {
				cost = 1
			}

			d[i][j] = min(
				min(d[i-1][j]+1, d[i][j-1]+1), // deletion vs insertion
				d[i-1][j-1]+cost,              // substitution
			)
		}
	}

	return d[rows-1][cols-1]
}

// CalculateSimilarity calculates similarity percentage between two strings
func (ls *LevenshteinService) CalculateSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}

	distance := ls.CalculateDistance(s1, s2)
	maxLen := math.Max(float64(len(s1)), float64(len(s2)))

	return 1.0 - (float64(distance) / maxLen)
}

// CalculateNormalizedDistance calculates normalized distance (0-1)
func (ls *LevenshteinService) CalculateNormalizedDistance(s1, s2 string) float64 {
	distance := ls.CalculateDistance(s1, s2)
	maxLen := math.Max(float64(len(s1)), float64(len(s2)))

	if maxLen == 0 {
		return 0.0
	}

	return float64(distance) / maxLen
}

// FindClosestMatch finds the closest match from a list of candidates
func (ls *LevenshteinService) FindClosestMatch(query string, candidates []string) (string, float64) {
	if len(candidates) == 0 {
		return "", 0.0
	}

	bestMatch := candidates[0]
	bestSimilarity := ls.CalculateSimilarity(query, bestMatch)

	for _, candidate := range candidates[1:] {
		similarity := ls.CalculateSimilarity(query, candidate)
		if similarity > bestSimilarity {
			bestMatch = candidate
			bestSimilarity = similarity
		}
	}

	return bestMatch, bestSimilarity
}

// FindMatchesWithinThreshold finds all matches within a similarity threshold
func (ls *LevenshteinService) FindMatchesWithinThreshold(query string, candidates []string, threshold float64) []MatchResult {
	var matches []MatchResult

	for _, candidate := range candidates {
		similarity := ls.CalculateSimilarity(query, candidate)
		if similarity >= threshold {
			matches = append(matches, MatchResult{
				Candidate:  candidate,
				Similarity: similarity,
				Distance:   ls.CalculateDistance(query, candidate),
			})
		}
	}

	return matches
}

// CalculateWeightedDistance calculates weighted Levenshtein distance
func (ls *LevenshteinService) CalculateWeightedDistance(s1, s2 string, weights EditWeights) int {
	r1, r2 := []rune(s1), []rune(s2)
	rows := len(r1) + 1
	cols := len(r2) + 1

	d := make([][]int, rows)
	for i := range d {
		d[i] = make([]int, cols)
	}

	// Initialize with weighted costs
	for i := 0; i < rows; i++ {
		d[i][0] = i * weights.Deletion
	}
	for j := 0; j < cols; j++ {
		d[0][j] = j * weights.Insertion
	}

	// Fill the matrix with weighted costs
	for i := 1; i < rows; i++ {
		for j := 1; j < cols; j++ {
			cost := 0
			if r1[i-1] != r2[j-1] {
				cost = weights.Substitution
			}

			d[i][j] = min(
				min(d[i-1][j]+weights.Deletion, d[i][j-1]+weights.Insertion), // deletion vs insertion
				d[i-1][j-1]+cost, // substitution
			)
		}
	}

	return d[rows-1][cols-1]
}

// CalculateJaroWinklerDistance calculates Jaro-Winkler distance
func (ls *LevenshteinService) CalculateJaroWinklerDistance(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}

	// Calculate Jaro distance
	jaro := ls.calculateJaroDistance(s1, s2)

	// Apply Winkler modification
	prefix := ls.calculateCommonPrefix(s1, s2)
	if prefix > 4 {
		prefix = 4
	}

	return jaro + (0.1 * float64(prefix) * (1.0 - jaro))
}

// calculateJaroDistance calculates Jaro distance
func (ls *LevenshteinService) calculateJaroDistance(s1, s2 string) float64 {
	if len(s1) == 0 && len(s2) == 0 {
		return 1.0
	}

	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}

	matchWindow := int(math.Max(float64(len(s1)), float64(len(s2)))/2) - 1
	if matchWindow < 0 {
		matchWindow = 0
	}

	s1Matches := make([]bool, len(s1))
	s2Matches := make([]bool, len(s2))

	matches := 0
	transpositions := 0

	// Find matches
	for i := 0; i < len(s1); i++ {
		start := i - matchWindow
		if start < 0 {
			start = 0
		}
		end := i + matchWindow + 1
		if end > len(s2) {
			end = len(s2)
		}

		for j := start; j < end; j++ {
			if s2Matches[j] || s1[i] != s2[j] {
				continue
			}
			s1Matches[i] = true
			s2Matches[j] = true
			matches++
			break
		}
	}

	if matches == 0 {
		return 0.0
	}

	// Count transpositions
	k := 0
	for i := 0; i < len(s1); i++ {
		if !s1Matches[i] {
			continue
		}
		for !s2Matches[k] {
			k++
		}
		if s1[i] != s2[k] {
			transpositions++
		}
		k++
	}

	return (float64(matches)/float64(len(s1)) +
		float64(matches)/float64(len(s2)) +
		float64(matches-transpositions/2)/float64(matches)) / 3.0
}

// calculateCommonPrefix calculates common prefix length
func (ls *LevenshteinService) calculateCommonPrefix(s1, s2 string) int {
	prefix := 0
	minLen := int(math.Min(float64(len(s1)), float64(len(s2))))

	for i := 0; i < minLen; i++ {
		if s1[i] == s2[i] {
			prefix++
		} else {
			break
		}
	}

	return prefix
}

// MatchResult represents a match result with similarity score
type MatchResult struct {
	Candidate  string  `json:"candidate"`
	Similarity float64 `json:"similarity"`
	Distance   int     `json:"distance"`
}

// EditWeights represents weights for different edit operations
type EditWeights struct {
	Insertion    int `json:"insertion"`
	Deletion     int `json:"deletion"`
	Substitution int `json:"substitution"`
}

// DefaultEditWeights returns default edit weights
func DefaultEditWeights() EditWeights {
	return EditWeights{
		Insertion:    1,
		Deletion:     1,
		Substitution: 1,
	}
}

// min returns the minimum of two integers
