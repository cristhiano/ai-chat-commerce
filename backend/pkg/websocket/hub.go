package websocket

import (
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients
	clients map[string]*Client

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Broadcast messages to all clients
	broadcast chan *WebSocketMessage

	// Broadcast to specific session
	sessionBroadcast chan SessionMessage

	// Broadcast to specific user
	userBroadcast chan UserMessage

	// Mutex for thread-safe operations
	mu sync.RWMutex
}

// SessionMessage represents a message targeted to a specific session
type SessionMessage struct {
	SessionID string
	Message   *WebSocketMessage
}

// UserMessage represents a message targeted to a specific user
type UserMessage struct {
	UserID  uuid.UUID
	Message *WebSocketMessage
}

// NewHub creates a new WebSocket hub
func NewHub() *Hub {
	return &Hub{
		clients:          make(map[string]*Client),
		register:         make(chan *Client),
		unregister:       make(chan *Client),
		broadcast:        make(chan *WebSocketMessage),
		sessionBroadcast: make(chan SessionMessage),
		userBroadcast:    make(chan UserMessage),
	}
}

// Run starts the hub and handles client connections
func (h *Hub) Run() {
	ticker := time.NewTicker(30 * time.Second) // Ping interval
	defer ticker.Stop()

	for {
		select {
		case client := <-h.register:
			h.mu.Lock()
			h.clients[client.ID] = client
			h.mu.Unlock()
			log.Printf("Client %s registered. Total clients: %d", client.ID, len(h.clients))

		case client := <-h.unregister:
			h.mu.Lock()
			if _, ok := h.clients[client.ID]; ok {
				delete(h.clients, client.ID)
				close(client.Send)
			}
			h.mu.Unlock()
			log.Printf("Client %s unregistered. Total clients: %d", client.ID, len(h.clients))

		case message := <-h.broadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client.ID)
				}
			}
			h.mu.RUnlock()

		case sessionMsg := <-h.sessionBroadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				if client.SessionID == sessionMsg.SessionID {
					select {
					case client.Send <- sessionMsg.Message:
					default:
						close(client.Send)
						delete(h.clients, client.ID)
					}
				}
			}
			h.mu.RUnlock()

		case userMsg := <-h.userBroadcast:
			h.mu.RLock()
			for _, client := range h.clients {
				if client.UserID != nil && *client.UserID == userMsg.UserID {
					select {
					case client.Send <- userMsg.Message:
					default:
						close(client.Send)
						delete(h.clients, client.ID)
					}
				}
			}
			h.mu.RUnlock()

		case <-ticker.C:
			// Send ping to all clients
			h.mu.RLock()
			for _, client := range h.clients {
				pingMsg := NewMessageFactory().CreatePing(0)
				select {
				case client.Send <- pingMsg:
				default:
					close(client.Send)
					delete(h.clients, client.ID)
				}
			}
			h.mu.RUnlock()
		}
	}
}

// RegisterClient registers a new client with the hub
func (h *Hub) RegisterClient(client *Client) {
	h.register <- client
}

// UnregisterClient unregisters a client from the hub
func (h *Hub) UnregisterClient(client *Client) {
	h.unregister <- client
}

// BroadcastMessage broadcasts a message to all connected clients
func (h *Hub) BroadcastMessage(message *WebSocketMessage) {
	h.broadcast <- message
}

// BroadcastToSession broadcasts a message to all clients in a specific session
func (h *Hub) BroadcastToSession(sessionID string, message *WebSocketMessage) {
	h.sessionBroadcast <- SessionMessage{
		SessionID: sessionID,
		Message:   message,
	}
}

// BroadcastToUser broadcasts a message to all clients for a specific user
func (h *Hub) BroadcastToUser(userID uuid.UUID, message *WebSocketMessage) {
	h.userBroadcast <- UserMessage{
		UserID:  userID,
		Message: message,
	}
}

// GetClientCount returns the number of connected clients
func (h *Hub) GetClientCount() int {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return len(h.clients)
}

// GetClientsBySession returns all clients for a specific session
func (h *Hub) GetClientsBySession(sessionID string) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	var clients []*Client
	for _, client := range h.clients {
		if client.SessionID == sessionID {
			clients = append(clients, client)
		}
	}
	return clients
}

// GetClientsByUser returns all clients for a specific user
func (h *Hub) GetClientsByUser(userID uuid.UUID) []*Client {
	h.mu.RLock()
	defer h.mu.RUnlock()
	
	var clients []*Client
	for _, client := range h.clients {
		if client.UserID != nil && *client.UserID == userID {
			clients = append(clients, client)
		}
	}
	return clients
}

// Stop stops the hub
func (h *Hub) Stop() {
	h.mu.Lock()
	defer h.mu.Unlock()
	
	// Close all client connections
	for _, client := range h.clients {
		close(client.Send)
	}
	
	// Clear clients map
	h.clients = make(map[string]*Client)
}
