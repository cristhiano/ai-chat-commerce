package handlers

import (
	"chat-ecommerce-backend/internal/handlers"
	"chat-ecommerce-backend/internal/models"
	"chat-ecommerce-backend/internal/services"
	"encoding/json"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupChatTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal("Failed to connect to test database:", err)
	}

	// Auto-migrate all required models
	err = db.AutoMigrate(
		&models.ChatSession{},
		&models.ChatMessage{},
		&models.Product{},
		&models.Category{},
		&models.Inventory{},
		&models.User{},
		&models.ShoppingCart{},
	)
	if err != nil {
		t.Fatal("Failed to migrate test database:", err)
	}

	return db
}

func setupChatHandler(t *testing.T) (*handlers.ChatHandler, *gorm.DB) {
	db := setupChatTestDB(t)

	productService := services.NewProductService(db)
	cartService := services.NewShoppingCartService(db)
	chatService := services.NewChatService(db, productService, cartService)
	handler := handlers.NewChatHandler(chatService)

	return handler, db
}

// TestWebSocketMessage_ProductSuggestionSerialization tests that product suggestions
// sent via WebSocket include complete product data with category
func TestWebSocketMessage_ProductSuggestionSerialization(t *testing.T) {
	// Setup test database
	db := setupChatTestDB(t)

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
	err := db.Create(&category).Error
	assert.NoError(t, err)

	// Create multiple test products with complete data
	products := []models.Product{
		{
			ID:          uuid.New(),
			Name:        "Wireless Headphones",
			Description: "High-quality wireless headphones with noise cancellation",
			Price:       199.99,
			CategoryID:  categoryID,
			SKU:         "WH-001",
			Status:      "active",
		},
		{
			ID:          uuid.New(),
			Name:        "Bluetooth Speaker",
			Description: "Portable bluetooth speaker with excellent sound quality",
			Price:       89.99,
			CategoryID:  categoryID,
			SKU:         "BS-001",
			Status:      "active",
		},
	}

	for _, product := range products {
		err = db.Create(&product).Error
		assert.NoError(t, err)

		// Create inventory for each product
		inventory := models.Inventory{
			ID:                uuid.New(),
			ProductID:         product.ID,
			WarehouseLocation: "Warehouse A",
			QuantityAvailable: 50,
			QuantityReserved:  5,
			LowStockThreshold: 10,
			ReorderPoint:      20,
		}
		err = db.Create(&inventory).Error
		assert.NoError(t, err)
	}

	// Load products with relationships (simulating what chat service does)
	var loadedProducts []models.Product
	err = db.Preload("Category").Preload("Inventory").Find(&loadedProducts).Error
	assert.NoError(t, err)
	assert.Len(t, loadedProducts, 2)

	// Create product suggestions (simulating chat service response)
	suggestions := []services.ProductSuggestion{
		{
			Product:    &loadedProducts[0],
			Reason:     "Matches your search for audio products",
			Confidence: 0.95,
		},
		{
			Product:    &loadedProducts[1],
			Reason:     "Popular in Electronics category",
			Confidence: 0.85,
		},
	}

	// Create WebSocket message structure that handler would send
	type WebSocketMessage struct {
		Type      string                       `json:"type"`
		Data      []services.ProductSuggestion `json:"data"`
		SessionID string                       `json:"session_id"`
	}

	wsMessage := WebSocketMessage{
		Type:      "suggestions",
		Data:      suggestions,
		SessionID: "test-session-123",
	}

	// Serialize to JSON (this is what gets sent over WebSocket)
	jsonData, err := json.Marshal(wsMessage)
	assert.NoError(t, err)
	assert.NotEmpty(t, jsonData)

	// Deserialize back to verify structure
	var deserialized map[string]interface{}
	err = json.Unmarshal(jsonData, &deserialized)
	assert.NoError(t, err)

	// Verify top-level WebSocket message fields
	assert.Equal(t, "suggestions", deserialized["type"])
	assert.Equal(t, "test-session-123", deserialized["session_id"])
	assert.Contains(t, deserialized, "data")

	// Verify suggestions array
	suggestionsData := deserialized["data"].([]interface{})
	assert.Len(t, suggestionsData, 2)

	// Verify first suggestion
	firstSuggestion := suggestionsData[0].(map[string]interface{})
	assert.Contains(t, firstSuggestion, "product")
	assert.Contains(t, firstSuggestion, "reason")
	assert.Contains(t, firstSuggestion, "confidence")

	// Verify product data in first suggestion
	productData := firstSuggestion["product"].(map[string]interface{})
	assert.Contains(t, productData, "id")
	assert.Contains(t, productData, "name")
	assert.Contains(t, productData, "description")
	assert.Contains(t, productData, "price")
	assert.Contains(t, productData, "category_id")
	assert.Contains(t, productData, "sku")
	assert.Contains(t, productData, "status")

	// CRITICAL: Verify category relationship is included
	assert.Contains(t, productData, "category")
	categoryData := productData["category"].(map[string]interface{})
	assert.Contains(t, categoryData, "id")
	assert.Contains(t, categoryData, "name")
	assert.Equal(t, "Electronics", categoryData["name"])
	assert.Contains(t, categoryData, "slug")
	assert.Equal(t, "electronics", categoryData["slug"])

	// CRITICAL: Verify inventory relationship is included
	assert.Contains(t, productData, "inventory")
	inventoryData := productData["inventory"].([]interface{})
	assert.Greater(t, len(inventoryData), 0, "Inventory should be included")
	firstInventory := inventoryData[0].(map[string]interface{})
	assert.Contains(t, firstInventory, "id")
	assert.Contains(t, firstInventory, "product_id")
	assert.Contains(t, firstInventory, "quantity_available")
	assert.Contains(t, firstInventory, "warehouse_location")
	assert.Equal(t, float64(50), firstInventory["quantity_available"])
	assert.Equal(t, "Warehouse A", firstInventory["warehouse_location"])

	// Verify suggestion metadata
	assert.Equal(t, "Matches your search for audio products", firstSuggestion["reason"])
	assert.Equal(t, 0.95, firstSuggestion["confidence"])

	// Verify second suggestion has same structure
	secondSuggestion := suggestionsData[1].(map[string]interface{})
	assert.Contains(t, secondSuggestion, "product")
	secondProduct := secondSuggestion["product"].(map[string]interface{})
	assert.Contains(t, secondProduct, "category")
	assert.Contains(t, secondProduct, "inventory")
}

