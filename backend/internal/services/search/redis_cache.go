package search

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"chat-ecommerce-backend/internal/models/search"

	"github.com/redis/go-redis/v9"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// RedisCacheService handles Redis-based caching for search results
type RedisCacheService struct {
	redisClient *redis.Client
	db          *gorm.DB
	defaultTTL  time.Duration
}

// NewRedisCacheService creates a new Redis cache service
func NewRedisCacheService(redisClient *redis.Client, db *gorm.DB) *RedisCacheService {
	return &RedisCacheService{
		redisClient: redisClient,
		db:          db,
		defaultTTL:  5 * time.Minute, // Default cache TTL
	}
}

// CacheSearchResults caches search results in Redis
func (rcs *RedisCacheService) CacheSearchResults(ctx context.Context, query string, filters map[string]interface{}, results *SearchResult, ttl time.Duration) error {
	if ttl <= 0 {
		ttl = rcs.defaultTTL
	}

	// Create cache key
	cacheKey := rcs.createCacheKey(query, filters)

	// Serialize results
	resultsJSON, err := json.Marshal(results)
	if err != nil {
		return fmt.Errorf("failed to marshal search results: %w", err)
	}

	// Store in Redis
	err = rcs.redisClient.Set(ctx, cacheKey, resultsJSON, ttl).Err()
	if err != nil {
		return fmt.Errorf("failed to cache search results in Redis: %w", err)
	}

	// Also store metadata in database for analytics
	go rcs.storeCacheMetadata(context.Background(), query, filters, len(results.Results), ttl)

	return nil
}

// GetCachedResults retrieves cached search results from Redis
func (rcs *RedisCacheService) GetCachedResults(ctx context.Context, query string, filters map[string]interface{}) (*SearchResult, error) {
	cacheKey := rcs.createCacheKey(query, filters)

	// Get from Redis
	resultJSON, err := rcs.redisClient.Get(ctx, cacheKey).Result()
	if err != nil {
		if err == redis.Nil {
			return nil, nil // Cache miss
		}
		return nil, fmt.Errorf("failed to get cached results from Redis: %w", err)
	}

	// Deserialize results
	var results SearchResult
	err = json.Unmarshal([]byte(resultJSON), &results)
	if err != nil {
		return nil, fmt.Errorf("failed to unmarshal cached results: %w", err)
	}

	// Update cache hit metrics
	go rcs.updateCacheHitMetrics(context.Background(), query, filters)

	return &results, nil
}

// InvalidateCache removes cached results for a query
func (rcs *RedisCacheService) InvalidateCache(ctx context.Context, query string) error {
	// Create pattern for all cache keys related to this query
	pattern := fmt.Sprintf("search:%s:*", query)

	// Find all matching keys
	keys, err := rcs.redisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return fmt.Errorf("failed to find cache keys: %w", err)
	}

	// Delete all matching keys
	if len(keys) > 0 {
		err = rcs.redisClient.Del(ctx, keys...).Err()
		if err != nil {
			return fmt.Errorf("failed to delete cache keys: %w", err)
		}
	}

	return nil
}

// WarmCache pre-populates cache with popular queries
func (rcs *RedisCacheService) WarmCache(ctx context.Context, popularQueries []string) error {
	for _, query := range popularQueries {
		// Check if already cached
		cacheKey := rcs.createCacheKey(query, map[string]interface{}{})
		exists, err := rcs.redisClient.Exists(ctx, cacheKey).Result()
		if err != nil {
			continue // Skip on error
		}

		if exists > 0 {
			continue // Already cached
		}

		// Perform search and cache results
		processor := NewSearchProcessor(rcs.db)
		results, err := processor.ProcessQuery(ctx, query, map[string]interface{}{}, 1, 20)
		if err != nil {
			continue // Skip on error
		}

		// Cache the results
		err = rcs.CacheSearchResults(ctx, query, map[string]interface{}{}, results, rcs.defaultTTL)
		if err != nil {
			continue // Skip on error
		}
	}

	return nil
}

// GetCacheStats returns cache statistics
func (rcs *RedisCacheService) GetCacheStats(ctx context.Context) (*RedisCacheStats, error) {
	// Get Redis info
	info, err := rcs.redisClient.Info(ctx, "memory").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to get Redis info: %w", err)
	}

	// Count search cache keys
	keys, err := rcs.redisClient.Keys(ctx, "search:*").Result()
	if err != nil {
		return nil, fmt.Errorf("failed to count cache keys: %w", err)
	}

	// Get database cache stats
	var totalCount int64
	var expiredCount int64

	err = rcs.db.WithContext(ctx).Model(&search.SearchCache{}).Count(&totalCount).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count database cache entries: %w", err)
	}

	err = rcs.db.WithContext(ctx).Model(&search.SearchCache{}).Where("expires_at < ?", time.Now()).Count(&expiredCount).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count expired cache entries: %w", err)
	}

	return &RedisCacheStats{
		RedisKeys:       len(keys),
		TotalEntries:    int(totalCount),
		ExpiredEntries:  int(expiredCount),
		ActiveEntries:   int(totalCount - expiredCount),
		RedisMemoryInfo: info,
	}, nil
}

// CleanupExpiredCache removes expired cache entries
func (rcs *RedisCacheService) CleanupExpiredCache(ctx context.Context) error {
	// Clean up database cache
	err := rcs.db.WithContext(ctx).Where("expires_at < ?", time.Now()).Delete(&search.SearchCache{}).Error
	if err != nil {
		return fmt.Errorf("failed to cleanup database cache: %w", err)
	}

	// Redis TTL is handled automatically, but we can clean up any orphaned keys
	// This is optional as Redis handles TTL automatically
	return nil
}

// createCacheKey creates a consistent cache key
func (rcs *RedisCacheService) createCacheKey(query string, filters map[string]interface{}) string {
	// Create filter hash
	filterHash := "default"
	if len(filters) > 0 {
		filterJSON, _ := json.Marshal(filters)
		filterHash = fmt.Sprintf("%x", filterJSON)
	}

	return fmt.Sprintf("search:%s:%s", query, filterHash)
}

// storeCacheMetadata stores cache metadata in database
func (rcs *RedisCacheService) storeCacheMetadata(ctx context.Context, query string, filters map[string]interface{}, resultCount int, ttl time.Duration) {
	filterJSON, _ := json.Marshal(filters)
	cache := search.SearchCache{
		Query:       query,
		Filters:     datatypes.JSON(filterJSON),
		ResultCount: resultCount,
		ExpiresAt:   time.Now().Add(ttl),
	}

	rcs.db.WithContext(ctx).Create(&cache)
}

// updateCacheHitMetrics updates cache hit metrics
func (rcs *RedisCacheService) updateCacheHitMetrics(ctx context.Context, query string, filters map[string]interface{}) {
	// This would typically update metrics in a monitoring system
	// For now, we'll just log the cache hit
	fmt.Printf("Cache hit for query: %s\n", query)
}

// RedisCacheStats represents Redis-specific cache statistics
type RedisCacheStats struct {
	RedisKeys       int    `json:"redis_keys"`
	TotalEntries    int    `json:"total_entries"`
	ExpiredEntries  int    `json:"expired_entries"`
	ActiveEntries   int    `json:"active_entries"`
	RedisMemoryInfo string `json:"redis_memory_info"`
}
