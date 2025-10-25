package search

import (
	"context"
	"strings"
	"time"

	"chat-ecommerce-backend/internal/models/search"

	"gorm.io/gorm"
)

// AutocompleteService handles autocomplete functionality
type AutocompleteService struct {
	db *gorm.DB
}

// NewAutocompleteService creates a new autocomplete service
func NewAutocompleteService(db *gorm.DB) *AutocompleteService {
	return &AutocompleteService{db: db}
}

// GetSuggestions returns autocomplete suggestions for a query
func (as *AutocompleteService) GetSuggestions(ctx context.Context, query string, limit int) ([]string, error) {
	if len(query) < 2 {
		return []string{}, nil
	}

	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	queryLower := strings.ToLower(strings.TrimSpace(query))

	// Get suggestions from multiple sources
	var suggestions []string

	// 1. Get cached suggestions
	cachedSuggestions, err := as.getCachedSuggestions(ctx, queryLower)
	if err == nil && len(cachedSuggestions) > 0 {
		return cachedSuggestions[:min(limit, len(cachedSuggestions))], nil
	}

	// 2. Get product name suggestions
	productSuggestions, err := as.getProductSuggestions(ctx, queryLower, limit)
	if err == nil {
		suggestions = append(suggestions, productSuggestions...)
	}

	// 3. Get category suggestions
	categorySuggestions, err := as.getCategorySuggestions(ctx, queryLower, limit/2)
	if err == nil {
		suggestions = append(suggestions, categorySuggestions...)
	}

	// 4. Get brand suggestions
	brandSuggestions, err := as.getBrandSuggestions(ctx, queryLower, limit/2)
	if err == nil {
		suggestions = append(suggestions, brandSuggestions...)
	}

	// 5. Get popular search suggestions
	popularSuggestions, err := as.getPopularSuggestions(ctx, queryLower, limit/2)
	if err == nil {
		suggestions = append(suggestions, popularSuggestions...)
	}

	// Remove duplicates and limit results
	suggestions = as.removeDuplicates(suggestions)
	if len(suggestions) > limit {
		suggestions = suggestions[:limit]
	}

	// Cache suggestions for future use
	go as.cacheSuggestions(context.Background(), queryLower, suggestions)

	return suggestions, nil
}

// getProductSuggestions gets suggestions from product names
func (as *AutocompleteService) getProductSuggestions(ctx context.Context, query string, limit int) ([]string, error) {
	var suggestions []string

	err := as.db.WithContext(ctx).
		Table("products").
		Select("DISTINCT name").
		Where("name ILIKE ? AND status = ?", "%"+query+"%", "active").
		Order("popularity DESC, name ASC").
		Limit(limit).
		Pluck("name", &suggestions).Error

	return suggestions, err
}

// getCategorySuggestions gets suggestions from category names
func (as *AutocompleteService) getCategorySuggestions(ctx context.Context, query string, limit int) ([]string, error) {
	var suggestions []string

	err := as.db.WithContext(ctx).
		Table("categories").
		Select("DISTINCT name").
		Where("name ILIKE ?", "%"+query+"%").
		Order("name ASC").
		Limit(limit).
		Pluck("name", &suggestions).Error

	return suggestions, err
}

// getBrandSuggestions gets suggestions from product brands
func (as *AutocompleteService) getBrandSuggestions(ctx context.Context, query string, limit int) ([]string, error) {
	var suggestions []string

	err := as.db.WithContext(ctx).
		Table("products").
		Select("DISTINCT metadata->>'brand' as brand").
		Where("metadata->>'brand' ILIKE ? AND status = ?", "%"+query+"%", "active").
		Order("brand ASC").
		Limit(limit).
		Pluck("brand", &suggestions).Error

	return suggestions, err
}

// getPopularSuggestions gets suggestions from popular search queries
func (as *AutocompleteService) getPopularSuggestions(ctx context.Context, query string, limit int) ([]string, error) {
	var suggestions []string

	err := as.db.WithContext(ctx).
		Table("search_analytics").
		Select("DISTINCT query").
		Where("query ILIKE ? AND result_count > 0", "%"+query+"%").
		Where("created_at > ?", time.Now().AddDate(0, 0, -30)).
		Group("query").
		Order("COUNT(*) DESC").
		Limit(limit).
		Pluck("query", &suggestions).Error

	return suggestions, err
}

