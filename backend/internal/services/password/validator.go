package password

import (
	"errors"
	"regexp"
	"strings"
	"unicode"
)

// ValidationResult contains the result of password validation
type ValidationResult struct {
	IsValid bool
	Errors  []string
}

// Service handles password validation
type Validator struct{}

// NewValidator creates a new password validator
func NewValidator() *Validator {
	return &Validator{}
}

// ValidatePassword checks if a password meets strength requirements
func (v *Validator) ValidatePassword(password string) ValidationResult {
	errors := []string{}

	if len(password) < 8 {
		errors = append(errors, "password must be at least 8 characters long")
	}

	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
			hasSpecial = true
		}
	}

	if !hasUpper {
		errors = append(errors, "password must contain at least one uppercase letter")
	}

	if !hasLower {
		errors = append(errors, "password must contain at least one lowercase letter")
	}

	if !hasNumber {
		errors = append(errors, "password must contain at least one number")
	}

	if !hasSpecial {
		errors = append(errors, "password must contain at least one special character (!@#$%^&*()_+-=[]{}|;:,.<>?)")
	}

	return ValidationResult{
		IsValid: len(errors) == 0,
		Errors:  errors,
	}
}

// ValidatePasswordPattern uses regex to validate password format
func (v *Validator) ValidatePasswordPattern(password string) error {
	// Pattern requires: at least 8 chars, 1 uppercase, 1 lowercase, 1 number, 1 special char
	pattern := `^(?=.*[a-z])(?=.*[A-Z])(?=.*\d)(?=.*[@$!%*?&])[A-Za-z\d@$!%*?&]+$`
	matched, err := regexp.MatchString(pattern, password)
	if err != nil {
		return err
	}

	if !matched {
		return errors.New("password must contain uppercase, lowercase, number, and special character")
	}

	if len(password) < 8 {
		return errors.New("password must be at least 8 characters long")
	}

	return nil
}

// GetPasswordStrength returns a strength rating from 0-4
func (v *Validator) GetPasswordStrength(password string) int {
	strength := 0

	if len(password) >= 8 {
		strength++
	}
	if len(password) >= 12 {
		strength++
	}

	hasUpper := false
	hasLower := false
	hasNumber := false
	hasSpecial := false

	for _, char := range password {
		switch {
		case unicode.IsUpper(char):
			hasUpper = true
		case unicode.IsLower(char):
			hasLower = true
		case unicode.IsNumber(char):
			hasNumber = true
		case strings.ContainsRune("!@#$%^&*()_+-=[]{}|;:,.<>?", char):
			hasSpecial = true
		}
	}

	if hasUpper {
		strength++
	}
	if hasLower {
		strength++
	}
	if hasNumber {
		strength++
	}
	if hasSpecial {
		strength++
	}

	return strength
}
