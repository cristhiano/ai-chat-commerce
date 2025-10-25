package websocket

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// ConnectionManager manages WebSocket connections and cleanup
type ConnectionManager struct {
	// Active connections
	connections map[string]*ConnectionInfo
	
	// Mutex for thread-safe operations
	mu sync.RWMutex
	
	// Hub reference
	hub *Hub
	
	// Configuration
	maxConnections     int
	connectionTimeout  time.Duration
	cleanupInterval    time.Duration
	
	// Context for cancellation
	ctx context.Context
	cancel context.CancelFunc
	
	// Cleanup control
	cleanupRunning bool
}

// ConnectionInfo stores information about a WebSocket connection
type ConnectionInfo struct {
	ID          string
	Client      *Client
	SessionID   string
	UserID      *uuid.UUID
	IPAddress   string
	UserAgent   string
	ConnectedAt time.Time
	LastPing    time.Time
	IsActive    bool
	Metadata    map[string]interface{}
}

// NewConnectionManager creates a new connection manager
func NewConnectionManager(hub *Hub, maxConnections int, connectionTimeout, cleanupInterval time.Duration) *ConnectionManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &ConnectionManager{
		connections:       make(map[string]*ConnectionInfo),
		hub:              hub,
		maxConnections:   maxConnections,
		connectionTimeout: connectionTimeout,
		cleanupInterval:  cleanupInterval,
		ctx:              ctx,
		cancel:           cancel,
	}
}

// RegisterConnection registers a new WebSocket connection
func (cm *ConnectionManager) RegisterConnection(client *Client, sessionID, ipAddress, userAgent string, userID *uuid.UUID) (*ConnectionInfo, error) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	// Check connection limit
	if len(cm.connections) >= cm.maxConnections {
		return nil, ErrMaxConnectionsReached
	}
	
	// Create connection info
	connInfo := &ConnectionInfo{
		ID:          client.ID,
		Client:      client,
		SessionID:   sessionID,
		UserID:      userID,
		IPAddress:   ipAddress,
		UserAgent:   userAgent,
		ConnectedAt: time.Now(),
		LastPing:    time.Now(),
		IsActive:    true,
		Metadata:    make(map[string]interface{}),
	}
	
	// Store connection info
	cm.connections[client.ID] = connInfo
	
	// Start cleanup if not already running
	if !cm.cleanupRunning {
		go cm.startCleanup()
		cm.cleanupRunning = true
	}
	
	log.Printf("Connection registered: %s (Session: %s, IP: %s)", 
		client.ID, sessionID, ipAddress)
	
	return connInfo, nil
}

// UnregisterConnection unregisters a WebSocket connection
func (cm *ConnectionManager) UnregisterConnection(clientID string) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	if connInfo, exists := cm.connections[clientID]; exists {
		connInfo.IsActive = false
		delete(cm.connections, clientID)
		
		log.Printf("Connection unregistered: %s (Session: %s)", 
			clientID, connInfo.SessionID)
	}
}

// UpdateLastPing updates the last ping time for a connection
func (cm *ConnectionManager) UpdateLastPing(clientID string) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	if connInfo, exists := cm.connections[clientID]; exists {
		connInfo.LastPing = time.Now()
	}
}

// GetConnectionInfo retrieves connection information
func (cm *ConnectionManager) GetConnectionInfo(clientID string) (*ConnectionInfo, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	connInfo, exists := cm.connections[clientID]
	return connInfo, exists
}

// GetConnectionsBySession returns all connections for a specific session
func (cm *ConnectionManager) GetConnectionsBySession(sessionID string) []*ConnectionInfo {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	var connections []*ConnectionInfo
	for _, connInfo := range cm.connections {
		if connInfo.SessionID == sessionID && connInfo.IsActive {
			connections = append(connections, connInfo)
		}
	}
	return connections
}

// GetConnectionsByUser returns all connections for a specific user
func (cm *ConnectionManager) GetConnectionsByUser(userID uuid.UUID) []*ConnectionInfo {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	var connections []*ConnectionInfo
	for _, connInfo := range cm.connections {
		if connInfo.UserID != nil && *connInfo.UserID == userID && connInfo.IsActive {
			connections = append(connections, connInfo)
		}
	}
	return connections
}

// GetActiveConnectionCount returns the number of active connections
func (cm *ConnectionManager) GetActiveConnectionCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	count := 0
	for _, connInfo := range cm.connections {
		if connInfo.IsActive {
			count++
		}
	}
	return count
}

// GetTotalConnectionCount returns the total number of connections (including inactive)
func (cm *ConnectionManager) GetTotalConnectionCount() int {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return len(cm.connections)
}

