package services

import (
	"chat-ecommerce-backend/internal/models"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// InventoryService handles inventory management operations
type InventoryService struct {
	db *gorm.DB
}

// NewInventoryService creates a new InventoryService
func NewInventoryService(db *gorm.DB) *InventoryService {
	return &InventoryService{
		db: db,
	}
}

// InventoryUpdateRequest represents a request to update inventory
type InventoryUpdateRequest struct {
	ProductID uuid.UUID  `json:"product_id" binding:"required"`
	VariantID *uuid.UUID `json:"variant_id"`
	Quantity  int        `json:"quantity" binding:"required"`
	Location  string     `json:"location"`
	Operation string     `json:"operation" binding:"required"` // "add", "subtract", "set"
}

// InventoryReservationRequest represents a request to reserve inventory
type InventoryReservationRequest struct {
	ProductID uuid.UUID  `json:"product_id" binding:"required"`
	VariantID *uuid.UUID `json:"variant_id"`
	Quantity  int        `json:"quantity" binding:"required"`
	SessionID string     `json:"session_id" binding:"required"`
	ExpiresAt time.Time  `json:"expires_at"`
}

// InventoryAlert represents a low stock alert
type InventoryAlert struct {
	ID              uuid.UUID              `json:"id"`
	ProductID       uuid.UUID              `json:"product_id"`
	Product         *models.Product        `json:"product"`
	VariantID       *uuid.UUID             `json:"variant_id"`
	Variant         *models.ProductVariant `json:"variant"`
	CurrentQuantity int                    `json:"current_quantity"`
	Threshold       int                    `json:"threshold"`
	Location        string                 `json:"location"`
	AlertType       string                 `json:"alert_type"` // "low_stock", "out_of_stock", "overstock"
	CreatedAt       time.Time              `json:"created_at"`
	IsRead          bool                   `json:"is_read"`
}

// InventoryReport represents inventory reporting data
type InventoryReport struct {
	TotalProducts     int   `json:"total_products"`
	TotalQuantity     int   `json:"total_quantity"`
	LowStockItems     int64 `json:"low_stock_items"`
	OutOfStockItems   int64 `json:"out_of_stock_items"`
	OverstockItems    int64 `json:"overstock_items"`
	ReservedQuantity  int   `json:"reserved_quantity"`
	AvailableQuantity int   `json:"available_quantity"`
}

// UpdateInventory updates inventory levels
func (s *InventoryService) UpdateInventory(req InventoryUpdateRequest) error {
	// Find existing inventory record
	var inventory models.Inventory
	query := s.db.Where("product_id = ?", req.ProductID)
	if req.VariantID != nil {
		query = query.Where("variant_id = ?", *req.VariantID)
	} else {
		query = query.Where("variant_id IS NULL")
	}

	err := query.First(&inventory).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new inventory record
			inventory = models.Inventory{
				ProductID:         req.ProductID,
				VariantID:         req.VariantID,
				QuantityAvailable: 0,
				WarehouseLocation: req.Location,
				QuantityReserved:  0,
			}
		} else {
			return fmt.Errorf("failed to find inventory: %v", err)
		}
	}

	// Update quantity based on operation
	switch req.Operation {
	case "add":
		inventory.QuantityAvailable += req.Quantity
	case "subtract":
		inventory.QuantityAvailable -= req.Quantity
		if inventory.QuantityAvailable < 0 {
			inventory.QuantityAvailable = 0
		}
	case "set":
		inventory.QuantityAvailable = req.Quantity
	default:
		return fmt.Errorf("invalid operation: %s", req.Operation)
	}

	// Update location if provided
	if req.Location != "" {
		inventory.WarehouseLocation = req.Location
	}

	// Save inventory
	if err := s.db.Save(&inventory).Error; err != nil {
		return fmt.Errorf("failed to save inventory: %v", err)
	}

	// Check for alerts
	go s.checkInventoryAlerts(inventory)

	return nil
}

