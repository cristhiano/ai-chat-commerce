package websocket

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// MessageType represents the type of WebSocket message
type MessageType string

const (
	// Connection messages
	MessageTypeConnect    MessageType = "connect"
	MessageTypeDisconnect MessageType = "disconnect"
	MessageTypePing       MessageType = "ping"
	MessageTypePong       MessageType = "pong"

	// Authentication messages
	MessageTypeAuth        MessageType = "auth"
	MessageTypeAuthSuccess MessageType = "auth_success"
	MessageTypeAuthError   MessageType = "auth_error"

	// Chat messages
	MessageTypeChatMessage  MessageType = "chat_message"
	MessageTypeChatResponse MessageType = "chat_response"
	MessageTypeChatTyping   MessageType = "chat_typing"

	// Cart messages
	MessageTypeCartUpdate MessageType = "cart_update"
	MessageTypeCartSync   MessageType = "cart_sync"
	MessageTypeCartAdd    MessageType = "cart_add"
	MessageTypeCartRemove MessageType = "cart_remove"
	MessageTypeCartClear  MessageType = "cart_clear"

	// Inventory messages
	MessageTypeInventoryUpdate MessageType = "inventory_update"
	MessageTypeInventoryAlert  MessageType = "inventory_alert"
	MessageTypeInventorySync   MessageType = "inventory_sync"

	// Order messages
	MessageTypeOrderUpdate    MessageType = "order_update"
	MessageTypeOrderStatus    MessageType = "order_status"
	MessageTypeOrderCreated   MessageType = "order_created"
	MessageTypeOrderCompleted MessageType = "order_completed"

	// Notification messages
	MessageTypeNotification MessageType = "notification"
	MessageTypeSystemAlert  MessageType = "system_alert"
	MessageTypeUserAlert    MessageType = "user_alert"

	// Error messages
	MessageTypeError           MessageType = "error"
	MessageTypeValidationError MessageType = "validation_error"
)

// MessagePriority represents the priority of a message
type MessagePriority int

const (
	PriorityLow MessagePriority = iota
	PriorityNormal
	PriorityHigh
	PriorityCritical
)

