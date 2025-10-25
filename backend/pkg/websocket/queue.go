package websocket

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// QueueMessage represents a message in the queue
type QueueMessage struct {
	ID        string      `json:"id"`
	Type      MessageType `json:"type"`
	Data      interface{} `json:"data"`
	SessionID string      `json:"session_id"`
	UserID    *uuid.UUID  `json:"user_id"`
	Timestamp time.Time   `json:"timestamp"`
	Retries   int         `json:"retries"`
	MaxRetries int        `json:"max_retries"`
	Priority  int         `json:"priority"` // Higher number = higher priority
}

// MessageQueue handles reliable message delivery
type MessageQueue struct {
	// Queue storage
	queue []QueueMessage
	
	// Failed messages for retry
	failedMessages []QueueMessage
	
	// Mutex for thread-safe operations
	mu sync.RWMutex
	
	// Hub reference for message delivery
	hub *Hub
	
	// Configuration
	maxRetries int
	retryDelay time.Duration
	
	// Context for cancellation
	ctx context.Context
	cancel context.CancelFunc
	
	// Processing control
	processing bool
	stopChan   chan struct{}
}

// NewMessageQueue creates a new message queue
func NewMessageQueue(hub *Hub, maxRetries int, retryDelay time.Duration) *MessageQueue {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &MessageQueue{
		queue:       make([]QueueMessage, 0),
		failedMessages: make([]QueueMessage, 0),
		hub:         hub,
		maxRetries:  maxRetries,
		retryDelay:  retryDelay,
		ctx:         ctx,
		cancel:      cancel,
		stopChan:    make(chan struct{}),
	}
}

// Enqueue adds a message to the queue
func (mq *MessageQueue) Enqueue(message QueueMessage) error {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	
	// Set default values
	if message.ID == "" {
		message.ID = uuid.New().String()
	}
	if message.Timestamp.IsZero() {
		message.Timestamp = time.Now()
	}
	if message.MaxRetries == 0 {
		message.MaxRetries = mq.maxRetries
	}
	
	// Add to queue based on priority
	mq.insertByPriority(message)
	
	log.Printf("Message %s enqueued. Queue size: %d", message.ID, len(mq.queue))
	return nil
}

// insertByPriority inserts a message maintaining priority order
func (mq *MessageQueue) insertByPriority(message QueueMessage) {
	// Find insertion point based on priority
	insertIndex := len(mq.queue)
	for i, existing := range mq.queue {
		if message.Priority > existing.Priority {
			insertIndex = i
			break
		}
	}
	
	// Insert at the found position
	if insertIndex == len(mq.queue) {
		mq.queue = append(mq.queue, message)
	} else {
		mq.queue = append(mq.queue[:insertIndex+1], mq.queue[insertIndex:]...)
		mq.queue[insertIndex] = message
	}
}

// Dequeue removes and returns the highest priority message
func (mq *MessageQueue) Dequeue() (*QueueMessage, bool) {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	
	if len(mq.queue) == 0 {
		return nil, false
	}
	
	message := mq.queue[0]
	mq.queue = mq.queue[1:]
	
	return &message, true
}

// ProcessQueue processes messages from the queue
func (mq *MessageQueue) ProcessQueue() {
	mq.mu.Lock()
	mq.processing = true
	mq.mu.Unlock()
	
	defer func() {
		mq.mu.Lock()
		mq.processing = false
		mq.mu.Unlock()
	}()
	
	ticker := time.NewTicker(100 * time.Millisecond) // Process every 100ms
	defer ticker.Stop()
	
	for {
		select {
		case <-mq.ctx.Done():
			return
		case <-mq.stopChan:
			return
		case <-ticker.C:
			mq.processMessages()
		}
	}
}

// processMessages processes messages from the queue
func (mq *MessageQueue) processMessages() {
	// Process regular queue
	for {
		message, ok := mq.Dequeue()
		if !ok {
			break
		}
		
		if !mq.deliverMessage(*message) {
			// Message delivery failed, add to failed messages
			mq.addToFailedMessages(*message)
		}
	}
	
	// Process failed messages for retry
	mq.processFailedMessages()
}

// deliverMessage attempts to deliver a message
func (mq *MessageQueue) deliverMessage(message QueueMessage) bool {
	// Convert to regular message
	wsMessage := Message{
		ID:        message.ID,
		Type:      message.Type,
		Data:      message.Data,
		Timestamp: message.Timestamp,
		SessionID: message.SessionID,
		UserID:    message.UserID,
	}
	
	// Try to deliver based on target
	if message.SessionID != "" {
		mq.hub.BroadcastToSession(message.SessionID, wsMessage)
	} else if message.UserID != nil {
		mq.hub.BroadcastToUser(*message.UserID, wsMessage)
	} else {
		mq.hub.BroadcastMessage(wsMessage)
	}
	
	log.Printf("Message %s delivered successfully", message.ID)
	return true
}

// addToFailedMessages adds a message to the failed messages list
func (mq *MessageQueue) addToFailedMessages(message QueueMessage) {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	
	message.Retries++
	mq.failedMessages = append(mq.failedMessages, message)
	
	log.Printf("Message %s failed delivery, added to retry queue. Retries: %d/%d", 
		message.ID, message.Retries, message.MaxRetries)
}

