package search

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"chat-ecommerce-backend/internal/dto"
	"chat-ecommerce-backend/internal/services/search"

	"github.com/gin-gonic/gin"
)

// Handler handles search HTTP requests
type Handler struct {
	searchService *search.Service
}

// NewHandler creates a new search handler
func NewHandler(searchService *search.Service) *Handler {
	return &Handler{
		searchService: searchService,
	}
}

// SearchProducts handles POST /api/search
func (h *Handler) SearchProducts(c *gin.Context) {
	var req dto.SearchRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponseDTO{
			Error: dto.ErrorDetailDTO{
				Code:    "INVALID_REQUEST",
				Message: "Invalid search request",
				Details: err.Error(),
			},
		})
		return
	}

	// Validate search query
	if err := h.validateSearchRequest(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponseDTO{
			Error: dto.ErrorDetailDTO{
				Code:    "VALIDATION_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	// Convert to service request
	serviceReq := &search.SearchRequest{
		Query:    req.Query,
		Filters:  req.Filters,
		SortBy:   req.SortBy,
		Page:     req.Page,
		PageSize: req.PageSize,
	}

	// Perform search
	result, err := h.searchService.Search(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponseDTO{
			Error: dto.ErrorDetailDTO{
				Code:    "SEARCH_ERROR",
				Message: "Failed to perform search",
				Details: err.Error(),
			},
		})
		return
	}

	// Convert to response DTO
	response := dto.SearchResponseDTO{
		Results:        h.convertToProductResultDTOs(result.Results),
		Pagination:     h.convertToPaginationDTO(result.Pagination),
		Query:          result.Query,
		TotalResults:   result.TotalResults,
		ResponseTimeMs: result.ResponseTimeMs,
	}

	c.JSON(http.StatusOK, response)
}

// GetSuggestions handles GET /api/search/suggestions
func (h *Handler) GetSuggestions(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, dto.ErrorResponseDTO{
			Error: dto.ErrorDetailDTO{
				Code:    "MISSING_QUERY",
				Message: "Query parameter 'q' is required",
			},
		})
		return
	}

	limitStr := c.DefaultQuery("limit", "10")
	limit, err := strconv.Atoi(limitStr)
	if err != nil || limit < 1 || limit > 50 {
		limit = 10
	}

	suggestions, err := h.searchService.GetSuggestions(c.Request.Context(), query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponseDTO{
			Error: dto.ErrorDetailDTO{
				Code:    "SUGGESTIONS_ERROR",
				Message: "Failed to get suggestions",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuggestionsResponseDTO{
		Suggestions: suggestions,
	})
}

// GetFilterOptions handles GET /api/search/filters
func (h *Handler) GetFilterOptions(c *gin.Context) {
	categoryID := c.Query("category_id")

	options, err := h.searchService.GetFilterOptions(c.Request.Context(), categoryID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponseDTO{
			Error: dto.ErrorDetailDTO{
				Code:    "FILTERS_ERROR",
				Message: "Failed to get filter options",
				Details: err.Error(),
			},
		})
		return
	}

	response := dto.FilterOptionsDTO{
		PriceRanges:         h.convertToPriceRangeDTOs(options.PriceRanges),
		Categories:          h.convertToCategoryOptionDTOs(options.Categories),
		AvailabilityOptions: options.AvailabilityOptions,
	}

	c.JSON(http.StatusOK, response)
}

// LogAnalytics handles POST /api/search/analytics
func (h *Handler) LogAnalytics(c *gin.Context) {
	var req dto.AnalyticsRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, dto.ErrorResponseDTO{
			Error: dto.ErrorDetailDTO{
				Code:    "INVALID_REQUEST",
				Message: "Invalid analytics request",
				Details: err.Error(),
			},
		})
		return
	}

	// Convert to service request
	serviceReq := &search.AnalyticsRequest{
		Query:            req.Query,
		Filters:          req.Filters,
		ResultCount:      req.ResultCount,
		SelectedProducts: req.SelectedProducts,
		ResponseTimeMs:   req.ResponseTimeMs,
		CacheHit:         req.CacheHit,
		SessionID:        req.SessionID,
		UserID:           req.UserID,
	}

	err := h.searchService.LogAnalytics(c.Request.Context(), serviceReq)
	if err != nil {
		c.JSON(http.StatusInternalServerError, dto.ErrorResponseDTO{
			Error: dto.ErrorDetailDTO{
				Code:    "ANALYTICS_ERROR",
				Message: "Failed to log analytics",
				Details: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, dto.SuccessResponseDTO{
		Success: true,
		Message: "Analytics logged successfully",
	})
}

// validateSearchRequest validates search request
func (h *Handler) validateSearchRequest(req *dto.SearchRequestDTO) error {
	// Validate query
	req.Query = strings.TrimSpace(req.Query)
	if req.Query == "" {
		return fmt.Errorf("search query cannot be empty")
	}

	if len(req.Query) > 500 {
		return fmt.Errorf("search query too long (max 500 characters)")
	}

	// Validate page
	if req.Page < 1 {
		req.Page = 1
	}

	// Validate page size
	if req.PageSize < 1 {
		req.PageSize = 20
	}
	if req.PageSize > 100 {
		req.PageSize = 100
	}

	// Validate sort by
	validSortOptions := []string{"relevance", "price_asc", "price_desc", "popularity", "newest"}
	if req.SortBy != "" {
		valid := false
		for _, option := range validSortOptions {
			if req.SortBy == option {
				valid = true
				break
			}
		}
		if !valid {
			req.SortBy = "relevance"
		}
	} else {
		req.SortBy = "relevance"
	}

	return nil
}

// convertToProductResultDTOs converts service results to DTOs
func (h *Handler) convertToProductResultDTOs(results []search.ProductResult) []dto.ProductResultDTO {
	dtos := make([]dto.ProductResultDTO, len(results))
	for i, result := range results {
		dtos[i] = dto.ProductResultDTO{
			ID:             result.ID,
			Name:           result.Name,
			Description:    result.Description,
			Price:          result.Price,
			Category:       result.Category,
			SKU:            result.SKU,
			Availability:   result.Availability,
			RelevanceScore: result.RelevanceScore,
		}
	}
	return dtos
}

// convertToPaginationDTO converts service pagination to DTO
func (h *Handler) convertToPaginationDTO(pagination search.Pagination) dto.PaginationDTO {
	return dto.PaginationDTO{
		CurrentPage:  pagination.CurrentPage,
		PageSize:     pagination.PageSize,
		TotalPages:   pagination.TotalPages,
		TotalResults: pagination.TotalResults,
		HasNext:      pagination.HasNext,
		HasPrevious:  pagination.HasPrevious,
	}
}

// convertToPriceRangeDTOs converts service price ranges to DTOs
func (h *Handler) convertToPriceRangeDTOs(ranges []search.PriceRange) []dto.PriceRangeDTO {
	dtos := make([]dto.PriceRangeDTO, len(ranges))
	for i, r := range ranges {
		dtos[i] = dto.PriceRangeDTO{
			Min:   r.Min,
			Max:   r.Max,
			Label: r.Label,
		}
	}
	return dtos
}

// convertToCategoryOptionDTOs converts service category options to DTOs
func (h *Handler) convertToCategoryOptionDTOs(categories []search.CategoryOption) []dto.CategoryOptionDTO {
	dtos := make([]dto.CategoryOptionDTO, len(categories))
	for i, cat := range categories {
		dtos[i] = dto.CategoryOptionDTO{
			ID:           cat.ID,
			Name:         cat.Name,
			ProductCount: cat.ProductCount,
		}
	}
	return dtos
}
