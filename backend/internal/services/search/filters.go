package search

import (
	"context"
	"fmt"
	"strconv"

	"gorm.io/gorm"
)

// FilterService handles search filtering operations
type FilterService struct {
	db *gorm.DB
}

// NewFilterService creates a new filter service
func NewFilterService(db *gorm.DB) *FilterService {
	return &FilterService{db: db}
}

// ApplyPriceRangeFilter applies price range filtering to query
func (fs *FilterService) ApplyPriceRangeFilter(query *gorm.DB, minPrice, maxPrice float64) *gorm.DB {
	if minPrice > 0 {
		query = query.Where("price >= ?", minPrice)
	}
	if maxPrice > 0 {
		query = query.Where("price <= ?", maxPrice)
	}
	return query
}

// ApplyCategoryFilter applies category filtering to query
func (fs *FilterService) ApplyCategoryFilter(query *gorm.DB, categoryID string) *gorm.DB {
	if categoryID != "" {
		query = query.Where("category_id = ?", categoryID)
	}
	return query
}

// ApplyAvailabilityFilter applies availability filtering to query
func (fs *FilterService) ApplyAvailabilityFilter(query *gorm.DB, availability string) *gorm.DB {
	switch availability {
	case "in_stock":
		query = query.Where("total_available > 0")
	case "low_stock":
		query = query.Where("total_available > 0 AND total_available <= 10")
	case "out_of_stock":
		query = query.Where("total_available = 0")
	case "all":
		// No filter applied
	default:
		// Default to in_stock
		query = query.Where("total_available > 0")
	}
	return query
}

// ApplySortFilter applies sorting to query
func (fs *FilterService) ApplySortFilter(query *gorm.DB, sortBy string, searchQuery string) *gorm.DB {
	switch sortBy {
	case "price_asc":
		query = query.Order("price ASC")
	case "price_desc":
		query = query.Order("price DESC")
	case "popularity":
		query = query.Order("popularity DESC")
	case "newest":
		query = query.Order("created_at DESC")
	case "relevance":
		fallthrough
	default:
		// Use full-text search ranking
		if searchQuery != "" {
			query = query.Order(fmt.Sprintf("ts_rank(search_vector, plainto_tsquery('english', '%s')) DESC", searchQuery))
		} else {
			query = query.Order("popularity DESC")
		}
	}
	return query
}

// ApplyTagFilter applies tag filtering to query
func (fs *FilterService) ApplyTagFilter(query *gorm.DB, tags []string) *gorm.DB {
	if len(tags) > 0 {
		// Use PostgreSQL array contains operator
		for _, tag := range tags {
			query = query.Where("tags @> ?", fmt.Sprintf("{\"%s\"}", tag))
		}
	}
	return query
}

// ApplyBrandFilter applies brand filtering to query
func (fs *FilterService) ApplyBrandFilter(query *gorm.DB, brand string) *gorm.DB {
	if brand != "" {
		query = query.Where("metadata->>'brand' = ?", brand)
	}
	return query
}

// ApplyRatingFilter applies rating filtering to query
func (fs *FilterService) ApplyRatingFilter(query *gorm.DB, minRating float64) *gorm.DB {
	if minRating > 0 {
		query = query.Where("metadata->>'rating' >= ?", strconv.FormatFloat(minRating, 'f', 1, 64))
	}
	return query
}

