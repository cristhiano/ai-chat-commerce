package validation

import (
	"fmt"
	"regexp"
	"strings"
)

// SearchValidator handles search input validation
type SearchValidator struct {
	maxQueryLength int
	minQueryLength int
}

// NewSearchValidator creates a new search validator
func NewSearchValidator() *SearchValidator {
	return &SearchValidator{
		maxQueryLength: 500,
		minQueryLength: 1,
	}
}

// ValidateSearchQuery validates a search query
func (sv *SearchValidator) ValidateSearchQuery(query string) error {
	// Trim whitespace
	query = strings.TrimSpace(query)

	// Check if empty
	if query == "" {
		return fmt.Errorf("search query cannot be empty")
	}

	// Check minimum length
	if len(query) < sv.minQueryLength {
		return fmt.Errorf("search query must be at least %d character(s)", sv.minQueryLength)
	}

	// Check maximum length
	if len(query) > sv.maxQueryLength {
		return fmt.Errorf("search query cannot exceed %d characters", sv.maxQueryLength)
	}

	// Check for SQL injection patterns
	if sv.containsSQLInjection(query) {
		return fmt.Errorf("invalid characters in search query")
	}

	// Check for XSS patterns
	if sv.containsXSS(query) {
		return fmt.Errorf("invalid characters in search query")
	}

	return nil
}

// ValidateFilters validates search filters
func (sv *SearchValidator) ValidateFilters(filters map[string]interface{}) error {
	if filters == nil {
		return nil
	}

	// Validate category ID if present
	if categoryID, ok := filters["category_id"].(string); ok {
		if err := sv.validateUUID(categoryID); err != nil {
			return fmt.Errorf("invalid category_id: %w", err)
		}
	}

	// Validate price range if present
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

	// Validate price range consistency
	if priceMin, ok := filters["price_min"].(float64); ok {
		if priceMax, ok := filters["price_max"].(float64); ok {
			if priceMin > priceMax {
				return fmt.Errorf("price_min cannot be greater than price_max")
			}
		}
	}

	// Validate availability if present
	if availability, ok := filters["availability"].(string); ok {
		validOptions := []string{"all", "in_stock", "low_stock", "out_of_stock"}
		if !sv.isValidOption(availability, validOptions) {
			return fmt.Errorf("invalid availability option")
		}
	}

	return nil
}

// ValidatePagination validates pagination parameters
func (sv *SearchValidator) ValidatePagination(page, pageSize int) error {
	if page < 1 {
		return fmt.Errorf("page must be greater than 0")
	}

	if pageSize < 1 {
		return fmt.Errorf("page size must be greater than 0")
	}

	if pageSize > 100 {
		return fmt.Errorf("page size cannot exceed 100")
	}

	return nil
}

// ValidateSortBy validates sort parameter
func (sv *SearchValidator) ValidateSortBy(sortBy string) error {
	if sortBy == "" {
		return nil // Default sorting is allowed
	}

	validOptions := []string{"relevance", "price_asc", "price_desc", "popularity", "newest"}
	if !sv.isValidOption(sortBy, validOptions) {
		return fmt.Errorf("invalid sort option")
	}

	return nil
}

// containsSQLInjection checks for SQL injection patterns
func (sv *SearchValidator) containsSQLInjection(query string) bool {
	// Common SQL injection patterns
	patterns := []string{
		`(?i)(union|select|insert|update|delete|drop|create|alter|exec|execute)`,
		`(?i)(--|\/\*|\*\/)`,
		`(?i)(or|and)\s+\d+\s*=\s*\d+`,
		`(?i)(or|and)\s+'.*'\s*=\s*'.*'`,
		`(?i)(or|and)\s+".*"\s*=\s*".*"`,
		`(?i)(or|and)\s+\w+\s*=\s*\w+`,
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, query)
		if matched {
			return true
		}
	}

	return false
}

// containsXSS checks for XSS patterns
func (sv *SearchValidator) containsXSS(query string) bool {
	// Common XSS patterns
	patterns := []string{
		`(?i)<script`,
		`(?i)<iframe`,
		`(?i)<object`,
		`(?i)<embed`,
		`(?i)<link`,
		`(?i)<meta`,
		`(?i)javascript:`,
		`(?i)vbscript:`,
		`(?i)onload=`,
		`(?i)onerror=`,
		`(?i)onclick=`,
	}

	for _, pattern := range patterns {
		matched, _ := regexp.MatchString(pattern, query)
		if matched {
			return true
		}
	}

	return false
}

// validateUUID validates UUID format
func (sv *SearchValidator) validateUUID(uuid string) error {
	uuidPattern := `^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`
	matched, _ := regexp.MatchString(uuidPattern, strings.ToLower(uuid))
	if !matched {
		return fmt.Errorf("invalid UUID format")
	}
	return nil
}

// isValidOption checks if a value is in the list of valid options
func (sv *SearchValidator) isValidOption(value string, validOptions []string) bool {
	for _, option := range validOptions {
		if value == option {
			return true
		}
	}
	return false
}

// SanitizeQuery sanitizes a search query
func (sv *SearchValidator) SanitizeQuery(query string) string {
	// Remove extra whitespace
	query = strings.TrimSpace(query)

	// Replace multiple spaces with single space
	spaceRegex := regexp.MustCompile(`\s+`)
	query = spaceRegex.ReplaceAllString(query, " ")

	// Remove special characters that could be problematic
	specialCharRegex := regexp.MustCompile(`[<>'"&]`)
	query = specialCharRegex.ReplaceAllString(query, "")

	return query
}
