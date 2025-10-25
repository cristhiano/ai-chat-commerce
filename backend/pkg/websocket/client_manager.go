package websocket

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// ClientState represents the state of a WebSocket client
type ClientState int

const (
	ClientStateConnecting ClientState = iota
	ClientStateConnected
	ClientStateAuthenticated
	ClientStateDisconnected
)

// ClientInfo represents client information for the manager
type ClientInfo struct {
	// Connection details
	ID          string
	Conn        *websocket.Conn
	SessionID   string
	UserID      *uuid.UUID
	AuthLevel   AuthLevel
	Permissions []string

	// State management
	State        ClientState
	ConnectedAt  time.Time
	LastActivity time.Time

	// Communication channels
	Send      chan *WebSocketMessage
	Receive   chan *WebSocketMessage
	CloseChan chan bool

	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc

	// Configuration
	PingInterval   time.Duration
	PongTimeout    time.Duration
	WriteTimeout   time.Duration
	ReadTimeout    time.Duration
	MaxMessageSize int64

	// Statistics
	MessagesSent     int64
	MessagesReceived int64
	BytesSent        int64
	BytesReceived    int64

	// Metadata
	Metadata map[string]interface{}

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// ClientManager manages WebSocket client connections
type ClientManager struct {
	// Client storage
	clients  map[string]*ClientInfo
	sessions map[string][]*ClientInfo // sessionID -> clients
	users    map[string][]*ClientInfo // userID -> clients

	// Mutex for thread-safe operations
	mu sync.RWMutex

	// Configuration
	maxClients      int
	clientTimeout   time.Duration
	cleanupInterval time.Duration

	// Context for cancellation
	ctx    context.Context
	cancel context.CancelFunc

	// Cleanup control
	cleanupRunning bool

	// Event handlers
	onConnect    func(*ClientInfo)
	onDisconnect func(*ClientInfo)
	onMessage    func(*ClientInfo, *WebSocketMessage)
	onError      func(*ClientInfo, error)
}

// NewClientManager creates a new client manager
func NewClientManager(maxClients int, clientTimeout, cleanupInterval time.Duration) *ClientManager {
	ctx, cancel := context.WithCancel(context.Background())

	return &ClientManager{
		clients:         make(map[string]*ClientInfo),
		sessions:        make(map[string][]*ClientInfo),
		users:           make(map[string][]*ClientInfo),
		maxClients:      maxClients,
		clientTimeout:   clientTimeout,
		cleanupInterval: cleanupInterval,
		ctx:             ctx,
		cancel:          cancel,
	}
}

// AddClient adds a new client to the manager
func (cm *ClientManager) AddClient(conn *websocket.Conn, sessionID string) (*ClientInfo, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Check if we've reached the maximum number of clients
	if len(cm.clients) >= cm.maxClients {
		return nil, fmt.Errorf("maximum number of clients reached: %d", cm.maxClients)
	}

	// Create client context
	ctx, cancel := context.WithCancel(cm.ctx)

	// Create new client
	client := &ClientInfo{
		ID:             uuid.New().String(),
		Conn:           conn,
		SessionID:      sessionID,
		State:          ClientStateConnected,
		ConnectedAt:    time.Now(),
		LastActivity:   time.Now(),
		Send:           make(chan *WebSocketMessage, 256),
		Receive:        make(chan *WebSocketMessage, 256),
		CloseChan:      make(chan bool, 1),
		ctx:            ctx,
		cancel:         cancel,
		PingInterval:   30 * time.Second,
		PongTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		ReadTimeout:    60 * time.Second,
		MaxMessageSize: 1024 * 1024, // 1MB
		Metadata:       make(map[string]interface{}),
	}

	// Add client to storage
	cm.clients[client.ID] = client

	// Add to session mapping
	if cm.sessions[sessionID] == nil {
		cm.sessions[sessionID] = make([]*ClientInfo, 0)
	}
	cm.sessions[sessionID] = append(cm.sessions[sessionID], client)

	// Start client goroutines
	go client.handleRead()
	go client.handleWrite()
	go client.handlePing()

	// Start cleanup if not already running
	if !cm.cleanupRunning {
		go cm.startCleanup()
		cm.cleanupRunning = true
	}

	// Call connect handler
	if cm.onConnect != nil {
		go cm.onConnect(client)
	}

	log.Printf("Client connected: %s (Session: %s)", client.ID, sessionID)

	return client, nil
}

// RemoveClient removes a client from the manager
func (cm *ClientManager) RemoveClient(clientID string) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	client, exists := cm.clients[clientID]
	if !exists {
		return fmt.Errorf("client not found: %s", clientID)
	}

	// Update client state
	client.State = ClientStateDisconnected
	client.cancel()

	// Close connection
	client.Conn.Close()

	// Remove from storage
	delete(cm.clients, clientID)

	// Remove from session mapping
	if clients, exists := cm.sessions[client.SessionID]; exists {
		for i, c := range clients {
			if c.ID == clientID {
				cm.sessions[client.SessionID] = append(clients[:i], clients[i+1:]...)
				break
			}
		}
		if len(cm.sessions[client.SessionID]) == 0 {
			delete(cm.sessions, client.SessionID)
		}
	}

	// Remove from user mapping
	if client.UserID != nil {
		userID := client.UserID.String()
		if clients, exists := cm.users[userID]; exists {
			for i, c := range clients {
				if c.ID == clientID {
					cm.users[userID] = append(clients[:i], clients[i+1:]...)
					break
				}
			}
			if len(cm.users[userID]) == 0 {
				delete(cm.users, userID)
			}
		}
	}

	// Call disconnect handler
	if cm.onDisconnect != nil {
		go cm.onDisconnect(client)
	}

	log.Printf("Client disconnected: %s (Session: %s)", clientID, client.SessionID)

	return nil
}