// GetFilterOptions returns available filter options for current search
func (fs *FilterService) GetFilterOptions(ctx context.Context, baseQuery string, filters map[string]interface{}) (*FilterOptions, error) {
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

	// Get categories with product counts
	var categories []CategoryOption
	err := fs.db.WithContext(ctx).
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

	// Get price ranges with actual counts
	var priceRanges []PriceRangeWithCount
	err = fs.db.WithContext(ctx).
		Table("products").
		Select(`
			CASE 
				WHEN price < 50 THEN '0-50'
				WHEN price < 100 THEN '50-100'
				WHEN price < 200 THEN '100-200'
				WHEN price < 500 THEN '200-500'
				ELSE '500+'
			END as range_label,
			CASE 
				WHEN price < 50 THEN 0
				WHEN price < 100 THEN 50
				WHEN price < 200 THEN 100
				WHEN price < 500 THEN 200
				ELSE 500
			END as min_price,
			CASE 
				WHEN price < 50 THEN 50
				WHEN price < 100 THEN 100
				WHEN price < 200 THEN 200
				WHEN price < 500 THEN 500
				ELSE 0
			END as max_price,
			COUNT(*) as product_count
		`).
		Where("status = 'active'").
		Group("range_label, min_price, max_price").
		Order("min_price").
		Scan(&priceRanges).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get price ranges: %w", err)
	}

	// Convert to PriceRange format
	for _, pr := range priceRanges {
		options.PriceRanges = append(options.PriceRanges, PriceRange{
			Min:   pr.MinPrice,
			Max:   pr.MaxPrice,
			Label: fmt.Sprintf("%s (%d)", pr.RangeLabel, pr.ProductCount),
		})
	}

	// Get brands
	var brands []BrandOption
	err = fs.db.WithContext(ctx).
		Table("products").
		Select("metadata->>'brand' as brand, COUNT(*) as product_count").
		Where("status = 'active' AND metadata->>'brand' IS NOT NULL").
		Group("metadata->>'brand'").
		Order("product_count DESC").
		Limit(20).
		Scan(&brands).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get brands: %w", err)
	}

	options.Brands = brands

	return options, nil
}

// ValidateFilters validates filter parameters
func (fs *FilterService) ValidateFilters(filters map[string]interface{}) error {
	// Validate price range
	if priceMin, ok := filters["price_min"].(float64); ok {
		if priceMin < 0 {
			return fmt.Errorf("price_min cannot be negative")
		}
	}

	if priceMax, ok := filters["price_max"].(float64); ok {
		if priceMax < 0 {
			return fmt.Errorf("price_max cannot be negative")
		}
	}

	if priceMin, ok := filters["price_min"].(float64); ok {
		if priceMax, ok := filters["price_max"].(float64); ok {
			if priceMin > priceMax {
				return fmt.Errorf("price_min cannot be greater than price_max")
			}
		}
	}

	// Validate availability
	if availability, ok := filters["availability"].(string); ok {
		validOptions := []string{"all", "in_stock", "low_stock", "out_of_stock"}
		if !fs.isValidOption(availability, validOptions) {
			return fmt.Errorf("invalid availability option")
		}
	}

	// Validate sort option
	if sortBy, ok := filters["sort_by"].(string); ok {
		validOptions := []string{"relevance", "price_asc", "price_desc", "popularity", "newest"}
		if !fs.isValidOption(sortBy, validOptions) {
			return fmt.Errorf("invalid sort option")
		}
	}

	return nil
}

// isValidOption checks if a value is in the list of valid options
func (fs *FilterService) isValidOption(value string, validOptions []string) bool {
	for _, option := range validOptions {
		if value == option {
			return true
		}
	}
	return false
}

// FilterOptions represents available filter options
type FilterOptions struct {
	PriceRanges         []PriceRange     `json:"price_ranges"`
	Categories          []CategoryOption `json:"categories"`
	AvailabilityOptions []string         `json:"availability_options"`
	Brands              []BrandOption    `json:"brands"`
}

// PriceRange represents a price range filter
type PriceRange struct {
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Label string  `json:"label"`
}

// PriceRangeWithCount represents a price range with product count
type PriceRangeWithCount struct {
	RangeLabel   string  `json:"range_label"`
	MinPrice     float64 `json:"min_price"`
	MaxPrice     float64 `json:"max_price"`
	ProductCount int     `json:"product_count"`
}

// CategoryOption represents a category filter option
type CategoryOption struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ProductCount int    `json:"product_count"`
}

// BrandOption represents a brand filter option
type BrandOption struct {
	Brand        string `json:"brand"`
	ProductCount int    `json:"product_count"`
}
