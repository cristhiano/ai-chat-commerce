package websocket

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// InventoryUpdate represents an inventory level update
type InventoryUpdate struct {
	ProductID       uuid.UUID              `json:"product_id"`
	VariantID       *uuid.UUID             `json:"variant_id"`
	ProductName     string                 `json:"product_name"`
	VariantName     string                 `json:"variant_name"`
	QuantityAvailable int                  `json:"quantity_available"`
	QuantityReserved  int                  `json:"quantity_reserved"`
	WarehouseLocation string               `json:"warehouse_location"`
	LastUpdated     time.Time              `json:"last_updated"`
	UpdateType      string                 `json:"update_type"` // "stock_change", "reservation", "restock", "alert"
	Metadata        map[string]interface{} `json:"metadata"`
}

// InventoryAlert represents an inventory alert
type InventoryAlert struct {
	ID              uuid.UUID              `json:"id"`
	ProductID       uuid.UUID              `json:"product_id"`
	VariantID       *uuid.UUID             `json:"variant_id"`
	ProductName     string                 `json:"product_name"`
	VariantName     string                 `json:"variant_name"`
	CurrentQuantity int                    `json:"current_quantity"`
	Threshold       int                    `json:"threshold"`
	Location        string                 `json:"location"`
	AlertType       string                 `json:"alert_type"` // "low_stock", "out_of_stock", "overstock"
	Severity        string                 `json:"severity"`   // "low", "medium", "high", "critical"
	CreatedAt       time.Time              `json:"created_at"`
	Metadata        map[string]interface{} `json:"metadata"`
}

// InventoryBroadcastManager manages inventory update broadcasting
type InventoryBroadcastManager struct {
	// Hub for broadcasting updates
	hub *Hub
	
	// Message queue for reliable delivery
	queue *MessageQueue
	
	// Recent updates cache to prevent spam
	recentUpdates map[string]time.Time
	
	// Mutex for thread-safe operations
	mu sync.RWMutex
	
	// Configuration
	broadcastDelay time.Duration
	cacheDuration  time.Duration
}

// NewInventoryBroadcastManager creates a new inventory broadcast manager
func NewInventoryBroadcastManager(hub *Hub, queue *MessageQueue, broadcastDelay, cacheDuration time.Duration) *InventoryBroadcastManager {
	return &InventoryBroadcastManager{
		hub:            hub,
		queue:          queue,
		recentUpdates:  make(map[string]time.Time),
		broadcastDelay: broadcastDelay,
		cacheDuration:  cacheDuration,
	}
}

// BroadcastInventoryUpdate broadcasts an inventory update to relevant clients
func (ibm *InventoryBroadcastManager) BroadcastInventoryUpdate(update InventoryUpdate) error {
	// Create cache key for deduplication
	cacheKey := fmt.Sprintf("%s_%s_%s", update.ProductID, update.VariantID, update.UpdateType)
	
	ibm.mu.Lock()
	defer ibm.mu.Unlock()
	
	// Check if we've recently broadcasted this update
	if lastBroadcast, exists := ibm.recentUpdates[cacheKey]; exists {
		if time.Since(lastBroadcast) < ibm.cacheDuration {
			log.Printf("Skipping duplicate inventory update for product %s", update.ProductID)
			return nil
		}
	}
	
	// Update cache
	ibm.recentUpdates[cacheKey] = time.Now()
	
	// Create broadcast message
	message := CreateInventoryUpdateMessage(update, "", nil)
	
	// Broadcast to all connected clients
	ibm.hub.BroadcastMessage(message)
	
	// Also queue for reliable delivery
	err := ibm.queue.EnqueueInventoryUpdate(update, "", nil, 5) // High priority for inventory updates
	if err != nil {
		log.Printf("Failed to queue inventory update: %v", err)
	}
	
	log.Printf("Broadcasted inventory update for product %s: %s (Available: %d, Reserved: %d)", 
		update.ProductID, update.UpdateType, update.QuantityAvailable, update.QuantityReserved)
	
	return nil
}

