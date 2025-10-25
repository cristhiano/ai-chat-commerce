package websocket

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// WebSocketService manages the WebSocket server and all real-time functionality
type WebSocketService struct {
	// Core components
	hub              *Hub
	clientManager    *ClientManager
	authManager      *WebSocketAuthManager
	cartSyncManager  *CartSyncManager
	inventoryManager *InventoryBroadcastManager
	notificationManager *NotificationManager
	sessionManager   *SessionManager
	connectionManager *ConnectionManager
	
	// Configuration
	upgrader         websocket.Upgrader
	readBufferSize   int
	writeBufferSize  int
	checkOrigin      func(r *http.Request) bool
	
	// Context for cancellation
	ctx              context.Context
	cancel           context.CancelFunc
	
	// Statistics
	stats            *WebSocketStats
	
	// Mutex for thread-safe operations
	mu               sync.RWMutex
}

// WebSocketStats holds WebSocket service statistics
type WebSocketStats struct {
	TotalConnections    int64
	ActiveConnections   int64
	TotalMessages       int64
	MessagesPerSecond   float64
	AverageLatency      time.Duration
	ErrorCount          int64
	LastReset           time.Time
	mu                  sync.RWMutex
}

// NewWebSocketService creates a new WebSocket service
func NewWebSocketService(
	hub *Hub,
	clientManager *ClientManager,
	authManager *WebSocketAuthManager,
	cartSyncManager *CartSyncManager,
	inventoryManager *InventoryBroadcastManager,
	notificationManager *NotificationManager,
	sessionManager *SessionManager,
	connectionManager *ConnectionManager,
) *WebSocketService {
	ctx, cancel := context.WithCancel(context.Background())
	
	service := &WebSocketService{
		hub:               hub,
		clientManager:     clientManager,
		authManager:       authManager,
		cartSyncManager:   cartSyncManager,
		inventoryManager:  inventoryManager,
		notificationManager: notificationManager,
		sessionManager:    sessionManager,
		connectionManager: connectionManager,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			CheckOrigin: func(r *http.Request) bool {
				// In production, implement proper origin checking
				return true
			},
		},
		readBufferSize:  1024,
		writeBufferSize: 1024,
		ctx:             ctx,
		cancel:          cancel,
		stats: &WebSocketStats{
			LastReset: time.Now(),
		},
	}
	
	// Set up event handlers
	service.setupEventHandlers()
	
	return service
}

// HandleWebSocket handles WebSocket connection upgrades
func (ws *WebSocketService) HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	// Extract session ID from query parameters or headers
	sessionID := r.URL.Query().Get("session_id")
	if sessionID == "" {
		sessionID = r.Header.Get("X-Session-ID")
	}
	if sessionID == "" {
		sessionID = uuid.New().String()
	}
	
	// Upgrade HTTP connection to WebSocket
	conn, err := ws.upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade error: %v", err)
		ws.stats.incrementErrorCount()
		return
	}
	
	// Add client to manager
	client, err := ws.clientManager.AddClient(conn, sessionID)
	if err != nil {
		log.Printf("Failed to add client: %v", err)
		conn.Close()
		ws.stats.incrementErrorCount()
		return
	}
	
	// Register client with session manager
	ws.sessionManager.RegisterClient(client.ID, sessionID)
	
	// Start message processing for this client
	go ws.handleClientMessages(client)
	
	ws.stats.incrementTotalConnections()
	ws.stats.incrementActiveConnections()
	
	log.Printf("WebSocket connection established: %s (Session: %s)", client.ID, sessionID)
}

// handleClientMessages processes messages for a specific client
func (ws *WebSocketService) handleClientMessages(client *Client) {
	defer func() {
		// Cleanup when client disconnects
		ws.clientManager.RemoveClient(client.ID)
		ws.sessionManager.UnregisterClient(client.ID)
		ws.stats.decrementActiveConnections()
		
		log.Printf("Client message handler stopped: %s", client.ID)
	}()
	
	for {
		select {
		case <-client.ctx.Done():
			return
		case message := <-client.Receive:
			ws.processMessage(client, message)
		case <-client.Close:
			return
		}
	}
}

// processMessage processes an incoming message from a client
func (ws *WebSocketService) processMessage(client *Client, message *WebSocketMessage) {
	ws.stats.incrementTotalMessages()
	
	// Update client activity
	client.UpdateActivity()
	
	// Process message based on type
	switch message.Type {
	case MessageTypeAuth:
		ws.handleAuthMessage(client, message)
	case MessageTypePing:
		ws.handlePingMessage(client, message)
	case MessageTypeChatMessage:
		ws.handleChatMessage(client, message)
	case MessageTypeCartAdd, MessageTypeCartRemove, MessageTypeCartClear:
		ws.handleCartMessage(client, message)
	case MessageTypeInventorySync:
		ws.handleInventorySyncMessage(client, message)
	default:
		log.Printf("Unknown message type: %s", message.Type)
		ws.sendError(client, "unknown_message_type", fmt.Sprintf("Unknown message type: %s", message.Type))
	}
}

