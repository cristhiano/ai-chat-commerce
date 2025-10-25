package search

import (
	"context"
	"fmt"
	"math"
	"strings"

	"github.com/sahilm/fuzzy"
	"gorm.io/gorm"
)

// FuzzyService handles fuzzy matching for search queries
type FuzzyService struct {
	db *gorm.DB
}

// NewFuzzyService creates a new fuzzy service
func NewFuzzyService(db *gorm.DB) *FuzzyService {
	return &FuzzyService{db: db}
}

// FuzzyMatch performs fuzzy matching on search terms
func (fs *FuzzyService) FuzzyMatch(ctx context.Context, query string, candidates []string) []fuzzy.Match {
	if query == "" || len(candidates) == 0 {
		return []fuzzy.Match{}
	}

	matches := fuzzy.Find(query, candidates)
	return matches
}

// FindSimilarProducts finds products with similar names using fuzzy matching
func (fs *FuzzyService) FindSimilarProducts(ctx context.Context, query string, limit int) ([]SimilarProduct, error) {
	if len(query) < 2 {
		return []SimilarProduct{}, nil
	}

	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	// Get product names for fuzzy matching
	var productNames []string
	err := fs.db.WithContext(ctx).
		Table("products").
		Select("name").
		Where("status = ?", "active").
		Pluck("name", &productNames).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get product names: %w", err)
	}

	// Perform fuzzy matching
	matches := fs.FuzzyMatch(ctx, query, productNames)

	// Get top matches
	topMatches := matches
	if len(matches) > limit {
		topMatches = matches[:limit]
	}

	// Get full product details for matches
	var similarProducts []SimilarProduct
	for _, match := range topMatches {
		var product SimilarProduct
		err := fs.db.WithContext(ctx).
			Table("product_search_view").
			Select("id, name, description, price, category_name, sku, popularity").
			Where("name = ? AND status = ?", match.Str, "active").
			First(&product).Error

		if err != nil {
			continue // Skip if product not found
		}

		product.SimilarityScore = float64(match.Score) / 100.0 // Normalize score
		similarProducts = append(similarProducts, product)
	}

	return similarProducts, nil
}

// SuggestCorrections suggests corrections for misspelled queries
func (fs *FuzzyService) SuggestCorrections(ctx context.Context, query string) ([]string, error) {
	if len(query) < 3 {
		return []string{}, nil
	}

	// Get common search terms
	var commonTerms []string
	err := fs.db.WithContext(ctx).
		Table("search_analytics").
		Select("DISTINCT query").
		Where("result_count > 0 AND LENGTH(query) >= 3").
		Where("created_at > NOW() - INTERVAL '30 days'").
		Limit(1000).
		Pluck("query", &commonTerms).Error

	if err != nil {
		return []string{}, nil // Return empty on error
	}

	// Find fuzzy matches
	matches := fs.FuzzyMatch(ctx, query, commonTerms)

	// Extract suggestions
	var suggestions []string
	for _, match := range matches {
		if match.Score > 50 { // Only include good matches
			suggestions = append(suggestions, match.Str)
		}
	}

	// Limit suggestions
	if len(suggestions) > 5 {
		suggestions = suggestions[:5]
	}

	return suggestions, nil
}

// CalculateSimilarity calculates similarity between two strings
func (fs *FuzzyService) CalculateSimilarity(s1, s2 string) float64 {
	if s1 == s2 {
		return 1.0
	}

	if len(s1) == 0 || len(s2) == 0 {
		return 0.0
	}

	// Use Levenshtein distance for similarity
	distance := fs.levenshteinDistance(s1, s2)
	maxLen := math.Max(float64(len(s1)), float64(len(s2)))

	return 1.0 - (float64(distance) / maxLen)
}

// levenshteinDistance calculates Levenshtein distance between two strings
func (fs *FuzzyService) levenshteinDistance(s1, s2 string) int {
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

// min returns the minimum of two integers

// ProcessFuzzyQuery processes a query with fuzzy matching
func (fs *FuzzyService) ProcessFuzzyQuery(ctx context.Context, query string) (*FuzzyQueryResult, error) {
	// Normalize query
	normalizedQuery := strings.TrimSpace(strings.ToLower(query))
	if normalizedQuery == "" {
		return &FuzzyQueryResult{
			OriginalQuery:   query,
			ProcessedQuery:  query,
			Suggestions:     []string{},
			SimilarityScore: 0.0,
		}, nil
	}

	// Check if query needs correction
	suggestions, err := fs.SuggestCorrections(ctx, normalizedQuery)
	if err != nil {
		return nil, fmt.Errorf("failed to get suggestions: %w", err)
	}

	// Determine best query to use
	bestQuery := normalizedQuery
	similarityScore := 1.0

	if len(suggestions) > 0 {
		// Check if first suggestion is significantly better
		firstSuggestion := suggestions[0]
		if fs.CalculateSimilarity(normalizedQuery, firstSuggestion) > 0.8 {
			bestQuery = firstSuggestion
			similarityScore = fs.CalculateSimilarity(normalizedQuery, firstSuggestion)
		}
	}

	return &FuzzyQueryResult{
		OriginalQuery:   query,
		ProcessedQuery:  bestQuery,
		Suggestions:     suggestions,
		SimilarityScore: similarityScore,
	}, nil
}

// SimilarProduct represents a product found through fuzzy matching
type SimilarProduct struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	Description     string  `json:"description"`
	Price           float64 `json:"price"`
	Category        string  `json:"category"`
	SKU             string  `json:"sku"`
	Popularity      int     `json:"popularity"`
	SimilarityScore float64 `json:"similarity_score"`
}

// FuzzyQueryResult represents the result of fuzzy query processing
type FuzzyQueryResult struct {
	OriginalQuery   string   `json:"original_query"`
	ProcessedQuery  string   `json:"processed_query"`
	Suggestions     []string `json:"suggestions"`
	SimilarityScore float64  `json:"similarity_score"`
}