// ReserveInventory reserves inventory for a session
func (s *InventoryService) ReserveInventory(req InventoryReservationRequest) error {
	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Find inventory record
	var inventory models.Inventory
	query := tx.Where("product_id = ?", req.ProductID)
	if req.VariantID != nil {
		query = query.Where("variant_id = ?", *req.VariantID)
	} else {
		query = query.Where("variant_id IS NULL")
	}

	err := query.First(&inventory).Error
	if err != nil {
		tx.Rollback()
		return fmt.Errorf("inventory not found: %v", err)
	}

	// Check if enough quantity is available
	availableQuantity := inventory.QuantityAvailable - inventory.QuantityReserved
	if availableQuantity < req.Quantity {
		tx.Rollback()
		return fmt.Errorf("insufficient inventory: available %d, requested %d", availableQuantity, req.Quantity)
	}

	// Create reservation
	reservation := models.InventoryReservation{
		InventoryID:      inventory.ID,
		QuantityReserved: req.Quantity,
		SessionID:        req.SessionID,
		ExpiresAt:        req.ExpiresAt,
		Status:           "active",
	}

	if err := tx.Create(&reservation).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to create reservation: %v", err)
	}

	// Update reserved quantity
	inventory.QuantityReserved += req.Quantity
	if err := tx.Save(&inventory).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update reserved quantity: %v", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// ReleaseInventory releases reserved inventory
func (s *InventoryService) ReleaseInventory(sessionID string) error {
	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Find active reservations for session
	var reservations []models.InventoryReservation
	if err := tx.Where("session_id = ? AND status = ?", sessionID, "active").Find(&reservations).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to find reservations: %v", err)
	}

	// Release each reservation
	for _, reservation := range reservations {
		// Update reservation status
		reservation.Status = "released"
		if err := tx.Save(&reservation).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update reservation: %v", err)
		}

		// Update inventory reserved quantity
		var inventory models.Inventory
		if err := tx.Where("id = ?", reservation.InventoryID).First(&inventory).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to find inventory: %v", err)
		}

		inventory.QuantityReserved -= reservation.QuantityReserved
		if inventory.QuantityReserved < 0 {
			inventory.QuantityReserved = 0
		}

		if err := tx.Save(&inventory).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update inventory: %v", err)
		}
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// ConfirmInventory confirms reserved inventory (converts reservation to actual deduction)
func (s *InventoryService) ConfirmInventory(sessionID string) error {
	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Find active reservations for session
	var reservations []models.InventoryReservation
	if err := tx.Where("session_id = ? AND status = ?", sessionID, "active").Find(&reservations).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to find reservations: %v", err)
	}

	// Confirm each reservation
	for _, reservation := range reservations {
		// Update reservation status
		reservation.Status = "confirmed"
		if err := tx.Save(&reservation).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update reservation: %v", err)
		}

		// Update inventory quantities
		var inventory models.Inventory
		if err := tx.Where("id = ?", reservation.InventoryID).First(&inventory).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to find inventory: %v", err)
		}

		// Deduct from quantity and reserved
		inventory.QuantityAvailable -= reservation.QuantityReserved
		inventory.QuantityReserved -= reservation.QuantityReserved

		if inventory.QuantityAvailable < 0 {
			inventory.QuantityAvailable = 0
		}
		if inventory.QuantityReserved < 0 {
			inventory.QuantityReserved = 0
		}

		if err := tx.Save(&inventory).Error; err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to update inventory: %v", err)
		}

		// Check for alerts
		go s.checkInventoryAlerts(inventory)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// GetInventoryLevels returns current inventory levels
func (s *InventoryService) GetInventoryLevels(productID *uuid.UUID, variantID *uuid.UUID) ([]models.Inventory, error) {
	var inventory []models.Inventory

	query := s.db.Preload("Product").Preload("Variant")

	if productID != nil {
		query = query.Where("product_id = ?", *productID)
	}

	if variantID != nil {
		query = query.Where("variant_id = ?", *variantID)
	}

	if err := query.Find(&inventory).Error; err != nil {
		return nil, fmt.Errorf("failed to get inventory levels: %v", err)
	}

	return inventory, nil
}

// GetInventoryAlerts returns current inventory alerts
func (s *InventoryService) GetInventoryAlerts(isRead *bool) ([]InventoryAlert, error) {
	var alerts []InventoryAlert

	query := s.db.Table("inventory_alerts").
		Select("inventory_alerts.*, products.name as product_name, product_variants.variant_name, product_variants.variant_value").
		Joins("LEFT JOIN products ON inventory_alerts.product_id = products.id").
		Joins("LEFT JOIN product_variants ON inventory_alerts.variant_id = product_variants.id")

	if isRead != nil {
		query = query.Where("is_read = ?", *isRead)
	}

	query = query.Order("created_at DESC")

	if err := query.Find(&alerts).Error; err != nil {
		return nil, fmt.Errorf("failed to get inventory alerts: %v", err)
	}

	return alerts, nil
}

// MarkAlertAsRead marks an alert as read
func (s *InventoryService) MarkAlertAsRead(alertID uuid.UUID) error {
	if err := s.db.Model(&models.InventoryAlert{}).
		Where("id = ?", alertID).
		Update("is_read", true).Error; err != nil {
		return fmt.Errorf("failed to mark alert as read: %v", err)
	}

	return nil
}

