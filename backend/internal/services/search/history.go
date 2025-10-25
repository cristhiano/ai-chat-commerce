package search

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// HistoryService handles search history functionality
type HistoryService struct {
	db *gorm.DB
}

// NewHistoryService creates a new history service
func NewHistoryService(db *gorm.DB) *HistoryService {
	return &HistoryService{db: db}
}

// SaveSearchHistory saves a search query to history
func (hs *HistoryService) SaveSearchHistory(ctx context.Context, sessionID string, query string, filters map[string]interface{}, resultCount int, userID *uuid.UUID) error {
	history := SearchHistory{
		SessionID:   sessionID,
		Query:       query,
		ResultCount: resultCount,
		UserID:      userID,
		SearchedAt:  time.Now(),
	}

	// Add filters if provided
	if len(filters) > 0 {
		filterJSON, err := json.Marshal(filters)
		if err == nil {
			history.Filters = datatypes.JSON(filterJSON)
		}
	}

	return hs.db.WithContext(ctx).Create(&history).Error
}

// GetSearchHistory returns search history for a session or user
func (hs *HistoryService) GetSearchHistory(ctx context.Context, sessionID string, userID *uuid.UUID, limit int) ([]SearchHistory, error) {
	if limit <= 0 {
		limit = 20
	}
	if limit > 100 {
		limit = 100
	}

	var history []SearchHistory
	query := hs.db.WithContext(ctx)

	// Filter by session or user
	if sessionID != "" {
		query = query.Where("session_id = ?", sessionID)
	} else if userID != nil {
		query = query.Where("user_id = ?", userID)
	} else {
		return nil, fmt.Errorf("either session_id or user_id must be provided")
	}

	// Order by most recent first
	err := query.Order("searched_at DESC").Limit(limit).Find(&history).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get search history: %w", err)
	}

	return history, nil
}

// GetRecentSearches returns recent searches for a user
func (hs *HistoryService) GetRecentSearches(ctx context.Context, userID *uuid.UUID, timeRange time.Duration, limit int) ([]SearchHistory, error) {
	if limit <= 0 {
		limit = 10
	}

	since := time.Now().Add(-timeRange)

	var history []SearchHistory
	err := hs.db.WithContext(ctx).
		Where("user_id = ? AND searched_at > ?", userID, since).
		Order("searched_at DESC").
		Limit(limit).
		Find(&history).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get recent searches: %w", err)
	}

	return history, nil
}

// GetPopularSearches returns popular searches across all users
func (hs *HistoryService) GetPopularSearches(ctx context.Context, timeRange time.Duration, limit int) ([]PopularSearch, error) {
	if limit <= 0 {
		limit = 20
	}

	since := time.Now().Add(-timeRange)

	var popularSearches []PopularSearch
	err := hs.db.WithContext(ctx).
		Table("search_history").
		Select("query, COUNT(*) as frequency, AVG(result_count) as avg_results").
		Where("searched_at > ?", since).
		Group("query").
		Order("frequency DESC").
		Limit(limit).
		Scan(&popularSearches).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get popular searches: %w", err)
	}

	return popularSearches, nil
}

// GetSearchSuggestionsFromHistory returns suggestions based on search history
func (hs *HistoryService) GetSearchSuggestionsFromHistory(ctx context.Context, userID *uuid.UUID, query string, limit int) ([]string, error) {
	if limit <= 0 {
		limit = 10
	}

	var suggestions []string
	err := hs.db.WithContext(ctx).
		Table("search_history").
		Select("DISTINCT query").
		Where("user_id = ? AND query ILIKE ?", userID, "%"+query+"%").
		Order("searched_at DESC").
		Limit(limit).
		Pluck("query", &suggestions).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get search suggestions from history: %w", err)
	}

	return suggestions, nil
}

// ClearSearchHistory clears search history for a session or user
func (hs *HistoryService) ClearSearchHistory(ctx context.Context, sessionID string, userID *uuid.UUID) error {
	query := hs.db.WithContext(ctx)

	if sessionID != "" {
		query = query.Where("session_id = ?", sessionID)
	} else if userID != nil {
		query = query.Where("user_id = ?", userID)
	} else {
		return fmt.Errorf("either session_id or user_id must be provided")
	}

	err := query.Delete(&SearchHistory{}).Error
	if err != nil {
		return fmt.Errorf("failed to clear search history: %w", err)
	}

	return nil
}

