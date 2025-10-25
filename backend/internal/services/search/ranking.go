package search

import (
	"math"
	"strings"

	"github.com/sahilm/fuzzy"
)

// RankingAlgorithm handles search result ranking
type RankingAlgorithm struct {
}

// NewRankingAlgorithm creates a new ranking algorithm
func NewRankingAlgorithm() *RankingAlgorithm {
	return &RankingAlgorithm{}
}

// RankResults ranks search results based on relevance
func (ra *RankingAlgorithm) RankResults(query string, results []ProductResult) []ProductResult {
	if len(results) == 0 {
		return results
	}

	// Calculate TF-IDF scores for each result
	for i := range results {
		results[i].RelevanceScore = ra.calculateTFIDFScore(query, results[i])
	}

	// Sort by relevance score (highest first)
	// Simple bubble sort for small result sets
	for i := 0; i < len(results)-1; i++ {
		for j := 0; j < len(results)-i-1; j++ {
			if results[j].RelevanceScore < results[j+1].RelevanceScore {
				results[j], results[j+1] = results[j+1], results[j]
			}
		}
	}

	return results
}

// calculateTFIDFScore calculates TF-IDF score for a product
func (ra *RankingAlgorithm) calculateTFIDFScore(query string, product ProductResult) float64 {
	queryTerms := strings.Fields(strings.ToLower(query))

	// Combine product text fields
	productText := strings.ToLower(product.Name + " " + product.Description)

	totalScore := 0.0

	for _, term := range queryTerms {
		// Term frequency in product text
		tf := float64(strings.Count(productText, term))

		// Inverse document frequency (simplified)
		// In a real implementation, this would be calculated against the entire corpus
		idf := math.Log(1000.0 / (1.0 + tf)) // Simplified IDF

		// TF-IDF score
		tfidf := tf * idf

		// Weight by field importance
		if strings.Contains(strings.ToLower(product.Name), term) {
			tfidf *= 2.0 // Name matches are more important
		}

		totalScore += tfidf
	}

	return totalScore
}

// FuzzyMatch performs fuzzy matching on search terms
func (ra *RankingAlgorithm) FuzzyMatch(query string, candidates []string) []fuzzy.Match {
	if query == "" || len(candidates) == 0 {
		return []fuzzy.Match{}
	}

	matches := fuzzy.Find(query, candidates)
	return matches
}

// CalculateSimilarity calculates similarity between two strings
func (ra *RankingAlgorithm) CalculateSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}

	// Simple Levenshtein distance-based similarity
	distance := ra.levenshteinDistance(s1, s2)
	maxLen := math.Max(float64(len(s1)), float64(len(s2)))

	return 1.0 - (float64(distance) / maxLen)
}