// handleAuthMessage handles authentication messages
func (ws *WebSocketService) handleAuthMessage(client *Client, message *WebSocketMessage) {
	token, ok := message.Data["token"].(string)
	if !ok {
		ws.sendError(client, "invalid_auth_data", "Missing or invalid token")
		return
	}
	
	// Authenticate token
	authResult, err := ws.authManager.AuthenticateToken(token)
	if err != nil {
		ws.sendError(client, "auth_error", fmt.Sprintf("Authentication failed: %v", err))
		return
	}
	
	if !authResult.IsValid {
		ws.sendError(client, "auth_failed", "Invalid token")
		return
	}
	
	// Create auth session
	session, err := ws.authManager.CreateAuthSession(
		*authResult.UserID,
		authResult.SessionID,
		authResult.AuthLevel,
		authResult.Permissions,
	)
	if err != nil {
		ws.sendError(client, "session_creation_failed", fmt.Sprintf("Failed to create session: %v", err))
		return
	}
	
	// Authenticate client
	client.Authenticate(*authResult.UserID, authResult.AuthLevel, authResult.Permissions)
	
	// Send success response
	successMsg := NewMessageBuilder(MessageTypeAuthSuccess).
		WithSession(authResult.SessionID).
		WithUser(*authResult.UserID).
		WithDataField("session_id", authResult.SessionID).
		WithDataField("user_id", authResult.UserID).
		WithDataField("auth_level", authResult.AuthLevel).
		WithDataField("permissions", authResult.Permissions).
		Build()
	
	client.SendMessage(successMsg)
	
	log.Printf("Client authenticated: %s (User: %s)", client.ID, authResult.UserID)
}

// handlePingMessage handles ping messages
func (ws *WebSocketService) handlePingMessage(client *Client, message *WebSocketMessage) {
	sequence, _ := message.Data["sequence"].(int)
	
	// Calculate latency
	latency := time.Since(message.Timestamp)
	
	// Send pong response
	pongMsg := NewMessageFactory().CreatePong(sequence, latency.Milliseconds())
	pongMsg.SetSession(client.SessionID)
	
	client.SendMessage(pongMsg)
}

// handleChatMessage handles chat messages
func (ws *WebSocketService) handleChatMessage(client *Client, message *WebSocketMessage) {
	// Check if client is authenticated for chat
	if !ws.authManager.CheckPermission(client.SessionID, PermissionChatAccess) {
		ws.sendError(client, "permission_denied", "Chat access denied")
		return
	}
	
	// Process chat message (this would integrate with chat service)
	content, ok := message.Data["content"].(string)
	if !ok {
		ws.sendError(client, "invalid_chat_data", "Missing or invalid content")
		return
	}
	
	// Broadcast to other clients in the same session
	chatMsg := NewMessageBuilder(MessageTypeChatMessage).
		WithSession(client.SessionID).
		WithDataField("content", content).
		WithDataField("message_type", "user").
		WithDataField("timestamp", time.Now()).
		Build()
	
	ws.clientManager.BroadcastToSession(client.SessionID, chatMsg)
}

// handleCartMessage handles cart-related messages
func (ws *WebSocketService) handleCartMessage(client *Client, message *WebSocketMessage) {
	// Check cart permissions
	if !ws.authManager.CheckPermission(client.SessionID, PermissionWriteCart) {
		ws.sendError(client, "permission_denied", "Cart write access denied")
		return
	}
	
	// Process cart message through cart sync manager
	ws.cartSyncManager.ProcessCartMessage(client, message)
}

// handleInventorySyncMessage handles inventory sync messages
func (ws *WebSocketService) handleInventorySyncMessage(client *Client, message *WebSocketMessage) {
	// Check inventory read permissions
	if !ws.authManager.CheckPermission(client.SessionID, PermissionReadInventory) {
		ws.sendError(client, "permission_denied", "Inventory read access denied")
		return
	}
	
	// Process inventory sync through inventory manager
	ws.inventoryManager.ProcessSyncRequest(client, message)
}

// sendError sends an error message to a client
func (ws *WebSocketService) sendError(client *Client, code, message string) {
	errorMsg := NewMessageFactory().CreateError(code, message, client.SessionID, client.UserID)
	client.SendMessage(errorMsg)
	ws.stats.incrementErrorCount()
}

// BroadcastToSession broadcasts a message to all clients in a session
func (ws *WebSocketService) BroadcastToSession(sessionID string, message *WebSocketMessage) error {
	return ws.clientManager.BroadcastToSession(sessionID, message)
}