// DeleteSearchHistoryEntry deletes a specific search history entry
func (hs *HistoryService) DeleteSearchHistoryEntry(ctx context.Context, entryID uuid.UUID) error {
	err := hs.db.WithContext(ctx).Delete(&SearchHistory{}, entryID).Error
	if err != nil {
		return fmt.Errorf("failed to delete search history entry: %w", err)
	}

	return nil
}

// GetSearchHistoryStats returns search history statistics
func (hs *HistoryService) GetSearchHistoryStats(ctx context.Context, userID *uuid.UUID, timeRange time.Duration) (*SearchHistoryStats, error) {
	since := time.Now().Add(-timeRange)

	var stats SearchHistoryStats

	// Count total searches
	var totalSearches int64
	err := hs.db.WithContext(ctx).
		Model(&SearchHistory{}).
		Where("user_id = ? AND searched_at > ?", userID, since).
		Count(&totalSearches).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count total searches: %w", err)
	}

	// Count unique queries
	var uniqueQueries int64
	err = hs.db.WithContext(ctx).
		Model(&SearchHistory{}).
		Where("user_id = ? AND searched_at > ?", userID, since).
		Select("COUNT(DISTINCT query)").
		Scan(&uniqueQueries).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count unique queries: %w", err)
	}

	// Get most searched query
	var mostSearched string
	err = hs.db.WithContext(ctx).
		Table("search_history").
		Select("query").
		Where("user_id = ? AND searched_at > ?", userID, since).
		Group("query").
		Order("COUNT(*) DESC").
		Limit(1).
		Pluck("query", &mostSearched).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get most searched query: %w", err)
	}

	// Get average results per search
	var avgResults float64
	err = hs.db.WithContext(ctx).
		Model(&SearchHistory{}).
		Where("user_id = ? AND searched_at > ?", userID, since).
		Select("AVG(result_count)").
		Scan(&avgResults).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate average results: %w", err)
	}

	stats.UserID = userID
	stats.TotalSearches = int(totalSearches)
	stats.UniqueQueries = int(uniqueQueries)
	stats.MostSearchedQuery = mostSearched
	stats.AvgResultsPerSearch = avgResults
	stats.TimeRange = timeRange
	stats.GeneratedAt = time.Now()

	return &stats, nil
}

// CleanupOldHistory removes old search history entries
func (hs *HistoryService) CleanupOldHistory(ctx context.Context, retentionDays int) error {
	if retentionDays <= 0 {
		retentionDays = 30 // Default 30 days
	}

	cutoffDate := time.Now().AddDate(0, 0, -retentionDays)

	err := hs.db.WithContext(ctx).
		Where("searched_at < ?", cutoffDate).
		Delete(&SearchHistory{}).Error

	if err != nil {
		return fmt.Errorf("failed to cleanup old search history: %w", err)
	}

	return nil
}

// SearchHistory represents a search history entry
type SearchHistory struct {
	ID          uuid.UUID              `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SessionID   string                 `gorm:"size:100;not null;index"`
	Query       string                 `gorm:"size:500;not null;index"`
	Filters     map[string]interface{} `gorm:"type:jsonb"`
	ResultCount int                    `gorm:"not null"`
	UserID      *uuid.UUID             `gorm:"type:uuid;index"`
	SearchedAt  time.Time              `gorm:"not null;index"`
}

// TableName returns the table name for SearchHistory
func (SearchHistory) TableName() string {
	return "search_history"
}

// PopularSearch represents a popular search query
type PopularSearch struct {
	Query      string  `json:"query"`
	Frequency  int     `json:"frequency"`
	AvgResults float64 `json:"avg_results"`
}

// SearchHistoryStats represents search history statistics
type SearchHistoryStats struct {
	UserID              *uuid.UUID    `json:"user_id"`
	TotalSearches       int           `json:"total_searches"`
	UniqueQueries       int           `json:"unique_queries"`
	MostSearchedQuery   string        `json:"most_searched_query"`
	AvgResultsPerSearch float64       `json:"avg_results_per_search"`
	TimeRange           time.Duration `json:"time_range"`
	GeneratedAt         time.Time     `json:"generated_at"`
}