// BroadcastInventoryAlert broadcasts an inventory alert to relevant clients
func (ibm *InventoryBroadcastManager) BroadcastInventoryAlert(alert InventoryAlert) error {
	// Create cache key for deduplication
	cacheKey := fmt.Sprintf("alert_%s_%s_%s", alert.ProductID, alert.VariantID, alert.AlertType)
	
	ibm.mu.Lock()
	defer ibm.mu.Unlock()
	
	// Check if we've recently broadcasted this alert
	if lastBroadcast, exists := ibm.recentUpdates[cacheKey]; exists {
		if time.Since(lastBroadcast) < ibm.cacheDuration {
			log.Printf("Skipping duplicate inventory alert for product %s", alert.ProductID)
			return nil
		}
	}
	
	// Update cache
	ibm.recentUpdates[cacheKey] = time.Now()
	
	// Create notification message
	notificationData := map[string]interface{}{
		"type":        "inventory_alert",
		"alert":       alert,
		"severity":    alert.Severity,
		"action_required": ibm.getActionRequired(alert.AlertType),
	}
	
	message := CreateNotificationMessage(notificationData, "", nil)
	
	// Broadcast to all connected clients
	ibm.hub.BroadcastMessage(message)
	
	// Queue for reliable delivery with high priority
	err := ibm.queue.EnqueueNotification(notificationData, "", nil, 8) // Very high priority for alerts
	if err != nil {
		log.Printf("Failed to queue inventory alert: %v", err)
	}
	
	log.Printf("Broadcasted inventory alert for product %s: %s (Severity: %s)", 
		alert.ProductID, alert.AlertType, alert.Severity)
	
	return nil
}

// BroadcastStockChange broadcasts a stock level change
func (ibm *InventoryBroadcastManager) BroadcastStockChange(productID uuid.UUID, variantID *uuid.UUID, productName, variantName string, oldQuantity, newQuantity int, location string) error {
	update := InventoryUpdate{
		ProductID:         productID,
		VariantID:         variantID,
		ProductName:       productName,
		VariantName:       variantName,
		QuantityAvailable: newQuantity,
		WarehouseLocation: location,
		LastUpdated:       time.Now(),
		UpdateType:        "stock_change",
		Metadata: map[string]interface{}{
			"old_quantity": oldQuantity,
			"change_amount": newQuantity - oldQuantity,
		},
	}
	
	return ibm.BroadcastInventoryUpdate(update)
}

// BroadcastReservationUpdate broadcasts a reservation change
func (ibm *InventoryBroadcastManager) BroadcastReservationUpdate(productID uuid.UUID, variantID *uuid.UUID, productName, variantName string, availableQuantity, reservedQuantity int, location string) error {
	update := InventoryUpdate{
		ProductID:         productID,
		VariantID:         variantID,
		ProductName:       productName,
		VariantName:       variantName,
		QuantityAvailable: availableQuantity,
		QuantityReserved:  reservedQuantity,
		WarehouseLocation: location,
		LastUpdated:       time.Now(),
		UpdateType:        "reservation",
		Metadata: map[string]interface{}{
			"available_quantity": availableQuantity,
			"reserved_quantity":  reservedQuantity,
		},
	}
	
	return ibm.BroadcastInventoryUpdate(update)
}

// BroadcastRestockNotification broadcasts a restock notification
func (ibm *InventoryBroadcastManager) BroadcastRestockNotification(productID uuid.UUID, variantID *uuid.UUID, productName, variantName string, newQuantity int, location string) error {
	update := InventoryUpdate{
		ProductID:         productID,
		VariantID:         variantID,
		ProductName:       productName,
		VariantName:       variantName,
		QuantityAvailable: newQuantity,
		WarehouseLocation: location,
		LastUpdated:       time.Now(),
		UpdateType:        "restock",
		Metadata: map[string]interface{}{
			"restock_quantity": newQuantity,
		},
	}
	
	return ibm.BroadcastInventoryUpdate(update)
}

// BroadcastLowStockAlert broadcasts a low stock alert
func (ibm *InventoryBroadcastManager) BroadcastLowStockAlert(productID uuid.UUID, variantID *uuid.UUID, productName, variantName string, currentQuantity, threshold int, location string) error {
	alert := InventoryAlert{
		ID:              uuid.New(),
		ProductID:       productID,
		VariantID:       variantID,
		ProductName:     productName,
		VariantName:     variantName,
		CurrentQuantity: currentQuantity,
		Threshold:       threshold,
		Location:        location,
		AlertType:       "low_stock",
		Severity:        ibm.calculateSeverity(currentQuantity, threshold),
		CreatedAt:       time.Now(),
		Metadata: map[string]interface{}{
			"current_quantity": currentQuantity,
			"threshold":        threshold,
			"shortage":         threshold - currentQuantity,
		},
	}
	
	return ibm.BroadcastInventoryAlert(alert)
}