// BroadcastToUser broadcasts a message to all clients for a user
func (ws *WebSocketService) BroadcastToUser(userID uuid.UUID, message *WebSocketMessage) error {
	return ws.clientManager.BroadcastToUser(userID, message)
}

// BroadcastToAll broadcasts a message to all connected clients
func (ws *WebSocketService) BroadcastToAll(message *WebSocketMessage) error {
	return ws.clientManager.BroadcastToAll(message)
}

// SendNotification sends a notification to a specific user or session
func (ws *WebSocketService) SendNotification(userID *uuid.UUID, sessionID string, notification *NotificationData) error {
	message := NewMessageBuilder(MessageTypeNotification).
		WithDataField("notification_id", notification.NotificationID).
		WithDataField("title", notification.Title).
		WithDataField("message", notification.Message).
		WithDataField("type", notification.Type).
		WithDataField("priority", notification.Priority).
		WithDataField("timestamp", time.Now()).
		Build()
	
	if userID != nil {
		message.SetUser(*userID)
		return ws.BroadcastToUser(*userID, message)
	} else if sessionID != "" {
		message.SetSession(sessionID)
		return ws.BroadcastToSession(sessionID, message)
	}
	
	return fmt.Errorf("either userID or sessionID must be provided")
}

// GetStats returns WebSocket service statistics
func (ws *WebSocketService) GetStats() map[string]interface{} {
	ws.stats.mu.RLock()
	defer ws.stats.mu.RUnlock()
	
	clientStats := map[string]interface{}{
		"total_clients":    ws.clientManager.GetClientCount(),
		"total_sessions":   ws.clientManager.GetSessionCount(),
		"total_users":      ws.clientManager.GetUserCount(),
	}
	
	authStats := ws.authManager.GetAuthStats()
	
	return map[string]interface{}{
		"websocket": map[string]interface{}{
			"total_connections":  ws.stats.TotalConnections,
			"active_connections": ws.stats.ActiveConnections,
			"total_messages":     ws.stats.TotalMessages,
			"messages_per_second": ws.stats.MessagesPerSecond,
			"average_latency_ms":  ws.stats.AverageLatency.Milliseconds(),
			"error_count":        ws.stats.ErrorCount,
			"uptime_seconds":     time.Since(ws.stats.LastReset).Seconds(),
		},
		"clients": clientStats,
		"auth":    authStats,
	}
}

// setupEventHandlers sets up event handlers for the WebSocket service
func (ws *WebSocketService) setupEventHandlers() {
	ws.clientManager.SetEventHandlers(
		// OnConnect
		func(client *Client) {
			log.Printf("Client connected: %s", client.ID)
		},
		// OnDisconnect
		func(client *Client) {
			log.Printf("Client disconnected: %s", client.ID)
		},
		// OnMessage
		func(client *Client, message *WebSocketMessage) {
			// Message processing is handled in handleClientMessages
		},
		// OnError
		func(client *Client, err error) {
			log.Printf("Client error: %s - %v", client.ID, err)
			ws.stats.incrementErrorCount()
		},
	)
}

// Stop stops the WebSocket service
func (ws *WebSocketService) Stop() {
	ws.cancel()
	ws.clientManager.Stop()
	ws.authManager.Stop()
}

// Statistics methods for WebSocketStats

func (stats *WebSocketStats) incrementTotalConnections() {
	stats.mu.Lock()
	defer stats.mu.Unlock()
	stats.TotalConnections++
}

func (stats *WebSocketStats) decrementTotalConnections() {
	stats.mu.Lock()
	defer stats.mu.Unlock()
	stats.TotalConnections--
}

func (stats *WebSocketStats) incrementActiveConnections() {
	stats.mu.Lock()
	defer stats.mu.Unlock()
	stats.ActiveConnections++
}

func (stats *WebSocketStats) decrementActiveConnections() {
	stats.mu.Lock()
	defer stats.mu.Unlock()
	stats.ActiveConnections--
}

func (stats *WebSocketStats) incrementTotalMessages() {
	stats.mu.Lock()
	defer stats.mu.Unlock()
	stats.TotalMessages++
}

func (stats *WebSocketStats) incrementErrorCount() {
	stats.mu.Lock()
	defer stats.mu.Unlock()
	stats.ErrorCount++
}

func (stats *WebSocketStats) updateLatency(latency time.Duration) {
	stats.mu.Lock()
	defer stats.mu.Unlock()
	stats.AverageLatency = latency
}

func (stats *WebSocketStats) reset() {
	stats.mu.Lock()
	defer stats.mu.Unlock()
	stats.TotalConnections = 0
	stats.ActiveConnections = 0
	stats.TotalMessages = 0
	stats.ErrorCount = 0
	stats.LastReset = time.Now()
}
