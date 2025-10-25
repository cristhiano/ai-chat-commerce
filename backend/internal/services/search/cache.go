package search

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"chat-ecommerce-backend/internal/models/search"

	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// CacheService handles search result caching
type CacheService struct {
	db *gorm.DB
}

// NewCacheService creates a new cache service
func NewCacheService(db *gorm.DB) *CacheService {
	return &CacheService{db: db}
}

// GetCachedResults retrieves cached search results
func (cs *CacheService) GetCachedResults(ctx context.Context, query string, filters map[string]interface{}) (*SearchResult, error) {
	var cache search.SearchCache

	// Create filter key
	filterKey := cs.createFilterKey(filters)

	// Query cache
	err := cs.db.WithContext(ctx).Where("query = ? AND filters = ? AND expires_at > ?",
		query, filterKey, time.Now()).First(&cache).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to query cache: %w", err)
	}

	// Parse cached results
	var result SearchResult
	if err := json.Unmarshal(cache.Results, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached results: %w", err)
	}

	return &result, nil
}

// SetCachedResults stores search results in cache
func (cs *CacheService) SetCachedResults(ctx context.Context, query string, filters map[string]interface{}, result *SearchResult, ttl time.Duration) error {
	// Create filter key
	filterKey := cs.createFilterKey(filters)

	// Serialize results
	resultsJSON, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal results: %w", err)
	}

	// Create cache entry
	cache := search.SearchCache{
		Query:       query,
		Filters:     datatypes.JSON(filterKey),
		Results:     datatypes.JSON(resultsJSON),
		ResultCount: len(result.Results),
		ExpiresAt:   time.Now().Add(ttl),
	}

	// Store in database
	if err := cs.db.WithContext(ctx).Create(&cache).Error; err != nil {
		return fmt.Errorf("failed to store cache: %w", err)
	}

	return nil
}

// InvalidateCache removes cached results for a query
func (cs *CacheService) InvalidateCache(ctx context.Context, query string) error {
	return cs.db.WithContext(ctx).Where("query = ?", query).Delete(&search.SearchCache{}).Error
}

// CleanExpiredCache removes expired cache entries
func (cs *CacheService) CleanExpiredCache(ctx context.Context) error {
	return cs.db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&search.SearchCache{}).Error
}

// GetCacheStats returns cache statistics
func (cs *CacheService) GetCacheStats(ctx context.Context) (*CacheStats, error) {
	var totalCount int64
	var expiredCount int64

	// Count total cache entries
	if err := cs.db.WithContext(ctx).Model(&search.SearchCache{}).Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count total cache entries: %w", err)
	}

	// Count expired cache entries
	if err := cs.db.WithContext(ctx).Model(&search.SearchCache{}).Where("expires_at < ?", time.Now()).Count(&expiredCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count expired cache entries: %w", err)
	}

	return &CacheStats{
		TotalEntries:   int(totalCount),
		ExpiredEntries: int(expiredCount),
		ActiveEntries:  int(totalCount - expiredCount),
	}, nil
}

// createFilterKey creates a consistent key for filters
func (cs *CacheService) createFilterKey(filters map[string]interface{}) string {
	if len(filters) == 0 {
		return "{}"
	}

	// Sort keys for consistent ordering
	filterJSON, err := json.Marshal(filters)
	if err != nil {
		return "{}"
	}

	return string(filterJSON)
}

// CacheStats represents cache statistics
type CacheStats struct {
	TotalEntries   int `json:"total_entries"`
	ExpiredEntries int `json:"expired_entries"`
	ActiveEntries  int `json:"active_entries"`
}

// StartCacheCleanup starts a background goroutine to clean expired cache entries
func (cs *CacheService) StartCacheCleanup(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := cs.CleanExpiredCache(ctx); err != nil {
					// Log error but continue
					fmt.Printf("Failed to clean expired cache: %v\n", err)
				}
			}
		}
	}()
}