// levenshteinDistance calculates Levenshtein distance between two strings
func (ra *RankingAlgorithm) levenshteinDistance(s1, s2 string) int {
	r1, r2 := []rune(s1), []rune(s2)
	rows := len(r1) + 1
	cols := len(r2) + 1

	d := make([][]int, rows)
	for i := range d {
		d[i] = make([]int, cols)
	}

	for i := 1; i < rows; i++ {
		d[i][0] = i
	}

	for j := 1; j < cols; j++ {
		d[0][j] = j
	}

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

// EnhancedTFIDFScore calculates enhanced TF-IDF score with field weighting
func (ra *RankingAlgorithm) EnhancedTFIDFScore(query string, product ProductResult) float64 {
	queryTerms := strings.Fields(strings.ToLower(query))

	// Combine product text fields with weights
	nameText := strings.ToLower(product.Name)
	descText := strings.ToLower(product.Description)
	categoryText := strings.ToLower(product.Category)
	skuText := strings.ToLower(product.SKU)

	totalScore := 0.0

	for _, term := range queryTerms {
		// Calculate term frequency for each field
		nameTF := float64(strings.Count(nameText, term))
		descTF := float64(strings.Count(descText, term))
		categoryTF := float64(strings.Count(categoryText, term))
		skuTF := float64(strings.Count(skuText, term))

		// Field weights (higher = more important)
		nameWeight := 3.0
		descWeight := 1.0
		categoryWeight := 2.0
		skuWeight := 2.5

		// Calculate weighted term frequency
		weightedTF := (nameTF * nameWeight) + (descTF * descWeight) +
			(categoryTF * categoryWeight) + (skuTF * skuWeight)

		// Calculate IDF (simplified - in production, this would be pre-calculated)
		idf := math.Log(1000.0 / (1.0 + weightedTF))

		// TF-IDF score
		tfidf := weightedTF * idf

		// Boost for exact matches
		if nameText == term {
			tfidf *= 2.0
		} else if strings.HasPrefix(nameText, term) {
			tfidf *= 1.5
		}

		totalScore += tfidf
	}

	return totalScore
}

// CalculateBM25Score calculates BM25 score (alternative to TF-IDF)
func (ra *RankingAlgorithm) CalculateBM25Score(query string, product ProductResult) float64 {
	queryTerms := strings.Fields(strings.ToLower(query))

	// Document length (total words in product text)
	docText := strings.ToLower(product.Name + " " + product.Description)
	docLength := len(strings.Fields(docText))

	// Average document length (simplified)
	avgDocLength := 50.0

	// BM25 parameters
	k1 := 1.2
	b := 0.75

	totalScore := 0.0

	for _, term := range queryTerms {
		// Term frequency in document
		tf := float64(strings.Count(docText, term))

		// Document frequency (simplified - would be pre-calculated in production)
		df := 100.0         // Assume term appears in 100 documents
		totalDocs := 1000.0 // Total documents

		// IDF component
		idf := math.Log((totalDocs - df + 0.5) / (df + 0.5))

		// BM25 score
		bm25 := idf * (tf * (k1 + 1)) / (tf + k1*(1-b+b*(float64(docLength)/avgDocLength)))

		totalScore += bm25
	}

	return totalScore
}

// CalculateRelevanceScore calculates comprehensive relevance score
func (ra *RankingAlgorithm) CalculateRelevanceScore(query string, product ProductResult) float64 {
	// Get base scores
	tfidfScore := ra.EnhancedTFIDFScore(query, product)
	bm25Score := ra.CalculateBM25Score(query, product)

	// Combine scores with weights
	combinedScore := (tfidfScore * 0.6) + (bm25Score * 0.4)

	// Apply field-specific boosts
	combinedScore = ra.ApplyFieldBoosts(query, product, combinedScore)

	// Apply popularity boost
	combinedScore = ra.ApplyPopularityBoost(product, combinedScore)

	return combinedScore
}

// ApplyFieldBoosts applies field-specific relevance boosts
func (ra *RankingAlgorithm) ApplyFieldBoosts(query string, product ProductResult, baseScore float64) float64 {
	queryLower := strings.ToLower(query)
	nameLower := strings.ToLower(product.Name)
	categoryLower := strings.ToLower(product.Category)
	skuLower := strings.ToLower(product.SKU)

	score := baseScore

	// Exact name match gets highest boost
	if nameLower == queryLower {
		score *= 3.0
	} else if strings.HasPrefix(nameLower, queryLower) {
		score *= 2.0
	} else if strings.Contains(nameLower, queryLower) {
		score *= 1.5
	}

	// SKU match gets high boost
	if skuLower == queryLower {
		score *= 2.5
	} else if strings.Contains(skuLower, queryLower) {
		score *= 1.8
	}

	// Category match gets moderate boost
	if strings.Contains(categoryLower, queryLower) {
		score *= 1.3
	}

	return score
}

// ApplyPopularityBoost applies popularity-based boost
func (ra *RankingAlgorithm) ApplyPopularityBoost(product ProductResult, baseScore float64) float64 {
	// Normalize popularity (assuming max popularity is 1000)
	normalizedPopularity := math.Min(float64(product.Popularity)/1000.0, 1.0)

	// Apply logarithmic boost to prevent dominance
	popularityBoost := 1.0 + (math.Log(1.0+normalizedPopularity) * 0.2)

	return baseScore * popularityBoost
}
