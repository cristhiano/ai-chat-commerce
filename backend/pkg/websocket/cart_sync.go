package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// CartState represents the current state of a shopping cart
type CartState struct {
	SessionID     string                 `json:"session_id"`
	UserID        *uuid.UUID             `json:"user_id"`
	Items         []CartItem             `json:"items"`
	Subtotal      float64                `json:"subtotal"`
	TaxAmount     float64                `json:"tax_amount"`
	ShippingAmount float64               `json:"shipping_amount"`
	TotalAmount   float64                `json:"total_amount"`
	Currency      string                 `json:"currency"`
	LastUpdated   time.Time              `json:"last_updated"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// CartItem represents an item in the shopping cart
type CartItem struct {
	ProductID   uuid.UUID              `json:"product_id"`
	VariantID   *uuid.UUID             `json:"variant_id"`
	Quantity    int                    `json:"quantity"`
	UnitPrice   float64                `json:"unit_price"`
	TotalPrice  float64                `json:"total_price"`
	ProductName string                 `json:"product_name"`
	VariantName string                 `json:"variant_name"`
	ImageURL    string                 `json:"image_url"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// CartSyncManager manages cart state synchronization across interfaces
type CartSyncManager struct {
	// In-memory cart states
	cartStates map[string]*CartState
	
	// Mutex for thread-safe operations
	mu sync.RWMutex
	
	// Hub for broadcasting updates
	hub *Hub
	
	// Message queue for reliable delivery
	queue *MessageQueue
	
	// Configuration
	syncInterval time.Duration
}

// NewCartSyncManager creates a new cart synchronization manager
func NewCartSyncManager(hub *Hub, queue *MessageQueue, syncInterval time.Duration) *CartSyncManager {
	return &CartSyncManager{
		cartStates:   make(map[string]*CartState),
		hub:          hub,
		queue:        queue,
		syncInterval: syncInterval,
	}
}

// UpdateCartState updates the cart state and broadcasts changes
func (csm *CartSyncManager) UpdateCartState(cartState *CartState) error {
	csm.mu.Lock()
	defer csm.mu.Unlock()
	
	// Update the cart state
	cartState.LastUpdated = time.Now()
	csm.cartStates[cartState.SessionID] = cartState
	
	// Create update message
	updateMessage := CreateCartUpdateMessage(cartState, cartState.SessionID, cartState.UserID)
	
	// Broadcast to all clients in the session
	csm.hub.BroadcastToSession(cartState.SessionID, updateMessage)
	
	// If user is logged in, also broadcast to all their sessions
	if cartState.UserID != nil {
		csm.hub.BroadcastToUser(*cartState.UserID, updateMessage)
	}
	
	log.Printf("Cart state updated for session %s. Items: %d, Total: %.2f", 
		cartState.SessionID, len(cartState.Items), cartState.TotalAmount)
	
	return nil
}

// GetCartState retrieves the current cart state for a session
func (csm *CartSyncManager) GetCartState(sessionID string) (*CartState, bool) {
	csm.mu.RLock()
	defer csm.mu.RUnlock()
	
	cartState, exists := csm.cartStates[sessionID]
	if !exists {
		return nil, false
	}
	
	// Return a copy to prevent external modifications
	cartStateCopy := *cartState
	return &cartStateCopy, true
}

// AddItemToCart adds an item to the cart and syncs across interfaces
func (csm *CartSyncManager) AddItemToCart(sessionID string, userID *uuid.UUID, item CartItem) error {
	csm.mu.Lock()
	defer csm.mu.Unlock()
	
	cartState, exists := csm.cartStates[sessionID]
	if !exists {
		cartState = &CartState{
			SessionID: sessionID,
			UserID:    userID,
			Items:     make([]CartItem, 0),
			Currency:  "USD",
			Metadata:  make(map[string]interface{}),
		}
		csm.cartStates[sessionID] = cartState
	}
	
	// Check if item already exists
	itemIndex := -1
	for i, existingItem := range cartState.Items {
		if existingItem.ProductID == item.ProductID && 
		   ((existingItem.VariantID == nil && item.VariantID == nil) ||
		    (existingItem.VariantID != nil && item.VariantID != nil && *existingItem.VariantID == *item.VariantID)) {
			itemIndex = i
			break
		}
	}
	
	if itemIndex >= 0 {
		// Update existing item quantity
		cartState.Items[itemIndex].Quantity += item.Quantity
		cartState.Items[itemIndex].TotalPrice = float64(cartState.Items[itemIndex].Quantity) * cartState.Items[itemIndex].UnitPrice
	} else {
		// Add new item
		cartState.Items = append(cartState.Items, item)
	}
	
	// Recalculate totals
	csm.recalculateCartTotals(cartState)
	cartState.LastUpdated = time.Now()
	
	// Broadcast update
	updateMessage := CreateCartUpdateMessage(cartState, sessionID, userID)
	csm.hub.BroadcastToSession(sessionID, updateMessage)
	
	if userID != nil {
		csm.hub.BroadcastToUser(*userID, updateMessage)
	}
	
	log.Printf("Item added to cart for session %s. Product: %s, Quantity: %d", 
		sessionID, item.ProductName, item.Quantity)
	
	return nil
}

// RemoveItemFromCart removes an item from the cart and syncs across interfaces
func (csm *CartSyncManager) RemoveItemFromCart(sessionID string, userID *uuid.UUID, productID uuid.UUID, variantID *uuid.UUID) error {
	csm.mu.Lock()
	defer csm.mu.Unlock()
	
	cartState, exists := csm.cartStates[sessionID]
	if !exists {
		return fmt.Errorf("cart not found for session %s", sessionID)
	}
	
	// Find and remove the item
	for i, item := range cartState.Items {
		if item.ProductID == productID && 
		   ((item.VariantID == nil && variantID == nil) ||
		    (item.VariantID != nil && variantID != nil && *item.VariantID == *variantID)) {
			cartState.Items = append(cartState.Items[:i], cartState.Items[i+1:]...)
			break
		}
	}
	
	// Recalculate totals
	csm.recalculateCartTotals(cartState)
	cartState.LastUpdated = time.Now()
	
	// Broadcast update
	updateMessage := CreateCartUpdateMessage(cartState, sessionID, userID)
	csm.hub.BroadcastToSession(sessionID, updateMessage)
	
	if userID != nil {
		csm.hub.BroadcastToUser(*userID, updateMessage)
	}
	
	log.Printf("Item removed from cart for session %s. Product: %s", sessionID, productID)
	
	return nil
}

// UpdateItemQuantity updates the quantity of an item in the cart
func (csm *CartSyncManager) UpdateItemQuantity(sessionID string, userID *uuid.UUID, productID uuid.UUID, variantID *uuid.UUID, quantity int) error {
	csm.mu.Lock()
	defer csm.mu.Unlock()
	
	cartState, exists := csm.cartStates[sessionID]
	if !exists {
		return fmt.Errorf("cart not found for session %s", sessionID)
	}
	
	// Find and update the item
	for i, item := range cartState.Items {
		if item.ProductID == productID && 
		   ((item.VariantID == nil && variantID == nil) ||
		    (item.VariantID != nil && variantID != nil && *item.VariantID == *variantID)) {
			if quantity <= 0 {
				// Remove item if quantity is 0 or negative
				cartState.Items = append(cartState.Items[:i], cartState.Items[i+1:]...)
			} else {
				cartState.Items[i].Quantity = quantity
				cartState.Items[i].TotalPrice = float64(quantity) * cartState.Items[i].UnitPrice
			}
			break
		}
	}
	
	// Recalculate totals
	csm.recalculateCartTotals(cartState)
	cartState.LastUpdated = time.Now()
	
	// Broadcast update
	updateMessage := CreateCartUpdateMessage(cartState, sessionID, userID)
	csm.hub.BroadcastToSession(sessionID, updateMessage)
	
	if userID != nil {
		csm.hub.BroadcastToUser(*userID, updateMessage)
	}
	
	log.Printf("Item quantity updated for session %s. Product: %s, Quantity: %d", 
		sessionID, productID, quantity)
	
	return nil
}

// ClearCart clears all items from the cart
func (csm *CartSyncManager) ClearCart(sessionID string, userID *uuid.UUID) error {
	csm.mu.Lock()
	defer csm.mu.Unlock()
	
	cartState, exists := csm.cartStates[sessionID]
	if !exists {
		return fmt.Errorf("cart not found for session %s", sessionID)
	}
	
	cartState.Items = make([]CartItem, 0)
	csm.recalculateCartTotals(cartState)
	cartState.LastUpdated = time.Now()
	
	// Broadcast update
	updateMessage := CreateCartUpdateMessage(cartState, sessionID, userID)
	csm.hub.BroadcastToSession(sessionID, updateMessage)
	
	if userID != nil {
		csm.hub.BroadcastToUser(*userID, updateMessage)
	}
	
	log.Printf("Cart cleared for session %s", sessionID)
	
	return nil
}

// recalculateCartTotals recalculates cart totals
func (csm *CartSyncManager) recalculateCartTotals(cartState *CartState) {
	cartState.Subtotal = 0
	
	for _, item := range cartState.Items {
		cartState.Subtotal += item.TotalPrice
	}
	
	// Calculate tax (simplified - 8.5% tax rate)
	cartState.TaxAmount = cartState.Subtotal * 0.085
	
	// Calculate shipping (simplified - free shipping over $50, otherwise $5.99)
	if cartState.Subtotal >= 50.0 {
		cartState.ShippingAmount = 0
	} else {
		cartState.ShippingAmount = 5.99
	}
	
	cartState.TotalAmount = cartState.Subtotal + cartState.TaxAmount + cartState.ShippingAmount
}

// MergeCartStates merges cart states from different sessions for the same user
func (csm *CartSyncManager) MergeCartStates(userID uuid.UUID, primarySessionID string) error {
	csm.mu.Lock()
	defer csm.mu.Unlock()
	
	var sessionsToMerge []string
	
	// Find all sessions for this user
	for sessionID, cartState := range csm.cartStates {
		if cartState.UserID != nil && *cartState.UserID == userID && sessionID != primarySessionID {
			sessionsToMerge = append(sessionsToMerge, sessionID)
		}
	}
	
	if len(sessionsToMerge) == 0 {
		return nil // No sessions to merge
	}
	
	primaryCart := csm.cartStates[primarySessionID]
	if primaryCart == nil {
		return fmt.Errorf("primary cart not found for session %s", primarySessionID)
	}
	
	// Merge items from other sessions
	for _, sessionID := range sessionsToMerge {
		cartToMerge := csm.cartStates[sessionID]
		if cartToMerge == nil {
			continue
		}
		
		for _, item := range cartToMerge.Items {
			csm.AddItemToCart(primarySessionID, &userID, item)
		}
		
		// Remove the merged session
		delete(csm.cartStates, sessionID)
	}
	
	log.Printf("Merged %d cart sessions for user %s into session %s", 
		len(sessionsToMerge), userID, primarySessionID)
	
	return nil
}

// GetCartStats returns cart synchronization statistics
func (csm *CartSyncManager) GetCartStats() map[string]interface{} {
	csm.mu.RLock()
	defer csm.mu.RUnlock()
	
	totalItems := 0
	totalValue := 0.0
	
	for _, cartState := range csm.cartStates {
		totalItems += len(cartState.Items)
		totalValue += cartState.TotalAmount
	}
	
	return map[string]interface{}{
		"active_carts":    len(csm.cartStates),
		"total_items":     totalItems,
		"total_value":     totalValue,
		"sync_interval_ms": csm.syncInterval.Milliseconds(),
	}
}

// CleanupExpiredCarts removes carts that haven't been updated recently
func (csm *CartSyncManager) CleanupExpiredCarts(maxAge time.Duration) {
	csm.mu.Lock()
	defer csm.mu.Unlock()
	
	cutoff := time.Now().Add(-maxAge)
	var expiredSessions []string
	
	for sessionID, cartState := range csm.cartStates {
		if cartState.LastUpdated.Before(cutoff) {
			expiredSessions = append(expiredSessions, sessionID)
		}
	}
	
	for _, sessionID := range expiredSessions {
		delete(csm.cartStates, sessionID)
		log.Printf("Cleaned up expired cart for session %s", sessionID)
	}
	
	if len(expiredSessions) > 0 {
		log.Printf("Cleaned up %d expired carts", len(expiredSessions))
	}
}

// SerializeCartState serializes cart state to JSON
func (csm *CartSyncManager) SerializeCartState(cartState *CartState) ([]byte, error) {
	return json.Marshal(cartState)
}

// DeserializeCartState deserializes cart state from JSON
func (csm *CartSyncManager) DeserializeCartState(data []byte) (*CartState, error) {
	var cartState CartState
	err := json.Unmarshal(data, &cartState)
	if err != nil {
		return nil, err
	}
	return &cartState, nil
}
