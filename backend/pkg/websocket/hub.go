package websocket

import (
	"chat-ecommerce-backend/pkg/redis"
	"context"
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Hub maintains the set of active clients and broadcasts messages
type Hub struct {
	// Registered clients
	clients map[*Client]bool

	// Inbound messages from clients
	broadcast chan []byte

	// Register requests from clients
	register chan *Client

	// Unregister requests from clients
	unregister chan *Client

	// Redis client for pub/sub
	redisClient *redis.Client

	// Mutex for thread safety
	mutex sync.RWMutex
}

// Client represents a websocket client
type Client struct {
	hub *Hub

	// The websocket connection
	conn *websocket.Conn

	// Buffered channel of outbound messages
	send chan []byte

	// Client session ID
	sessionID string

	// User ID (if authenticated)
	userID string

	// Client channels for different types of messages
	channels []string
}

// Message represents a websocket message
type Message struct {
	Type      string      `json:"type"`
	Data      interface{} `json:"data"`
	SessionID string      `json:"session_id,omitempty"`
	UserID    string      `json:"user_id,omitempty"`
	Timestamp time.Time   `json:"timestamp"`
}

// NewHub creates a new websocket hub
func NewHub(redisClient *redis.Client) *Hub {
	return &Hub{
		clients:     make(map[*Client]bool),
		broadcast:   make(chan []byte),
		register:    make(chan *Client),
		unregister:  make(chan *Client),
		redisClient: redisClient,
	}
}

// Run starts the hub
func (h *Hub) Run() {
	// Start Redis pub/sub for cross-instance communication
	go h.startRedisSubscriber()

	for {
		select {
		case client := <-h.register:
			h.registerClient(client)

		case client := <-h.unregister:
			h.unregisterClient(client)

		case message := <-h.broadcast:
			h.broadcastMessage(message)
		}
	}
}

// RegisterClient registers a new client
func (h *Hub) RegisterClient(conn *websocket.Conn, sessionID, userID string) *Client {
	client := &Client{
		hub:       h,
		conn:      conn,
		send:      make(chan []byte, 256),
		sessionID: sessionID,
		userID:    userID,
		channels:  []string{"general"},
	}

	h.register <- client
	return client
}

// registerClient handles client registration
func (h *Hub) registerClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	h.clients[client] = true
	log.Printf("Client registered: %s (User: %s)", client.sessionID, client.userID)

	// Send welcome message
	welcomeMsg := Message{
		Type:      "welcome",
		Data:      map[string]string{"message": "Connected to chat ecommerce"},
		SessionID: client.sessionID,
		UserID:    client.userID,
		Timestamp: time.Now(),
	}

	if data, err := json.Marshal(welcomeMsg); err == nil {
		select {
		case client.send <- data:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

// unregisterClient handles client unregistration
func (h *Hub) unregisterClient(client *Client) {
	h.mutex.Lock()
	defer h.mutex.Unlock()

	if _, ok := h.clients[client]; ok {
		delete(h.clients, client)
		close(client.send)
		log.Printf("Client unregistered: %s", client.sessionID)
	}
}

// broadcastMessage broadcasts a message to all clients
func (h *Hub) broadcastMessage(message []byte) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for client := range h.clients {
		select {
		case client.send <- message:
		default:
			close(client.send)
			delete(h.clients, client)
		}
	}
}

// BroadcastToSession broadcasts a message to a specific session
func (h *Hub) BroadcastToSession(sessionID string, message Message) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	message.SessionID = sessionID
	message.Timestamp = time.Now()

	if data, err := json.Marshal(message); err == nil {
		for client := range h.clients {
			if client.sessionID == sessionID {
				select {
				case client.send <- data:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// BroadcastToUser broadcasts a message to a specific user
func (h *Hub) BroadcastToUser(userID string, message Message) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	message.UserID = userID
	message.Timestamp = time.Now()

	if data, err := json.Marshal(message); err == nil {
		for client := range h.clients {
			if client.userID == userID {
				select {
				case client.send <- data:
				default:
					close(client.send)
					delete(h.clients, client)
				}
			}
		}
	}
}

// BroadcastToChannel broadcasts a message to clients subscribed to a channel
func (h *Hub) BroadcastToChannel(channel string, message Message) {
	h.mutex.RLock()
	defer h.mutex.RUnlock()

	message.Timestamp = time.Now()

	if data, err := json.Marshal(message); err == nil {
		for client := range h.clients {
			for _, clientChannel := range client.channels {
				if clientChannel == channel {
					select {
					case client.send <- data:
					default:
						close(client.send)
						delete(h.clients, client)
					}
					break
				}
			}
		}
	}
}

// GetClientCount returns the number of connected clients
func (h *Hub) GetClientCount() int {
	h.mutex.RLock()
	defer h.mutex.RUnlock()
	return len(h.clients)
}

// startRedisSubscriber starts Redis pub/sub for cross-instance communication
func (h *Hub) startRedisSubscriber() {
	ctx := context.Background()
	pubsub := h.redisClient.Subscribe(ctx, "websocket_broadcast")

	for {
		msg, err := pubsub.ReceiveMessage(ctx)
		if err != nil {
			log.Printf("Redis pub/sub error: %v", err)
			continue
		}

		// Broadcast the message to all connected clients
		h.broadcast <- []byte(msg.Payload)
	}
}

// PublishToRedis publishes a message to Redis for cross-instance broadcasting
func (h *Hub) PublishToRedis(message Message) error {
	ctx := context.Background()
	data, err := json.Marshal(message)
	if err != nil {
		return err
	}

	return h.redisClient.Publish(ctx, "websocket_broadcast", data).Err()
}

// Client methods

// ReadPump pumps messages from the websocket connection to the hub
func (c *Client) ReadPump() {
	defer func() {
		c.hub.unregister <- c
		c.conn.Close()
	}()

	c.conn.SetReadLimit(512)
	c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.conn.SetPongHandler(func(string) error {
		c.conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket error: %v", err)
			}
			break
		}

		// Process the message
		var msg Message
		if err := json.Unmarshal(message, &msg); err == nil {
			msg.SessionID = c.sessionID
			msg.UserID = c.userID
			msg.Timestamp = time.Now()

			// Handle different message types
			c.handleMessage(msg)
		}
	}
}

// WritePump pumps messages from the hub to the websocket connection
func (c *Client) WritePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Add queued chat messages to the current websocket message
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write([]byte{'\n'})
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// handleMessage processes incoming messages
func (c *Client) handleMessage(msg Message) {
	switch msg.Type {
	case "ping":
		// Respond with pong
		pongMsg := Message{
			Type:      "pong",
			Data:      map[string]string{"message": "pong"},
			SessionID: c.sessionID,
			UserID:    c.userID,
			Timestamp: time.Now(),
		}
		if data, err := json.Marshal(pongMsg); err == nil {
			c.send <- data
		}

	case "subscribe":
		// Subscribe to a channel
		if channel, ok := msg.Data.(string); ok {
			c.subscribeToChannel(channel)
		}

	case "unsubscribe":
		// Unsubscribe from a channel
		if channel, ok := msg.Data.(string); ok {
			c.unsubscribeFromChannel(channel)
		}

	default:
		// Broadcast the message to all clients
		if data, err := json.Marshal(msg); err == nil {
			c.hub.broadcast <- data
		}
	}
}

// subscribeToChannel subscribes the client to a channel
func (c *Client) subscribeToChannel(channel string) {
	for _, ch := range c.channels {
		if ch == channel {
			return // Already subscribed
		}
	}
	c.channels = append(c.channels, channel)
}

// unsubscribeFromChannel unsubscribes the client from a channel
func (c *Client) unsubscribeFromChannel(channel string) {
	for i, ch := range c.channels {
		if ch == channel {
			c.channels = append(c.channels[:i], c.channels[i+1:]...)
			break
		}
	}
}
