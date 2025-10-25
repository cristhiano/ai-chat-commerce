package services

import (
	"chat-ecommerce-backend/internal/models"
	"chat-ecommerce-backend/internal/services"
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
	userID := uuid.New()

	// Get conversation history (should be empty)
	history, err := service.GetConversationHistory(sessionID, 10)
	assert.NoError(t, err)
	assert.NotNil(t, history)
	assert.Len(t, history, 0)
}
