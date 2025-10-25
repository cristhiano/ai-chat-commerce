package validation

import (
	"errors"
	"regexp"
	"strings"
	"unicode"

	"github.com/google/uuid"
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors represents multiple validation errors
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	var messages []string
	for _, err := range ve {
		messages = append(messages, err.Message)
	}
	return strings.Join(messages, "; ")
}

// ValidateEmail validates an email address
func ValidateEmail(email string) error {
	if email == "" {
		return errors.New("email is required")
	}

	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	if !emailRegex.MatchString(email) {
		return errors.New("invalid email format")
	}

	return nil
}

// ValidatePassword validates a password
func ValidatePassword(password string) error {
	if password == "" {
		return errors.New("password is required")
	}

	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	var hasUpper, hasLower, hasDigit, hasSpecial bool
	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsDigit(char):
			hasDigit = true
		case unicode.IsPunct(char) || unicode.IsSymbol(char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		return errors.New("password must contain at least one uppercase letter")
	}
	if !hasLower {
		return errors.New("password must contain at least one lowercase letter")
	}
	if !hasDigit {
		return errors.New("password must contain at least one digit")
	}
	if !hasSpecial {
		return errors.New("password must contain at least one special character")
	}

	return nil
}

// ValidateName validates a name field
func ValidateName(name, fieldName string) error {
	if name == "" {
		return errors.New(fieldName + " is required")
	}

	if len(name) < 2 {
		return errors.New(fieldName + " must be at least 2 characters long")
	}

	if len(name) > 50 {
		return errors.New(fieldName + " must be no more than 50 characters long")
	}

	// Check for valid characters (letters, spaces, hyphens, apostrophes)
	nameRegex := regexp.MustCompile(`^[a-zA-Z\s\-']+$`)
	if !nameRegex.MatchString(name) {
		return errors.New(fieldName + " contains invalid characters")
	}

	return nil
}

// ValidatePhone validates a phone number
func ValidatePhone(phone string) error {
	if phone == "" {
		return nil // Phone is optional
	}

	// Remove all non-digit characters
	digits := regexp.MustCompile(`\D`).ReplaceAllString(phone, "")
	
	if len(digits) < 10 || len(digits) > 15 {
		return errors.New("phone number must be between 10 and 15 digits")
	}

	return nil
}

// ValidateUUID validates a UUID string
func ValidateUUID(id, fieldName string) error {
	if id == "" {
		return errors.New(fieldName + " is required")
	}

	if _, err := uuid.Parse(id); err != nil {
		return errors.New("invalid " + fieldName + " format")
	}

	return nil
}

// ValidatePrice validates a price value
func ValidatePrice(price float64, fieldName string) error {
	if price < 0 {
		return errors.New(fieldName + " must be non-negative")
	}

	if price > 999999.99 {
		return errors.New(fieldName + " must be less than 1,000,000")
	}

	return nil
}

// ValidateQuantity validates a quantity value
func ValidateQuantity(quantity int, fieldName string) error {
	if quantity <= 0 {
		return errors.New(fieldName + " must be positive")
	}

	if quantity > 10000 {
		return errors.New(fieldName + " must be less than 10,000")
	}

	return nil
}

// ValidateSKU validates a SKU string
func ValidateSKU(sku string) error {
	if sku == "" {
		return errors.New("SKU is required")
	}

	if len(sku) < 3 || len(sku) > 100 {
		return errors.New("SKU must be between 3 and 100 characters")
	}

	// SKU should contain only alphanumeric characters, hyphens, and underscores
	skuRegex := regexp.MustCompile(`^[a-zA-Z0-9\-_]+$`)
	if !skuRegex.MatchString(sku) {
		return errors.New("SKU contains invalid characters")
	}

	return nil
}

// ValidateRequired validates that a field is not empty
func ValidateRequired(value, fieldName string) error {
	if strings.TrimSpace(value) == "" {
		return errors.New(fieldName + " is required")
	}
	return nil
}

// ValidateLength validates string length
func ValidateLength(value, fieldName string, min, max int) error {
	length := len(strings.TrimSpace(value))
	if length < min {
		return errors.New(fieldName + " must be at least " + string(rune(min)) + " characters")
	}
	if length > max {
		return errors.New(fieldName + " must be no more than " + string(rune(max)) + " characters")
	}
	return nil
}

// ValidateStatus validates a status field
func ValidateStatus(status string, validStatuses []string) error {
	if status == "" {
		return errors.New("status is required")
	}

	for _, validStatus := range validStatuses {
		if status == validStatus {
			return nil
		}
	}

	return errors.New("invalid status value")
}

// ValidatePagination validates pagination parameters
func ValidatePagination(page, limit int) error {
	if page < 1 {
		return errors.New("page must be at least 1")
	}
	if limit < 1 || limit > 100 {
		return errors.New("limit must be between 1 and 100")
	}
	return nil
}