// startCleanup starts the cleanup routine
func (cm *ConnectionManager) startCleanup() {
	ticker := time.NewTicker(cm.cleanupInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-cm.ctx.Done():
			return
		case <-ticker.C:
			cm.cleanupStaleConnections()
		}
	}
}

// cleanupStaleConnections removes stale connections
func (cm *ConnectionManager) cleanupStaleConnections() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	cutoff := time.Now().Add(-cm.connectionTimeout)
	var staleConnections []string
	
	for clientID, connInfo := range cm.connections {
		if connInfo.LastPing.Before(cutoff) {
			staleConnections = append(staleConnections, clientID)
		}
	}
	
	for _, clientID := range staleConnections {
		connInfo := cm.connections[clientID]
		connInfo.IsActive = false
		
		// Close the WebSocket connection
		if connInfo.Client != nil {
			connInfo.Client.Close()
		}
		
		delete(cm.connections, clientID)
		
		log.Printf("Cleaned up stale connection: %s (Session: %s, Last ping: %v)", 
			clientID, connInfo.SessionID, connInfo.LastPing)
	}
	
	if len(staleConnections) > 0 {
		log.Printf("Cleaned up %d stale connections", len(staleConnections))
	}
}

// GetConnectionStats returns connection statistics
func (cm *ConnectionManager) GetConnectionStats() map[string]interface{} {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	activeCount := 0
	sessionCount := make(map[string]int)
	userCount := make(map[string]int)
	
	for _, connInfo := range cm.connections {
		if connInfo.IsActive {
			activeCount++
			sessionCount[connInfo.SessionID]++
			
			if connInfo.UserID != nil {
				userCount[connInfo.UserID.String()]++
			}
		}
	}
	
	return map[string]interface{}{
		"total_connections":    len(cm.connections),
		"active_connections":   activeCount,
		"unique_sessions":      len(sessionCount),
		"unique_users":         len(userCount),
		"max_connections":      cm.maxConnections,
		"connection_timeout_s": cm.connectionTimeout.Seconds(),
		"cleanup_interval_s":   cm.cleanupInterval.Seconds(),
	}
}

// CloseAllConnections closes all active connections
func (cm *ConnectionManager) CloseAllConnections() {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	for clientID, connInfo := range cm.connections {
		if connInfo.IsActive && connInfo.Client != nil {
			connInfo.Client.Close()
			connInfo.IsActive = false
			
			log.Printf("Closed connection: %s", clientID)
		}
	}
	
	log.Printf("Closed all %d connections", len(cm.connections))
}

// Stop stops the connection manager
func (cm *ConnectionManager) Stop() {
	cm.cancel()
	cm.CloseAllConnections()
}

// SetConnectionMetadata sets metadata for a connection
func (cm *ConnectionManager) SetConnectionMetadata(clientID string, key string, value interface{}) {
	cm.mu.Lock()
	defer cm.mu.Unlock()
	
	if connInfo, exists := cm.connections[clientID]; exists {
		connInfo.Metadata[key] = value
	}
}

// GetConnectionMetadata retrieves metadata for a connection
func (cm *ConnectionManager) GetConnectionMetadata(clientID string, key string) (interface{}, bool) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	if connInfo, exists := cm.connections[clientID]; exists {
		value, exists := connInfo.Metadata[key]
		return value, exists
	}
	return nil, false
}

// BroadcastToConnections broadcasts a message to specific connections
func (cm *ConnectionManager) BroadcastToConnections(clientIDs []string, message Message) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	for _, clientID := range clientIDs {
		if connInfo, exists := cm.connections[clientID]; exists && connInfo.IsActive {
			connInfo.Client.SendMessage(message)
		}
	}
}

// GetConnectionsByIP returns all connections from a specific IP address
func (cm *ConnectionManager) GetConnectionsByIP(ipAddress string) []*ConnectionInfo {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	var connections []*ConnectionInfo
	for _, connInfo := range cm.connections {
		if connInfo.IPAddress == ipAddress && connInfo.IsActive {
			connections = append(connections, connInfo)
		}
	}
	return connections
}

// GetConnectionsByUserAgent returns all connections with a specific user agent
func (cm *ConnectionManager) GetConnectionsByUserAgent(userAgent string) []*ConnectionInfo {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	
	var connections []*ConnectionInfo
	for _, connInfo := range cm.connections {
		if connInfo.UserAgent == userAgent && connInfo.IsActive {
			connections = append(connections, connInfo)
		}
	}
	return connections
}

// Custom errors
var (
	ErrMaxConnectionsReached = &ConnectionError{
		Code:    "MAX_CONNECTIONS_REACHED",
		Message: "Maximum number of connections reached",
	}
)

// ConnectionError represents a connection-related error
type ConnectionError struct {
	Code    string
	Message string
}

func (e *ConnectionError) Error() string {
	return e.Message
}
