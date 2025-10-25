package services

import (
	"chat-ecommerce-backend/internal/models"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// AlertService handles inventory alert management
type AlertService struct {
	db *gorm.DB
}

// NewAlertService creates a new AlertService
func NewAlertService(db *gorm.DB) *AlertService {
	return &AlertService{
		db: db,
	}
}

// AlertConfig represents alert configuration
type AlertConfig struct {
	ID           uuid.UUID  `json:"id"`
	ProductID    *uuid.UUID `json:"product_id"`
	CategoryID   *uuid.UUID `json:"category_id"`
	AlertType    string     `json:"alert_type"` // "low_stock", "out_of_stock", "overstock"
	Threshold    int        `json:"threshold"`
	IsEnabled    bool       `json:"is_enabled"`
	EmailEnabled bool       `json:"email_enabled"`
	WebhookURL   string     `json:"webhook_url"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}

// AlertNotification represents a notification to be sent
type AlertNotification struct {
	ID        uuid.UUID  `json:"id"`
	AlertID   uuid.UUID  `json:"alert_id"`
	Type      string     `json:"type"` // "email", "webhook", "dashboard"
	Recipient string     `json:"recipient"`
	Subject   string     `json:"subject"`
	Message   string     `json:"message"`
	Status    string     `json:"status"` // "pending", "sent", "failed"
	CreatedAt time.Time  `json:"created_at"`
	SentAt    *time.Time `json:"sent_at"`
}

// AlertSummary represents a summary of alerts
type AlertSummary struct {
	TotalAlerts      int64                   `json:"total_alerts"`
	UnreadAlerts     int64                   `json:"unread_alerts"`
	LowStockAlerts   int64                   `json:"low_stock_alerts"`
	OutOfStockAlerts int64                   `json:"out_of_stock_alerts"`
	OverstockAlerts  int64                   `json:"overstock_alerts"`
	RecentAlerts     []models.InventoryAlert `json:"recent_alerts"`
}

// CreateAlertConfig creates a new alert configuration
func (s *AlertService) CreateAlertConfig(config AlertConfig) (*AlertConfig, error) {
	config.ID = uuid.New()
	config.CreatedAt = time.Now()
	config.UpdatedAt = time.Now()

	if err := s.db.Create(&config).Error; err != nil {
		return nil, fmt.Errorf("failed to create alert config: %v", err)
	}

	return &config, nil
}

// UpdateAlertConfig updates an existing alert configuration
func (s *AlertService) UpdateAlertConfig(id uuid.UUID, config AlertConfig) (*AlertConfig, error) {
	config.UpdatedAt = time.Now()

	if err := s.db.Model(&AlertConfig{}).Where("id = ?", id).Updates(config).Error; err != nil {
		return nil, fmt.Errorf("failed to update alert config: %v", err)
	}

	return &config, nil
}

// GetAlertConfigs returns all alert configurations
func (s *AlertService) GetAlertConfigs() ([]AlertConfig, error) {
	var configs []AlertConfig

	if err := s.db.Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("failed to get alert configs: %v", err)
	}

	return configs, nil
}

// GetAlertConfig returns a specific alert configuration
func (s *AlertService) GetAlertConfig(id uuid.UUID) (*AlertConfig, error) {
	var config AlertConfig

	if err := s.db.First(&config, id).Error; err != nil {
		return nil, fmt.Errorf("alert config not found: %v", err)
	}

	return &config, nil
}

// DeleteAlertConfig deletes an alert configuration
func (s *AlertService) DeleteAlertConfig(id uuid.UUID) error {
	if err := s.db.Delete(&AlertConfig{}, id).Error; err != nil {
		return fmt.Errorf("failed to delete alert config: %v", err)
	}

	return nil
}

// ProcessInventoryAlerts processes inventory changes and creates alerts
func (s *AlertService) ProcessInventoryAlerts(inventory models.Inventory) error {
	// Get applicable alert configurations
	configs, err := s.getApplicableConfigs(inventory)
	if err != nil {
		return fmt.Errorf("failed to get applicable configs: %v", err)
	}

	// Check each configuration
	for _, config := range configs {
		if !config.IsEnabled {
			continue
		}

		shouldAlert := false
		alertType := ""

		switch config.AlertType {
		case "low_stock":
			if inventory.QuantityAvailable < config.Threshold && inventory.QuantityAvailable > 0 {
				shouldAlert = true
				alertType = "low_stock"
			}
		case "out_of_stock":
			if inventory.QuantityAvailable == 0 {
				shouldAlert = true
				alertType = "out_of_stock"
			}
		case "overstock":
			if inventory.QuantityAvailable > config.Threshold {
				shouldAlert = true
				alertType = "overstock"
			}
		}

		if shouldAlert {
			if err := s.createAlert(inventory, alertType, config.Threshold, config); err != nil {
				log.Printf("Failed to create alert: %v", err)
			}
		}
	}

	return nil
}

// getApplicableConfigs returns alert configurations that apply to the given inventory
func (s *AlertService) getApplicableConfigs(inventory models.Inventory) ([]AlertConfig, error) {
	var configs []AlertConfig

	// Get product-specific configs
	if err := s.db.Where("product_id = ?", inventory.ProductID).Find(&configs).Error; err != nil {
		return nil, fmt.Errorf("failed to get product-specific configs: %v", err)
	}

	// Get category-specific configs
	var product models.Product
	if err := s.db.First(&product, inventory.ProductID).Error; err == nil {
		var categoryConfigs []AlertConfig
		if err := s.db.Where("category_id = ? AND product_id IS NULL", product.CategoryID).Find(&categoryConfigs).Error; err == nil {
			configs = append(configs, categoryConfigs...)
		}
	}

	// Get global configs
	var globalConfigs []AlertConfig
	if err := s.db.Where("product_id IS NULL AND category_id IS NULL").Find(&globalConfigs).Error; err == nil {
		configs = append(configs, globalConfigs...)
	}

	return configs, nil
}

// createAlert creates an inventory alert
func (s *AlertService) createAlert(inventory models.Inventory, alertType string, threshold int, config AlertConfig) error {
	// Check if alert already exists and is unread
	var existingAlert models.InventoryAlert
	err := s.db.Where("product_id = ? AND variant_id = ? AND alert_type = ? AND is_read = ?",
		inventory.ProductID, inventory.VariantID, alertType, false).
		First(&existingAlert).Error

	if err == nil {
		// Alert already exists, update it
		existingAlert.CurrentQuantity = inventory.QuantityAvailable
		existingAlert.CreatedAt = time.Now()
		if err := s.db.Save(&existingAlert).Error; err != nil {
			return fmt.Errorf("failed to update existing alert: %v", err)
		}
		return nil
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
		return fmt.Errorf("failed to create alert: %v", err)
	}

	// Create notifications if configured
	if config.EmailEnabled {
		if err := s.createEmailNotification(alert, config); err != nil {
			log.Printf("Failed to create email notification: %v", err)
		}
	}

	if config.WebhookURL != "" {
		if err := s.createWebhookNotification(alert, config); err != nil {
			log.Printf("Failed to create webhook notification: %v", err)
		}
	}

	return nil
}

// createEmailNotification creates an email notification
func (s *AlertService) createEmailNotification(alert models.InventoryAlert, config AlertConfig) error {
	// Get product details
	var product models.Product
	if err := s.db.First(&product, alert.ProductID).Error; err != nil {
		return fmt.Errorf("failed to get product: %v", err)
	}

	subject := fmt.Sprintf("Inventory Alert: %s", alert.AlertType)
	message := fmt.Sprintf("Product: %s\nCurrent Quantity: %d\nThreshold: %d\nLocation: %s",
		product.Name, alert.CurrentQuantity, alert.Threshold, alert.Location)

	notification := AlertNotification{
		ID:        uuid.New(),
		AlertID:   alert.ID,
		Type:      "email",
		Recipient: "admin@example.com", // This should come from config
		Subject:   subject,
		Message:   message,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	if err := s.db.Create(&notification).Error; err != nil {
		return fmt.Errorf("failed to create email notification: %v", err)
	}

	return nil
}

// createWebhookNotification creates a webhook notification
func (s *AlertService) createWebhookNotification(alert models.InventoryAlert, config AlertConfig) error {
	// Get product details
	var product models.Product
	if err := s.db.First(&product, alert.ProductID).Error; err != nil {
		return fmt.Errorf("failed to get product: %v", err)
	}

	message := fmt.Sprintf("Inventory Alert: %s for product %s", alert.AlertType, product.Name)

	notification := AlertNotification{
		ID:        uuid.New(),
		AlertID:   alert.ID,
		Type:      "webhook",
		Recipient: config.WebhookURL,
		Subject:   "",
		Message:   message,
		Status:    "pending",
		CreatedAt: time.Now(),
	}

	if err := s.db.Create(&notification).Error; err != nil {
		return fmt.Errorf("failed to create webhook notification: %v", err)
	}

	return nil
}

// GetAlertSummary returns a summary of current alerts
func (s *AlertService) GetAlertSummary() (*AlertSummary, error) {
	summary := &AlertSummary{}

	// Total alerts
	if err := s.db.Model(&models.InventoryAlert{}).Count(&summary.TotalAlerts).Error; err != nil {
		return nil, fmt.Errorf("failed to count total alerts: %v", err)
	}

	// Unread alerts
	if err := s.db.Model(&models.InventoryAlert{}).Where("is_read = ?", false).Count(&summary.UnreadAlerts).Error; err != nil {
		return nil, fmt.Errorf("failed to count unread alerts: %v", err)
	}

	// Low stock alerts
	if err := s.db.Model(&models.InventoryAlert{}).Where("alert_type = ? AND is_read = ?", "low_stock", false).Count(&summary.LowStockAlerts).Error; err != nil {
		return nil, fmt.Errorf("failed to count low stock alerts: %v", err)
	}

	// Out of stock alerts
	if err := s.db.Model(&models.InventoryAlert{}).Where("alert_type = ? AND is_read = ?", "out_of_stock", false).Count(&summary.OutOfStockAlerts).Error; err != nil {
		return nil, fmt.Errorf("failed to count out of stock alerts: %v", err)
	}

	// Overstock alerts
	if err := s.db.Model(&models.InventoryAlert{}).Where("alert_type = ? AND is_read = ?", "overstock", false).Count(&summary.OverstockAlerts).Error; err != nil {
		return nil, fmt.Errorf("failed to count overstock alerts: %v", err)
	}

	// Recent alerts (last 10)
	if err := s.db.Preload("Product").Preload("Variant").
		Order("created_at DESC").
		Limit(10).
		Find(&summary.RecentAlerts).Error; err != nil {
		return nil, fmt.Errorf("failed to get recent alerts: %v", err)
	}

	return summary, nil
}

// MarkAlertsAsRead marks multiple alerts as read
func (s *AlertService) MarkAlertsAsRead(alertIDs []uuid.UUID) error {
	if err := s.db.Model(&models.InventoryAlert{}).
		Where("id IN ?", alertIDs).
		Update("is_read", true).Error; err != nil {
		return fmt.Errorf("failed to mark alerts as read: %v", err)
	}

	return nil
}

// GetPendingNotifications returns pending notifications
func (s *AlertService) GetPendingNotifications() ([]AlertNotification, error) {
	var notifications []AlertNotification

	if err := s.db.Where("status = ?", "pending").Find(&notifications).Error; err != nil {
		return nil, fmt.Errorf("failed to get pending notifications: %v", err)
	}

	return notifications, nil
}

// MarkNotificationAsSent marks a notification as sent
func (s *AlertService) MarkNotificationAsSent(notificationID uuid.UUID) error {
	now := time.Now()
	if err := s.db.Model(&AlertNotification{}).
		Where("id = ?", notificationID).
		Updates(map[string]interface{}{
			"status":  "sent",
			"sent_at": &now,
		}).Error; err != nil {
		return fmt.Errorf("failed to mark notification as sent: %v", err)
	}

	return nil
}

// MarkNotificationAsFailed marks a notification as failed
func (s *AlertService) MarkNotificationAsFailed(notificationID uuid.UUID, errorMessage string) error {
	if err := s.db.Model(&AlertNotification{}).
		Where("id = ?", notificationID).
		Updates(map[string]interface{}{
			"status":  "failed",
			"message": errorMessage,
		}).Error; err != nil {
		return fmt.Errorf("failed to mark notification as failed: %v", err)
	}

	return nil
}

// CleanupOldAlerts removes alerts older than specified days
func (s *AlertService) CleanupOldAlerts(days int) error {
	cutoffDate := time.Now().AddDate(0, 0, -days)

	if err := s.db.Where("created_at < ? AND is_read = ?", cutoffDate, true).Delete(&models.InventoryAlert{}).Error; err != nil {
		return fmt.Errorf("failed to cleanup old alerts: %v", err)
	}

	return nil
}
