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

// AnalyticsService handles search analytics and logging
type AnalyticsService struct {
	db *gorm.DB
}

// NewAnalyticsService creates a new analytics service
func NewAnalyticsService(db *gorm.DB) *AnalyticsService {
	return &AnalyticsService{db: db}
}

// LogSearchQuery logs a search query for analytics
func (as *AnalyticsService) LogSearchQuery(ctx context.Context, query string, filters map[string]interface{}, resultCount int, responseTime int, sessionID string, userID *uuid.UUID) error {
	analytics := SearchAnalytics{
		Query:        query,
		ResultCount:  resultCount,
		ResponseTime: responseTime,
		SessionID:    sessionID,
		UserID:       userID,
		CreatedAt:    time.Now(),
	}

	// Add filters if provided
	if len(filters) > 0 {
		filterJSON, err := json.Marshal(filters)
		if err == nil {
			analytics.Filters = datatypes.JSON(filterJSON)
		}
	}

	return as.db.WithContext(ctx).Create(&analytics).Error
}

// LogSearchClick logs when a user clicks on a search result
func (as *AnalyticsService) LogSearchClick(ctx context.Context, query string, productID string, position int, sessionID string, userID *uuid.UUID) error {
	click := SearchClick{
		Query:     query,
		ProductID: productID,
		Position:  position,
		SessionID: sessionID,
		UserID:    userID,
		CreatedAt: time.Now(),
	}

	return as.db.WithContext(ctx).Create(&click).Error
}

// LogNoResultsQuery logs queries that return no results
func (as *AnalyticsService) LogNoResultsQuery(ctx context.Context, query string, sessionID string, userID *uuid.UUID) error {
	analytics := SearchAnalytics{
		Query:        query,
		ResultCount:  0,
		ResponseTime: 0,
		SessionID:    sessionID,
		UserID:       userID,
		CreatedAt:    time.Now(),
	}

	return as.db.WithContext(ctx).Create(&analytics).Error
}

// GetSearchStats returns search statistics
func (as *AnalyticsService) GetSearchStats(ctx context.Context, timeRange time.Duration) (*SearchStats, error) {
	since := time.Now().Add(-timeRange)

	var totalQueries int64
	var totalResults int64
	var noResultsQueries int64
	var avgResponseTime float64

	// Count total queries
	err := as.db.WithContext(ctx).
		Model(&SearchAnalytics{}).
		Where("created_at > ?", since).
		Count(&totalQueries).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count total queries: %w", err)
	}

	// Sum total results
	err = as.db.WithContext(ctx).
		Model(&SearchAnalytics{}).
		Where("created_at > ?", since).
		Select("SUM(result_count)").
		Scan(&totalResults).Error
	if err != nil {
		return nil, fmt.Errorf("failed to sum total results: %w", err)
	}

	// Count no results queries
	err = as.db.WithContext(ctx).
		Model(&SearchAnalytics{}).
		Where("created_at > ? AND result_count = 0", since).
		Count(&noResultsQueries).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count no results queries: %w", err)
	}

	// Calculate average response time
	err = as.db.WithContext(ctx).
		Model(&SearchAnalytics{}).
		Where("created_at > ? AND response_time > 0", since).
		Select("AVG(response_time)").
		Scan(&avgResponseTime).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate average response time: %w", err)
	}

	return &SearchStats{
		TotalQueries:     int(totalQueries),
		TotalResults:     int(totalResults),
		NoResultsQueries: int(noResultsQueries),
		AvgResponseTime:  avgResponseTime,
		TimeRange:        timeRange,
		GeneratedAt:      time.Now(),
	}, nil
}

// GetPopularQueries returns popular search queries
func (as *AnalyticsService) GetPopularQueries(ctx context.Context, limit int, timeRange time.Duration) ([]PopularQuery, error) {
	since := time.Now().Add(-timeRange)

	var queries []PopularQuery

	err := as.db.WithContext(ctx).
		Table("search_analytics").
		Select("query, COUNT(*) as frequency, AVG(result_count) as avg_results").
		Where("created_at > ?", since).
		Group("query").
		Order("frequency DESC").
		Limit(limit).
		Scan(&queries).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get popular queries: %w", err)
	}

	return queries, nil
}