// processFailedMessages processes failed messages for retry
func (mq *MessageQueue) processFailedMessages() {
	mq.mu.Lock()
	defer mq.mu.Unlock()
	
	var remainingFailed []QueueMessage
	
	for _, message := range mq.failedMessages {
		if message.Retries >= message.MaxRetries {
			log.Printf("Message %s exceeded max retries, dropping", message.ID)
			continue
		}
		
		// Check if enough time has passed since last retry
		if time.Since(message.Timestamp) < mq.retryDelay {
			remainingFailed = append(remainingFailed, message)
			continue
		}
		
		// Retry the message
		message.Timestamp = time.Now()
		mq.insertByPriority(message)
	}
	
	mq.failedMessages = remainingFailed
}

// Stop stops the message queue processing
func (mq *MessageQueue) Stop() {
	mq.cancel()
	close(mq.stopChan)
}

// GetQueueSize returns the current queue size
func (mq *MessageQueue) GetQueueSize() int {
	mq.mu.RLock()
	defer mq.mu.RUnlock()
	return len(mq.queue)
}

// GetFailedMessagesCount returns the number of failed messages
func (mq *MessageQueue) GetFailedMessagesCount() int {
	mq.mu.RLock()
	defer mq.mu.RUnlock()
	return len(mq.failedMessages)
}

// GetQueueStats returns queue statistics
func (mq *MessageQueue) GetQueueStats() map[string]interface{} {
	mq.mu.RLock()
	defer mq.mu.RUnlock()
	
	return map[string]interface{}{
		"queue_size":         len(mq.queue),
		"failed_messages":    len(mq.failedMessages),
		"processing":         mq.processing,
		"max_retries":        mq.maxRetries,
		"retry_delay_ms":     mq.retryDelay.Milliseconds(),
	}
}

// EnqueueCartUpdate enqueues a cart update message
func (mq *MessageQueue) EnqueueCartUpdate(cartData interface{}, sessionID string, userID *uuid.UUID, priority int) error {
	message := QueueMessage{
		Type:      MessageTypeCartUpdate,
		Data:      cartData,
		SessionID: sessionID,
		UserID:    userID,
		Priority:  priority,
	}
	return mq.Enqueue(message)
}

// EnqueueInventoryUpdate enqueues an inventory update message
func (mq *MessageQueue) EnqueueInventoryUpdate(inventoryData interface{}, sessionID string, userID *uuid.UUID, priority int) error {
	message := QueueMessage{
		Type:      MessageTypeInventoryUpdate,
		Data:      inventoryData,
		SessionID: sessionID,
		UserID:    userID,
		Priority:  priority,
	}
	return mq.Enqueue(message)
}

// EnqueueNotification enqueues a notification message
func (mq *MessageQueue) EnqueueNotification(notificationData interface{}, sessionID string, userID *uuid.UUID, priority int) error {
	message := QueueMessage{
		Type:      MessageTypeNotification,
		Data:      notificationData,
		SessionID: sessionID,
		UserID:    userID,
		Priority:  priority,
	}
	return mq.Enqueue(message)
}

// EnqueueBroadcast enqueues a broadcast message to all clients
func (mq *MessageQueue) EnqueueBroadcast(messageType MessageType, data interface{}, priority int) error {
	message := QueueMessage{
		Type:     messageType,
		Data:     data,
		Priority: priority,
	}
	return mq.Enqueue(message)
}

// Persistence interface for message queue persistence
type QueuePersistence interface {
	SaveMessage(message QueueMessage) error
	LoadMessages() ([]QueueMessage, error)
	DeleteMessage(messageID string) error
}

// PersistentMessageQueue extends MessageQueue with persistence
type PersistentMessageQueue struct {
	*MessageQueue
	persistence QueuePersistence
}

// NewPersistentMessageQueue creates a new persistent message queue
func NewPersistentMessageQueue(hub *Hub, persistence QueuePersistence, maxRetries int, retryDelay time.Duration) *PersistentMessageQueue {
	baseQueue := NewMessageQueue(hub, maxRetries, retryDelay)
	
	return &PersistentMessageQueue{
		MessageQueue: baseQueue,
		persistence:  persistence,
	}
}

// EnqueueWithPersistence enqueues a message with persistence
func (pmq *PersistentMessageQueue) EnqueueWithPersistence(message QueueMessage) error {
	// Save to persistence layer
	if err := pmq.persistence.SaveMessage(message); err != nil {
		return fmt.Errorf("failed to persist message: %w", err)
	}
	
	// Add to in-memory queue
	return pmq.Enqueue(message)
}

// LoadPersistedMessages loads messages from persistence layer
func (pmq *PersistentMessageQueue) LoadPersistedMessages() error {
	messages, err := pmq.persistence.LoadMessages()
	if err != nil {
		return fmt.Errorf("failed to load persisted messages: %w", err)
	}
	
	pmq.mu.Lock()
	defer pmq.mu.Unlock()
	
	for _, message := range messages {
		pmq.insertByPriority(message)
	}
	
	log.Printf("Loaded %d persisted messages", len(messages))
	return nil
}

// DeletePersistedMessage deletes a message from persistence layer
func (pmq *PersistentMessageQueue) DeletePersistedMessage(messageID string) error {
	return pmq.persistence.DeleteMessage(messageID)
}