// GetInventoryReport returns inventory reporting data
func (s *InventoryService) GetInventoryReport() (*InventoryReport, error) {
	report := &InventoryReport{}

	// Total products with inventory
	if err := s.db.Model(&models.Inventory{}).
		Select("COUNT(DISTINCT product_id)").
		Scan(&report.TotalProducts).Error; err != nil {
		return nil, fmt.Errorf("failed to count products: %v", err)
	}

	// Total quantity
	if err := s.db.Model(&models.Inventory{}).
		Select("COALESCE(SUM(quantity_available), 0)").
		Scan(&report.TotalQuantity).Error; err != nil {
		return nil, fmt.Errorf("failed to sum quantity: %v", err)
	}

	// Reserved quantity
	if err := s.db.Model(&models.Inventory{}).
		Select("COALESCE(SUM(quantity_reserved), 0)").
		Scan(&report.ReservedQuantity).Error; err != nil {
		return nil, fmt.Errorf("failed to sum reserved: %v", err)
	}

	// Available quantity
	report.AvailableQuantity = report.TotalQuantity - report.ReservedQuantity

	// Low stock items (quantity < 10)
	if err := s.db.Model(&models.Inventory{}).
		Where("quantity_available < ?", 10).
		Count(&report.LowStockItems).Error; err != nil {
		return nil, fmt.Errorf("failed to count low stock items: %v", err)
	}

	// Out of stock items
	if err := s.db.Model(&models.Inventory{}).
		Where("quantity_available = ?", 0).
		Count(&report.OutOfStockItems).Error; err != nil {
		return nil, fmt.Errorf("failed to count out of stock items: %v", err)
	}

	// Overstock items (quantity > 100)
	if err := s.db.Model(&models.Inventory{}).
		Where("quantity_available > ?", 100).
		Count(&report.OverstockItems).Error; err != nil {
		return nil, fmt.Errorf("failed to count overstock items: %v", err)
	}

	return report, nil
}

// checkInventoryAlerts checks if inventory levels trigger alerts
func (s *InventoryService) checkInventoryAlerts(inventory models.Inventory) {
	// Check for low stock alert (quantity < 10)
	if inventory.QuantityAvailable < 10 && inventory.QuantityAvailable > 0 {
		s.createAlert(inventory, "low_stock", 10)
	}

	// Check for out of stock alert (quantity = 0)
	if inventory.QuantityAvailable == 0 {
		s.createAlert(inventory, "out_of_stock", 0)
	}

	// Check for overstock alert (quantity > 100)
	if inventory.QuantityAvailable > 100 {
		s.createAlert(inventory, "overstock", 100)
	}
}

// createAlert creates an inventory alert
func (s *InventoryService) createAlert(inventory models.Inventory, alertType string, threshold int) {
	// Check if alert already exists
	var existingAlert models.InventoryAlert
	err := s.db.Where("product_id = ? AND variant_id = ? AND alert_type = ? AND is_read = ?",
		inventory.ProductID, inventory.VariantID, alertType, false).
		First(&existingAlert).Error

	if err == nil {
		// Alert already exists, update it
		existingAlert.CurrentQuantity = inventory.QuantityAvailable
		existingAlert.CreatedAt = time.Now()
		s.db.Save(&existingAlert)
		return
	}

	// Create new alert
	alert := models.InventoryAlert{
		ProductID:       inventory.ProductID,
		VariantID:       inventory.VariantID,
		CurrentQuantity: inventory.QuantityAvailable,
		Threshold:       threshold,
		Location:        inventory.WarehouseLocation,
		AlertType:       alertType,
		IsRead:          false,
	}

	if err := s.db.Create(&alert).Error; err != nil {
		log.Printf("Failed to create inventory alert: %v", err)
	}
}

// CleanupExpiredReservations removes expired inventory reservations
func (s *InventoryService) CleanupExpiredReservations() error {
	now := time.Now()

	// Find expired reservations
	var expiredReservations []models.InventoryReservation
	if err := s.db.Where("expires_at < ? AND status = ?", now, "active").Find(&expiredReservations).Error; err != nil {
		return fmt.Errorf("failed to find expired reservations: %v", err)
	}

	// Release expired reservations
	for _, reservation := range expiredReservations {
		if err := s.ReleaseInventory(reservation.SessionID); err != nil {
			log.Printf("Failed to release expired reservation %s: %v", reservation.ID, err)
		}
	}

	return nil
}
