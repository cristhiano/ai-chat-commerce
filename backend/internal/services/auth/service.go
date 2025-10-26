package auth

import (
	"errors"
	"fmt"
	"time"

	"chat-ecommerce-backend/internal/models"
	"chat-ecommerce-backend/internal/models/auth"
	"chat-ecommerce-backend/internal/services/password"
	"chat-ecommerce-backend/internal/services/session"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Service handles core authentication operations
type Service struct {
	db                *gorm.DB
	passwordService   *password.Service
	passwordValidator *password.Validator
	sessionService    *session.Service
	sessionValidator  *session.Validator
}

// NewService creates a new authentication service
func NewService(
	db *gorm.DB,
	passwordService *password.Service,
	passwordValidator *password.Validator,
	sessionService *session.Service,
	sessionValidator *session.Validator,
) *Service {
	return &Service{
		db:                db,
		passwordService:   passwordService,
		passwordValidator: passwordValidator,
		sessionService:    sessionService,
		sessionValidator:  sessionValidator,
	}
}

// RegisterRequest represents registration input
type RegisterRequest struct {
	Email    string
	Password string
}

// RegisterResponse represents registration output
type RegisterResponse struct {
	UserID  uuid.UUID
	Email   string
	Message string
}

// Register creates a new user account
func (s *Service) Register(req *RegisterRequest) (*RegisterResponse, error) {
	// Validate password strength
	validation := s.passwordValidator.ValidatePassword(req.Password)
	if !validation.IsValid {
		return nil, fmt.Errorf("password validation failed: %v", validation.Errors)
	}

	// Check if email already exists
	var existingUser models.User
	if err := s.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, errors.New("email already exists")
	}

	// Hash password
	passwordHash, err := s.passwordService.HashPassword(req.Password)
	if err != nil {
		return nil, fmt.Errorf("failed to hash password: %w", err)
	}

	// Create user
	user := &models.User{
		ID:                  uuid.New(),
		Email:               req.Email,
		PasswordHash:        passwordHash,
		AccountState:        "active",
		FailedLoginAttempts: 0,
		FirstName:           "",
		LastName:            "",
		CreatedAt:           time.Now(),
		UpdatedAt:           time.Now(),
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, fmt.Errorf("failed to create user: %w", err)
	}

	return &RegisterResponse{
		UserID:  user.ID,
		Email:   user.Email,
		Message: "Account created successfully",
	}, nil
}

// LoginRequest represents login input
type LoginRequest struct {
	Email    string
	Password string
}

// LoginResponse represents login output
type LoginResponse struct {
	Token     string
	UserID    uuid.UUID
	Email     string
	ExpiresAt time.Time
}

// Login authenticates a user and creates a session
func (s *Service) Login(req *LoginRequest) (*LoginResponse, error) {
	// Find user by email
	var user models.User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, fmt.Errorf("failed to find user: %w", err)
	}

	// Check if account is locked
	if user.IsLocked() {
		return nil, fmt.Errorf("account is locked, try again in %d minutes", user.GetLockoutRemainingMinutes())
	}

	// Verify password
	if !s.passwordService.VerifyPassword(user.PasswordHash, req.Password) {
		// Increment failed login attempts
		user.FailedLoginAttempts++

		// Lock account after 5 failed attempts
		if user.FailedLoginAttempts >= 5 {
			lockoutUntil := time.Now().Add(15 * time.Minute)
			user.LockoutUntil = &lockoutUntil
			user.AccountState = "locked"
		}

		s.db.Save(&user)
		return nil, errors.New("invalid email or password")
	}

	// Reset failed login attempts on successful login
	user.FailedLoginAttempts = 0
	user.LockoutUntil = nil
	user.AccountState = "active"
	now := time.Now()
	user.LastLoginAt = &now

	if err := s.db.Save(&user).Error; err != nil {
		return nil, fmt.Errorf("failed to update user: %w", err)
	}

	// Generate session token
	token, err := s.sessionService.GenerateToken(user.ID.String(), user.Email)
	if err != nil {
		return nil, fmt.Errorf("failed to generate token: %w", err)
	}

	// Calculate expiration
	expiresAt := time.Now().Add(s.sessionService.GetTokenExpiration())

	// Store session in database
	sessionRecord := &auth.Session{
		ID:           uuid.New(),
		UserID:       user.ID,
		Token:        token,
		LastAccessAt: time.Now(),
		ExpiresAt:    expiresAt,
		CreatedAt:    time.Now(),
	}

	if err := s.db.Create(sessionRecord).Error; err != nil {
		return nil, fmt.Errorf("failed to create session: %w", err)
	}

	return &LoginResponse{
		Token:     token,
		UserID:    user.ID,
		Email:     user.Email,
		ExpiresAt: expiresAt,
	}, nil
}

// LogoutRequest represents logout input
type LogoutRequest struct {
	Token string
}

// LogoutResponse represents logout output
type LogoutResponse struct {
	Message string
}

// Logout terminates a user session
func (s *Service) Logout(token string) (*LogoutResponse, error) {
	// Delete session from database
	if err := s.db.Where("token = ?", token).Delete(&auth.Session{}).Error; err != nil {
		return nil, fmt.Errorf("failed to delete session: %w", err)
	}

	return &LogoutResponse{
		Message: "Logged out successfully",
	}, nil
}
