package search

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// TrackingService handles search query tracking and user behavior analysis
type TrackingService struct {
	db *gorm.DB
}

// NewTrackingService creates a new tracking service
func NewTrackingService(db *gorm.DB) *TrackingService {
	return &TrackingService{db: db}
}

// TrackSearchSession tracks a complete search session
func (ts *TrackingService) TrackSearchSession(ctx context.Context, sessionID string, userID *uuid.UUID, queries []string) error {
	session := SearchSession{
		SessionID: sessionID,
		UserID:    userID,
		Queries:   queries,
		StartedAt: time.Now(),
		EndedAt:   time.Now(),
	}

	return ts.db.WithContext(ctx).Create(&session).Error
}

// TrackQueryRefinement tracks when a user refines their search
func (ts *TrackingService) TrackQueryRefinement(ctx context.Context, sessionID string, originalQuery string, refinedQuery string, userID *uuid.UUID) error {
	refinement := QueryRefinement{
		SessionID:     sessionID,
		OriginalQuery: originalQuery,
		RefinedQuery:  refinedQuery,
		UserID:        userID,
		RefinedAt:     time.Now(),
	}

	return ts.db.WithContext(ctx).Create(&refinement).Error
}

// TrackFilterUsage tracks when users apply filters
func (ts *TrackingService) TrackFilterUsage(ctx context.Context, sessionID string, query string, filters map[string]interface{}, userID *uuid.UUID) error {
	filterUsage := FilterUsage{
		SessionID: sessionID,
		Query:     query,
		Filters:   filters,
		UserID:    userID,
		UsedAt:    time.Now(),
	}

	return ts.db.WithContext(ctx).Create(&filterUsage).Error
}

// TrackSearchAbandonment tracks when users abandon search
func (ts *TrackingService) TrackSearchAbandonment(ctx context.Context, sessionID string, query string, position int, userID *uuid.UUID) error {
	abandonment := SearchAbandonment{
		SessionID:   sessionID,
		Query:       query,
		Position:    position,
		UserID:      userID,
		AbandonedAt: time.Now(),
	}

	return ts.db.WithContext(ctx).Create(&abandonment).Error
}

// GetUserSearchBehavior returns user search behavior analysis
func (ts *TrackingService) GetUserSearchBehavior(ctx context.Context, userID uuid.UUID, timeRange time.Duration) (*UserSearchBehavior, error) {
	since := time.Now().Add(-timeRange)

	var behavior UserSearchBehavior

	// Get total searches
	var totalSearches int64
	err := ts.db.WithContext(ctx).
		Model(&SearchAnalytics{}).
		Where("user_id = ? AND created_at > ?", userID, since).
		Count(&totalSearches).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count total searches: %w", err)
	}

	// Get average queries per session
	var avgQueriesPerSession float64
	err = ts.db.WithContext(ctx).
		Table("search_sessions").
		Where("user_id = ? AND started_at > ?", userID, since).
		Select("AVG(array_length(queries, 1))").
		Scan(&avgQueriesPerSession).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate avg queries per session: %w", err)
	}

	// Get most used filters
	var filterUsage []FilterUsageStats
	err = ts.db.WithContext(ctx).
		Table("filter_usage").
		Where("user_id = ? AND used_at > ?", userID, since).
		Select("filters, COUNT(*) as usage_count").
		Group("filters").
		Order("usage_count DESC").
		Limit(10).
		Scan(&filterUsage).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get filter usage: %w", err)
	}

	// Get search patterns
	var patterns []SearchPattern
	err = ts.db.WithContext(ctx).
		Table("query_refinements").
		Where("user_id = ? AND refined_at > ?", userID, since).
		Select("original_query, refined_query, COUNT(*) as frequency").
		Group("original_query, refined_query").
		Order("frequency DESC").
		Limit(10).
		Scan(&patterns).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get search patterns: %w", err)
	}

	behavior.UserID = userID
	behavior.TotalSearches = int(totalSearches)
	behavior.AvgQueriesPerSession = avgQueriesPerSession
	behavior.FilterUsage = filterUsage
	behavior.SearchPatterns = patterns
	behavior.TimeRange = timeRange
	behavior.GeneratedAt = time.Now()

	return &behavior, nil
}

