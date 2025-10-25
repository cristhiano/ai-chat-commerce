package search

import (
	"context"
	"fmt"
	"strings"
	"time"

	"chat-ecommerce-backend/internal/models/search"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// NoResultsService handles no-results scenarios
type NoResultsService struct {
	db *gorm.DB
}

// NewNoResultsService creates a new no-results service
func NewNoResultsService(db *gorm.DB) *NoResultsService {
	return &NoResultsService{db: db}
}

// HandleNoResults handles cases where search returns no results
func (nrs *NoResultsService) HandleNoResults(ctx context.Context, query string) (*NoResultsResponse, error) {
	// Get suggestions for similar queries
	suggestions, err := nrs.getSimilarQueries(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("failed to get similar queries: %w", err)
	}

	// Get popular products as fallback
	popularProducts, err := nrs.getPopularProducts(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get popular products: %w", err)
	}

	// Get trending searches
	trendingSearches, err := nrs.getTrendingSearches(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get trending searches: %w", err)
	}

	return &NoResultsResponse{
		Message:          nrs.generateNoResultsMessage(query),
		Suggestions:      suggestions,
		PopularProducts:  popularProducts,
		TrendingSearches: trendingSearches,
		HelpText:         nrs.generateHelpText(),
	}, nil
}

// getSimilarQueries finds similar queries that returned results
func (nrs *NoResultsService) getSimilarQueries(ctx context.Context, query string) ([]string, error) {
	var suggestions []string

	// Split query into words
	words := strings.Fields(strings.ToLower(query))
	if len(words) == 0 {
		return suggestions, nil
	}

	// Find queries that contain some of the same words
	var similarQueries []string
	err := nrs.db.WithContext(ctx).
		Table("search_analytics").
		Select("DISTINCT query").
		Where("result_count > 0 AND created_at > ?", time.Now().AddDate(0, 0, -30)).
		Where("query ILIKE ?", "%"+words[0]+"%").
		Limit(5).
		Pluck("query", &similarQueries).Error

	if err != nil {
		return suggestions, err
	}

	return similarQueries, nil
}

// getPopularProducts gets popular products as fallback
func (nrs *NoResultsService) getPopularProducts(ctx context.Context) ([]PopularProduct, error) {
	var products []PopularProduct

	err := nrs.db.WithContext(ctx).
		Table("product_search_view").
		Select("id, name, description, price, category_name, popularity").
		Where("status = ?", "active").
		Order("popularity DESC").
		Limit(6).
		Scan(&products).Error

	if err != nil {
		return nil, err
	}

	return products, nil
}

// getTrendingSearches gets trending search terms
func (nrs *NoResultsService) getTrendingSearches(ctx context.Context) ([]string, error) {
	var trending []string

	err := nrs.db.WithContext(ctx).
		Table("search_analytics").
		Select("query").
		Where("created_at > ?", time.Now().AddDate(0, 0, -7)).
		Where("result_count > 0").
		Group("query").
		Order("COUNT(*) DESC").
		Limit(5).
		Pluck("query", &trending).Error

	if err != nil {
		return nil, err
	}

	return trending, nil
}

// generateNoResultsMessage generates a helpful no-results message
func (nrs *NoResultsService) generateNoResultsMessage(query string) string {
	messages := []string{
		fmt.Sprintf("No products found for \"%s\"", query),
		fmt.Sprintf("We couldn't find any products matching \"%s\"", query),
		fmt.Sprintf("No results for \"%s\" - try a different search term", query),
	}

	// Return a simple message for now
	return messages[0]
}

// generateHelpText generates help text for users
func (nrs *NoResultsService) generateHelpText() string {
	return "Try searching with different keywords, check your spelling, or browse our popular products below."
}

// NoResultsResponse represents the response for no results
type NoResultsResponse struct {
	Message          string           `json:"message"`
	Suggestions      []string         `json:"suggestions"`
	PopularProducts  []PopularProduct `json:"popular_products"`
	TrendingSearches []string         `json:"trending_searches"`
	HelpText         string           `json:"help_text"`
}

// PopularProduct represents a popular product
type PopularProduct struct {
	ID          string  `json:"id"`
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Price       float64 `json:"price"`
	Category    string  `json:"category"`
	Popularity  int     `json:"popularity"`
}

// LogNoResultsQuery logs queries that return no results for analytics
func (nrs *NoResultsService) LogNoResultsQuery(ctx context.Context, query string, sessionID string, userID *uuid.UUID) error {
	analytics := search.SearchAnalytics{
		Query:       query,
		ResultCount: 0,
		SessionID:   sessionID,
		CreatedAt:   time.Now(),
	}

	if userID != nil {
		analytics.UserID = userID
	}

	return nrs.db.WithContext(ctx).Create(&analytics).Error
}
