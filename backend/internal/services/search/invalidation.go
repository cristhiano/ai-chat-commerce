package search

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

// CacheInvalidationService handles cache invalidation strategies
type CacheInvalidationService struct {
	redisClient *redis.Client
	db          *gorm.DB
}

// NewCacheInvalidationService creates a new cache invalidation service
func NewCacheInvalidationService(redisClient *redis.Client, db *gorm.DB) *CacheInvalidationService {
	return &CacheInvalidationService{
		redisClient: redisClient,
		db:          db,
	}
}

// InvalidateProductCache invalidates cache when product data changes
func (cis *CacheInvalidationService) InvalidateProductCache(ctx context.Context, productID string) error {
	// Get all search cache entries that might contain this product
	var cacheEntries []SearchCache
	err := cis.db.WithContext(ctx).
		Where("results::text LIKE ?", "%"+productID+"%").
		Find(&cacheEntries).Error

	if err != nil {
		return fmt.Errorf("failed to find cache entries for product: %w", err)
	}

	// Invalidate Redis cache for each entry
	for _, entry := range cacheEntries {
		cacheKey := fmt.Sprintf("search:%s:%s", entry.Query, string(entry.Filters))
		err = cis.redisClient.Del(ctx, cacheKey).Err()
		if err != nil {
			// Log error but continue with other entries
			fmt.Printf("Failed to invalidate cache key %s: %v\n", cacheKey, err)
		}
	}

	// Delete database cache entries
	err = cis.db.WithContext(ctx).
		Where("results::text LIKE ?", "%"+productID+"%").
		Delete(&SearchCache{}).Error

	if err != nil {
		return fmt.Errorf("failed to delete database cache entries: %w", err)
	}

	return nil
}

// InvalidateCategoryCache invalidates cache when category data changes
func (cis *CacheInvalidationService) InvalidateCategoryCache(ctx context.Context, categoryID string) error {
	// Get all search cache entries that might contain this category
	var cacheEntries []SearchCache
	err := cis.db.WithContext(ctx).
		Where("filters::text LIKE ?", "%"+categoryID+"%").
		Find(&cacheEntries).Error

	if err != nil {
		return fmt.Errorf("failed to find cache entries for category: %w", err)
	}

	// Invalidate Redis cache for each entry
	for _, entry := range cacheEntries {
		cacheKey := fmt.Sprintf("search:%s:%s", entry.Query, string(entry.Filters))
		err = cis.redisClient.Del(ctx, cacheKey).Err()
		if err != nil {
			fmt.Printf("Failed to invalidate cache key %s: %v\n", cacheKey, err)
		}
	}

	// Delete database cache entries
	err = cis.db.WithContext(ctx).
		Where("filters::text LIKE ?", "%"+categoryID+"%").
		Delete(&SearchCache{}).Error

	if err != nil {
		return fmt.Errorf("failed to delete database cache entries: %w", err)
	}

	return nil
}

// InvalidateQueryCache invalidates cache for specific query patterns
func (cis *CacheInvalidationService) InvalidateQueryCache(ctx context.Context, queryPattern string) error {
	// Find cache entries matching the pattern
	var cacheEntries []SearchCache
	err := cis.db.WithContext(ctx).
		Where("query ILIKE ?", "%"+queryPattern+"%").
		Find(&cacheEntries).Error

	if err != nil {
		return fmt.Errorf("failed to find cache entries for query pattern: %w", err)
	}

	// Invalidate Redis cache for each entry
	for _, entry := range cacheEntries {
		cacheKey := fmt.Sprintf("search:%s:%s", entry.Query, string(entry.Filters))
		err = cis.redisClient.Del(ctx, cacheKey).Err()
		if err != nil {
			fmt.Printf("Failed to invalidate cache key %s: %v\n", cacheKey, err)
		}
	}

	// Delete database cache entries
	err = cis.db.WithContext(ctx).
		Where("query ILIKE ?", "%"+queryPattern+"%").
		Delete(&SearchCache{}).Error

	if err != nil {
		return fmt.Errorf("failed to delete database cache entries: %w", err)
	}

	return nil
}

