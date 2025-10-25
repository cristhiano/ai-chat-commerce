package contracts

import (
	"chat-ecommerce-backend/internal/handlers"
	"chat-ecommerce-backend/internal/models"
	"chat-ecommerce-backend/internal/services"
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

type ProductAPIContractTestSuite struct {
	suite.Suite
	db             *gorm.DB
	router         *gin.Engine
	productService *services.ProductService
	testCategory   *models.Category
	testProduct    *models.Product
}

func (suite *ProductAPIContractTestSuite) SetupSuite() {
	// Setup in-memory SQLite database for testing
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		suite.T().Fatal("Failed to connect to test database:", err)
	}

	// Auto-migrate the database
	err = db.AutoMigrate(
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
	suite.productService = services.NewProductService(db)

	// Setup test data
	suite.setupTestData()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	suite.setupRoutes()
}

func (suite *ProductAPIContractTestSuite) setupTestData() {
	// Create test category
	suite.testCategory = &models.Category{
		ID:   uuid.New(),
		Name: "Electronics",
		Slug: "electronics",
	}
	suite.db.Create(suite.testCategory)

	// Create test product
	suite.testProduct = &models.Product{
		ID:          uuid.New(),
		Name:        "Test Product",
		Description: "A test product for API contract testing",
		Price:       99.99,
		CategoryID:  suite.testCategory.ID,
		SKU:         "TEST-001",
		Status:      "active",
	}
	suite.db.Create(suite.testProduct)

	// Create product variant
	variant := &models.ProductVariant{
		ID:            uuid.New(),
		ProductID:     suite.testProduct.ID,
		VariantName:   "Color",
		VariantValue:  "Red",
		PriceModifier: 10.0,
	}
	suite.db.Create(variant)

	// Create product image
	image := &models.ProductImage{
		ID:        uuid.New(),
		ProductID: suite.testProduct.ID,
		URL:       "https://example.com/image.jpg",
		AltText:   "Test Product Image",
		SortOrder: 1,
	}
	suite.db.Create(image)

	// Create inventory
	inventory := &models.Inventory{
		ID:        uuid.New(),
		ProductID: suite.testProduct.ID,
		Quantity:  100,
	}
	suite.db.Create(inventory)
}

func (suite *ProductAPIContractTestSuite) setupRoutes() {
	productHandler := handlers.NewProductHandler(suite.productService)

	api := suite.router.Group("/api/v1")
	{
		// Public product routes
		products := api.Group("/products")
		{
			products.GET("/", productHandler.GetProducts)
			products.GET("/:id", productHandler.GetProductByID)
			products.GET("/sku/:sku", productHandler.GetProductBySKU)
			products.GET("/search", productHandler.SearchProducts)
			products.GET("/featured", productHandler.GetFeaturedProducts)
			products.GET("/:id/related", productHandler.GetRelatedProducts)
		}

		// Public category routes
		categories := api.Group("/categories")
		{
			categories.GET("/", productHandler.GetCategories)
			categories.GET("/:id", productHandler.GetCategoryByID)
			categories.GET("/slug/:slug", productHandler.GetCategoryBySlug)
		}
	}
}

func (suite *ProductAPIContractTestSuite) TearDownSuite() {
	sqlDB, _ := suite.db.DB()
	sqlDB.Close()
}

// Test GET /api/v1/products - Get all products
func (suite *ProductAPIContractTestSuite) TestGetProducts() {
	req, _ := http.NewRequest("GET", "/api/v1/products", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	// Verify response structure
	assert.True(suite.T(), response["success"].(bool))

	data := response["data"].(map[string]interface{})
	assert.Contains(suite.T(), data, "products")
	assert.Contains(suite.T(), data, "total")
	assert.Contains(suite.T(), data, "page")
	assert.Contains(suite.T(), data, "limit")
	assert.Contains(suite.T(), data, "total_pages")

	products := data["products"].([]interface{})
	assert.GreaterOrEqual(suite.T(), len(products), 1)

	// Verify product structure
	if len(products) > 0 {
		product := products[0].(map[string]interface{})
		assert.Contains(suite.T(), product, "id")
		assert.Contains(suite.T(), product, "name")
		assert.Contains(suite.T(), product, "description")
		assert.Contains(suite.T(), product, "price")
		assert.Contains(suite.T(), product, "sku")
		assert.Contains(suite.T(), product, "status")
		assert.Contains(suite.T(), product, "created_at")
		assert.Contains(suite.T(), product, "updated_at")
	}
}

// Test GET /api/v1/products with query parameters
func (suite *ProductAPIContractTestSuite) TestGetProductsWithFilters() {
	req, _ := http.NewRequest("GET", "/api/v1/products?search=test&category_id="+suite.testCategory.ID.String()+"&min_price=50&max_price=150&status=active&page=1&limit=10&sort_by=name&sort_order=asc", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
}

// Test GET /api/v1/products/:id - Get product by ID
func (suite *ProductAPIContractTestSuite) TestGetProductByID() {
	req, _ := http.NewRequest("GET", "/api/v1/products/"+suite.testProduct.ID.String(), nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)

	// Verify response structure
	assert.True(suite.T(), response["success"].(bool))

	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), suite.testProduct.ID.String(), data["id"])
	assert.Equal(suite.T(), suite.testProduct.Name, data["name"])
	assert.Equal(suite.T(), suite.testProduct.Description, data["description"])
	assert.Equal(suite.T(), suite.testProduct.Price, data["price"])
	assert.Equal(suite.T(), suite.testProduct.SKU, data["sku"])
	assert.Equal(suite.T(), suite.testProduct.Status, data["status"])

	// Verify related data is included
	assert.Contains(suite.T(), data, "category")
	assert.Contains(suite.T(), data, "variants")
	assert.Contains(suite.T(), data, "images")
	assert.Contains(suite.T(), data, "inventory")
}

// Test GET /api/v1/products/:id with invalid ID
func (suite *ProductAPIContractTestSuite) TestGetProductByIDNotFound() {
	invalidID := uuid.New()
	req, _ := http.NewRequest("GET", "/api/v1/products/"+invalidID.String(), nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response["success"].(bool))
	assert.Contains(suite.T(), response["message"], "not found")
}

// Test GET /api/v1/products/sku/:sku - Get product by SKU
func (suite *ProductAPIContractTestSuite) TestGetProductBySKU() {
	req, _ := http.NewRequest("GET", "/api/v1/products/sku/"+suite.testProduct.SKU, nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))

	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), suite.testProduct.SKU, data["sku"])
}

