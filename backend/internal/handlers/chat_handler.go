package handlers

import (
	"chat-ecommerce-backend/internal/services"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// ChatHandler handles chat-related HTTP requests and WebSocket connections
type ChatHandler struct {
	chatService *services.ChatService
	upgrader    websocket.Upgrader
}

// NewChatHandler creates a new ChatHandler
func NewChatHandler(chatService *services.ChatService) *ChatHandler {
	return &ChatHandler{
		chatService: chatService,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true // Allow all origins in development
			},
		},
	}
}

// ChatMessage represents a chat message
type ChatMessage struct {
	ID        string                 `json:"id"`
	SessionID string                 `json:"session_id"`
	UserID    *string                `json:"user_id,omitempty"`
	Role      string                 `json:"role"`
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	Timestamp string                 `json:"timestamp"`
}

// ChatRequest represents a chat request
type ChatRequest struct {
	Message   string `json:"message" binding:"required"`
	SessionID string `json:"session_id"`
}

// ChatResponse represents a chat response
type ChatResponse struct {
	Message     string                       `json:"message"`
	Actions     []services.ChatAction        `json:"actions,omitempty"`
	Suggestions []services.ProductSuggestion `json:"suggestions,omitempty"`
	Context     map[string]interface{}       `json:"context,omitempty"`
	Error       string                       `json:"error,omitempty"`
}

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	Type      string      `json:"type"` // "message", "typing", "error"
	Data      interface{} `json:"data"`
	SessionID string      `json:"session_id"`
	UserID    *string     `json:"user_id,omitempty"`
}

// HandleWebSocket handles WebSocket connections for real-time chat
func (h *ChatHandler) HandleWebSocket(c *gin.Context) {
	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket connection: %v", err)
		return
	}
	defer conn.Close()

	// Get session ID from query parameters
	sessionID := c.Query("session_id")
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	// Get user ID from context (if authenticated)
	var userID *uuid.UUID
	if userIDStr, exists := c.Get("user_id"); exists {
		if id, ok := userIDStr.(uuid.UUID); ok {
			userID = &id
		}
	}

	log.Printf("WebSocket connection established for session: %s", sessionID)

	// Send welcome message
	welcomeMsg := WebSocketMessage{
		Type: "message",
		Data: ChatMessage{
			ID:        uuid.New().String(),
			SessionID: sessionID,
			Role:      "assistant",
			Content:   "Hello! I'm your shopping assistant. How can I help you today?",
			Timestamp: time.Now().Format(time.RFC3339),
		},
		SessionID: sessionID,
	}

	if err := conn.WriteJSON(welcomeMsg); err != nil {
		log.Printf("Failed to send welcome message: %v", err)
		return
	}

	// Handle incoming messages
	for {
		var wsMsg WebSocketMessage
		err := conn.ReadJSON(&wsMsg)
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Handle different message types
		switch wsMsg.Type {
		case "message":
			h.handleChatMessage(conn, wsMsg, sessionID, userID)
		case "typing":
			h.handleTypingIndicator(conn, wsMsg)
		default:
			log.Printf("Unknown message type: %s", wsMsg.Type)
		}
	}
}

// handleChatMessage processes a chat message
func (h *ChatHandler) handleChatMessage(conn *websocket.Conn, wsMsg WebSocketMessage, sessionID string, userID *uuid.UUID) {
	// Extract message content
	msgData, ok := wsMsg.Data.(map[string]interface{})
	if !ok {
		h.sendError(conn, "Invalid message format", sessionID)
		return
	}

	content, ok := msgData["content"].(string)
	if !ok {
		h.sendError(conn, "Missing message content", sessionID)
		return
	}

	// Send typing indicator
	h.sendTypingIndicator(conn, sessionID, true)

	// Process message with chat service
	response, err := h.chatService.ProcessMessage(sessionID, userID, content)
	if err != nil {
		log.Printf("Failed to process chat message: %v", err)
		h.sendError(conn, "Failed to process message", sessionID)
		return
	}

	// Stop typing indicator
	h.sendTypingIndicator(conn, sessionID, false)

	// Send response
	responseMsg := WebSocketMessage{
		Type: "message",
		Data: ChatMessage{
			ID:        uuid.New().String(),
			SessionID: sessionID,
			Role:      "assistant",
			Content:   response.Message,
			Metadata: map[string]interface{}{
				"actions":     response.Actions,
				"suggestions": response.Suggestions,
			},
			Timestamp: time.Now().Format(time.RFC3339),
		},
		SessionID: sessionID,
	}

	if err := conn.WriteJSON(responseMsg); err != nil {
		log.Printf("Failed to send response: %v", err)
		return
	}

	// Send actions if any
	if len(response.Actions) > 0 {
		actionsMsg := WebSocketMessage{
			Type:      "actions",
			Data:      response.Actions,
			SessionID: sessionID,
		}
		conn.WriteJSON(actionsMsg)
	}

	// Send suggestions if any
	if len(response.Suggestions) > 0 {
		suggestionsMsg := WebSocketMessage{
			Type:      "suggestions",
			Data:      response.Suggestions,
			SessionID: sessionID,
		}
		conn.WriteJSON(suggestionsMsg)
	}
}