// TestWebSocketMessage_CompleteProductDataFields validates that all required
// product fields are present in WebSocket messages
func TestWebSocketMessage_CompleteProductDataFields(t *testing.T) {
	// Setup test database
	db := setupChatTestDB(t)

	// Create category
	categoryID := uuid.New()
	category := models.Category{
		ID:          categoryID,
		Name:        "Books",
		Description: "Books and literature",
		Slug:        "books",
		SortOrder:   2,
		IsActive:    true,
	}
	err := db.Create(&category).Error
	assert.NoError(t, err)

	// Create product with all fields
	productID := uuid.New()
	product := models.Product{
		ID:          productID,
		Name:        "The Go Programming Language",
		Description: "A comprehensive guide to Go programming",
		Price:       49.99,
		CategoryID:  categoryID,
		SKU:         "BOOK-GO-001",
		Status:      "active",
	}
	err = db.Create(&product).Error
	assert.NoError(t, err)

	// Create inventory
	inventory := models.Inventory{
		ID:                uuid.New(),
		ProductID:         productID,
		WarehouseLocation: "Warehouse B",
		QuantityAvailable: 25,
		QuantityReserved:  3,
		LowStockThreshold: 5,
		ReorderPoint:      10,
	}
	err = db.Create(&inventory).Error
	assert.NoError(t, err)

	// Load product with all relationships
	var loadedProduct models.Product
	err = db.Preload("Category").Preload("Inventory").First(&loadedProduct, productID).Error
	assert.NoError(t, err)

	// Verify all relationships are loaded
	assert.NotNil(t, loadedProduct.Category)
	assert.NotEmpty(t, loadedProduct.Inventory)

	// Create suggestion
	suggestion := services.ProductSuggestion{
		Product:    &loadedProduct,
		Reason:     "Best seller in Books category",
		Confidence: 0.92,
	}

	// Serialize
	jsonData, err := json.Marshal(suggestion)
	assert.NoError(t, err)

	// Verify JSON contains all required fields
	jsonString := string(jsonData)

	// Product fields
	assert.Contains(t, jsonString, "\"id\":")
	assert.Contains(t, jsonString, "\"name\":")
	assert.Contains(t, jsonString, "\"description\":")
	assert.Contains(t, jsonString, "\"price\":")
	assert.Contains(t, jsonString, "\"category_id\":")
	assert.Contains(t, jsonString, "\"sku\":")
	assert.Contains(t, jsonString, "\"status\":")

	// Category fields
	assert.Contains(t, jsonString, "\"category\":")
	assert.Contains(t, jsonString, "Books")

	// Inventory fields
	assert.Contains(t, jsonString, "\"inventory\":")
	assert.Contains(t, jsonString, "\"quantity_available\":")
	assert.Contains(t, jsonString, "\"warehouse_location\":")

	// Suggestion fields
	assert.Contains(t, jsonString, "\"reason\":")
	assert.Contains(t, jsonString, "\"confidence\":")
	assert.Contains(t, jsonString, "Best seller in Books category")
}

