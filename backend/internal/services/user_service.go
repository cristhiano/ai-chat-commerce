package services

import (
	"chat-ecommerce-backend/internal/models"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// UserService handles user-related business logic
type UserService struct {
	db *gorm.DB
}

// NewUserService creates a new UserService
func NewUserService(db *gorm.DB) *UserService {
	return &UserService{
		db: db,
	}
}

// RegisterRequest represents user registration data
type RegisterRequest struct {
	Email     string `json:"email" binding:"required,email"`
	Password  string `json:"password" binding:"required,min=8"`
	FirstName string `json:"first_name" binding:"required"`
	LastName  string `json:"last_name" binding:"required"`
	Phone     string `json:"phone"`
}

// LoginRequest represents user login data
type LoginRequest struct {
	Email    string `json:"email" binding:"required,email"`
	Password string `json:"password" binding:"required"`
}

// UpdateProfileRequest represents user profile update data
type UpdateProfileRequest struct {
	FirstName   string                 `json:"first_name"`
	LastName    string                 `json:"last_name"`
	Phone       string                 `json:"phone"`
	Preferences map[string]interface{} `json:"preferences"`
}

// User represents a user in the system (alias for models.User)
type User = models.User

// Register creates a new user account
func (s *UserService) Register(req *RegisterRequest) (*User, error) {
	// Check if user already exists
	var existingUser User
	if err := s.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return nil, errors.New("user with this email already exists")
	}

	// Hash password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	// Create user
	user := &User{
		ID:            uuid.New(),
		Email:         req.Email,
		PasswordHash:  string(hashedPassword),
		FirstName:     req.FirstName,
		LastName:      req.LastName,
		Phone:         req.Phone,
		Status:        "active",
		EmailVerified: false,
		CreatedAt:     time.Now(),
		UpdatedAt:     time.Now(),
	}

	if err := s.db.Create(user).Error; err != nil {
		return nil, errors.New("failed to create user")
	}

	// Clear password hash from response
	user.PasswordHash = ""
	return user, nil
}

// Login authenticates a user and returns user data
func (s *UserService) Login(req *LoginRequest) (*User, error) {
	var user User
	if err := s.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("invalid email or password")
		}
		return nil, errors.New("failed to find user")
	}

	// Check password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
		return nil, errors.New("invalid email or password")
	}

	// Check if user is active
	if user.Status != "active" {
		return nil, errors.New("account is not active")
	}

	// Clear password hash from response
	user.PasswordHash = ""
	return &user, nil
}

// GetUserByID retrieves a user by ID
func (s *UserService) GetUserByID(userID uuid.UUID) (*User, error) {
	var user User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, errors.New("failed to retrieve user")
	}

	// Clear password hash from response
	user.PasswordHash = ""
	return &user, nil
}

// GetUserByEmail retrieves a user by email
func (s *UserService) GetUserByEmail(email string) (*User, error) {
	var user User
	if err := s.db.Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, errors.New("failed to retrieve user")
	}

	// Clear password hash from response
	user.PasswordHash = ""
	return &user, nil
}

// UpdateProfile updates user profile information
func (s *UserService) UpdateProfile(userID uuid.UUID, req *UpdateProfileRequest) (*User, error) {
	var user User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("user not found")
		}
		return nil, errors.New("failed to find user")
	}

	// Update fields
	if req.FirstName != "" {
		user.FirstName = req.FirstName
	}
	if req.LastName != "" {
		user.LastName = req.LastName
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.Preferences != nil {
		preferencesJSON, err := json.Marshal(req.Preferences)
		if err != nil {
			return nil, errors.New("failed to marshal preferences")
		}
		user.Preferences = datatypes.JSON(preferencesJSON)
	}
	user.UpdatedAt = time.Now()

	if err := s.db.Save(&user).Error; err != nil {
		return nil, errors.New("failed to update user profile")
	}

	// Clear password hash from response
	user.PasswordHash = ""
	return &user, nil
}

// ChangePassword changes user password
func (s *UserService) ChangePassword(userID uuid.UUID, currentPassword, newPassword string) error {
	var user User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return errors.New("failed to find user")
	}

	// Verify current password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(currentPassword)); err != nil {
		return errors.New("current password is incorrect")
	}

	// Hash new password
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return errors.New("failed to hash new password")
	}

	// Update password
	user.PasswordHash = string(hashedPassword)
	user.UpdatedAt = time.Now()

	if err := s.db.Save(&user).Error; err != nil {
		return errors.New("failed to update password")
	}

	return nil
}

// DeleteUser soft deletes a user account
func (s *UserService) DeleteUser(userID uuid.UUID) error {
	var user User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return errors.New("failed to find user")
	}

	// Soft delete by updating status
	user.Status = "deleted"
	user.UpdatedAt = time.Now()

	if err := s.db.Save(&user).Error; err != nil {
		return errors.New("failed to delete user")
	}

	return nil
}

// VerifyEmail marks user email as verified
func (s *UserService) VerifyEmail(userID uuid.UUID) error {
	var user User
	if err := s.db.Where("id = ?", userID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("user not found")
		}
		return errors.New("failed to find user")
	}

	user.EmailVerified = true
	user.UpdatedAt = time.Now()

	if err := s.db.Save(&user).Error; err != nil {
		return errors.New("failed to verify email")
	}

	return nil
}
