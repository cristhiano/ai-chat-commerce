package search

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// SearchCache represents cached search results
type SearchCache struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Query       string         `gorm:"size:500;not null;index"`
	Filters     datatypes.JSON `gorm:"type:jsonb"`
	Results     datatypes.JSON `gorm:"type:jsonb;not null"`
	ResultCount int            `gorm:"not null"`
	CreatedAt   time.Time
	ExpiresAt   time.Time `gorm:"not null;index"`
}

// TableName returns the table name for SearchCache
func (SearchCache) TableName() string {
	return "search_cache"
}

// SearchSuggestion represents autocomplete suggestions
type SearchSuggestion struct {
	ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Query       string    `gorm:"size:200;not null;index"`
	Suggestions string    `gorm:"type:text;not null"`
	Occurrences int       `gorm:"default:1;index"`
	LastUsed    time.Time
	CreatedAt   time.Time
}

// TableName returns the table name for SearchSuggestion
func (SearchSuggestion) TableName() string {
	return "search_suggestions"
}

// SearchAnalytics represents search query analytics
type SearchAnalytics struct {
	ID               uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
	Query            string         `gorm:"size:500;not null;index"`
	Filters          datatypes.JSON `gorm:"type:jsonb"`
	ResultCount      int            `gorm:"not null"`
	SelectedProducts string         `gorm:"type:text"`
	ResponseTime     int            `gorm:"not null"` // milliseconds
	CacheHit         bool           `gorm:"default:false"`
	UserID           *uuid.UUID     `gorm:"type:uuid;index"`
	SessionID        string         `gorm:"size:100;index"`
	CreatedAt        time.Time
}

// TableName returns the table name for SearchAnalytics
func (SearchAnalytics) TableName() string {
	return "search_analytics"
}