// getCachedSuggestions gets cached suggestions
func (as *AutocompleteService) getCachedSuggestions(ctx context.Context, query string) ([]string, error) {
	var suggestion search.SearchSuggestion

	err := as.db.WithContext(ctx).
		Where("query = ?", query).
		First(&suggestion).Error

	if err != nil {
		return nil, err
	}

	// Parse suggestions from JSON string
	suggestions := strings.Split(suggestion.Suggestions, ",")

	// Clean up suggestions
	var cleanSuggestions []string
	for _, s := range suggestions {
		clean := strings.TrimSpace(s)
		if clean != "" {
			cleanSuggestions = append(cleanSuggestions, clean)
		}
	}

	return cleanSuggestions, nil
}

// cacheSuggestions caches suggestions for future use
func (as *AutocompleteService) cacheSuggestions(ctx context.Context, query string, suggestions []string) {
	if len(suggestions) == 0 {
		return
	}

	// Convert suggestions to comma-separated string
	suggestionsStr := strings.Join(suggestions, ",")

	// Check if suggestion already exists
	var existing search.SearchSuggestion
	err := as.db.WithContext(ctx).
		Where("query = ?", query).
		First(&existing).Error

	if err == gorm.ErrRecordNotFound {
		// Create new suggestion
		suggestion := search.SearchSuggestion{
			Query:       query,
			Suggestions: suggestionsStr,
			Occurrences: 1,
			LastUsed:    time.Now(),
		}
		as.db.WithContext(ctx).Create(&suggestion)
	} else if err == nil {
		// Update existing suggestion
		existing.Suggestions = suggestionsStr
		existing.Occurrences++
		existing.LastUsed = time.Now()
		as.db.WithContext(ctx).Save(&existing)
	}
}

// removeDuplicates removes duplicate suggestions
func (as *AutocompleteService) removeDuplicates(suggestions []string) []string {
	seen := make(map[string]bool)
	var result []string

	for _, suggestion := range suggestions {
		if !seen[suggestion] {
			seen[suggestion] = true
			result = append(result, suggestion)
		}
	}

	return result
}

// LogSuggestionUsage logs when a suggestion is used
func (as *AutocompleteService) LogSuggestionUsage(ctx context.Context, query string, suggestion string) error {
	// Update occurrence count
	err := as.db.WithContext(ctx).
		Model(&search.SearchSuggestion{}).
		Where("query = ?", query).
		Updates(map[string]interface{}{
			"occurrences": gorm.Expr("occurrences + 1"),
			"last_used":   time.Now(),
		}).Error

	return err
}

// GetTrendingSuggestions gets trending suggestions based on recent usage
func (as *AutocompleteService) GetTrendingSuggestions(ctx context.Context, limit int) ([]TrendingSuggestion, error) {
	if limit <= 0 {
		limit = 10
	}

	var suggestions []TrendingSuggestion

	err := as.db.WithContext(ctx).
		Table("search_suggestions").
		Select("query, occurrences, last_used").
		Where("last_used > ?", time.Now().AddDate(0, 0, -7)).
		Order("occurrences DESC, last_used DESC").
		Limit(limit).
		Scan(&suggestions).Error

	return suggestions, err
}

// CleanupOldSuggestions removes old, unused suggestions
func (as *AutocompleteService) CleanupOldSuggestions(ctx context.Context) error {
	// Remove suggestions not used in the last 30 days with low occurrence count
	err := as.db.WithContext(ctx).
		Where("last_used < ? AND occurrences < ?",
			time.Now().AddDate(0, 0, -30), 2).
		Delete(&search.SearchSuggestion{}).Error

	return err
}

// TrendingSuggestion represents a trending suggestion
type TrendingSuggestion struct {
	Query       string    `json:"query"`
	Occurrences int       `json:"occurrences"`
	LastUsed    time.Time `json:"last_used"`
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
