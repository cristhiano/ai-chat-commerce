package services

import (
	"chat-ecommerce-backend/internal/models"
	"chat-ecommerce-backend/internal/services/search"
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDBWithSearch(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal("Failed to connect to test database:", err)
	}

	// Auto-migrate the database
	err = db.AutoMigrate(
		&models.Product{},
		&models.Category{},
	)
	if err != nil {
		t.Fatal("Failed to migrate test database:", err)
	}

	// Create test categories
	category := models.Category{
		ID:          uuid.New(),
		Name:        "Electronics",
		Description: "Electronic products",
		Slug:        "electronics",
	}
	db.Create(&category)

	// Create test products
	products := []models.Product{
		{
			ID:          uuid.New(),
			Name:        "Gaming Laptop",
			Description: "High-performance gaming laptop",
			Price:       1299.99,
			CategoryID:  category.ID,
			SKU:         "LAP001",
			Status:      "active",
		},
		{
			ID:          uuid.New(),
			Name:        "Wireless Mouse",
			Description: "Ergonomic wireless mouse",
			Price:       29.99,
			CategoryID:  category.ID,
			SKU:         "MOU001",
			Status:      "active",
		},
		{
			ID:          uuid.New(),
			Name:        "Mechanical Keyboard",
			Description: "RGB mechanical keyboard",
			Price:       99.99,
			CategoryID:  category.ID,
			SKU:         "KEY001",
			Status:      "active",
		},
	}

	for _, product := range products {
		db.Create(&product)
	}

	return db
}

func TestSearchService_Search(t *testing.T) {
	db := setupTestDBWithSearch(t)
	service := search.NewService(db)

	ctx := context.Background()

	t.Run("Successful search with results", func(t *testing.T) {
		req := &search.SearchRequest{
			Query:    "laptop",
			Page:     1,
			PageSize: 10,
		}

		result, err := service.Search(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Greater(t, result.TotalResults, 0)
		assert.Greater(t, len(result.Results), 0)
		assert.Equal(t, "laptop", result.Query)
	})

	t.Run("Search with no results", func(t *testing.T) {
		req := &search.SearchRequest{
			Query:    "nonexistentproductxyz",
			Page:     1,
			PageSize: 10,
		}

		result, err := service.Search(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 0, result.TotalResults)
		assert.Equal(t, 0, len(result.Results))
	})

	t.Run("Search with pagination", func(t *testing.T) {
		req := &search.SearchRequest{
			Query:    "laptop",
			Page:     1,
			PageSize: 1,
		}

		result, err := service.Search(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, 1, len(result.Results))
		assert.Equal(t, req.PageSize, result.Pagination.PageSize)
	})
}

func TestSearchService_GetSuggestions(t *testing.T) {
	db := setupTestDBWithSearch(t)
	service := search.NewService(db)

	ctx := context.Background()

	t.Run("Get suggestions with results", func(t *testing.T) {
		suggestions, err := service.GetSuggestions(ctx, "lap", 10)
		assert.NoError(t, err)
		assert.NotNil(t, suggestions)
		assert.Greater(t, len(suggestions), 0)
	})

	t.Run("Get suggestions with no results", func(t *testing.T) {
		suggestions, err := service.GetSuggestions(ctx, "xyzabc123", 10)
		assert.NoError(t, err)
		assert.NotNil(t, suggestions)
		assert.Equal(t, 0, len(suggestions))
	})

	t.Run("Limit suggestions count", func(t *testing.T) {
		suggestions, err := service.GetSuggestions(ctx, "laptop", 5)
		assert.NoError(t, err)
		assert.NotNil(t, suggestions)
		assert.LessOrEqual(t, len(suggestions), 5)
	})
}

func TestSearchService_GetFilterOptions(t *testing.T) {
	db := setupTestDBWithSearch(t)
	service := search.NewService(db)

	ctx := context.Background()

	t.Run("Get filter options", func(t *testing.T) {
		options, err := service.GetFilterOptions(ctx, "")
		assert.NoError(t, err)
		assert.NotNil(t, options)
		assert.NotNil(t, options.Categories)
	})

	t.Run("Get filter options for specific category", func(t *testing.T) {
		// Get first category
		var category models.Category
		db.First(&category)

		options, err := service.GetFilterOptions(ctx, category.ID.String())
		assert.NoError(t, err)
		assert.NotNil(t, options)
	})
}

func TestSearchService_LogAnalytics(t *testing.T) {
	db := setupTestDBWithSearch(t)
	service := search.NewService(db)

	ctx := context.Background()

	t.Run("Log analytics successfully", func(t *testing.T) {
		req := &search.AnalyticsRequest{
			Query:            "laptop",
			ResultCount:      5,
			ResponseTimeMs:   100,
			CacheHit:         false,
			SessionID:        "test-session-123",
			SelectedProducts: "",
		}

		err := service.LogAnalytics(ctx, req)
		assert.NoError(t, err)
	})

	t.Run("Log analytics with user ID", func(t *testing.T) {
		userID := uuid.New()
		req := &search.AnalyticsRequest{
			Query:            "mouse",
			ResultCount:      3,
			ResponseTimeMs:   50,
			CacheHit:         true,
			SessionID:        "test-session-456",
			UserID:           userID.String(),
			SelectedProducts: "",
		}

		err := service.LogAnalytics(ctx, req)
		assert.NoError(t, err)
	})
}

func TestSearchService_ResponseTime(t *testing.T) {
	db := setupTestDBWithSearch(t)
	service := search.NewService(db)

	ctx := context.Background()

	t.Run("Search response time should be reasonable", func(t *testing.T) {
		req := &search.SearchRequest{
			Query:    "laptop",
			Page:     1,
			PageSize: 10,
		}

		start := time.Now()
		result, err := service.Search(ctx, req)
		duration := time.Since(start)

		assert.NoError(t, err)
		assert.NotNil(t, result)
		// Response should be within 1 second (target from requirements)
		assert.Less(t, duration, 1*time.Second)
	})
}

func TestSearchService_FuzzyMatching(t *testing.T) {
	db := setupTestDBWithSearch(t)
	service := search.NewService(db)

	ctx := context.Background()

	t.Run("Handle typos gracefully", func(t *testing.T) {
		req := &search.SearchRequest{
			Query:    "lapotp", // Typo for "laptop"
			Page:     1,
			PageSize: 10,
		}

		result, err := service.Search(ctx, req)
		assert.NoError(t, err)
		assert.NotNil(t, result)
		// Should still find relevant results
		assert.Greater(t, result.TotalResults, 0)
	})
}
