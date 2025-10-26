package services

import (
	"chat-ecommerce-backend/internal/models"
	"chat-ecommerce-backend/internal/services"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal("Failed to connect to test database:", err)
	}

	// Auto-migrate the database
	err = db.AutoMigrate(
		&models.ChatSession{},
		&models.ChatMessage{},
	)
	if err != nil {
		t.Fatal("Failed to migrate test database:", err)
	}

	return db
}

func TestChatService_GetChatSession(t *testing.T) {
	db := setupTestDB(t)

	// Create mock services
	productService := services.NewProductService(db)
	cartService := services.NewShoppingCartService(db)
	service := services.NewChatService(db, productService, cartService)

	sessionID := "test-session-123"
	userID := uuid.New()

	// Test creating new session
	session, err := service.GetChatSession(sessionID, &userID)
	assert.NoError(t, err)
	assert.NotNil(t, session)
	assert.Equal(t, sessionID, session.ID)
	assert.Equal(t, &userID, session.UserID)

	// Test retrieving existing session
	session2, err := service.GetChatSession(sessionID, &userID)
	assert.NoError(t, err)
	assert.NotNil(t, session2)
	assert.Equal(t, sessionID, session2.ID)
}

func TestChatService_SearchProducts(t *testing.T) {
	db := setupTestDB(t)

	// Create mock services
	productService := services.NewProductService(db)
	cartService := services.NewShoppingCartService(db)
	service := services.NewChatService(db, productService, cartService)

	// Test search products
	suggestions, err := service.SearchProducts("test", 5)
	assert.NoError(t, err)
	assert.NotNil(t, suggestions)
	// Should return empty array since no products in test DB
	assert.Len(t, suggestions, 0)
}

func TestChatService_GetProductRecommendations(t *testing.T) {
	db := setupTestDB(t)

	// Create mock services
	productService := services.NewProductService(db)
	cartService := services.NewShoppingCartService(db)
	service := services.NewChatService(db, productService, cartService)

	sessionID := "test-session"
	userID := uuid.New()

	// Test getting recommendations
	suggestions, err := service.GetProductRecommendations(sessionID, &userID, 5)
	assert.NoError(t, err)
	assert.NotNil(t, suggestions)
	// Should return empty array since no products in test DB
	assert.Len(t, suggestions, 0)
}

func TestChatService_GetConversationHistory(t *testing.T) {
	db := setupTestDB(t)

	// Create mock services
	productService := services.NewProductService(db)
	cartService := services.NewShoppingCartService(db)
	service := services.NewChatService(db, productService, cartService)

	sessionID := "test-session-history"

	// Get conversation history (should be empty)
	history, err := service.GetConversationHistory(sessionID, 10)
	assert.NoError(t, err)
	assert.NotNil(t, history)
	assert.Len(t, history, 0)
}

func TestProductSuggestion_JSONSerialization(t *testing.T) {
	// Setup test database with products
	db := setupTestDB(t)

	// Migrate Product, Category, and Inventory models
	err := db.AutoMigrate(&models.Product{}, &models.Category{}, &models.Inventory{})
	assert.NoError(t, err)

	// Create a test category
	categoryID := uuid.New()
	category := models.Category{
		ID:          categoryID,
		Name:        "Electronics",
		Description: "Electronic products",
		Slug:        "electronics",
		SortOrder:   1,
		IsActive:    true,
	}
	err = db.Create(&category).Error
	assert.NoError(t, err)

	// Create a test product with complete data
	productID := uuid.New()
	product := models.Product{
		ID:          productID,
		Name:        "Wireless Headphones",
		Description: "High-quality wireless headphones with noise cancellation",
		Price:       199.99,
		CategoryID:  categoryID,
		SKU:         "WH-001",
		Status:      "active",
	}
	err = db.Create(&product).Error
	assert.NoError(t, err)

	// Create inventory for the product
	inventory := models.Inventory{
		ID:                uuid.New(),
		ProductID:         productID,
		WarehouseLocation: "Warehouse A",
		QuantityAvailable: 100,
		QuantityReserved:  10,
		LowStockThreshold: 20,
		ReorderPoint:      30,
	}
	err = db.Create(&inventory).Error
	assert.NoError(t, err)

	// Load product with relationships
	var loadedProduct models.Product
	err = db.Preload("Category").Preload("Inventory").First(&loadedProduct, productID).Error
	assert.NoError(t, err)

	// Create ProductSuggestion
	suggestion := services.ProductSuggestion{
		Product:    &loadedProduct,
		Reason:     "Matches your search query",
		Confidence: 0.95,
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(suggestion)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Deserialize back to verify structure
	var deserialized map[string]interface{}
	err = json.Unmarshal(jsonData, &deserialized)
	assert.NoError(t, err)

	// Verify all fields are present
	assert.Contains(t, deserialized, "product")
	assert.Contains(t, deserialized, "reason")
	assert.Contains(t, deserialized, "confidence")

	// Verify product fields
	productData := deserialized["product"].(map[string]interface{})
	assert.Contains(t, productData, "id")
	assert.Contains(t, productData, "name")
	assert.Contains(t, productData, "description")
	assert.Contains(t, productData, "price")
	assert.Contains(t, productData, "category_id")
	assert.Contains(t, productData, "sku")
	assert.Contains(t, productData, "status")

	// Verify category relationship is loaded
	assert.Contains(t, productData, "category")
	categoryData := productData["category"].(map[string]interface{})
	assert.Contains(t, categoryData, "id")
	assert.Contains(t, categoryData, "name")
	assert.Equal(t, "Electronics", categoryData["name"])

	// Verify inventory relationship is loaded
	assert.Contains(t, productData, "inventory")
	inventoryData := productData["inventory"].([]interface{})
	assert.Len(t, inventoryData, 1)
	firstInventory := inventoryData[0].(map[string]interface{})
	assert.Contains(t, firstInventory, "quantity_available")
	assert.Equal(t, float64(100), firstInventory["quantity_available"])

	// Verify suggestion fields
	assert.Equal(t, "Matches your search query", deserialized["reason"])
	assert.Equal(t, 0.95, deserialized["confidence"])
}
