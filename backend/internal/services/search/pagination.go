package search

import (
	"context"
	"fmt"
	"math"

	"gorm.io/gorm"
)

// PaginationService handles pagination logic
type PaginationService struct {
}

// NewPaginationService creates a new pagination service
func NewPaginationService() *PaginationService {
	return &PaginationService{}
}

// CalculatePagination calculates pagination information
func (ps *PaginationService) CalculatePagination(totalResults int, page, pageSize int) PaginationInfo {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	if pageSize > 100 {
		pageSize = 100
	}

	totalPages := int(math.Ceil(float64(totalResults) / float64(pageSize)))
	if totalPages == 0 {
		totalPages = 1
	}

	// Ensure page doesn't exceed total pages
	if page > totalPages {
		page = totalPages
	}

	return PaginationInfo{
		CurrentPage:  page,
		PageSize:     pageSize,
		TotalPages:   totalPages,
		TotalResults: totalResults,
		HasNext:      page < totalPages,
		HasPrevious:  page > 1,
		Offset:       (page - 1) * pageSize,
	}
}

// ValidatePagination validates pagination parameters
func (ps *PaginationService) ValidatePagination(page, pageSize int) (int, int, error) {
	if page < 1 {
		return 1, pageSize, fmt.Errorf("page must be greater than 0")
	}

	if pageSize < 1 {
		return page, 20, fmt.Errorf("page size must be greater than 0")
	}

	if pageSize > 100 {
		return page, 100, fmt.Errorf("page size cannot exceed 100")
	}

	return page, pageSize, nil
}

// GetPageRange returns the range of items for a given page
func (ps *PaginationService) GetPageRange(page, pageSize, totalResults int) (int, int) {
	offset := (page - 1) * pageSize
	limit := pageSize

	// Ensure we don't exceed total results
	if offset >= totalResults {
		return 0, 0
	}

	if offset+limit > totalResults {
		limit = totalResults - offset
	}

	return offset, limit
}

// PaginationInfo represents pagination information
type PaginationInfo struct {
	CurrentPage  int  `json:"current_page"`
	PageSize     int  `json:"page_size"`
	TotalPages   int  `json:"total_pages"`
	TotalResults int  `json:"total_results"`
	HasNext      bool `json:"has_next"`
	HasPrevious  bool `json:"has_previous"`
	Offset       int  `json:"offset"`
}

// ApplyPaginationToQuery applies pagination to a GORM query
func (ps *PaginationService) ApplyPaginationToQuery(ctx context.Context, query *gorm.DB, page, pageSize int) (*gorm.DB, error) {
	validatedPage, validatedPageSize, err := ps.ValidatePagination(page, pageSize)
	if err != nil {
		return nil, err
	}

	offset := (validatedPage - 1) * validatedPageSize
	return query.Offset(offset).Limit(validatedPageSize), nil
}

// GetPaginationMetadata returns pagination metadata for API responses
func (ps *PaginationService) GetPaginationMetadata(totalResults int, page, pageSize int) map[string]interface{} {
	pagination := ps.CalculatePagination(totalResults, page, pageSize)

	return map[string]interface{}{
		"current_page":  pagination.CurrentPage,
		"page_size":     pagination.PageSize,
		"total_pages":   pagination.TotalPages,
		"total_results": pagination.TotalResults,
		"has_next":      pagination.HasNext,
		"has_previous":  pagination.HasPrevious,
	}
}
