package integration

import (
	"bytes"
	"chat-ecommerce-backend/internal/handlers"
	"chat-ecommerce-backend/internal/middleware"
	"chat-ecommerce-backend/internal/models"
	"chat-ecommerce-backend/internal/services"
	"context"
	"encoding/json"
	"fmt"
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

type CartIntegrationTestSuite struct {
	suite.Suite
	db             *gorm.DB
	router         *gin.Engine
	cartService    *services.ShoppingCartService
	productService *services.ProductService
	userID         uuid.UUID
	sessionID      string
}

func (suite *CartIntegrationTestSuite) SetupSuite() {
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
		&models.ShoppingCart{},
		&models.CartItem{},
	)
	if err != nil {
		suite.T().Fatal("Failed to migrate test database:", err)
	}

	suite.db = db
	suite.cartService = services.NewShoppingCartService(db)
	suite.productService = services.NewProductService(db)

	// Setup test user and session
	suite.userID = uuid.New()
	suite.sessionID = "test-session-" + uuid.New().String()

	// Setup Gin router
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()
	suite.setupRoutes()
}

func (suite *CartIntegrationTestSuite) setupRoutes() {
	// Setup cart routes
	cartHandler := handlers.NewCartHandler(suite.cartService)

	api := suite.router.Group("/api/v1")
	{
		cart := api.Group("/cart")
		cart.Use(middleware.AuthMiddleware())
		{
			cart.GET("/", cartHandler.GetCart)
			cart.POST("/add", cartHandler.AddToCart)
			cart.PUT("/update", cartHandler.UpdateCartItem)
			cart.DELETE("/remove/:product_id", cartHandler.RemoveFromCart)
			cart.DELETE("/clear", cartHandler.ClearCart)
			cart.POST("/calculate", cartHandler.CalculateTotals)
			cart.GET("/count", cartHandler.GetCartItemCount)
		}
	}
}

func (suite *CartIntegrationTestSuite) SetupTest() {
	// Clean up cart items before each test
	suite.db.Where("user_id = ? OR session_id = ?", suite.userID, suite.sessionID).Delete(&models.CartItem{})
	suite.db.Where("user_id = ? OR session_id = ?", suite.userID, suite.sessionID).Delete(&models.ShoppingCart{})
}

func (suite *CartIntegrationTestSuite) TearDownSuite() {
	sqlDB, _ := suite.db.DB()
	sqlDB.Close()
}

func (suite *CartIntegrationTestSuite) createTestProduct() *models.Product {
	category := &models.Category{
		ID:   uuid.New(),
		Name: "Test Category",
		Slug: "test-category",
	}
	suite.db.Create(category)

	product := &models.Product{
		ID:          uuid.New(),
		Name:        "Test Product",
		Description: "Test Product Description",
		Price:       99.99,
		CategoryID:  category.ID,
		SKU:         "TEST-001",
		Status:      "active",
	}
	suite.db.Create(product)

	// Create inventory
	inventory := &models.Inventory{
		ID:        uuid.New(),
		ProductID: product.ID,
		Quantity:  100,
	}
	suite.db.Create(inventory)

	return product
}

func (suite *CartIntegrationTestSuite) TestAddToCart() {
	product := suite.createTestProduct()

	// Test data
	addToCartRequest := services.AddToCartRequest{
		ProductID: product.ID,
		Quantity:  2,
	}

	jsonData, _ := json.Marshal(addToCartRequest)

	// Create request
	req, _ := http.NewRequest("POST", "/api/v1/cart/add", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	// Mock user context
	req = req.WithContext(context.WithValue(req.Context(), "user_id", suite.userID))

	// Execute request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))

	// Verify cart item was created
	var cartItem models.CartItem
	err = suite.db.Where("user_id = ? AND product_id = ?", suite.userID, product.ID).First(&cartItem).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(2), cartItem.Quantity)
	assert.Equal(suite.T(), product.Price, cartItem.UnitPrice)
}

func (suite *CartIntegrationTestSuite) TestGetCart() {
	product := suite.createTestProduct()

	// Add item to cart
	cartItem := &models.CartItem{
		ID:        uuid.New(),
		UserID:    &suite.userID,
		ProductID: product.ID,
		Quantity:  3,
		UnitPrice: product.Price,
	}
	suite.db.Create(cartItem)

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/cart/", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Mock user context
	req = req.WithContext(context.WithValue(req.Context(), "user_id", suite.userID))

	// Execute request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))

	cartData := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), float64(1), cartData["item_count"])
	assert.Equal(suite.T(), float64(3), cartData["total_quantity"])
}

func (suite *CartIntegrationTestSuite) TestUpdateCartItem() {
	product := suite.createTestProduct()

	// Add item to cart
	cartItem := &models.CartItem{
		ID:        uuid.New(),
		UserID:    &suite.userID,
		ProductID: product.ID,
		Quantity:  2,
		UnitPrice: product.Price,
	}
	suite.db.Create(cartItem)

	// Test data
	updateRequest := services.UpdateCartItemRequest{
		ProductID: product.ID,
		Quantity:  5,
	}

	jsonData, _ := json.Marshal(updateRequest)

	// Create request
	req, _ := http.NewRequest("PUT", "/api/v1/cart/update", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	// Mock user context
	req = req.WithContext(context.WithValue(req.Context(), "user_id", suite.userID))

	// Execute request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))

	// Verify cart item was updated
	var updatedCartItem models.CartItem
	err = suite.db.Where("user_id = ? AND product_id = ?", suite.userID, product.ID).First(&updatedCartItem).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(5), updatedCartItem.Quantity)
}