// Test GET /api/v1/products/sku/:sku with invalid SKU
func (suite *ProductAPIContractTestSuite) TestGetProductBySKUNotFound() {
	req, _ := http.NewRequest("GET", "/api/v1/products/sku/INVALID-SKU", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusNotFound, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response["success"].(bool))
}

// Test GET /api/v1/products/search - Search products
func (suite *ProductAPIContractTestSuite) TestSearchProducts() {
	req, _ := http.NewRequest("GET", "/api/v1/products/search?q=test&limit=10", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))

	data := response["data"].([]interface{})
	assert.GreaterOrEqual(suite.T(), len(data), 1)
}

// Test GET /api/v1/products/featured - Get featured products
func (suite *ProductAPIContractTestSuite) TestGetFeaturedProducts() {
	req, _ := http.NewRequest("GET", "/api/v1/products/featured?limit=5", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))

	data := response["data"].([]interface{})
	assert.GreaterOrEqual(suite.T(), len(data), 1)
}

// Test GET /api/v1/products/:id/related - Get related products
func (suite *CartIntegrationTestSuite) TestGetRelatedProducts() {
	req, _ := http.NewRequest("GET", "/api/v1/products/"+suite.testProduct.ID.String()+"/related?limit=5", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))

	data := response["data"].([]interface{})
	// Should return empty array since no other products in same category
	assert.Equal(suite.T(), 0, len(data))
}

// Test GET /api/v1/categories - Get all categories
func (suite *ProductAPIContractTestSuite) TestGetCategories() {
	req, _ := http.NewRequest("GET", "/api/v1/categories", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))

	data := response["data"].([]interface{})
	assert.GreaterOrEqual(suite.T(), len(data), 1)

	// Verify category structure
	if len(data) > 0 {
		category := data[0].(map[string]interface{})
		assert.Contains(suite.T(), category, "id")
		assert.Contains(suite.T(), category, "name")
		assert.Contains(suite.T(), category, "slug")
		assert.Contains(suite.T(), category, "created_at")
		assert.Contains(suite.T(), category, "updated_at")
	}
}

// Test GET /api/v1/categories/:id - Get category by ID
func (suite *ProductAPIContractTestSuite) TestGetCategoryByID() {
	req, _ := http.NewRequest("GET", "/api/v1/categories/"+suite.testCategory.ID.String(), nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))

	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), suite.testCategory.ID.String(), data["id"])
	assert.Equal(suite.T(), suite.testCategory.Name, data["name"])
	assert.Equal(suite.T(), suite.testCategory.Slug, data["slug"])
}

// Test GET /api/v1/categories/slug/:slug - Get category by slug
func (suite *ProductAPIContractTestSuite) TestGetCategoryBySlug() {
	req, _ := http.NewRequest("GET", "/api/v1/categories/slug/"+suite.testCategory.Slug, nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))

	data := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), suite.testCategory.Slug, data["slug"])
}

// Test error responses
func (suite *ProductAPIContractTestSuite) TestErrorResponses() {
	// Test invalid UUID format
	req, _ := http.NewRequest("GET", "/api/v1/products/invalid-uuid", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusBadRequest, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.False(suite.T(), response["success"].(bool))
	assert.Contains(suite.T(), response["message"], "invalid")
}

// Test pagination contract
func (suite *ProductAPIContractTestSuite) TestPaginationContract() {
	req, _ := http.NewRequest("GET", "/api/v1/products?page=1&limit=1", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))

	data := response["data"].(map[string]interface{})

	// Verify pagination fields
	assert.Equal(suite.T(), float64(1), data["page"])
	assert.Equal(suite.T(), float64(1), data["limit"])
	assert.Contains(suite.T(), data, "total_pages")
	assert.Contains(suite.T(), data, "has_next")
	assert.Contains(suite.T(), data, "has_previous")
}

// Run the test suite
func TestProductAPIContractTestSuite(t *testing.T) {
	suite.Run(t, new(ProductAPIContractTestSuite))
}
