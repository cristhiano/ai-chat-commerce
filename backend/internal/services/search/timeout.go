package search

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// TimeoutService handles query timeout and cancellation
type TimeoutService struct {
	db *gorm.DB
}

// NewTimeoutService creates a new timeout service
func NewTimeoutService(db *gorm.DB) *TimeoutService {
	return &TimeoutService{db: db}
}

// ExecuteWithTimeout executes a database query with timeout
func (ts *TimeoutService) ExecuteWithTimeout(ctx context.Context, timeout time.Duration, query func(ctx context.Context) error) error {
	// Create a context with timeout
	timeoutCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	// Execute the query
	err := query(timeoutCtx)
	if err != nil {
		// Check if it's a timeout error
		if timeoutCtx.Err() == context.DeadlineExceeded {
			return fmt.Errorf("query timeout after %v: %w", timeout, err)
		}
		return err
	}

	return nil
}

// SearchWithTimeout performs a search with timeout
func (ts *TimeoutService) SearchWithTimeout(ctx context.Context, timeout time.Duration, query string, filters map[string]interface{}, page, pageSize int) (*SearchResult, error) {
	var result *SearchResult
	var err error

	err = ts.ExecuteWithTimeout(ctx, timeout, func(timeoutCtx context.Context) error {
		processor := NewSearchProcessor(ts.db)
		result, err = processor.ProcessQuery(timeoutCtx, query, filters, page, pageSize)
		return err
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

// GetSuggestionsWithTimeout gets suggestions with timeout
func (ts *TimeoutService) GetSuggestionsWithTimeout(ctx context.Context, timeout time.Duration, query string, limit int) ([]string, error) {
	var suggestions []string
	var err error

	err = ts.ExecuteWithTimeout(ctx, timeout, func(timeoutCtx context.Context) error {
		autocomplete := NewAutocompleteService(ts.db)
		suggestions, err = autocomplete.GetSuggestions(timeoutCtx, query, limit)
		return err
	})

	if err != nil {
		return nil, err
	}

	return suggestions, nil
}

// GetFilterOptionsWithTimeout gets filter options with timeout
func (ts *TimeoutService) GetFilterOptionsWithTimeout(ctx context.Context, timeout time.Duration) (*FilterOptions, error) {
	var options *FilterOptions
	var err error

	err = ts.ExecuteWithTimeout(ctx, timeout, func(timeoutCtx context.Context) error {
		filterService := NewFilterService(ts.db)
		options, err = filterService.GetFilterOptions(timeoutCtx, "", map[string]interface{}{})
		return err
	})

	if err != nil {
		return nil, err
	}

	return options, nil
}

// GetAnalyticsWithTimeout gets analytics with timeout
func (ts *TimeoutService) GetAnalyticsWithTimeout(ctx context.Context, timeout time.Duration, timeRange time.Duration) (*SearchStats, error) {
	var stats *SearchStats
	var err error

	err = ts.ExecuteWithTimeout(ctx, timeout, func(timeoutCtx context.Context) error {
		analytics := NewAnalyticsService(ts.db)
		stats, err = analytics.GetSearchStats(timeoutCtx, timeRange)
		return err
	})

	if err != nil {
		return nil, err
	}

	return stats, nil
}

// GetDefaultTimeouts returns default timeout configurations
func (ts *TimeoutService) GetDefaultTimeouts() TimeoutConfig {
	return TimeoutConfig{
		SearchQuery:    5 * time.Second,
		Suggestions:    2 * time.Second,
		FilterOptions:  3 * time.Second,
		Analytics:      10 * time.Second,
		CacheOperation: 1 * time.Second,
		DatabaseQuery:  3 * time.Second,
	}
}

// TimeoutConfig represents timeout configuration
type TimeoutConfig struct {
	SearchQuery    time.Duration `json:"search_query"`
	Suggestions    time.Duration `json:"suggestions"`
	FilterOptions  time.Duration `json:"filter_options"`
	Analytics      time.Duration `json:"analytics"`
	CacheOperation time.Duration `json:"cache_operation"`
	DatabaseQuery  time.Duration `json:"database_query"`
}

// GetTimeoutForOperation returns timeout for a specific operation
func (ts *TimeoutService) GetTimeoutForOperation(operation string) time.Duration {
	config := ts.GetDefaultTimeouts()

	switch operation {
	case "search":
		return config.SearchQuery
	case "suggestions":
		return config.Suggestions
	case "filters":
		return config.FilterOptions
	case "analytics":
		return config.Analytics
	case "cache":
		return config.CacheOperation
	case "database":
		return config.DatabaseQuery
	default:
		return config.DatabaseQuery
	}
}

// MonitorQueryPerformance monitors query performance
func (ts *TimeoutService) MonitorQueryPerformance(ctx context.Context, operation string, query func(ctx context.Context) error) error {
	start := time.Now()
	timeout := ts.GetTimeoutForOperation(operation)

	err := ts.ExecuteWithTimeout(ctx, timeout, query)
	duration := time.Since(start)

	// Log performance metrics
	fmt.Printf("Operation %s took %v (timeout: %v)\n", operation, duration, timeout)

	if err != nil {
		fmt.Printf("Operation %s failed: %v\n", operation, err)
		return err
	}

	// Check if query was close to timeout
	if duration > timeout*0.8 {
		fmt.Printf("Warning: Operation %s took %v (80%% of timeout %v)\n", operation, duration, timeout)
	}

	return nil
}

// GetPerformanceMetrics returns performance metrics
func (ts *TimeoutService) GetPerformanceMetrics(ctx context.Context) (*PerformanceMetrics, error) {
	// Get database connection stats
	sqlDB, err := ts.db.DB()
	if err != nil {
		return nil, fmt.Errorf("failed to get database connection: %w", err)
	}

	stats := sqlDB.Stats()

	metrics := &PerformanceMetrics{
		OpenConnections:   stats.OpenConnections,
		InUse:             stats.InUse,
		Idle:              stats.Idle,
		WaitCount:         stats.WaitCount,
		WaitDuration:      stats.WaitDuration,
		MaxIdleClosed:     stats.MaxIdleClosed,
		MaxIdleTimeClosed: stats.MaxIdleTimeClosed,
		MaxLifetimeClosed: stats.MaxLifetimeClosed,
		GeneratedAt:       time.Now(),
	}

	return metrics, nil
}

// PerformanceMetrics represents database performance metrics
type PerformanceMetrics struct {
	OpenConnections   int           `json:"open_connections"`
	InUse             int           `json:"in_use"`
	Idle              int           `json:"idle"`
	WaitCount         int64         `json:"wait_count"`
	WaitDuration      time.Duration `json:"wait_duration"`
	MaxIdleClosed     int64         `json:"max_idle_closed"`
	MaxIdleTimeClosed int64         `json:"max_idle_time_closed"`
	MaxLifetimeClosed int64         `json:"max_lifetime_closed"`
	GeneratedAt       time.Time     `json:"generated_at"`
}
