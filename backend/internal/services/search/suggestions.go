package search

import (
	"context"
	"strings"
	"time"

	"chat-ecommerce-backend/internal/models/search"

	"gorm.io/gorm"
)

// SuggestionsService handles suggestion caching and management
type SuggestionsService struct {
	db *gorm.DB
}

// NewSuggestionsService creates a new suggestions service
func NewSuggestionsService(db *gorm.DB) *SuggestionsService {
	return &SuggestionsService{db: db}
}

// CacheSuggestions caches suggestions for a query
func (ss *SuggestionsService) CacheSuggestions(ctx context.Context, query string, suggestions []string, ttl time.Duration) error {
	if len(suggestions) == 0 {
		return nil
	}

	// Convert suggestions to JSON string
	suggestionsStr := strings.Join(suggestions, ",")

	// Check if suggestion already exists
	var existing search.SearchSuggestion
	err := ss.db.WithContext(ctx).
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
		return ss.db.WithContext(ctx).Create(&suggestion).Error
	} else if err == nil {
		// Update existing suggestion
		existing.Suggestions = suggestionsStr
		existing.Occurrences++
		existing.LastUsed = time.Now()
		return ss.db.WithContext(ctx).Save(&existing).Error
	}

	return err
}

// GetCachedSuggestions retrieves cached suggestions
func (ss *SuggestionsService) GetCachedSuggestions(ctx context.Context, query string) ([]string, error) {
	var suggestion search.SearchSuggestion

	err := ss.db.WithContext(ctx).
		Where("query = ?", query).
		First(&suggestion).Error

	if err != nil {
		return nil, err
	}

	// Parse suggestions from comma-separated string
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

// UpdateSuggestionUsage updates suggestion usage statistics
func (ss *SuggestionsService) UpdateSuggestionUsage(ctx context.Context, query string) error {
	return ss.db.WithContext(ctx).
		Model(&search.SearchSuggestion{}).
		Where("query = ?", query).
		Updates(map[string]interface{}{
			"occurrences": gorm.Expr("occurrences + 1"),
			"last_used":   time.Now(),
		}).Error
}

// GetPopularSuggestions gets popular suggestions based on usage
func (ss *SuggestionsService) GetPopularSuggestions(ctx context.Context, limit int) ([]PopularSuggestion, error) {
	if limit <= 0 {
		limit = 10
	}

	var suggestions []PopularSuggestion

	err := ss.db.WithContext(ctx).
		Table("search_suggestions").
		Select("query, occurrences, last_used").
		Order("occurrences DESC, last_used DESC").
		Limit(limit).
		Scan(&suggestions).Error

	return suggestions, err
}

// GetRecentSuggestions gets recently used suggestions
func (ss *SuggestionsService) GetRecentSuggestions(ctx context.Context, limit int) ([]RecentSuggestion, error) {
	if limit <= 0 {
		limit = 10
	}

	var suggestions []RecentSuggestion

	err := ss.db.WithContext(ctx).
		Table("search_suggestions").
		Select("query, last_used").
		Where("last_used > ?", time.Now().AddDate(0, 0, -7)).
		Order("last_used DESC").
		Limit(limit).
		Scan(&suggestions).Error

	return suggestions, err
}

// CleanupExpiredSuggestions removes old, unused suggestions
func (ss *SuggestionsService) CleanupExpiredSuggestions(ctx context.Context) error {
	// Remove suggestions not used in the last 30 days with low occurrence count
	return ss.db.WithContext(ctx).
		Where("last_used < ? AND occurrences < ?",
			time.Now().AddDate(0, 0, -30), 2).
		Delete(&search.SearchSuggestion{}).Error
}

// GetSuggestionStats returns suggestion statistics
func (ss *SuggestionsService) GetSuggestionStats(ctx context.Context) (*SuggestionStats, error) {
	var totalCount int64
	var activeCount int64
	var expiredCount int64

	// Count total suggestions
	err := ss.db.WithContext(ctx).
		Model(&search.SearchSuggestion{}).
		Count(&totalCount).Error
	if err != nil {
		return nil, err
	}

	// Count active suggestions (used in last 7 days)
	err = ss.db.WithContext(ctx).
		Model(&search.SearchSuggestion{}).
		Where("last_used > ?", time.Now().AddDate(0, 0, -7)).
		Count(&activeCount).Error
	if err != nil {
		return nil, err
	}

	// Count expired suggestions
	err = ss.db.WithContext(ctx).
		Model(&search.SearchSuggestion{}).
		Where("last_used < ?", time.Now().AddDate(0, 0, -30)).
		Count(&expiredCount).Error
	if err != nil {
		return nil, err
	}

	return &SuggestionStats{
		TotalSuggestions:   int(totalCount),
		ActiveSuggestions:  int(activeCount),
		ExpiredSuggestions: int(expiredCount),
	}, nil
}

// PopularSuggestion represents a popular suggestion
type PopularSuggestion struct {
	Query       string    `json:"query"`
	Occurrences int       `json:"occurrences"`
	LastUsed    time.Time `json:"last_used"`
}

// RecentSuggestion represents a recent suggestion
type RecentSuggestion struct {
	Query    string    `json:"query"`
	LastUsed time.Time `json:"last_used"`
}

// SuggestionStats represents suggestion statistics
type SuggestionStats struct {
	TotalSuggestions   int `json:"total_suggestions"`
	ActiveSuggestions  int `json:"active_suggestions"`
	ExpiredSuggestions int `json:"expired_suggestions"`
}