// WebSocketMessage represents a WebSocket message
type WebSocketMessage struct {
	ID          string                 `json:"id"`
	Type        MessageType            `json:"type"`
	Priority    MessagePriority        `json:"priority"`
	Timestamp   time.Time              `json:"timestamp"`
	SessionID   string                 `json:"session_id,omitempty"`
	UserID      *uuid.UUID             `json:"user_id,omitempty"`
	Channel     string                 `json:"channel,omitempty"`
	Data        map[string]interface{} `json:"data"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	RequiresAck bool                   `json:"requires_ack,omitempty"`
	AckID       string                 `json:"ack_id,omitempty"`
}

// NewWebSocketMessage creates a new WebSocket message
func NewWebSocketMessage(msgType MessageType, data map[string]interface{}) *WebSocketMessage {
	return &WebSocketMessage{
		ID:        uuid.New().String(),
		Type:      msgType,
		Priority:  PriorityNormal,
		Timestamp: time.Now(),
		Data:      data,
		Metadata:  make(map[string]interface{}),
	}
}

// NewWebSocketMessageWithPriority creates a new WebSocket message with priority
func NewWebSocketMessageWithPriority(msgType MessageType, priority MessagePriority, data map[string]interface{}) *WebSocketMessage {
	return &WebSocketMessage{
		ID:        uuid.New().String(),
		Type:      msgType,
		Priority:  priority,
		Timestamp: time.Now(),
		Data:      data,
		Metadata:  make(map[string]interface{}),
	}
}

// SetSession sets the session ID for the message
func (msg *WebSocketMessage) SetSession(sessionID string) *WebSocketMessage {
	msg.SessionID = sessionID
	return msg
}

// SetUser sets the user ID for the message
func (msg *WebSocketMessage) SetUser(userID uuid.UUID) *WebSocketMessage {
	msg.UserID = &userID
	return msg
}

// SetChannel sets the channel for the message
func (msg *WebSocketMessage) SetChannel(channel string) *WebSocketMessage {
	msg.Channel = channel
	return msg
}

// SetRequiresAck marks the message as requiring acknowledgment
func (msg *WebSocketMessage) SetRequiresAck(ackID string) *WebSocketMessage {
	msg.RequiresAck = true
	msg.AckID = ackID
	return msg
}

// AddMetadata adds metadata to the message
func (msg *WebSocketMessage) AddMetadata(key string, value interface{}) *WebSocketMessage {
	if msg.Metadata == nil {
		msg.Metadata = make(map[string]interface{})
	}
	msg.Metadata[key] = value
	return msg
}

// ToJSON converts the message to JSON bytes
func (msg *WebSocketMessage) ToJSON() ([]byte, error) {
	return json.Marshal(msg)
}

// FromJSON creates a message from JSON bytes
func FromJSON(data []byte) (*WebSocketMessage, error) {
	var msg WebSocketMessage
	if err := json.Unmarshal(data, &msg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal message: %w", err)
	}
	return &msg, nil
}

// ChatMessageData represents chat message data
type ChatMessageData struct {
	MessageID   string                 `json:"message_id"`
	Content     string                 `json:"content"`
	MessageType string                 `json:"message_type"` // "user", "assistant", "system"
	Timestamp   time.Time              `json:"timestamp"`
	SessionID   string                 `json:"session_id"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// CartUpdateData represents cart update data
type CartUpdateData struct {
	SessionID string                 `json:"session_id"`
	UserID    *uuid.UUID             `json:"user_id,omitempty"`
	Items     []CartItemData         `json:"items"`
	Total     float64                `json:"total"`
	Currency  string                 `json:"currency"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// CartItemData represents cart item data
type CartItemData struct {
	ProductID uuid.UUID `json:"product_id"`
	Quantity  int       `json:"quantity"`
	Price     float64   `json:"price"`
	Name      string    `json:"name"`
	ImageURL  string    `json:"image_url,omitempty"`
}

// InventoryUpdateData represents inventory update data
type InventoryUpdateData struct {
	ProductID   uuid.UUID              `json:"product_id"`
	Quantity    int                    `json:"quantity"`
	Reserved    int                    `json:"reserved"`
	Available   int                    `json:"available"`
	Location    string                 `json:"location,omitempty"`
	LastUpdated time.Time              `json:"last_updated"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
}

// OrderUpdateData represents order update data
type OrderUpdateData struct {
	OrderID   uuid.UUID              `json:"order_id"`
	UserID    uuid.UUID              `json:"user_id"`
	Status    string                 `json:"status"`
	Total     float64                `json:"total"`
	Currency  string                 `json:"currency"`
	UpdatedAt time.Time              `json:"updated_at"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// NotificationData represents notification data
type NotificationData struct {
	NotificationID string                 `json:"notification_id"`
	Title          string                 `json:"title"`
	Message        string                 `json:"message"`
	Type           string                 `json:"type"` // "info", "warning", "error", "success"
	Priority       MessagePriority        `json:"priority"`
	UserID         *uuid.UUID             `json:"user_id,omitempty"`
	SessionID      string                 `json:"session_id,omitempty"`
	Channel        string                 `json:"channel,omitempty"`
	ExpiresAt      *time.Time             `json:"expires_at,omitempty"`
	Actions        []NotificationAction   `json:"actions,omitempty"`
	Metadata       map[string]interface{} `json:"metadata,omitempty"`
}

// NotificationAction represents a notification action
type NotificationAction struct {
	ID    string `json:"id"`
	Label string `json:"label"`
	Type  string `json:"type"` // "button", "link"
	URL   string `json:"url,omitempty"`
}

// ErrorData represents error data
type ErrorData struct {
	ErrorID   string                 `json:"error_id"`
	Code      string                 `json:"code"`
	Message   string                 `json:"message"`
	Details   map[string]interface{} `json:"details,omitempty"`
	Timestamp time.Time              `json:"timestamp"`
	SessionID string                 `json:"session_id,omitempty"`
	UserID    *uuid.UUID             `json:"user_id,omitempty"`
}

// AuthData represents authentication data
type AuthData struct {
	Token     string                 `json:"token"`
	SessionID string                 `json:"session_id"`
	UserID    *uuid.UUID             `json:"user_id,omitempty"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
}

// PingData represents ping data
type PingData struct {
	Timestamp time.Time `json:"timestamp"`
	Sequence  int       `json:"sequence"`
}

// PongData represents pong data
type PongData struct {
	Timestamp time.Time `json:"timestamp"`
	Sequence  int       `json:"sequence"`
	Latency   int64     `json:"latency_ms"`
}

// MessageBuilder provides a fluent interface for building messages
type MessageBuilder struct {
	msg *WebSocketMessage
}

// NewMessageBuilder creates a new message builder
func NewMessageBuilder(msgType MessageType) *MessageBuilder {
	return &MessageBuilder{
		msg: NewWebSocketMessage(msgType, make(map[string]interface{})),
	}
}

// WithPriority sets the message priority
func (mb *MessageBuilder) WithPriority(priority MessagePriority) *MessageBuilder {
	mb.msg.Priority = priority
	return mb
}

// WithSession sets the session ID
func (mb *MessageBuilder) WithSession(sessionID string) *MessageBuilder {
	mb.msg.SetSession(sessionID)
	return mb
}

// WithUser sets the user ID
func (mb *MessageBuilder) WithUser(userID uuid.UUID) *MessageBuilder {
	mb.msg.SetUser(userID)
	return mb
}

// WithChannel sets the channel
func (mb *MessageBuilder) WithChannel(channel string) *MessageBuilder {
	mb.msg.SetChannel(channel)
	return mb
}

// WithData sets the message data
func (mb *MessageBuilder) WithData(data map[string]interface{}) *MessageBuilder {
	mb.msg.Data = data
	return mb
}

// WithDataField sets a specific data field
func (mb *MessageBuilder) WithDataField(key string, value interface{}) *MessageBuilder {
	if mb.msg.Data == nil {
		mb.msg.Data = make(map[string]interface{})
	}
	mb.msg.Data[key] = value
	return mb
}

// WithMetadata sets metadata
func (mb *MessageBuilder) WithMetadata(metadata map[string]interface{}) *MessageBuilder {
	mb.msg.Metadata = metadata
	return mb
}

// WithMetadataField sets a specific metadata field
func (mb *MessageBuilder) WithMetadataField(key string, value interface{}) *MessageBuilder {
	mb.msg.AddMetadata(key, value)
	return mb
}

// RequireAck marks the message as requiring acknowledgment
func (mb *MessageBuilder) RequireAck(ackID string) *MessageBuilder {
	mb.msg.SetRequiresAck(ackID)
	return mb
}

// Build returns the constructed message
func (mb *MessageBuilder) Build() *WebSocketMessage {
	return mb.msg
}

// MessageFactory provides factory methods for common message types
type MessageFactory struct{}

// NewMessageFactory creates a new message factory
func NewMessageFactory() *MessageFactory {
	return &MessageFactory{}
}

// CreateChatMessage creates a chat message
func (mf *MessageFactory) CreateChatMessage(content string, messageType string, sessionID string) *WebSocketMessage {
	return NewMessageBuilder(MessageTypeChatMessage).
		WithSession(sessionID).
		WithDataField("content", content).
		WithDataField("message_type", messageType).
		WithDataField("timestamp", time.Now()).
		Build()
}

// CreateCartUpdate creates a cart update message
func (mf *MessageFactory) CreateCartUpdate(sessionID string, userID *uuid.UUID, items []CartItemData, total float64) *WebSocketMessage {
	data := map[string]interface{}{
		"session_id": sessionID,
		"items":      items,
		"total":      total,
		"currency":   "USD",
	}

	if userID != nil {
		data["user_id"] = userID
	}

	return NewMessageBuilder(MessageTypeCartUpdate).
		WithSession(sessionID).
		WithData(data).
		Build()
}

// CreateInventoryUpdate creates an inventory update message
func (mf *MessageFactory) CreateInventoryUpdate(productID uuid.UUID, quantity, reserved, available int, location string) *WebSocketMessage {
	return NewMessageBuilder(MessageTypeInventoryUpdate).
		WithPriority(PriorityHigh).
		WithDataField("product_id", productID).
		WithDataField("quantity", quantity).
		WithDataField("reserved", reserved).
		WithDataField("available", available).
		WithDataField("location", location).
		WithDataField("last_updated", time.Now()).
		Build()
}

// CreateNotification creates a notification message
func (mf *MessageFactory) CreateNotification(title, message, notificationType string, priority MessagePriority) *WebSocketMessage {
	return NewMessageBuilder(MessageTypeNotification).
		WithPriority(priority).
		WithDataField("title", title).
		WithDataField("message", message).
		WithDataField("type", notificationType).
		WithDataField("priority", priority).
		WithDataField("timestamp", time.Now()).
		Build()
}

// CreateError creates an error message
func (mf *MessageFactory) CreateError(code, message string, sessionID string, userID *uuid.UUID) *WebSocketMessage {
	return NewMessageBuilder(MessageTypeError).
		WithPriority(PriorityHigh).
		WithSession(sessionID).
		WithDataField("code", code).
		WithDataField("message", message).
		WithDataField("timestamp", time.Now()).
		Build()
}

// CreatePing creates a ping message
func (mf *MessageFactory) CreatePing(sequence int) *WebSocketMessage {
	return NewMessageBuilder(MessageTypePing).
		WithDataField("timestamp", time.Now()).
		WithDataField("sequence", sequence).
		Build()
}

// CreateCartUpdateMessage creates a cart update message
func CreateCartUpdateMessage(cartData interface{}, sessionID string, userID *uuid.UUID) *WebSocketMessage {
	return NewMessageBuilder(MessageTypeCartUpdate).
		WithSession(sessionID).
		WithDataField("cart_data", cartData).
		Build()
}

// CreateInventoryUpdateMessage creates an inventory update message
func CreateInventoryUpdateMessage(inventoryData interface{}, sessionID string, userID *uuid.UUID) *WebSocketMessage {
	return NewMessageBuilder(MessageTypeInventoryUpdate).
		WithSession(sessionID).
		WithDataField("inventory_data", inventoryData).
		Build()
}

// CreateNotificationMessage creates a notification message
func CreateNotificationMessage(notificationData interface{}, sessionID string, userID *uuid.UUID) *WebSocketMessage {
	return NewMessageBuilder(MessageTypeNotification).
		WithSession(sessionID).
		WithDataField("notification_data", notificationData).
		Build()
}
