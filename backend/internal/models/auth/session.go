package auth

import (
	"time"

	"github.com/google/uuid"
)

// Session represents an active user authentication session
type Session struct {
	ID           uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	UserID       uuid.UUID `gorm:"type:uuid;not null;index"`
	Token        string    `gorm:"size:512;uniqueIndex;not null"`
	DeviceInfo   string    `gorm:"type:text"` // JSON: browser, OS, IP
	LastAccessAt time.Time `gorm:"index"`
	ExpiresAt    time.Time `gorm:"index;not null"`
	CreatedAt    time.Time
}

// TableName specifies the table name for the Session model
func (Session) TableName() string {
	return "sessions"
}

// IsExpired checks if the session has expired
func (s *Session) IsExpired() bool {
	return time.Now().After(s.ExpiresAt)
}

// IsActive checks if the session is currently active
func (s *Session) IsActive() bool {
	return !s.IsExpired()
}

// GetRemainingMinutes returns the remaining session time in minutes
func (s *Session) GetRemainingMinutes() int {
	remaining := time.Until(s.ExpiresAt)
	if remaining <= 0 {
		return 0
	}
	return int(remaining.Minutes())
}

// UpdateLastAccess updates the last access time
func (s *Session) UpdateLastAccess() {
	s.LastAccessAt = time.Now()
}