// GetTrendingQueries returns trending search queries
func (as *AnalyticsService) GetTrendingQueries(ctx context.Context, limit int) ([]TrendingQuery, error) {
	// Get queries from last 7 days
	since := time.Now().AddDate(0, 0, -7)

	var queries []TrendingQuery

	err := as.db.WithContext(ctx).
		Table("search_analytics").
		Select("query, COUNT(*) as frequency, MAX(created_at) as last_searched").
		Where("created_at > ?", since).
		Group("query").
		Order("frequency DESC").
		Limit(limit).
		Scan(&queries).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get trending queries: %w", err)
	}

	return queries, nil
}

// GetSearchPerformance returns search performance metrics
func (as *AnalyticsService) GetSearchPerformance(ctx context.Context, timeRange time.Duration) (*SearchPerformance, error) {
	since := time.Now().Add(-timeRange)

	var performance SearchPerformance

	// Get response time percentiles
	var p50, p90, p95, p99 float64

	err := as.db.WithContext(ctx).
		Model(&SearchAnalytics{}).
		Where("created_at > ? AND response_time > 0", since).
		Select("PERCENTILE_CONT(0.5) WITHIN GROUP (ORDER BY response_time)").
		Scan(&p50).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get p50 response time: %w", err)
	}

	err = as.db.WithContext(ctx).
		Model(&SearchAnalytics{}).
		Where("created_at > ? AND response_time > 0", since).
		Select("PERCENTILE_CONT(0.9) WITHIN GROUP (ORDER BY response_time)").
		Scan(&p90).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get p90 response time: %w", err)
	}

	err = as.db.WithContext(ctx).
		Model(&SearchAnalytics{}).
		Where("created_at > ? AND response_time > 0", since).
		Select("PERCENTILE_CONT(0.95) WITHIN GROUP (ORDER BY response_time)").
		Scan(&p95).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get p95 response time: %w", err)
	}

	err = as.db.WithContext(ctx).
		Model(&SearchAnalytics{}).
		Where("created_at > ? AND response_time > 0", since).
		Select("PERCENTILE_CONT(0.99) WITHIN GROUP (ORDER BY response_time)").
		Scan(&p99).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get p99 response time: %w", err)
	}

	performance.ResponseTimePercentiles = ResponseTimePercentiles{
		P50: p50,
		P90: p90,
		P95: p95,
		P99: p99,
	}

	// Get cache hit rate
	var cacheHits int64
	var totalQueries int64

	err = as.db.WithContext(ctx).
		Model(&SearchAnalytics{}).
		Where("created_at > ?", since).
		Count(&totalQueries).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count total queries: %w", err)
	}

	err = as.db.WithContext(ctx).
		Model(&SearchAnalytics{}).
		Where("created_at > ? AND cache_hit = true", since).
		Count(&cacheHits).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count cache hits: %w", err)
	}

	if totalQueries > 0 {
		performance.CacheHitRate = float64(cacheHits) / float64(totalQueries) * 100
	}

	performance.TimeRange = timeRange
	performance.GeneratedAt = time.Now()

	return &performance, nil
}

// SearchStats represents search statistics
type SearchStats struct {
	TotalQueries     int           `json:"total_queries"`
	TotalResults     int           `json:"total_results"`
	NoResultsQueries int           `json:"no_results_queries"`
	AvgResponseTime  float64       `json:"avg_response_time"`
	TimeRange        time.Duration `json:"time_range"`
	GeneratedAt      time.Time     `json:"generated_at"`
}

// TrendingQuery represents a trending search query
type TrendingQuery struct {
	Query        string    `json:"query"`
	Frequency    int       `json:"frequency"`
	LastSearched time.Time `json:"last_searched"`
}

// SearchPerformance represents search performance metrics
type SearchPerformance struct {
	ResponseTimePercentiles ResponseTimePercentiles `json:"response_time_percentiles"`
	CacheHitRate            float64                 `json:"cache_hit_rate"`
	TimeRange               time.Duration           `json:"time_range"`
	GeneratedAt             time.Time               `json:"generated_at"`
}

// ResponseTimePercentiles represents response time percentiles
type ResponseTimePercentiles struct {
	P50 float64 `json:"p50"`
	P90 float64 `json:"p90"`
	P95 float64 `json:"p95"`
	P99 float64 `json:"p99"`
}

// SearchClick represents a search result click
type SearchClick struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Query     string     `gorm:"size:500;not null;index"`
	ProductID string     `gorm:"size:255;not null;index"`
	Position  int        `gorm:"not null"`
	SessionID string     `gorm:"size:100;index"`
	UserID    *uuid.UUID `gorm:"type:uuid;index"`
	CreatedAt time.Time
}

// TableName returns the table name for SearchClick
func (SearchClick) TableName() string {
	return "search_clicks"
}