// GetClient retrieves a client by ID
func (cm *ClientManager) GetClient(clientID string) (*ClientInfo, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	client, exists := cm.clients[clientID]
	return client, exists
}

// GetClientsBySession retrieves all clients for a session
func (cm *ClientManager) GetClientsBySession(sessionID string) []*ClientInfo {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	clients, exists := cm.sessions[sessionID]
	if !exists {
		return nil
	}

	// Return a copy to prevent external modifications
	result := make([]*ClientInfo, len(clients))
	copy(result, clients)
	return result
}

// GetClientsByUser retrieves all clients for a user
func (cm *ClientManager) GetClientsByUser(userID uuid.UUID) []*ClientInfo {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	clients, exists := cm.users[userID.String()]
	if !exists {
		return nil
	}

	// Return a copy to prevent external modifications
	result := make([]*ClientInfo, len(clients))
	copy(result, clients)
	return result
}

// BroadcastToSession broadcasts a message to all clients in a session
func (cm *ClientManager) BroadcastToSession(sessionID string, message *WebSocketMessage) error {
	cm.mu.RLock()
	clients := cm.sessions[sessionID]
	cm.mu.RUnlock()

	if clients == nil {
		return fmt.Errorf("no clients found for session: %s", sessionID)
	}

	var errors []error
	for _, client := range clients {
		if client.State == ClientStateConnected || client.State == ClientStateAuthenticated {
			if err := client.SendMessage(message); err != nil {
				errors = append(errors, fmt.Errorf("failed to send to client %s: %w", client.ID, err))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("broadcast errors: %v", errors)
	}

	return nil
}

// BroadcastToUser broadcasts a message to all clients for a user
func (cm *ClientManager) BroadcastToUser(userID uuid.UUID, message *WebSocketMessage) error {
	cm.mu.RLock()
	clients := cm.users[userID.String()]
	cm.mu.RUnlock()

	if clients == nil {
		return fmt.Errorf("no clients found for user: %s", userID)
	}

	var errors []error
	for _, client := range clients {
		if client.State == ClientStateConnected || client.State == ClientStateAuthenticated {
			if err := client.SendMessage(message); err != nil {
				errors = append(errors, fmt.Errorf("failed to send to client %s: %w", client.ID, err))
			}
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("broadcast errors: %v", errors)
	}

	return nil
}

// BroadcastToAll broadcasts a message to all connected clients
func (cm *ClientManager) BroadcastToAll(message *WebSocketMessage) error {
	cm.mu.RLock()
	clients := make([]*ClientInfo, 0, len(cm.clients))
	for _, client := range cm.clients {
		if client.State == ClientStateConnected || client.State == ClientStateAuthenticated {
			clients = append(clients, client)
		}
	}
	cm.mu.RUnlock()

	var errors []error
	for _, client := range clients {
		if err := client.SendMessage(message); err != nil {
			errors = append(errors, fmt.Errorf("failed to send to client %s: %w", client.ID, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("broadcast errors: %v", errors)
	}

	return nil
}

// GetClientCount returns the number of connected clients
func (cm *ClientManager) GetClientCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	count := 0
	for _, client := range cm.clients {
		if client.State == ClientStateConnected || client.State == ClientStateAuthenticated {
			count++
		}
	}
	return count
}

// GetSessionCount returns the number of active sessions
func (cm *ClientManager) GetSessionCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.sessions)
}

// GetUserCount returns the number of unique users
func (cm *ClientManager) GetUserCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.users)
}

// SetEventHandlers sets the event handlers
func (cm *ClientManager) SetEventHandlers(
	onConnect func(*ClientInfo),
	onDisconnect func(*ClientInfo),
	onMessage func(*ClientInfo, *WebSocketMessage),
	onError func(*ClientInfo, error),
) {
	cm.onConnect = onConnect
	cm.onDisconnect = onDisconnect
	cm.onMessage = onMessage
	cm.onError = onError
}

// handleRead handles reading messages from the WebSocket connection
func (c *ClientInfo) handleRead() {
	defer func() {
		c.CloseChan <- true
	}()

	c.Conn.SetReadLimit(c.MaxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(c.ReadTimeout))
	c.Conn.SetPongHandler(func(string) error {
		c.mu.Lock()
		c.LastActivity = time.Now()
		c.mu.Unlock()
		c.Conn.SetReadDeadline(time.Now().Add(c.ReadTimeout))
		return nil
	})

	for {
		select {
		case <-c.ctx.Done():
			return
		default:
			var msg WebSocketMessage
			err := c.Conn.ReadJSON(&msg)
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					log.Printf("WebSocket read error for client %s: %v", c.ID, err)
				}
				return
			}

			c.mu.Lock()
			c.LastActivity = time.Now()
			c.MessagesReceived++
			if jsonData, err := msg.ToJSON(); err == nil {
				c.BytesReceived += int64(len(jsonData))
			}
			c.mu.Unlock()

			c.Receive <- &msg
		}
	}
}

