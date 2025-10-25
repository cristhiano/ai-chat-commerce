package integration

import (
	"bytes"
	"chat-ecommerce-backend/internal/handlers"
	"chat-ecommerce-backend/internal/models"
	"chat-ecommerce-backend/internal/services"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type ChatIntegrationTestSuite struct {
	suite.Suite
	db             *gorm.DB
	router         *gin.Engine
	chatService    *services.ChatService
	chatHandler    *handlers.ChatHandler
	productService *services.ProductService
	cartService    *services.ShoppingCartService
	sessionID      string
	userID         uuid.UUID
}

func (suite *ChatIntegrationTestSuite) SetupSuite() {
	// Setup test database
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	suite.Require().NoError(err)

	// Auto-migrate
	err = db.AutoMigrate(
		&models.ChatSession{},
		&models.ChatMessage{},
		&models.Product{},
		&models.Category{},
		&models.ShoppingCart{},
		&models.CartItem{},
	)
	suite.Require().NoError(err)

	suite.db = db

	// Create test data
	suite.setupTestData()

	// Setup services
	suite.productService = services.NewProductService(db)
	suite.cartService = services.NewShoppingCartService(db)
	suite.chatService = services.NewChatService(db, suite.productService, suite.cartService)
	suite.chatHandler = handlers.NewChatHandler(suite.chatService)

	// Setup router
	suite.setupRoutes()

	// Generate test session and user
	suite.sessionID = fmt.Sprintf("test-session-%d", time.Now().Unix())
	suite.userID = uuid.New()
}

func (suite *ChatIntegrationTestSuite) setupTestData() {
	// Create test category
	category := models.Category{
		ID:          uuid.New(),
		Name:        "Test Category",
		Description: "Test category for chat tests",
		Slug:        "test-category",
		IsActive:    true,
	}
	suite.db.Create(&category)

	// Create test products
	products := []models.Product{
		{
			ID:          uuid.New(),
			Name:        "Test Product 1",
			Description: "First test product",
			Price:       99.99,
			CategoryID:  category.ID,
			SKU:         "TEST-001",
			Status:      "active",
		},
		{
			ID:          uuid.New(),
			Name:        "Test Product 2",
			Description: "Second test product",
			Price:       149.99,
			CategoryID:  category.ID,
			SKU:         "TEST-002",
			Status:      "active",
		},
	}

	for _, product := range products {
		suite.db.Create(&product)
	}
}

func (suite *ChatIntegrationTestSuite) setupRoutes() {
	gin.SetMode(gin.TestMode)
	suite.router = gin.New()

	// Setup chat routes
	api := suite.router.Group("/api/v1")
	{
		chat := api.Group("chat")
		{
			chat.GET("/ws", suite.chatHandler.HandleWebSocket)
			chat.POST("/message", suite.chatHandler.SendMessage)
			chat.GET("/history/:session_id", suite.chatHandler.GetChatHistory)
			chat.GET("/suggestions", suite.chatHandler.GetProductSuggestions)
			chat.GET("/search", suite.chatHandler.SearchProducts)
			chat.GET("/session/:session_id", suite.chatHandler.GetChatSession)
		}
	}
}

func (suite *ChatIntegrationTestSuite) TearDownSuite() {
	if suite.db != nil {
		sqlDB, _ := suite.db.DB()
		sqlDB.Close()
	}
}

func (suite *ChatIntegrationTestSuite) TestSendMessage() {
	// Test HTTP POST message endpoint
	reqBody := map[string]interface{}{
		"message":    "Hello, I need help finding products",
		"session_id": suite.sessionID,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/chat/message", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	assert.Contains(suite.T(), response, "data")
}

func (suite *ChatIntegrationTestSuite) TestGetChatHistory() {
	// First send a message to create history
	reqBody := map[string]interface{}{
		"message":    "Test message for history",
		"session_id": suite.sessionID,
	}

	jsonBody, _ := json.Marshal(reqBody)
	req := httptest.NewRequest("POST", "/api/v1/chat/message", bytes.NewBuffer(jsonBody))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)
	assert.Equal(suite.T(), http.StatusOK, w.Code)

	// Now get chat history
	req = httptest.NewRequest("GET", fmt.Sprintf("/api/v1/chat/history/%s", suite.sessionID), nil)
	w = httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	assert.Contains(suite.T(), response, "data")
}

func (suite *ChatIntegrationTestSuite) TestGetProductSuggestions() {
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/chat/suggestions?session_id=%s", suite.sessionID), nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	assert.Contains(suite.T(), response, "data")
}

func (suite *ChatIntegrationTestSuite) TestSearchProducts() {
	req := httptest.NewRequest("GET", "/api/v1/chat/search?q=test&limit=5", nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	assert.Contains(suite.T(), response, "data")
}

func (suite *ChatIntegrationTestSuite) TestGetChatSession() {
	req := httptest.NewRequest("GET", fmt.Sprintf("/api/v1/chat/session/%s", suite.sessionID), nil)
	w := httptest.NewRecorder()
	suite.router.ServeHTTP(w, req)

	assert.Equal(suite.T(), http.StatusOK, w.Code)

	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(suite.T(), err)
	assert.True(suite.T(), response["success"].(bool))
	assert.Contains(suite.T(), response, "data")
}

func (suite *ChatIntegrationTestSuite) TestWebSocketConnection() {
	// Create WebSocket connection
	server := httptest.NewServer(suite.router)
	defer server.Close()

	wsURL := fmt.Sprintf("ws://%s/api/v1/chat/ws?session_id=%s", server.URL[7:], suite.sessionID)
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	suite.Require().NoError(err)
	defer conn.Close()

	// Test sending a message
	message := map[string]interface{}{
		"type": "message",
		"data": map[string]interface{}{
			"content":    "Hello from WebSocket",
			"session_id": suite.sessionID,
			"user_id":    suite.userID.String(),
		},
	}

	err = conn.WriteJSON(message)
	suite.Require().NoError(err)

	// Read response
	var response map[string]interface{}
	err = conn.ReadJSON(&response)
	suite.Require().NoError(err)

	assert.Contains(suite.T(), response, "type")
	assert.Contains(suite.T(), response, "data")
}

func (suite *ChatIntegrationTestSuite) TestWebSocketTypingIndicator() {
	// Create WebSocket connection
	server := httptest.NewServer(suite.router)
	defer server.Close()

	wsURL := fmt.Sprintf("ws://%s/api/v1/chat/ws?session_id=%s", server.URL[7:], suite.sessionID)
	conn, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	suite.Require().NoError(err)
	defer conn.Close()

	// Test typing indicator
	typingMessage := map[string]interface{}{
		"type": "typing",
		"data": map[string]interface{}{
			"is_typing": true,
		},
		"sessionId": suite.sessionID,
	}

	err = conn.WriteJSON(typingMessage)
	suite.Require().NoError(err)

	// Should not error
	assert.NoError(suite.T(), err)
}

func TestChatIntegrationTestSuite(t *testing.T) {
	suite.Run(t, new(ChatIntegrationTestSuite))
}
