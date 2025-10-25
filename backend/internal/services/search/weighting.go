package search

import (
	"context"
	"math"
	"strings"

	"gorm.io/gorm"
)

// WeightingService handles field weighting for relevance scoring
type WeightingService struct {
	db *gorm.DB
}

// NewWeightingService creates a new weighting service
func NewWeightingService(db *gorm.DB) *WeightingService {
	return &WeightingService{db: db}
}

// FieldWeights represents the weights for different fields
type FieldWeights struct {
	Name        float64 `json:"name"`
	Description float64 `json:"description"`
	Category    float64 `json:"category"`
	SKU         float64 `json:"sku"`
	Tags        float64 `json:"tags"`
}

// DefaultFieldWeights returns default field weights
func (ws *WeightingService) DefaultFieldWeights() FieldWeights {
	return FieldWeights{
		Name:        3.0,
		Description: 1.0,
		Category:    2.0,
		SKU:         2.5,
		Tags:        1.5,
	}
}

// CalculateWeightedScore calculates weighted relevance score
func (ws *WeightingService) CalculateWeightedScore(query string, product ProductResult, weights FieldWeights) float64 {
	queryTerms := strings.Fields(strings.ToLower(query))

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

		// Apply field weights
		fieldScore := (nameTF * weights.Name) +
			(descTF * weights.Description) +
			(categoryTF * weights.Category) +
			(skuTF * weights.SKU)

		// Apply term importance boost
		termScore := ws.calculateTermImportance(term, fieldScore)

		totalScore += termScore
	}

	return totalScore
}

// calculateTermImportance calculates importance boost for terms
func (ws *WeightingService) calculateTermImportance(term string, baseScore float64) float64 {
	// Boost for longer terms (more specific)
	lengthBoost := math.Log(float64(len(term)) + 1.0)

	// Boost for rare terms (assuming shorter terms are more common)
	rarityBoost := 1.0 + (1.0 / math.Max(float64(len(term)), 1.0))

	return baseScore * lengthBoost * rarityBoost
}

// ApplyDynamicWeights applies dynamic weights based on query characteristics
func (ws *WeightingService) ApplyDynamicWeights(query string) FieldWeights {
	weights := ws.DefaultFieldWeights()

	queryLower := strings.ToLower(query)

	// If query is short, boost name and SKU weights
	if len(strings.Fields(query)) <= 2 {
		weights.Name *= 1.2
		weights.SKU *= 1.1
	}

	// If query contains numbers, boost SKU weight
	if ws.containsNumbers(query) {
		weights.SKU *= 1.5
	}

	// If query is long, boost description weight
	if len(strings.Fields(query)) > 3 {
		weights.Description *= 1.3
	}

	// If query looks like a category, boost category weight
	if ws.looksLikeCategory(queryLower) {
		weights.Category *= 1.4
	}

	return weights
}

// containsNumbers checks if string contains numbers
func (ws *WeightingService) containsNumbers(s string) bool {
	for _, r := range s {
		if r >= '0' && r <= '9' {
			return true
		}
	}
	return false
}

// looksLikeCategory checks if query looks like a category name
func (ws *WeightingService) looksLikeCategory(query string) bool {
	// Simple heuristic: if query is a single word and starts with capital letter
	words := strings.Fields(query)
	if len(words) == 1 {
		word := words[0]
		if len(word) > 0 && word[0] >= 'A' && word[0] <= 'Z' {
			return true
		}
	}
	return false
}

// CalculateFieldRelevance calculates relevance for each field separately
func (ws *WeightingService) CalculateFieldRelevance(query string, product ProductResult) map[string]float64 {
	queryTerms := strings.Fields(strings.ToLower(query))

	nameText := strings.ToLower(product.Name)
	descText := strings.ToLower(product.Description)
	categoryText := strings.ToLower(product.Category)
	skuText := strings.ToLower(product.SKU)

	relevance := make(map[string]float64)

	for _, term := range queryTerms {
		// Name relevance
		if strings.Contains(nameText, term) {
			relevance["name"] += ws.calculateTermRelevance(term, nameText)
		}

		// Description relevance
		if strings.Contains(descText, term) {
			relevance["description"] += ws.calculateTermRelevance(term, descText)
		}

		// Category relevance
		if strings.Contains(categoryText, term) {
			relevance["category"] += ws.calculateTermRelevance(term, categoryText)
		}

		// SKU relevance
		if strings.Contains(skuText, term) {
			relevance["sku"] += ws.calculateTermRelevance(term, skuText)
		}
	}

	return relevance
}

// calculateTermRelevance calculates relevance score for a term in a field
func (ws *WeightingService) calculateTermRelevance(term, fieldText string) float64 {
	// Count occurrences
	count := float64(strings.Count(fieldText, term))

	// Boost for exact match
	if fieldText == term {
		return count * 2.0
	}

	// Boost for prefix match
	if strings.HasPrefix(fieldText, term) {
		return count * 1.5
	}

	// Boost for word boundary match
	words := strings.Fields(fieldText)
	for _, word := range words {
		if word == term {
			return count * 1.3
		}
	}

	return count
}

// GetOptimalWeights returns optimal weights based on search analytics
func (ws *WeightingService) GetOptimalWeights(ctx context.Context) (FieldWeights, error) {
	// This would typically analyze search analytics to determine optimal weights
	// For now, return default weights
	return ws.DefaultFieldWeights(), nil
}

// UpdateWeights updates field weights based on search performance
func (ws *WeightingService) UpdateWeights(ctx context.Context, weights FieldWeights) error {
	// This would typically store weights in database or config
	// For now, just return nil
	return nil
}
