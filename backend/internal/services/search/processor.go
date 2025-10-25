package search

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// SearchProcessor handles search query processing
type SearchProcessor struct {
	db *gorm.DB
}

// NewSearchProcessor creates a new search processor
func NewSearchProcessor(db *gorm.DB) *SearchProcessor {
	return &SearchProcessor{db: db}
}

// ProcessQuery processes a search query and returns results
func (sp *SearchProcessor) ProcessQuery(ctx context.Context, query string, filters map[string]interface{}, page, pageSize int) (*SearchResult, error) {
	// Normalize query
	normalizedQuery := strings.TrimSpace(strings.ToLower(query))
	if normalizedQuery == "" {
		return &SearchResult{
			Results: []ProductResult{},
			Pagination: Pagination{
				CurrentPage:  page,
				PageSize:     pageSize,
				TotalPages:   0,
				TotalResults: 0,
				HasNext:      false,
				HasPrevious:  false,
			},
			Query: normalizedQuery,
		}, nil
	}

	// Build search query
	var products []ProductSearchResult
	queryBuilder := sp.db.WithContext(ctx).Table("product_search_view").
		Select("id, name, description, price, category_name, sku, status, popularity, total_available").
		Where("status = ?", "active")

	// Add full-text search
	if normalizedQuery != "" {
		queryBuilder = queryBuilder.Where("search_vector @@ plainto_tsquery('english', ?)", normalizedQuery)
	}

	// Apply filters
	if categoryID, ok := filters["category_id"].(string); ok && categoryID != "" {
		queryBuilder = queryBuilder.Where("category_id = ?", categoryID)
	}

	if priceMin, ok := filters["price_min"].(float64); ok {
		queryBuilder = queryBuilder.Where("price >= ?", priceMin)
	}

	if priceMax, ok := filters["price_max"].(float64); ok {
		queryBuilder = queryBuilder.Where("price <= ?", priceMax)
	}

	if availability, ok := filters["availability"].(string); ok && availability != "all" {
		switch availability {
		case "in_stock":
			queryBuilder = queryBuilder.Where("total_available > 0")
		case "low_stock":
			queryBuilder = queryBuilder.Where("total_available > 0 AND total_available <= 10")
		case "out_of_stock":
			queryBuilder = queryBuilder.Where("total_available = 0")
		}
	}

	// Get total count
	var totalCount int64
	countQuery := queryBuilder
	if err := countQuery.Count(&totalCount).Error; err != nil {
		return nil, fmt.Errorf("failed to count search results: %w", err)
	}

	// Apply pagination and ordering
	offset := (page - 1) * pageSize
	queryBuilder = queryBuilder.Order(fmt.Sprintf("ts_rank(search_vector, plainto_tsquery('english', '%s')) DESC, popularity DESC", normalizedQuery)).
		Offset(offset).Limit(pageSize)

	// Execute query
	if err := queryBuilder.Find(&products).Error; err != nil {
		return nil, fmt.Errorf("failed to execute search query: %w", err)
	}

	// Convert to result format
	results := make([]ProductResult, len(products))
	for i, product := range products {
		results[i] = ProductResult{
			ID:             product.ID,
			Name:           product.Name,
			Description:    product.Description,
			Price:          product.Price,
			Category:       product.CategoryName,
			SKU:            product.SKU,
			Availability:   getAvailabilityStatus(product.TotalAvailable),
			RelevanceScore: calculateRelevanceScore(normalizedQuery, product.Name, product.Description),
		}
	}

	// Calculate pagination
	totalPages := int((totalCount + int64(pageSize) - 1) / int64(pageSize))

	return &SearchResult{
		Results: results,
		Pagination: Pagination{
			CurrentPage:  page,
			PageSize:     pageSize,
			TotalPages:   totalPages,
			TotalResults: int(totalCount),
			HasNext:      page < totalPages,
			HasPrevious:  page > 1,
		},
		Query: normalizedQuery,
	}, nil
}

// ProductSearchResult represents a product from search view
type ProductSearchResult struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Price          float64   `json:"price"`
	CategoryName   string    `json:"category_name"`
	SKU            string    `json:"sku"`
	Status         string    `json:"status"`
	Popularity     int       `json:"popularity"`
	TotalAvailable int       `json:"total_available"`
}

// SearchResult represents search results
type SearchResult struct {
	Results    []ProductResult `json:"results"`
	Pagination Pagination      `json:"pagination"`
	Query      string          `json:"query"`
}

// ProductResult represents a product in search results
type ProductResult struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Price          float64   `json:"price"`
	Category       string    `json:"category"`
	SKU            string    `json:"sku"`
	Availability   string    `json:"availability"`
	Popularity     int       `json:"popularity"`
	RelevanceScore float64   `json:"relevance_score"`
}

// Pagination represents pagination information
type Pagination struct {
	CurrentPage  int  `json:"current_page"`
	PageSize     int  `json:"page_size"`
	TotalPages   int  `json:"total_pages"`
	TotalResults int  `json:"total_results"`
	HasNext      bool `json:"has_next"`
	HasPrevious  bool `json:"has_previous"`
}

// getAvailabilityStatus returns availability status based on quantity
func getAvailabilityStatus(quantity int) string {
	if quantity == 0 {
		return "out_of_stock"
	} else if quantity <= 10 {
		return "low_stock"
	}
	return "in_stock"
}

// calculateRelevanceScore calculates a simple relevance score
func calculateRelevanceScore(query, name, description string) float64 {
	score := 0.0
	queryLower := strings.ToLower(query)
	nameLower := strings.ToLower(name)
	descLower := strings.ToLower(description)

	// Exact name match gets highest score
	if strings.Contains(nameLower, queryLower) {
		score += 1.0
	}

	// Description match gets lower score
	if strings.Contains(descLower, queryLower) {
		score += 0.5
	}

	// Word boundary matches get bonus
	words := strings.Fields(queryLower)
	for _, word := range words {
		if strings.Contains(nameLower, word) {
			score += 0.3
		}
		if strings.Contains(descLower, word) {
			score += 0.1
		}
	}

	return score
}
