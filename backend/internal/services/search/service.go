package search

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	searchmodels "chat-ecommerce-backend/internal/models/search"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// Service handles search operations
type Service struct {
	db        *gorm.DB
	processor *SearchProcessor
	ranking   *RankingAlgorithm
	cache     *CacheService
	cacheTTL  time.Duration
}

// NewService creates a new search service
func NewService(db *gorm.DB) *Service {
	return &Service{
		db:        db,
		processor: NewSearchProcessor(db),
		ranking:   NewRankingAlgorithm(),
		cache:     NewCacheService(db),
		cacheTTL:  5 * time.Minute, // Default cache TTL
	}
}

// SearchRequest represents a search request
type SearchRequest struct {
	Query    string                 `json:"query" binding:"required"`
	Filters  map[string]interface{} `json:"filters,omitempty"`
	SortBy   string                 `json:"sort_by,omitempty"`
	Page     int                    `json:"page,omitempty"`
	PageSize int                    `json:"page_size,omitempty"`
}

// SearchResponse represents a search response
type SearchResponse struct {
	Results        []ProductResult `json:"results"`
	Pagination     Pagination      `json:"pagination"`
	Query          string          `json:"query"`
	TotalResults   int             `json:"total_results"`
	ResponseTimeMs int             `json:"response_time_ms"`
}

// Search performs a product search
func (s *Service) Search(ctx context.Context, req *SearchRequest) (*SearchResponse, error) {
	startTime := time.Now()

	// Set defaults
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100 // Limit max page size
	}

	// Check cache first
	if cachedResult, err := s.cache.GetCachedResults(ctx, req.Query, req.Filters); err == nil && cachedResult != nil {
		return &SearchResponse{
			Results:        cachedResult.Results,
			Pagination:     cachedResult.Pagination,
			Query:          req.Query,
			TotalResults:   cachedResult.Pagination.TotalResults,
			ResponseTimeMs: int(time.Since(startTime).Milliseconds()),
		}, nil
	}

	// Process search query
	result, err := s.processor.ProcessQuery(ctx, req.Query, req.Filters, req.Page, req.PageSize)
	if err != nil {
		return nil, fmt.Errorf("failed to process search query: %w", err)
	}

	// Apply ranking
	result.Results = s.ranking.RankResults(req.Query, result.Results)

	// Apply field weights
	for i := range result.Results {
		result.Results[i].RelevanceScore = s.ranking.ApplyFieldBoosts(req.Query, result.Results[i], result.Results[i].RelevanceScore)
	}

	// Cache results
	go func() {
		cacheCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		s.cache.SetCachedResults(cacheCtx, req.Query, req.Filters, result, s.cacheTTL)
	}()

	responseTime := int(time.Since(startTime).Milliseconds())

	return &SearchResponse{
		Results:        result.Results,
		Pagination:     result.Pagination,
		Query:          req.Query,
		TotalResults:   result.Pagination.TotalResults,
		ResponseTimeMs: responseTime,
	}, nil
}

// GetSuggestions returns autocomplete suggestions
func (s *Service) GetSuggestions(ctx context.Context, query string, limit int) ([]string, error) {
	if len(query) < 2 {
		return []string{}, nil
	}

	if limit <= 0 {
		limit = 10
	}
	if limit > 50 {
		limit = 50
	}

	// Simple suggestion based on product names
	var suggestions []string
	err := s.db.WithContext(ctx).
		Table("products").
		Select("DISTINCT name").
		Where("name ILIKE ? AND status = ?", "%"+query+"%", "active").
		Limit(limit).
		Pluck("name", &suggestions).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get suggestions: %w", err)
	}

	return suggestions, nil
}

// GetFilterOptions returns available filter options
func (s *Service) GetFilterOptions(ctx context.Context, categoryID string) (*FilterOptions, error) {
	options := &FilterOptions{
		PriceRanges: []PriceRange{
			{Min: 0, Max: 50, Label: "Under $50"},
			{Min: 50, Max: 100, Label: "$50 - $100"},
			{Min: 100, Max: 200, Label: "$100 - $200"},
			{Min: 200, Max: 500, Label: "$200 - $500"},
			{Min: 500, Max: 0, Label: "Over $500"},
		},
		AvailabilityOptions: []string{"all", "in_stock", "low_stock", "out_of_stock"},
	}

	// Get categories
	var categories []CategoryOption
	err := s.db.WithContext(ctx).
		Table("categories").
		Select("id, name, COUNT(products.id) as product_count").
		Joins("LEFT JOIN products ON categories.id = products.category_id AND products.status = 'active'").
		Group("categories.id, categories.name").
		Order("categories.name").
		Scan(&categories).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get categories: %w", err)
	}

	options.Categories = categories
	return options, nil
}

// LogAnalytics logs search analytics
func (s *Service) LogAnalytics(ctx context.Context, analytics *AnalyticsRequest) error {
	// Convert filters map to JSON
	filtersJSON, err := json.Marshal(analytics.Filters)
	if err != nil {
		return fmt.Errorf("failed to marshal filters: %w", err)
	}

	analyticsRecord := searchmodels.SearchAnalytics{
		Query:            analytics.Query,
		Filters:          datatypes.JSON(filtersJSON),
		ResultCount:      analytics.ResultCount,
		SelectedProducts: analytics.SelectedProducts,
		ResponseTime:     analytics.ResponseTimeMs,
		CacheHit:         analytics.CacheHit,
		SessionID:        analytics.SessionID,
	}

	if analytics.UserID != "" {
		userID, err := uuid.Parse(analytics.UserID)
		if err == nil {
			analyticsRecord.UserID = &userID
		}
	}

	return s.db.WithContext(ctx).Create(&analyticsRecord).Error
}

// AnalyticsRequest represents analytics data
type AnalyticsRequest struct {
	Query            string                 `json:"query"`
	Filters          map[string]interface{} `json:"filters,omitempty"`
	ResultCount      int                    `json:"result_count"`
	SelectedProducts string                 `json:"selected_products,omitempty"`
	ResponseTimeMs   int                    `json:"response_time_ms"`
	CacheHit         bool                   `json:"cache_hit"`
	SessionID        string                 `json:"session_id"`
	UserID           string                 `json:"user_id,omitempty"`
}
