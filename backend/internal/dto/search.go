package dto

import (
	"time"

	"github.com/google/uuid"
)

// SearchRequestDTO represents the search request DTO
type SearchRequestDTO struct {
	Query    string                 `json:"query" binding:"required"`
	Filters  map[string]interface{} `json:"filters,omitempty"`
	SortBy   string                 `json:"sort_by,omitempty"`
	Page     int                    `json:"page,omitempty"`
	PageSize int                    `json:"page_size,omitempty"`
}

// SearchResponseDTO represents the search response DTO
type SearchResponseDTO struct {
	Results        []ProductResultDTO `json:"results"`
	Pagination     PaginationDTO      `json:"pagination"`
	Query          string             `json:"query"`
	TotalResults   int                `json:"total_results"`
	ResponseTimeMs int                `json:"response_time_ms"`
}

// ProductResultDTO represents a product in search results
type ProductResultDTO struct {
	ID             uuid.UUID `json:"id"`
	Name           string    `json:"name"`
	Description    string    `json:"description"`
	Price          float64   `json:"price"`
	Category       string    `json:"category"`
	SKU            string    `json:"sku"`
	Availability   string    `json:"availability"`
	RelevanceScore float64   `json:"relevance_score"`
	ImageURL       string    `json:"image_url,omitempty"`
}

// PaginationDTO represents pagination information
type PaginationDTO struct {
	CurrentPage  int  `json:"current_page"`
	PageSize     int  `json:"page_size"`
	TotalPages   int  `json:"total_pages"`
	TotalResults int  `json:"total_results"`
	HasNext      bool `json:"has_next"`
	HasPrevious  bool `json:"has_previous"`
}

// SuggestionsResponseDTO represents autocomplete suggestions response
type SuggestionsResponseDTO struct {
	Suggestions []string `json:"suggestions"`
}

// FilterOptionsDTO represents available filter options
type FilterOptionsDTO struct {
	PriceRanges         []PriceRangeDTO     `json:"price_ranges"`
	Categories          []CategoryOptionDTO `json:"categories"`
	AvailabilityOptions []string            `json:"availability_options"`
}

// PriceRangeDTO represents a price range filter
type PriceRangeDTO struct {
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Label string  `json:"label"`
}

// CategoryOptionDTO represents a category filter option
type CategoryOptionDTO struct {
	ID           string `json:"id"`
	Name         string `json:"name"`
	ProductCount int    `json:"product_count"`
}

// AnalyticsRequestDTO represents analytics data
type AnalyticsRequestDTO struct {
	Query            string                 `json:"query"`
	Filters          map[string]interface{} `json:"filters,omitempty"`
	ResultCount      int                    `json:"result_count"`
	SelectedProducts string                 `json:"selected_products,omitempty"`
	ResponseTimeMs   int                    `json:"response_time_ms"`
	CacheHit         bool                   `json:"cache_hit"`
	SessionID        string                 `json:"session_id"`
	UserID           string                 `json:"user_id,omitempty"`
}

// ErrorResponseDTO represents an error response
type ErrorResponseDTO struct {
	Error ErrorDetailDTO `json:"error"`
}

// ErrorDetailDTO represents error details
type ErrorDetailDTO struct {
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// SuccessResponseDTO represents a success response
type SuccessResponseDTO struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
}

// SearchHistoryDTO represents search history
type SearchHistoryDTO struct {
	ID          uuid.UUID              `json:"id"`
	Query       string                 `json:"query"`
	Filters     map[string]interface{} `json:"filters,omitempty"`
	ResultCount int                    `json:"result_count"`
	CreatedAt   time.Time              `json:"created_at"`
}

// SearchStatsDTO represents search statistics
type SearchStatsDTO struct {
	TotalSearches       int      `json:"total_searches"`
	AverageResponseTime float64  `json:"average_response_time"`
	CacheHitRate        float64  `json:"cache_hit_rate"`
	TopQueries          []string `json:"top_queries"`
}
