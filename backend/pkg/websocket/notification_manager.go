package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// NotificationType represents the type of notification
type NotificationType string

const (
	NotificationTypeInfo     NotificationType = "info"
	NotificationTypeSuccess  NotificationType = "success"
	NotificationTypeWarning  NotificationType = "warning"
	NotificationTypeError    NotificationType = "error"
	NotificationTypeAlert    NotificationType = "alert"
)

// NotificationPriority represents the priority of a notification
type NotificationPriority int

const (
	PriorityLow    NotificationPriority = 1
	PriorityMedium NotificationPriority = 5
	PriorityHigh   NotificationPriority = 8
	PriorityCritical NotificationPriority = 10
)

// Notification represents a real-time notification
type Notification struct {
	ID          string                 `json:"id"`
	Type        NotificationType       `json:"type"`
	Title       string                 `json:"title"`
	Message     string                 `json:"message"`
	Priority    NotificationPriority   `json:"priority"`
	Category    string                 `json:"category"`
	ActionURL   string                 `json:"action_url,omitempty"`
	ActionText  string                 `json:"action_text,omitempty"`
	ExpiresAt   *time.Time             `json:"expires_at,omitempty"`
	CreatedAt   time.Time              `json:"created_at"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// NotificationManager manages real-time notifications
type NotificationManager struct {
	// Hub for broadcasting notifications
	hub *Hub
	
	// Message queue for reliable delivery
	queue *MessageQueue
	
	// Active notifications cache
	activeNotifications map[string]*Notification
	
	// User notification preferences
	userPreferences map[string]*NotificationPreferences
	
	// Mutex for thread-safe operations
	mu sync.RWMutex
	
	// Configuration
	defaultExpiry time.Duration
	maxNotifications int
}

// NotificationPreferences represents user notification preferences
type NotificationPreferences struct {
	UserID              uuid.UUID            `json:"user_id"`
	EnabledCategories   map[string]bool      `json:"enabled_categories"`
	EnabledTypes        map[NotificationType]bool `json:"enabled_types"`
	MinPriority         NotificationPriority `json:"min_priority"`
	QuietHoursStart     *time.Time           `json:"quiet_hours_start,omitempty"`
	QuietHoursEnd       *time.Time           `json:"quiet_hours_end,omitempty"`
	MaxNotifications    int                  `json:"max_notifications"`
	Metadata            map[string]interface{} `json:"metadata"`
}

// NewNotificationManager creates a new notification manager
func NewNotificationManager(hub *Hub, queue *MessageQueue, defaultExpiry time.Duration, maxNotifications int) *NotificationManager {
	return &NotificationManager{
		hub:                 hub,
		queue:               queue,
		activeNotifications: make(map[string]*Notification),
		userPreferences:     make(map[string]*NotificationPreferences),
		defaultExpiry:       defaultExpiry,
		maxNotifications:    maxNotifications,
	}
}

// SendNotification sends a notification to specific targets
func (nm *NotificationManager) SendNotification(notification *Notification, targets NotificationTargets) error {
	// Set default values
	if notification.ID == "" {
		notification.ID = uuid.New().String()
	}
	if notification.CreatedAt.IsZero() {
		notification.CreatedAt = time.Now()
	}
	if notification.ExpiresAt == nil {
		expiry := time.Now().Add(nm.defaultExpiry)
		notification.ExpiresAt = &expiry
	}
	
	// Store active notification
	nm.mu.Lock()
	nm.activeNotifications[notification.ID] = notification
	nm.mu.Unlock()
	
	// Create notification message
	notificationData := map[string]interface{}{
		"notification": notification,
		"targets":      targets,
	}
	
	message := CreateNotificationMessage(notificationData, "", nil)
	
	// Send to targets
	if targets.BroadcastToAll {
		nm.hub.BroadcastMessage(message)
	}
	
	if targets.SessionID != "" {
		nm.hub.BroadcastToSession(targets.SessionID, message)
	}
	
	if targets.UserID != nil {
		nm.hub.BroadcastToUser(*targets.UserID, message)
	}
	
	// Queue for reliable delivery
	err := nm.queue.EnqueueNotification(notificationData, targets.SessionID, targets.UserID, int(notification.Priority))
	if err != nil {
		log.Printf("Failed to queue notification: %v", err)
	}
	
	log.Printf("Notification sent: %s (Type: %s, Priority: %d)", 
		notification.ID, notification.Type, notification.Priority)
	
	return nil
}

// NotificationTargets represents the targets for a notification
type NotificationTargets struct {
	BroadcastToAll bool
	SessionID      string
	UserID         *uuid.UUID
	ClientIDs      []string
}

// SendInfoNotification sends an info notification
func (nm *NotificationManager) SendInfoNotification(title, message, category string, targets NotificationTargets) error {
	notification := &Notification{
		Type:     NotificationTypeInfo,
		Title:    title,
		Message:  message,
		Priority: PriorityLow,
		Category: category,
		Metadata: make(map[string]interface{}),
	}
	
	return nm.SendNotification(notification, targets)
}

// SendSuccessNotification sends a success notification
func (nm *NotificationManager) SendSuccessNotification(title, message, category string, targets NotificationTargets) error {
	notification := &Notification{
		Type:     NotificationTypeSuccess,
		Title:    title,
		Message:  message,
		Priority: PriorityMedium,
		Category: category,
		Metadata: make(map[string]interface{}),
	}
	
	return nm.SendNotification(notification, targets)
}

// SendWarningNotification sends a warning notification
func (nm *NotificationManager) SendWarningNotification(title, message, category string, targets NotificationTargets) error {
	notification := &Notification{
		Type:     NotificationTypeWarning,
		Title:    title,
		Message:  message,
		Priority: PriorityHigh,
		Category: category,
		Metadata: make(map[string]interface{}),
	}
	
	return nm.SendNotification(notification, targets)
}

// SendErrorNotification sends an error notification
func (nm *NotificationManager) SendErrorNotification(title, message, category string, targets NotificationTargets) error {
	notification := &Notification{
		Type:     NotificationTypeError,
		Title:    title,
		Message:  message,
		Priority: PriorityHigh,
		Category: category,
		Metadata: make(map[string]interface{}),
	}
	
	return nm.SendNotification(notification, targets)
}

// SendAlertNotification sends an alert notification
func (nm *NotificationManager) SendAlertNotification(title, message, category string, targets NotificationTargets) error {
	notification := &Notification{
		Type:     NotificationTypeAlert,
		Title:    title,
		Message:  message,
		Priority: PriorityCritical,
		Category: category,
		Metadata: make(map[string]interface{}),
	}
	
	return nm.SendNotification(notification, targets)
}

// SendCartNotification sends a cart-related notification
func (nm *NotificationManager) SendCartNotification(notificationType NotificationType, title, message string, sessionID string, userID *uuid.UUID) error {
	targets := NotificationTargets{
		SessionID: sessionID,
		UserID:    userID,
	}
	
	var priority NotificationPriority
	switch notificationType {
	case NotificationTypeSuccess:
		priority = PriorityMedium
	case NotificationTypeWarning:
		priority = PriorityHigh
	case NotificationTypeError:
		priority = PriorityHigh
	default:
		priority = PriorityLow
	}
	
	notification := &Notification{
		Type:     notificationType,
		Title:    title,
		Message:  message,
		Priority: priority,
		Category: "cart",
		Metadata: make(map[string]interface{}),
	}
	
	return nm.SendNotification(notification, targets)
}

// SendInventoryNotification sends an inventory-related notification
func (nm *NotificationManager) SendInventoryNotification(notificationType NotificationType, title, message string, sessionID string, userID *uuid.UUID) error {
	targets := NotificationTargets{
		SessionID: sessionID,
		UserID:    userID,
	}
	
	var priority NotificationPriority
	switch notificationType {
	case NotificationTypeAlert:
		priority = PriorityCritical
	case NotificationTypeWarning:
		priority = PriorityHigh
	default:
		priority = PriorityMedium
	}
	
	notification := &Notification{
		Type:     notificationType,
		Title:    title,
		Message:  message,
		Priority: priority,
		Category: "inventory",
		Metadata: make(map[string]interface{}),
	}
	
	return nm.SendNotification(notification, targets)
}

// SendOrderNotification sends an order-related notification
func (nm *NotificationManager) SendOrderNotification(notificationType NotificationType, title, message string, sessionID string, userID *uuid.UUID) error {
	targets := NotificationTargets{
		SessionID: sessionID,
		UserID:    userID,
	}
	
	var priority NotificationPriority
	switch notificationType {
	case NotificationTypeSuccess:
		priority = PriorityHigh
	case NotificationTypeError:
		priority = PriorityCritical
	default:
		priority = PriorityMedium
	}
	
	notification := &Notification{
		Type:     notificationType,
		Title:    title,
		Message:  message,
		Priority: priority,
		Category: "order",
		Metadata: make(map[string]interface{}),
	}
	
	return nm.SendNotification(notification, targets)
}

// SetUserPreferences sets notification preferences for a user
func (nm *NotificationManager) SetUserPreferences(userID uuid.UUID, preferences *NotificationPreferences) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	
	preferences.UserID = userID
	nm.userPreferences[userID.String()] = preferences
}

// GetUserPreferences retrieves notification preferences for a user
func (nm *NotificationManager) GetUserPreferences(userID uuid.UUID) (*NotificationPreferences, bool) {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	
	preferences, exists := nm.userPreferences[userID.String()]
	if !exists {
		return nil, false
	}
	
	// Return a copy to prevent external modifications
	preferencesCopy := *preferences
	return &preferencesCopy, true
}

// ShouldSendNotification checks if a notification should be sent based on user preferences
func (nm *NotificationManager) ShouldSendNotification(userID uuid.UUID, notification *Notification) bool {
	preferences, exists := nm.GetUserPreferences(userID)
	if !exists {
		return true // Send by default if no preferences set
	}
	
	// Check if notification type is enabled
	if enabled, exists := preferences.EnabledTypes[notification.Type]; exists && !enabled {
		return false
	}
	
	// Check if category is enabled
	if enabled, exists := preferences.EnabledCategories[notification.Category]; exists && !enabled {
		return false
	}
	
	// Check minimum priority
	if notification.Priority < preferences.MinPriority {
		return false
	}
	
	// Check quiet hours
	if preferences.QuietHoursStart != nil && preferences.QuietHoursEnd != nil {
		now := time.Now()
		start := *preferences.QuietHoursStart
		end := *preferences.QuietHoursEnd
		
		if now.After(start) && now.Before(end) {
			return false
		}
	}
	
	return true
}

// CleanupExpiredNotifications removes expired notifications
func (nm *NotificationManager) CleanupExpiredNotifications() {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	
	now := time.Now()
	var expiredIDs []string
	
	for id, notification := range nm.activeNotifications {
		if notification.ExpiresAt != nil && notification.ExpiresAt.Before(now) {
			expiredIDs = append(expiredIDs, id)
		}
	}
	
	for _, id := range expiredIDs {
		delete(nm.activeNotifications, id)
	}
	
	if len(expiredIDs) > 0 {
		log.Printf("Cleaned up %d expired notifications", len(expiredIDs))
	}
}

// GetActiveNotifications returns all active notifications
func (nm *NotificationManager) GetActiveNotifications() []*Notification {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	
	var notifications []*Notification
	for _, notification := range nm.activeNotifications {
		notifications = append(notifications, notification)
	}
	return notifications
}

// GetNotificationStats returns notification statistics
func (nm *NotificationManager) GetNotificationStats() map[string]interface{} {
	nm.mu.RLock()
	defer nm.mu.RUnlock()
	
	typeCount := make(map[NotificationType]int)
	categoryCount := make(map[string]int)
	priorityCount := make(map[NotificationPriority]int)
	
	for _, notification := range nm.activeNotifications {
		typeCount[notification.Type]++
		categoryCount[notification.Category]++
		priorityCount[notification.Priority]++
	}
	
	return map[string]interface{}{
		"active_notifications": len(nm.activeNotifications),
		"user_preferences":     len(nm.userPreferences),
		"type_distribution":    typeCount,
		"category_distribution": categoryCount,
		"priority_distribution": priorityCount,
		"default_expiry_s":     nm.defaultExpiry.Seconds(),
		"max_notifications":    nm.maxNotifications,
	}
}

// CreateDefaultPreferences creates default notification preferences
func CreateDefaultPreferences() *NotificationPreferences {
	return &NotificationPreferences{
		EnabledCategories: map[string]bool{
			"cart":      true,
			"inventory": true,
			"order":     true,
			"system":    true,
		},
		EnabledTypes: map[NotificationType]bool{
			NotificationTypeInfo:     true,
			NotificationTypeSuccess:  true,
			NotificationTypeWarning:  true,
			NotificationTypeError:    true,
			NotificationTypeAlert:    true,
		},
		MinPriority:      PriorityLow,
		MaxNotifications: 50,
		Metadata:         make(map[string]interface{}),
	}
}