// handleTypingIndicator handles typing indicators
func (h *ChatHandler) handleTypingIndicator(conn *websocket.Conn, wsMsg WebSocketMessage) {
	// Echo typing indicator to other clients (in a real app, you'd broadcast to other clients)
	// For now, just acknowledge
}

// sendTypingIndicator sends a typing indicator
func (h *ChatHandler) sendTypingIndicator(conn *websocket.Conn, sessionID string, isTyping bool) {
	typingMsg := WebSocketMessage{
		Type: "typing",
		Data: map[string]interface{}{
			"is_typing": isTyping,
		},
		SessionID: sessionID,
	}
	conn.WriteJSON(typingMsg)
}

// sendError sends an error message
func (h *ChatHandler) sendError(conn *websocket.Conn, message string, sessionID string) {
	errorMsg := WebSocketMessage{
		Type: "error",
		Data: map[string]interface{}{
			"message": message,
		},
		SessionID: sessionID,
	}
	conn.WriteJSON(errorMsg)
}

// SendMessage handles HTTP POST requests for sending messages
func (h *ChatHandler) SendMessage(c *gin.Context) {
	var req ChatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context (if authenticated)
	var userID *uuid.UUID
	if userIDStr, exists := c.Get("user_id"); exists {
		if id, ok := userIDStr.(uuid.UUID); ok {
			userID = &id
		}
	}

	// Generate session ID if not provided
	sessionID := req.SessionID
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	// Process message
	response, err := h.chatService.ProcessMessage(sessionID, userID, req.Message)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetChatHistory retrieves chat history for a session
func (h *ChatHandler) GetChatHistory(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	// Get conversation history
	history, err := h.chatService.GetConversationHistory(sessionID, 50)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data": gin.H{
			"session_id": sessionID,
			"messages":   history,
		},
	})
}

// GetProductSuggestions gets product suggestions based on chat context
func (h *ChatHandler) GetProductSuggestions(c *gin.Context) {
	sessionID := c.Query("session_id")
	if sessionID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "session_id is required"})
		return
	}

	// Get user ID from context (if authenticated)
	var userID *uuid.UUID
	if userIDStr, exists := c.Get("user_id"); exists {
		if id, ok := userIDStr.(uuid.UUID); ok {
			userID = &id
		}
	}

	// Get recommendations
	suggestions, err := h.chatService.GetProductRecommendations(sessionID, userID, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    suggestions,
	})
}

// SearchProducts searches for products based on natural language query
func (h *ChatHandler) SearchProducts(c *gin.Context) {
	query := c.Query("q")
	if query == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "query parameter 'q' is required"})
		return
	}

	limit := 10
	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := parseInt(limitStr); err == nil {
			limit = l
		}
	}

	// Search products
	suggestions, err := h.chatService.SearchProducts(query, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    suggestions,
	})
}

// GetChatSession retrieves or creates a chat session
func (h *ChatHandler) GetChatSession(c *gin.Context) {
	sessionID := c.Param("session_id")
	if sessionID == "" {
		sessionID = uuid.New().String()
	}

	// Get user ID from context (if authenticated)
	var userID *uuid.UUID
	if userIDStr, exists := c.Get("user_id"); exists {
		if id, ok := userIDStr.(uuid.UUID); ok {
			userID = &id
		}
	}

	// Get or create session
	session, err := h.chatService.GetChatSession(sessionID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    session,
	})
}

// Helper function to parse integer from string
func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}