// InvalidateExpiredCache removes expired cache entries
func (cis *CacheInvalidationService) InvalidateExpiredCache(ctx context.Context) error {
	// Get expired cache entries
	var expiredEntries []SearchCache
	err := cis.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Find(&expiredEntries).Error

	if err != nil {
		return fmt.Errorf("failed to find expired cache entries: %w", err)
	}

	// Delete from Redis
	for _, entry := range expiredEntries {
		cacheKey := fmt.Sprintf("search:%s:%s", entry.Query, string(entry.Filters))
		cis.redisClient.Del(ctx, cacheKey)
	}

	// Delete from database
	err = cis.db.WithContext(ctx).
		Where("expires_at < ?", time.Now()).
		Delete(&SearchCache{}).Error

	if err != nil {
		return fmt.Errorf("failed to delete expired cache entries: %w", err)
	}

	return nil
}

// InvalidateAllCache clears all search cache
func (cis *CacheInvalidationService) InvalidateAllCache(ctx context.Context) error {
	// Clear Redis cache
	keys, err := cis.redisClient.Keys(ctx, "search:*").Result()
	if err != nil {
		return fmt.Errorf("failed to get cache keys: %w", err)
	}

	if len(keys) > 0 {
		err = cis.redisClient.Del(ctx, keys...).Err()
		if err != nil {
			return fmt.Errorf("failed to clear Redis cache: %w", err)
		}
	}

	// Clear database cache
	err = cis.db.WithContext(ctx).Delete(&SearchCache{}).Error
	if err != nil {
		return fmt.Errorf("failed to clear database cache: %w", err)
	}

	return nil
}

// ScheduleCacheCleanup schedules periodic cache cleanup
func (cis *CacheInvalidationService) ScheduleCacheCleanup(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				err := cis.InvalidateExpiredCache(ctx)
				if err != nil {
					fmt.Printf("Failed to cleanup expired cache: %v\n", err)
				}
			}
		}
	}()
}

// GetInvalidationStats returns invalidation statistics
func (cis *CacheInvalidationService) GetInvalidationStats(ctx context.Context) (*InvalidationStats, error) {
	var totalEntries int64
	var expiredEntries int64
	var recentInvalidations int64

	// Count total entries
	err := cis.db.WithContext(ctx).Model(&SearchCache{}).Count(&totalEntries).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count total entries: %w", err)
	}

	// Count expired entries
	err = cis.db.WithContext(ctx).Model(&SearchCache{}).Where("expires_at < ?", time.Now()).Count(&expiredEntries).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count expired entries: %w", err)
	}

	// Count recent invalidations (entries deleted in last hour)
	err = cis.db.WithContext(ctx).Model(&SearchCache{}).
		Where("updated_at > ? AND deleted_at IS NOT NULL", time.Now().Add(-time.Hour)).
		Count(&recentInvalidations).Error
	if err != nil {
		// This might fail if soft deletes aren't enabled, so we'll set to 0
		recentInvalidations = 0
	}

	return &InvalidationStats{
		TotalEntries:        int(totalEntries),
		ExpiredEntries:      int(expiredEntries),
		RecentInvalidations: int(recentInvalidations),
		CacheHitRate:        0.0, // Would be calculated from metrics
		LastCleanupTime:     time.Now(),
	}, nil
}

// InvalidationStats represents cache invalidation statistics
type InvalidationStats struct {
	TotalEntries        int       `json:"total_entries"`
	ExpiredEntries      int       `json:"expired_entries"`
	RecentInvalidations int       `json:"recent_invalidations"`
	CacheHitRate        float64   `json:"cache_hit_rate"`
	LastCleanupTime     time.Time `json:"last_cleanup_time"`
}
