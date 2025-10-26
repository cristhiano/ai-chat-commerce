package auth

import (
	"time"

	"github.com/google/uuid"
)

// PasswordResetToken represents a one-time use token for password reset
type PasswordResetToken struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
	Token     string    `gorm:"size:128;uniqueIndex;not null"`
	ExpiresAt time.Time `gorm:"index;not null"`
	Used      bool      `gorm:"default:false"`
	CreatedAt time.Time
}

// TableName specifies the table name for the PasswordResetToken model
func (PasswordResetToken) TableName() string {
	return "password_reset_tokens"
}

// IsExpired checks if the token has expired
func (p *PasswordResetToken) IsExpired() bool {
	return time.Now().After(p.ExpiresAt)
}

// IsValid checks if the token is valid (not used and not expired)
func (p *PasswordResetToken) IsValid() bool {
	return !p.Used && !p.IsExpired()
}

// MarkAsUsed marks the token as used
func (p *PasswordResetToken) MarkAsUsed() {
	p.Used = true
}

// GetRemainingMinutes returns the remaining time until expiration in minutes
func (p *PasswordResetToken) GetRemainingMinutes() int {
	remaining := time.Until(p.ExpiresAt)
	if remaining <= 0 {
		return 0
	}
	return int(remaining.Minutes())
}