// T084: Test WebSocket message size stays under 50KB for product list
func TestWebSocketMessage_MessageSizeLimit(t *testing.T) {
	// Setup test database
	db := setupChatTestDB(t)

	// Create category
	categoryID := uuid.New()
	category := models.Category{
		ID:          categoryID,
		Name:        "Electronics",
		Description: "Electronic products and gadgets",
		Slug:        "electronics",
		SortOrder:   1,
		IsActive:    true,
	}
	err := db.Create(&category).Error
	assert.NoError(t, err)

	// Create a large list of products (simulate maximum response)
	products := make([]models.Product, 10)
	for i := 0; i < 10; i++ {
		product := models.Product{
			ID:          uuid.New(),
			Name:        "Product Name" + string(rune('A'+i)),
			Description: "Detailed product description with features and specifications that simulates realistic data sent over WebSocket connections",
			Price:       99.99 + float64(i)*10,
			CategoryID:  categoryID,
			SKU:         "PROD-00" + string(rune('0'+i)),
			Status:      "active",
		}
		err = db.Create(&product).Error
		assert.NoError(t, err)

		// Add inventory
		inventory := models.Inventory{
			ID:                uuid.New(),
			ProductID:         product.ID,
			WarehouseLocation: "Warehouse A",
			QuantityAvailable: 50 + i*10,
			QuantityReserved:  5,
			LowStockThreshold: 10,
			ReorderPoint:      20,
		}
		err = db.Create(&inventory).Error
		assert.NoError(t, err)

		products[i] = product
	}

	// Load products with relationships
	var loadedProducts []models.Product
	err = db.Preload("Category").Preload("Inventory").Find(&loadedProducts).Error
	assert.NoError(t, err)

	// Create suggestions
	suggestions := make([]services.ProductSuggestion, len(loadedProducts))
	for i, product := range loadedProducts {
		suggestions[i] = services.ProductSuggestion{
			Product:    &product,
			Reason:     "Matches your search query for electronics products",
			Confidence: 0.85 + float64(i)*0.01,
		}
	}

	// Create WebSocket message structure (as sent by handler)
	type WebSocketMessage struct {
		Type      string                       `json:"type"`
		Data      []services.ProductSuggestion `json:"data"`
		SessionID string                       `json:"session_id"`
	}

	wsMessage := WebSocketMessage{
		Type:      "suggestions",
		Data:      suggestions,
		SessionID: "test-session-perf-123",
	}

	// Serialize to JSON
	jsonData, err := json.Marshal(wsMessage)
	assert.NoError(t, err)

	// T084 VALIDATION: Verify message size is under 50KB
	messageSize := len(jsonData)
	maxSize := 50 * 1024 // 50KB in bytes

	assert.Less(t, messageSize, maxSize,
		"WebSocket message size (%d bytes) exceeds 50KB limit (%d bytes)",
		messageSize, maxSize)

	// Log actual size for monitoring
	t.Logf("✅ WebSocket message size: %d bytes (%.2f KB) for %d products",
		messageSize, float64(messageSize)/1024, len(suggestions))
}
