package handlers

import (
	"bytes"
	"chat-ecommerce-backend/internal/dto"
	searchhandlers "chat-ecommerce-backend/internal/handlers/search"
	searchservices "chat-ecommerce-backend/internal/services/search"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestHandler(t *testing.T) (*searchhandlers.Handler, *gorm.DB) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal("Failed to connect to test database:", err)
	}

	searchService := searchservices.NewService(db)
	handler := searchhandlers.NewHandler(searchService)

	return handler, db
}

func TestSearchHandler_SearchProducts(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Successful search request", func(t *testing.T) {
		handler, _ := setupTestHandler(t)

		reqBody := dto.SearchRequestDTO{
			Query:    "laptop",
			Page:     1,
			PageSize: 10,
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/search", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req

		handler.SearchProducts(ctx)

		assert.Equal(t, http.StatusOK, w.Code)
		var response dto.SearchResponseDTO
		json.Unmarshal(w.Body.Bytes(), &response)
		assert.NotNil(t, response)
	})

	t.Run("Invalid request - empty query", func(t *testing.T) {
		handler, _ := setupTestHandler(t)

		reqBody := dto.SearchRequestDTO{
			Query: "",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/search", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req

		handler.SearchProducts(ctx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Invalid request - malformed JSON", func(t *testing.T) {
		handler, _ := setupTestHandler(t)

		req := httptest.NewRequest(http.MethodPost, "/api/search", bytes.NewBufferString("invalid json"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req

		handler.SearchProducts(ctx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Valid search with filters", func(t *testing.T) {
		handler, _ := setupTestHandler(t)

		reqBody := dto.SearchRequestDTO{
			Query: "laptop",
			Filters: map[string]interface{}{
				"price_min": 500,
				"price_max": 2000,
			},
			Page:     1,
			PageSize: 20,
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/search", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req

		handler.SearchProducts(ctx)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestSearchHandler_GetSuggestions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Get suggestions - valid query", func(t *testing.T) {
		handler, _ := setupTestHandler(t)

		req := httptest.NewRequest(http.MethodGet, "/api/search/suggestions?q=laptop&limit=10", nil)
		w := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req

		handler.GetSuggestions(ctx)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Get suggestions - missing query", func(t *testing.T) {
		handler, _ := setupTestHandler(t)

		req := httptest.NewRequest(http.MethodGet, "/api/search/suggestions", nil)
		w := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req

		handler.GetSuggestions(ctx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})

	t.Run("Get suggestions - with limit", func(t *testing.T) {
		handler, _ := setupTestHandler(t)

		req := httptest.NewRequest(http.MethodGet, "/api/search/suggestions?q=mouse&limit=5", nil)
		w := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req

		handler.GetSuggestions(ctx)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestSearchHandler_GetFilterOptions(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Get filter options", func(t *testing.T) {
		handler, _ := setupTestHandler(t)

		req := httptest.NewRequest(http.MethodGet, "/api/search/filters", nil)
		w := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req

		handler.GetFilterOptions(ctx)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Get filter options with category", func(t *testing.T) {
		handler, _ := setupTestHandler(t)

		req := httptest.NewRequest(http.MethodGet, "/api/search/filters?category_id=test-category", nil)
		w := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req

		handler.GetFilterOptions(ctx)

		assert.Equal(t, http.StatusOK, w.Code)
	})
}

func TestSearchHandler_LogAnalytics(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("Log analytics successfully", func(t *testing.T) {
		handler, _ := setupTestHandler(t)

		reqBody := dto.AnalyticsRequestDTO{
			Query:          "laptop",
			ResultCount:    5,
			ResponseTimeMs: 100,
			CacheHit:       false,
			SessionID:      "test-session-123",
		}

		jsonBody, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/api/search/analytics", bytes.NewBuffer(jsonBody))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req

		handler.LogAnalytics(ctx)

		assert.Equal(t, http.StatusOK, w.Code)
	})

	t.Run("Log analytics - invalid request", func(t *testing.T) {
		handler, _ := setupTestHandler(t)

		req := httptest.NewRequest(http.MethodPost, "/api/search/analytics", bytes.NewBufferString("invalid"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()

		ctx, _ := gin.CreateTestContext(w)
		ctx.Request = req

		handler.LogAnalytics(ctx)

		assert.Equal(t, http.StatusBadRequest, w.Code)
	})
}
