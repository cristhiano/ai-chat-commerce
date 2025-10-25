package search

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// CacheWarmingService handles cache warming for popular queries
type CacheWarmingService struct {
	db           *gorm.DB
	cacheService *RedisCacheService
}

// NewCacheWarmingService creates a new cache warming service
func NewCacheWarmingService(db *gorm.DB, cacheService *RedisCacheService) *CacheWarmingService {
	return &CacheWarmingService{
		db:           db,
		cacheService: cacheService,
	}
}

// WarmPopularQueries warms cache with popular search queries
func (cws *CacheWarmingService) WarmPopularQueries(ctx context.Context, limit int) error {
	if limit <= 0 {
		limit = 50
	}

	// Get popular queries from analytics
	popularQueries, err := cws.getPopularQueries(ctx, limit)
	if err != nil {
		return fmt.Errorf("failed to get popular queries: %w", err)
	}

	// Warm cache for each query
	for _, query := range popularQueries {
		err = cws.warmQueryCache(ctx, query)
		if err != nil {
			// Log error but continue with other queries
			fmt.Printf("Failed to warm cache for query '%s': %v\n", query.Query, err)
		}
	}

	return nil
}

// WarmTrendingQueries warms cache with trending queries
func (cws *CacheWarmingService) WarmTrendingQueries(ctx context.Context, limit int) error {
	if limit <= 0 {
		limit = 20
	}

	// Get trending queries from recent analytics
	trendingQueries, err := cws.getTrendingQueries(ctx, limit)
	if err != nil {
		return fmt.Errorf("failed to get trending queries: %w", err)
	}

	// Warm cache for each query
	for _, query := range trendingQueries {
		err = cws.warmQueryCache(ctx, query)
		if err != nil {
			fmt.Printf("Failed to warm cache for trending query '%s': %v\n", query.Query, err)
		}
	}

	return nil
}

// WarmCategoryQueries warms cache with category-based queries
func (cws *CacheWarmingService) WarmCategoryQueries(ctx context.Context) error {
	// Get popular categories
	var categories []struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	err := cws.db.WithContext(ctx).
		Table("categories").
		Select("id, name").
		Order("name").
		Scan(&categories).Error

	if err != nil {
		return fmt.Errorf("failed to get categories: %w", err)
	}

	// Warm cache for each category
	for _, category := range categories {
		query := PopularQuery{
			Query: category.Name,
			Filters: map[string]interface{}{
				"category_id": category.ID,
			},
		}

		err = cws.warmQueryCache(ctx, query)
		if err != nil {
			fmt.Printf("Failed to warm cache for category '%s': %v\n", category.Name, err)
		}
	}

	return nil
}

// WarmBrandQueries warms cache with brand-based queries
func (cws *CacheWarmingService) WarmBrandQueries(ctx context.Context) error {
	// Get popular brands
	var brands []struct {
		Brand string `json:"brand"`
	}

	err := cws.db.WithContext(ctx).
		Table("products").
		Select("DISTINCT metadata->>'brand' as brand").
		Where("metadata->>'brand' IS NOT NULL AND status = ?", "active").
		Order("brand").
		Limit(20).
		Scan(&brands).Error

	if err != nil {
		return fmt.Errorf("failed to get brands: %w", err)
	}

	// Warm cache for each brand
	for _, brand := range brands {
		query := PopularQuery{
			Query: brand.Brand,
			Filters: map[string]interface{}{
				"brand": brand.Brand,
			},
		}

		err = cws.warmQueryCache(ctx, query)
		if err != nil {
			fmt.Printf("Failed to warm cache for brand '%s': %v\n", brand.Brand, err)
		}
	}

	return nil
}

// ScheduleCacheWarming schedules periodic cache warming
func (cws *CacheWarmingService) ScheduleCacheWarming(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				// Warm popular queries
				err := cws.WarmPopularQueries(ctx, 30)
				if err != nil {
					fmt.Printf("Failed to warm popular queries: %v\n", err)
				}

				// Warm trending queries
				err = cws.WarmTrendingQueries(ctx, 10)
				if err != nil {
					fmt.Printf("Failed to warm trending queries: %v\n", err)
				}
			}
		}
	}()
}

