package session

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

const (
	// DefaultExpiration is the default session expiration time (24 hours)
	DefaultExpiration = 24 * time.Hour
)

// Claims represents JWT claims
type Claims struct {
	UserID    string `json:"user_id"`
	Email     string `json:"email"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
	jwt.RegisteredClaims
}

// Service handles JWT token generation and validation
type Service struct {
	secretKey []byte
}

// NewService creates a new session service
func NewService(secretKey string) (*Service, error) {
	if secretKey == "" {
		return nil, errors.New("secret key cannot be empty")
	}
	return &Service{
		secretKey: []byte(secretKey),
	}, nil
}

// GenerateToken generates a JWT token for a user
func (s *Service) GenerateToken(userID, email string) (string, error) {
	now := time.Now()
	expiresAt := now.Add(DefaultExpiration)

	claims := Claims{
		UserID:    userID,
		Email:     email,
		IssuedAt:  now.Unix(),
		ExpiresAt: expiresAt.Unix(),
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// GenerateTokenWithExpiration generates a JWT token with custom expiration
func (s *Service) GenerateTokenWithExpiration(userID, email string, expiration time.Duration) (string, error) {
	now := time.Now()
	expiresAt := now.Add(expiration)

	claims := Claims{
		UserID:    userID,
		Email:     email,
		IssuedAt:  now.Unix(),
		ExpiresAt: expiresAt.Unix(),
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(expiresAt),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.secretKey)
}

// GetTokenExpiration returns the default token expiration time
func (s *Service) GetTokenExpiration() time.Duration {
	return DefaultExpiration
}