func (suite *CartIntegrationTestSuite) TestRemoveFromCart() {
	product := suite.createTestProduct()

	// Add item to cart
	cartItem := &models.CartItem{
		ID:        uuid.New(),
		UserID:    &suite.userID,
		ProductID: product.ID,
		Quantity:  2,
		UnitPrice: product.Price,
	}
	suite.db.Create(cartItem)

	// Create request
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("/api/v1/cart/remove/%s", product.ID), nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Mock user context
	req = req.WithContext(context.WithValue(req.Context(), "user_id", suite.userID))

	// Execute request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))

	// Verify cart item was removed
	var count int64
	err = suite.db.Model(&models.CartItem{}).Where("user_id = ? AND product_id = ?", suite.userID, product.ID).Count(&count).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(0), count)
}

func (suite *CartIntegrationTestSuite) TestClearCart() {
	product1 := suite.createTestProduct()
	product2 := suite.createTestProduct()

	// Add multiple items to cart
	cartItems := []models.CartItem{
		{
			ID:        uuid.New(),
			UserID:    &suite.userID,
			ProductID: product1.ID,
			Quantity:  2,
			UnitPrice: product1.Price,
		},
		{
			ID:        uuid.New(),
			UserID:    &suite.userID,
			ProductID: product2.ID,
			Quantity:  1,
			UnitPrice: product2.Price,
		},
	}
	suite.db.Create(&cartItems)

	// Create request
	req, _ := http.NewRequest("DELETE", "/api/v1/cart/clear", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Mock user context
	req = req.WithContext(context.WithValue(req.Context(), "user_id", suite.userID))

	// Execute request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))

	// Verify all cart items were removed
	var count int64
	err = suite.db.Model(&models.CartItem{}).Where("user_id = ?", suite.userID).Count(&count).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(0), count)
}

func (suite *CartIntegrationTestSuite) TestCalculateTotals() {
	product1 := suite.createTestProduct()
	product2 := suite.createTestProduct()

	// Add items to cart
	cartItems := []models.CartItem{
		{
			ID:        uuid.New(),
			UserID:    &suite.userID,
			ProductID: product1.ID,
			Quantity:  2,
			UnitPrice: product1.Price,
		},
		{
			ID:        uuid.New(),
			UserID:    &suite.userID,
			ProductID: product2.ID,
			Quantity:  1,
			UnitPrice: product2.Price,
		},
	}
	suite.db.Create(&cartItems)

	// Create request
	req, _ := http.NewRequest("POST", "/api/v1/cart/calculate", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Mock user context
	req = req.WithContext(context.WithValue(req.Context(), "user_id", suite.userID))

	// Execute request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))

	cartData := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), float64(2), cartData["item_count"])
	assert.Equal(suite.T(), float64(3), cartData["total_quantity"])

	// Verify totals are calculated correctly
	expectedSubtotal := (product1.Price * 2) + (product2.Price * 1)
	assert.Equal(suite.T(), expectedSubtotal, cartData["subtotal"])
}

func (suite *CartIntegrationTestSuite) TestGetCartItemCount() {
	product1 := suite.createTestProduct()
	product2 := suite.createTestProduct()

	// Add items to cart
	cartItems := []models.CartItem{
		{
			ID:        uuid.New(),
			UserID:    &suite.userID,
			ProductID: product1.ID,
			Quantity:  2,
			UnitPrice: product1.Price,
		},
		{
			ID:        uuid.New(),
			UserID:    &suite.userID,
			ProductID: product2.ID,
			Quantity:  1,
			UnitPrice: product2.Price,
		},
	}
	suite.db.Create(&cartItems)

	// Create request
	req, _ := http.NewRequest("GET", "/api/v1/cart/count", nil)
	req.Header.Set("Authorization", "Bearer test-token")

	// Mock user context
	req = req.WithContext(context.WithValue(req.Context(), "user_id", suite.userID))

	// Execute request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))

	countData := response["data"].(map[string]interface{})
	assert.Equal(suite.T(), float64(2), countData["item_count"])
	assert.Equal(suite.T(), float64(3), countData["total_quantity"])
}

func (suite *CartIntegrationTestSuite) TestCartWithSessionID() {
	product := suite.createTestProduct()

	// Test data for anonymous user
	addToCartRequest := services.AddToCartRequest{
		ProductID: product.ID,
		Quantity:  1,
	}

	jsonData, _ := json.Marshal(addToCartRequest)

	// Create request without user ID but with session ID
	req, _ := http.NewRequest("POST", "/api/v1/cart/add", bytes.NewBuffer(jsonData))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer test-token")

	// Mock session context
	req = req.WithContext(context.WithValue(req.Context(), "session_id", suite.sessionID))

	// Execute request
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	// Assertions
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))

	// Verify cart item was created with session ID
	var cartItem models.CartItem
	err = suite.db.Where("session_id = ? AND product_id = ?", suite.sessionID, product.ID).First(&cartItem).Error
	assert.NoError(suite.T(), err)
	assert.Equal(suite.T(), int64(1), cartItem.Quantity)
}

// Run the test suite
func TestCartIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(CartIntegrationTestSuite))
}