// handleWrite handles writing messages to the WebSocket connection
func (c *ClientInfo) handleWrite() {
	ticker := time.NewTicker(c.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case message := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(c.WriteTimeout))
			if err := c.Conn.WriteJSON(message); err != nil {
				log.Printf("WebSocket write error for client %s: %v", c.ID, err)
				return
			}

			c.mu.Lock()
			c.MessagesSent++
			if jsonData, err := message.ToJSON(); err == nil {
				c.BytesSent += int64(len(jsonData))
			}
			c.mu.Unlock()

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(c.WriteTimeout))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				log.Printf("WebSocket ping error for client %s: %v", c.ID, err)
				return
			}
		}
	}
}

// handlePing handles ping/pong for connection health
func (c *ClientInfo) handlePing() {
	ticker := time.NewTicker(c.PingInterval)
	defer ticker.Stop()

	for {
		select {
		case <-c.ctx.Done():
			return
		case <-ticker.C:
			c.mu.RLock()
			lastActivity := c.LastActivity
			c.mu.RUnlock()

			if time.Since(lastActivity) > c.PongTimeout {
				log.Printf("Client %s ping timeout", c.ID)
				c.CloseChan <- true
				return
			}
		}
	}
}

// SendMessage sends a message to the client
func (c *ClientInfo) SendMessage(message *WebSocketMessage) error {
	select {
	case c.Send <- message:
		return nil
	case <-c.ctx.Done():
		return fmt.Errorf("client context cancelled")
	default:
		return fmt.Errorf("client send channel full")
	}
}

// Authenticate authenticates the client
func (c *ClientInfo) Authenticate(userID uuid.UUID, authLevel AuthLevel, permissions []string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.UserID = &userID
	c.AuthLevel = authLevel
	c.Permissions = permissions
	c.State = ClientStateAuthenticated

	// Add to user mapping in client manager
	// This would need to be handled by the client manager
}

// UpdateActivity updates the last activity time
func (c *ClientInfo) UpdateActivity() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.LastActivity = time.Now()
}

// GetStats returns client statistics
func (c *ClientInfo) GetStats() map[string]interface{} {
	c.mu.RLock()
	defer c.mu.RUnlock()

	return map[string]interface{}{
		"id":                c.ID,
		"session_id":        c.SessionID,
		"user_id":           c.UserID,
		"state":             c.State,
		"connected_at":      c.ConnectedAt,
		"last_activity":     c.LastActivity,
		"messages_sent":     c.MessagesSent,
		"messages_received": c.MessagesReceived,
		"bytes_sent":        c.BytesSent,
		"bytes_received":    c.BytesReceived,
		"uptime_seconds":    time.Since(c.ConnectedAt).Seconds(),
	}
}

// startCleanup starts the cleanup routine
func (cm *ClientManager) startCleanup() {
	ticker := time.NewTicker(cm.cleanupInterval)
	defer ticker.Stop()

	for {
		select {
		case <-cm.ctx.Done():
			return
		case <-ticker.C:
			cm.cleanupInactiveClients()
		}
	}
}

// cleanupInactiveClients removes inactive clients
func (cm *ClientManager) cleanupInactiveClients() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	now := time.Now()
	var inactiveClients []string

	for clientID, client := range cm.clients {
		if now.Sub(client.LastActivity) > cm.clientTimeout {
			inactiveClients = append(inactiveClients, clientID)
		}
	}

	for _, clientID := range inactiveClients {
		client := cm.clients[clientID]
		client.State = ClientStateDisconnected
		client.cancel()
		client.Conn.Close()

		log.Printf("Cleaned up inactive client: %s", clientID)
	}

	if len(inactiveClients) > 0 {
		log.Printf("Cleaned up %d inactive clients", len(inactiveClients))
	}
}

// Stop stops the client manager
func (cm *ClientManager) Stop() {
	cm.cancel()

	cm.mu.Lock()
	defer cm.mu.Unlock()

	// Close all client connections
	for _, client := range cm.clients {
		client.cancel()
		client.Conn.Close()
	}
}