// getPopularQueries gets popular queries from analytics
func (cws *CacheWarmingService) getPopularQueries(ctx context.Context, limit int) ([]PopularQuery, error) {
	var queries []PopularQuery

	err := cws.db.WithContext(ctx).
		Table("search_analytics").
		Select("query, COUNT(*) as frequency").
		Where("result_count > 0 AND created_at > ?", time.Now().AddDate(0, 0, -30)).
		Group("query").
		Order("frequency DESC").
		Limit(limit).
		Scan(&queries).Error

	if err != nil {
		return nil, err
	}

	return queries, nil
}

// getTrendingQueries gets trending queries from recent analytics
func (cws *CacheWarmingService) getTrendingQueries(ctx context.Context, limit int) ([]PopularQuery, error) {
	var queries []PopularQuery

	err := cws.db.WithContext(ctx).
		Table("search_analytics").
		Select("query, COUNT(*) as frequency").
		Where("result_count > 0 AND created_at > ?", time.Now().AddDate(0, 0, -7)).
		Group("query").
		Order("frequency DESC").
		Limit(limit).
		Scan(&queries).Error

	if err != nil {
		return nil, err
	}

	return queries, nil
}

// warmQueryCache warms cache for a specific query
func (cws *CacheWarmingService) warmQueryCache(ctx context.Context, query PopularQuery) error {
	// Check if already cached
	cacheKey := cws.cacheService.createCacheKey(query.Query, query.Filters)
	exists, err := cws.cacheService.redisClient.Exists(ctx, cacheKey).Result()
	if err != nil {
		return err
	}

	if exists > 0 {
		return nil // Already cached
	}

	// Perform search
	processor := NewSearchProcessor(cws.db)
	results, err := processor.ProcessQuery(ctx, query.Query, query.Filters, 1, 20)
	if err != nil {
		return err
	}

	// Cache the results
	err = cws.cacheService.CacheSearchResults(ctx, query.Query, query.Filters, results, 10*time.Minute)
	if err != nil {
		return err
	}

	return nil
}

// GetWarmingStats returns cache warming statistics
func (cws *CacheWarmingService) GetWarmingStats(ctx context.Context) (*WarmingStats, error) {
	var totalQueries int64
	var warmedQueries int64
	var recentWarming int64

	// Count total popular queries
	err := cws.db.WithContext(ctx).
		Table("search_analytics").
		Where("created_at > ?", time.Now().AddDate(0, 0, -30)).
		Count(&totalQueries).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count total queries: %w", err)
	}

	// Count warmed queries (queries with cache entries)
	err = cws.db.WithContext(ctx).
		Table("search_cache").
		Where("created_at > ?", time.Now().AddDate(0, 0, -1)).
		Count(&warmedQueries).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count warmed queries: %w", err)
	}

	// Count recent warming activity
	err = cws.db.WithContext(ctx).
		Table("search_cache").
		Where("created_at > ?", time.Now().Add(-time.Hour)).
		Count(&recentWarming).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count recent warming: %w", err)
	}

	return &WarmingStats{
		TotalQueries:    int(totalQueries),
		WarmedQueries:   int(warmedQueries),
		RecentWarming:   int(recentWarming),
		WarmingRate:     float64(warmedQueries) / float64(totalQueries) * 100,
		LastWarmingTime: time.Now(),
	}, nil
}

// PopularQuery represents a popular search query
type PopularQuery struct {
	Query   string                 `json:"query"`
	Filters map[string]interface{} `json:"filters"`
}

// WarmingStats represents cache warming statistics
type WarmingStats struct {
	TotalQueries    int       `json:"total_queries"`
	WarmedQueries   int       `json:"warmed_queries"`
	RecentWarming   int       `json:"recent_warming"`
	WarmingRate     float64   `json:"warming_rate"`
	LastWarmingTime time.Time `json:"last_warming_time"`
}
