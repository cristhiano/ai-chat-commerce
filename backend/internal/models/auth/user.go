package auth

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user account for authentication
// This extends the base User model with authentication-specific fields
type User struct {
	ID                  uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Email               string     `gorm:"size:255;uniqueIndex;not null"`
	PasswordHash        string     `gorm:"type:text;not null"`
	AccountState        string     `gorm:"size:20;default:'active';index"`
	FailedLoginAttempts int        `gorm:"default:0;not null"`
	LockoutUntil        *time.Time `gorm:"index"`
	LastLoginAt         *time.Time
	CreatedAt           time.Time
	UpdatedAt           time.Time
}

// TableName specifies the table name for the User model
func (User) TableName() string {
	return "users"
}

// IsLocked checks if the user account is currently locked
func (u *User) IsLocked() bool {
	if u.LockoutUntil == nil {
		return false
	}
	return u.AccountState == "locked" && time.Now().Before(*u.LockoutUntil)
}

// IsActive checks if the user account is active
func (u *User) IsActive() bool {
	return u.AccountState == "active" && !u.IsLocked()
}

// GetLockoutRemainingMinutes returns the remaining lockout time in minutes
func (u *User) GetLockoutRemainingMinutes() int {
	if u.LockoutUntil == nil || !u.IsLocked() {
		return 0
	}

	remaining := time.Until(*u.LockoutUntil)
	if remaining <= 0 {
		return 0
	}
	return int(remaining.Minutes()) + 1 // Round up
}