// BroadcastOutOfStockAlert broadcasts an out of stock alert
func (ibm *InventoryBroadcastManager) BroadcastOutOfStockAlert(productID uuid.UUID, variantID *uuid.UUID, productName, variantName string, location string) error {
	alert := InventoryAlert{
		ID:              uuid.New(),
		ProductID:       productID,
		VariantID:       variantID,
		ProductName:     productName,
		VariantName:     variantName,
		CurrentQuantity: 0,
		Threshold:       0,
		Location:        location,
		AlertType:       "out_of_stock",
		Severity:        "critical",
		CreatedAt:       time.Now(),
		Metadata: map[string]interface{}{
			"current_quantity": 0,
			"urgent":          true,
		},
	}
	
	return ibm.BroadcastInventoryAlert(alert)
}

// BroadcastOverstockAlert broadcasts an overstock alert
func (ibm *InventoryBroadcastManager) BroadcastOverstockAlert(productID uuid.UUID, variantID *uuid.UUID, productName, variantName string, currentQuantity, threshold int, location string) error {
	alert := InventoryAlert{
		ID:              uuid.New(),
		ProductID:       productID,
		VariantID:       variantID,
		ProductName:     productName,
		VariantName:     variantName,
		CurrentQuantity: currentQuantity,
		Threshold:       threshold,
		Location:        location,
		AlertType:       "overstock",
		Severity:        "low",
		CreatedAt:       time.Now(),
		Metadata: map[string]interface{}{
			"current_quantity": currentQuantity,
			"threshold":        threshold,
			"excess":          currentQuantity - threshold,
		},
	}
	
	return ibm.BroadcastInventoryAlert(alert)
}

// getActionRequired determines what action is required based on alert type
func (ibm *InventoryBroadcastManager) getActionRequired(alertType string) string {
	switch alertType {
	case "low_stock":
		return "Consider reordering"
	case "out_of_stock":
		return "Urgent reorder required"
	case "overstock":
		return "Consider promotions or transfers"
	default:
		return "Review inventory levels"
	}
}

// calculateSeverity calculates alert severity based on current quantity and threshold
func (ibm *InventoryBroadcastManager) calculateSeverity(currentQuantity, threshold int) string {
	if currentQuantity == 0 {
		return "critical"
	}
	
	percentage := float64(currentQuantity) / float64(threshold)
	
	if percentage <= 0.1 {
		return "critical"
	} else if percentage <= 0.3 {
		return "high"
	} else if percentage <= 0.5 {
		return "medium"
	} else {
		return "low"
	}
}

// CleanupExpiredCache removes expired entries from the recent updates cache
func (ibm *InventoryBroadcastManager) CleanupExpiredCache() {
	ibm.mu.Lock()
	defer ibm.mu.Unlock()
	
	cutoff := time.Now().Add(-ibm.cacheDuration)
	var expiredKeys []string
	
	for key, timestamp := range ibm.recentUpdates {
		if timestamp.Before(cutoff) {
			expiredKeys = append(expiredKeys, key)
		}
	}
	
	for _, key := range expiredKeys {
		delete(ibm.recentUpdates, key)
	}
	
	if len(expiredKeys) > 0 {
		log.Printf("Cleaned up %d expired inventory update cache entries", len(expiredKeys))
	}
}

// GetBroadcastStats returns broadcast statistics
func (ibm *InventoryBroadcastManager) GetBroadcastStats() map[string]interface{} {
	ibm.mu.RLock()
	defer ibm.mu.RUnlock()
	
	return map[string]interface{}{
		"cached_updates":     len(ibm.recentUpdates),
		"broadcast_delay_ms": ibm.broadcastDelay.Milliseconds(),
		"cache_duration_ms":  ibm.cacheDuration.Milliseconds(),
	}
}
