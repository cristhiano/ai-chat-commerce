package password

import (
	"errors"

	"golang.org/x/crypto/bcrypt"
)

const (
	// DefaultCost is the default bcrypt cost factor (10)
	DefaultCost = 10
)

// Service handles password hashing and verification
type Service struct {
	cost int
}

// NewService creates a new password service with the specified cost
func NewService(cost int) *Service {
	if cost <= 0 {
		cost = DefaultCost
	}
	return &Service{
		cost: cost,
	}
}

// HashPassword hashes a plain text password using bcrypt
func (s *Service) HashPassword(password string) (string, error) {
	if password == "" {
		return "", errors.New("password cannot be empty")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), s.cost)
	if err != nil {
		return "", errors.New("failed to hash password")
	}

	return string(hash), nil
}

// VerifyPassword verifies a plain text password against a hash
func (s *Service) VerifyPassword(hashedPassword, password string) bool {
	if hashedPassword == "" || password == "" {
		return false
	}

	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
	return err == nil
}
