package session

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// Validator handles JWT token validation
type Validator struct {
	secretKey []byte
}

// NewValidator creates a new token validator
func NewValidator(secretKey string) (*Validator, error) {
	if secretKey == "" {
		return nil, errors.New("secret key cannot be empty")
	}
	return &Validator{
		secretKey: []byte(secretKey),
	}, nil
}

// ValidateToken validates a JWT token and extracts claims
func (v *Validator) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("invalid signing method")
		}
		return v.secretKey, nil
	})

	if err != nil {
		return nil, errors.New("invalid token")
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	// Check if token is expired
	if claims.ExpiresAt != 0 && time.Now().Unix() > claims.ExpiresAt {
		return nil, errors.New("token has expired")
	}

	return claims, nil
}

// ExtractClaims extracts claims from a token without validation (use with caution)
func (v *Validator) ExtractClaims(tokenString string) (*Claims, error) {
	parser := jwt.NewParser(jwt.WithoutClaimsValidation())
	token, _, err := parser.ParseUnverified(tokenString, &Claims{})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)
	if !ok {
		return nil, errors.New("invalid claims type")
	}

	return claims, nil
}

// IsTokenExpired checks if a token is expired without full validation
func (v *Validator) IsTokenExpired(tokenString string) bool {
	claims, err := v.ExtractClaims(tokenString)
	if err != nil {
		return true
	}

	if claims.ExpiresAt == 0 {
		return false
	}

	return time.Now().Unix() > claims.ExpiresAt
}

// GetTokenRemainingTime returns the remaining validity time of a token
func (v *Validator) GetTokenRemainingTime(tokenString string) (time.Duration, error) {
	claims, err := v.ValidateToken(tokenString)
	if err != nil {
		return 0, err
	}

	if claims.ExpiresAt == 0 {
		return 0, errors.New("token has no expiration")
	}

	remaining := time.Until(time.Unix(claims.ExpiresAt, 0))
	if remaining <= 0 {
		return 0, errors.New("token has expired")
	}

	return remaining, nil
}