// GetSearchInsights returns insights about search behavior
func (ts *TrackingService) GetSearchInsights(ctx context.Context, timeRange time.Duration) (*SearchInsights, error) {
	since := time.Now().Add(-timeRange)

	var insights SearchInsights

	// Get conversion rate (searches that led to clicks)
	var totalSearches int64
	var searchesWithClicks int64

	err := ts.db.WithContext(ctx).
		Model(&SearchAnalytics{}).
		Where("created_at > ?", since).
		Count(&totalSearches).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count total searches: %w", err)
	}

	err = ts.db.WithContext(ctx).
		Table("search_analytics sa").
		Joins("JOIN search_clicks sc ON sa.query = sc.query AND sa.session_id = sc.session_id").
		Where("sa.created_at > ?", since).
		Select("COUNT(DISTINCT sa.id)").
		Scan(&searchesWithClicks).Error
	if err != nil {
		return nil, fmt.Errorf("failed to count searches with clicks: %w", err)
	}

	if totalSearches > 0 {
		insights.ConversionRate = float64(searchesWithClicks) / float64(totalSearches) * 100
	}

	// Get average session length
	var avgSessionLength float64
	err = ts.db.WithContext(ctx).
		Table("search_sessions").
		Where("started_at > ?", since).
		Select("AVG(EXTRACT(EPOCH FROM (ended_at - started_at)))").
		Scan(&avgSessionLength).Error
	if err != nil {
		return nil, fmt.Errorf("failed to calculate avg session length: %w", err)
	}

	insights.AvgSessionLength = avgSessionLength

	// Get top performing queries
	var topQueries []QueryPerformance
	err = ts.db.WithContext(ctx).
		Table("search_analytics sa").
		Joins("LEFT JOIN search_clicks sc ON sa.query = sc.query AND sa.session_id = sc.session_id").
		Where("sa.created_at > ?", since).
		Select("sa.query, COUNT(DISTINCT sa.id) as search_count, COUNT(DISTINCT sc.id) as click_count").
		Group("sa.query").
		Order("click_count DESC").
		Limit(10).
		Scan(&topQueries).Error
	if err != nil {
		return nil, fmt.Errorf("failed to get top performing queries: %w", err)
	}

	insights.TopPerformingQueries = topQueries
	insights.TimeRange = timeRange
	insights.GeneratedAt = time.Now()

	return &insights, nil
}

// SearchSession represents a search session
type SearchSession struct {
	ID        uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SessionID string     `gorm:"size:100;not null;index"`
	UserID    *uuid.UUID `gorm:"type:uuid;index"`
	Queries   []string   `gorm:"type:text[]"`
	StartedAt time.Time
	EndedAt   time.Time
}

// TableName returns the table name for SearchSession
func (SearchSession) TableName() string {
	return "search_sessions"
}

// QueryRefinement represents a query refinement
type QueryRefinement struct {
	ID            uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SessionID     string     `gorm:"size:100;not null;index"`
	OriginalQuery string     `gorm:"size:500;not null"`
	RefinedQuery  string     `gorm:"size:500;not null"`
	UserID        *uuid.UUID `gorm:"type:uuid;index"`
	RefinedAt     time.Time
}

// TableName returns the table name for QueryRefinement
func (QueryRefinement) TableName() string {
	return "query_refinements"
}

// FilterUsage represents filter usage tracking
type FilterUsage struct {
	ID        uuid.UUID              `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SessionID string                 `gorm:"size:100;not null;index"`
	Query     string                 `gorm:"size:500;not null"`
	Filters   map[string]interface{} `gorm:"type:jsonb"`
	UserID    *uuid.UUID             `gorm:"type:uuid;index"`
	UsedAt    time.Time
}

// TableName returns the table name for FilterUsage
func (FilterUsage) TableName() string {
	return "filter_usage"
}

// SearchAbandonment represents search abandonment tracking
type SearchAbandonment struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	SessionID   string     `gorm:"size:100;not null;index"`
	Query       string     `gorm:"size:500;not null"`
	Position    int        `gorm:"not null"`
	UserID      *uuid.UUID `gorm:"type:uuid;index"`
	AbandonedAt time.Time
}

// TableName returns the table name for SearchAbandonment
func (SearchAbandonment) TableName() string {
	return "search_abandonment"
}

// UserSearchBehavior represents user search behavior analysis
type UserSearchBehavior struct {
	UserID               uuid.UUID          `json:"user_id"`
	TotalSearches        int                `json:"total_searches"`
	AvgQueriesPerSession float64            `json:"avg_queries_per_session"`
	FilterUsage          []FilterUsageStats `json:"filter_usage"`
	SearchPatterns       []SearchPattern    `json:"search_patterns"`
	TimeRange            time.Duration      `json:"time_range"`
	GeneratedAt          time.Time          `json:"generated_at"`
}

// FilterUsageStats represents filter usage statistics
type FilterUsageStats struct {
	Filters    map[string]interface{} `json:"filters"`
	UsageCount int                    `json:"usage_count"`
}

// SearchPattern represents a search pattern
type SearchPattern struct {
	OriginalQuery string `json:"original_query"`
	RefinedQuery  string `json:"refined_query"`
	Frequency     int    `json:"frequency"`
}

// SearchInsights represents search insights
type SearchInsights struct {
	ConversionRate       float64            `json:"conversion_rate"`
	AvgSessionLength     float64            `json:"avg_session_length"`
	TopPerformingQueries []QueryPerformance `json:"top_performing_queries"`
	TimeRange            time.Duration      `json:"time_range"`
	GeneratedAt          time.Time          `json:"generated_at"`
}

// QueryPerformance represents query performance metrics
type QueryPerformance struct {
	Query       string `json:"query"`
	SearchCount int    `json:"search_count"`
	ClickCount  int    `json:"click_count"`
}
