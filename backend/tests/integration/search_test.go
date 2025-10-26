package integration

import (
	"bytes"
	"chat-ecommerce-backend/internal/dto"
	searchhandlers "chat-ecommerce-backend/internal/handlers/search"
	"chat-ecommerce-backend/internal/models"
	searchservices "chat-ecommerce-backend/internal/services/search"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type SearchIntegrationTestSuite struct {
	suite.Suite
	db             *gorm.DB
	router         *gin.Engine
	searchService  *searchservices.Service
	searchHandler  *searchhandlers.Handler
	testUserID     uuid.UUID
	testCategoryID uuid.UUID
	testProductIDs []uuid.UUID
}

func (suite *SearchIntegrationTestSuite) SetupSuite() {
	// Setup in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		suite.T().Fatal("Failed to connect to test database:", err)
	}

	// Auto-migrate the database
	err = db.AutoMigrate(
		&models.User{},
		&models.Category{},
		&models.Product{},
		&models.ProductVariant{},
		&models.ProductImage{},
		&models.Inventory{},
	)
	if err != nil {
		suite.T().Fatal("Failed to migrate test database:", err)
	}

	suite.db = db
	suite.searchService = searchservices.NewService(db)
	suite.searchHandler = searchhandlers.NewHandler(suite.searchService)

	// Setup test data
	suite.setupTestData()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	suite.setupRoutes()
}

func (suite *SearchIntegrationTestSuite) setupTestData() {
	// Create test user
	testUser := &models.User{
		ID:        uuid.New(),
		Email:     "test@example.com",
		FirstName: "Test",
		LastName:  "User",
	}
	suite.db.Create(testUser)
	suite.testUserID = testUser.ID

	// Create test category
	category := &models.Category{
		ID:          uuid.New(),
		Name:        "Electronics",
		Description: "Electronic products",
		Slug:        "electronics",
	}
	suite.db.Create(category)
	suite.testCategoryID = category.ID

	// Create test products
	products := []models.Product{
		{
			ID:          uuid.New(),
			Name:        "Gaming Laptop Pro",
			Description: "High-performance gaming laptop with RTX graphics",
			Price:       1299.99,
			CategoryID:  category.ID,
			SKU:         "LAP001",
			Status:      "active",
		},
		{
			ID:          uuid.New(),
			Name:        "Wireless Mouse",
			Description: "Ergonomic wireless mouse for gaming",
			Price:       29.99,
			CategoryID:  category.ID,
			SKU:         "MOU001",
			Status:      "active",
		},
		{
			ID:          uuid.New(),
			Name:        "Mechanical Keyboard RGB",
			Description: "RGB mechanical keyboard with Cherry MX switches",
			Price:       99.99,
			CategoryID:  category.ID,
			SKU:         "KEY001",
			Status:      "active",
		},
		{
			ID:          uuid.New(),
			Name:        "Noise Cancelling Headphones",
			Description: "Premium headphones with active noise cancellation",
			Price:       249.99,
			CategoryID:  category.ID,
			SKU:         "AUD001",
			Status:      "active",
		},
		{
			ID:          uuid.New(),
			Name:        "USB-C Cable",
			Description: "Durable USB-C charging cable",
			Price:       19.99,
			CategoryID:  category.ID,
			SKU:         "CBL001",
			Status:      "active",
		},
	}

	for _, product := range products {
		suite.db.Create(&product)
		suite.testProductIDs = append(suite.testProductIDs, product.ID)
	}
}

func (suite *SearchIntegrationTestSuite) setupRoutes() {
	// Setup search routes
	api := suite.router.Group("/api")
	{
		search := api.Group("/search")
		{
			search.POST("", suite.searchHandler.SearchProducts)
			search.GET("/suggestions", suite.searchHandler.GetSuggestions)
			search.GET("/filters", suite.searchHandler.GetFilterOptions)
			search.POST("/analytics", suite.searchHandler.LogAnalytics)
		}
	}
}

func (suite *SearchIntegrationTestSuite) TestSearch_SuccessfulSearch() {
	reqBody := dto.SearchRequestDTO{
		Query:    "laptop",
		Page:     1,
		PageSize: 10,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/search", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response dto.SearchResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response.Results)
	assert.Greater(suite.T(), len(response.Results), 0)
	assert.Equal(suite.T(), "laptop", response.Query)
}

func (suite *SearchIntegrationTestSuite) TestSearch_NoResults() {
	reqBody := dto.SearchRequestDTO{
		Query:    "nonexistentproductxyz123",
		Page:     1,
		PageSize: 10,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/search", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response dto.SearchResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), 0, response.TotalResults)
}

func (suite *SearchIntegrationTestSuite) TestSearch_WithFilters() {
	reqBody := dto.SearchRequestDTO{
		Query: "laptop",
		Filters: map[string]interface{}{
			"price_min": 1000,
			"price_max": 1500,
		},
		Page:     1,
		PageSize: 10,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/search", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response dto.SearchResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	// Results should be within price range
	if len(response.Results) > 0 {
		for _, result := range response.Results {
			assert.GreaterOrEqual(suite.T(), result.Price, 1000.0)
			assert.LessOrEqual(suite.T(), result.Price, 1500.0)
		}
	}
}

func (suite *SearchIntegrationTestSuite) TestSearch_Pagination() {
	reqBody := dto.SearchRequestDTO{
		Query:    "electronics",
		Page:     1,
		PageSize: 2,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/api/search", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response dto.SearchResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.LessOrEqual(suite.T(), len(response.Results), 2)
	assert.Equal(suite.T(), 2, response.Pagination.PageSize)
}

func (suite *SearchIntegrationTestSuite) TestSearch_GetSuggestions() {
	req := httptest.NewRequest(http.MethodGet, "/api/search/suggestions?q=laptop&limit=5", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response dto.SuggestionsResponseDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response.Suggestions)
}

func (suite *SearchIntegrationTestSuite) TestSearch_GetFilterOptions() {
	req := httptest.NewRequest(http.MethodGet, "/api/search/filters", nil)
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response dto.FilterOptionsDTO
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.NotNil(suite.T(), response.PriceRanges)
	assert.NotNil(suite.T(), response.Categories)
}

func (suite *SearchIntegrationTestSuite) TestSearch_LogAnalytics() {
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

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)
}

func (suite *SearchIntegrationTestSuite) TestSearch_InvalidRequest() {
	req := httptest.NewRequest(http.MethodPost, "/api/search", bytes.NewBufferString("invalid json"))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()

	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)
}

func TestSearchIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(SearchIntegrationTestSuite))
}
