package services

import (
	"chat-ecommerce-backend/internal/models"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// OrderService handles order-related business logic
type OrderService struct {
	db *gorm.DB
}

// NewOrderService creates a new OrderService
func NewOrderService(db *gorm.DB) *OrderService {
	return &OrderService{
		db: db,
	}
}

// CreateOrderRequest represents the request payload for creating an order
type CreateOrderRequest struct {
	UserID          uuid.UUID              `json:"user_id"`
	SessionID       string                 `json:"session_id"`
	Items           []OrderItemRequest     `json:"items" binding:"required"`
	ShippingAddress map[string]interface{} `json:"shipping_address" binding:"required"`
	BillingAddress  map[string]interface{} `json:"billing_address" binding:"required"`
	PaymentMethod   string                 `json:"payment_method" binding:"required"`
	Notes           string                 `json:"notes"`
}

// OrderItemRequest represents an item in the order request
type OrderItemRequest struct {
	ProductID uuid.UUID  `json:"product_id" binding:"required"`
	VariantID *uuid.UUID `json:"variant_id"`
	Quantity  int        `json:"quantity" binding:"required,min=1"`
}

// UpdateOrderStatusRequest represents the request payload for updating order status
type UpdateOrderStatusRequest struct {
	Status string `json:"status" binding:"required"`
	Notes  string `json:"notes"`
}

// Order represents an order in the system (alias for models.Order)
type Order = models.Order

// OrderItem represents an order item in the system (alias for models.OrderItem)
type OrderItem = models.OrderItem

// CreateOrder creates a new order
func (s *OrderService) CreateOrder(req *CreateOrderRequest) (*Order, error) {
	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Generate order number
	orderNumber := s.generateOrderNumber()

	// Calculate totals
	var subtotal float64
	var orderItems []OrderItem

	for _, itemReq := range req.Items {
		// Get product details
		var product models.Product
		if err := tx.Where("id = ?", itemReq.ProductID).First(&product).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("product not found: %v", err)
		}

		// Check inventory
		if err := s.checkInventory(tx, itemReq.ProductID, itemReq.VariantID, itemReq.Quantity); err != nil {
			tx.Rollback()
			return nil, err
		}

		// Calculate item total
		unitPrice := product.Price
		totalPrice := unitPrice * float64(itemReq.Quantity)
		subtotal += totalPrice

		// Create order item
		orderItem := OrderItem{
			ID:         uuid.New(),
			OrderID:    uuid.New(), // Will be updated after order creation
			ProductID:  itemReq.ProductID,
			VariantID:  itemReq.VariantID,
			Quantity:   itemReq.Quantity,
			UnitPrice:  unitPrice,
			TotalPrice: totalPrice,
			CreatedAt:  time.Now(),
		}

		orderItems = append(orderItems, orderItem)
	}

	// Calculate tax and shipping (simplified)
	taxAmount := subtotal * 0.08 // 8% tax
	shippingAmount := 9.99       // Fixed shipping
	totalAmount := subtotal + taxAmount + shippingAmount

	// Marshal addresses to JSON
	shippingJSON, err := json.Marshal(req.ShippingAddress)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("failed to marshal shipping address")
	}

	billingJSON, err := json.Marshal(req.BillingAddress)
	if err != nil {
		tx.Rollback()
		return nil, errors.New("failed to marshal billing address")
	}

	// Create order
	order := &Order{
		ID:              uuid.New(),
		OrderNumber:     orderNumber,
		UserID:          req.UserID,
		SessionID:       req.SessionID,
		Status:          "pending",
		Subtotal:        subtotal,
		TaxAmount:       taxAmount,
		ShippingAmount:  shippingAmount,
		TotalAmount:     totalAmount,
		Currency:        "USD",
		PaymentStatus:   "pending",
		ShippingAddress: datatypes.JSON(shippingJSON),
		BillingAddress:  datatypes.JSON(billingJSON),
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
	}

	// Save order
	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("failed to create order")
	}

	// Update order items with order ID
	for i := range orderItems {
		orderItems[i].OrderID = order.ID
	}

	// Save order items
	if err := tx.Create(&orderItems).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("failed to create order items")
	}

	// Reserve inventory
	if err := s.reserveInventory(tx, orderItems); err != nil {
		tx.Rollback()
		return nil, err
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("failed to commit order transaction")
	}

	// Load order with items
	if err := s.db.Preload("Items").Preload("Items.Product").First(order, order.ID).Error; err != nil {
		return nil, errors.New("failed to load order details")
	}

	return order, nil
}

// GetOrderByID retrieves an order by ID
func (s *OrderService) GetOrderByID(orderID uuid.UUID) (*Order, error) {
	var order Order
	if err := s.db.Preload("Items").Preload("Items.Product").Preload("Items.Variant").Where("id = ?", orderID).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order not found")
		}
		return nil, errors.New("failed to retrieve order")
	}

	return &order, nil
}

// GetOrderByNumber retrieves an order by order number
func (s *OrderService) GetOrderByNumber(orderNumber string) (*Order, error) {
	var order Order
	if err := s.db.Preload("Items").Preload("Items.Product").Preload("Items.Variant").Where("order_number = ?", orderNumber).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order not found")
		}
		return nil, errors.New("failed to retrieve order")
	}

	return &order, nil
}

// GetUserOrders retrieves orders for a specific user
func (s *OrderService) GetUserOrders(userID uuid.UUID, page, limit int) ([]Order, int64, error) {
	var orders []Order
	var total int64

	// Count total orders
	if err := s.db.Model(&Order{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, errors.New("failed to count orders")
	}

	// Get orders with pagination
	offset := (page - 1) * limit
	if err := s.db.Preload("Items").Preload("Items.Product").
		Where("user_id = ?", userID).
		Order("created_at DESC").
		Offset(offset).
		Limit(limit).
		Find(&orders).Error; err != nil {
		return nil, 0, errors.New("failed to retrieve orders")
	}

	return orders, total, nil
}

// UpdateOrderStatus updates the status of an order
func (s *OrderService) UpdateOrderStatus(orderID uuid.UUID, req *UpdateOrderStatusRequest) (*Order, error) {
	var order Order
	if err := s.db.Where("id = ?", orderID).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order not found")
		}
		return nil, errors.New("failed to find order")
	}

	// Update status
	order.Status = req.Status
	order.UpdatedAt = time.Now()

	if err := s.db.Save(&order).Error; err != nil {
		return nil, errors.New("failed to update order status")
	}

	// Load updated order with items
	if err := s.db.Preload("Items").Preload("Items.Product").First(&order, order.ID).Error; err != nil {
		return nil, errors.New("failed to load updated order")
	}

	return &order, nil
}

// UpdatePaymentStatus updates the payment status of an order
func (s *OrderService) UpdatePaymentStatus(orderID uuid.UUID, paymentStatus string, paymentIntentID string) (*Order, error) {
	var order Order
	if err := s.db.Where("id = ?", orderID).First(&order).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order not found")
		}
		return nil, errors.New("failed to find order")
	}

	// Update payment status
	order.PaymentStatus = paymentStatus
	if paymentIntentID != "" {
		order.PaymentIntentID = paymentIntentID
	}
	order.UpdatedAt = time.Now()

	if err := s.db.Save(&order).Error; err != nil {
		return nil, errors.New("failed to update payment status")
	}

	return &order, nil
}

// CancelOrder cancels an order and releases inventory
func (s *OrderService) CancelOrder(orderID uuid.UUID) (*Order, error) {
	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	var order Order
	if err := tx.Preload("Items").Where("id = ?", orderID).First(&order).Error; err != nil {
		tx.Rollback()
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("order not found")
		}
		return nil, errors.New("failed to find order")
	}

	// Check if order can be cancelled
	if order.Status == "shipped" || order.Status == "delivered" {
		tx.Rollback()
		return nil, errors.New("cannot cancel shipped or delivered orders")
	}

	// Release inventory
	if err := s.releaseInventory(tx, order.Items); err != nil {
		tx.Rollback()
		return nil, err
	}

	// Update order status
	order.Status = "cancelled"
	order.UpdatedAt = time.Now()

	if err := tx.Save(&order).Error; err != nil {
		tx.Rollback()
		return nil, errors.New("failed to cancel order")
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, errors.New("failed to commit cancellation")
	}

	return &order, nil
}

// generateOrderNumber generates a unique order number
func (s *OrderService) generateOrderNumber() string {
	return fmt.Sprintf("ORD-%d", time.Now().Unix())
}

// checkInventory verifies inventory availability
func (s *OrderService) checkInventory(tx *gorm.DB, productID uuid.UUID, variantID *uuid.UUID, quantity int) error {
	var inventory models.Inventory
	query := tx.Where("product_id = ?", productID)

	if variantID != nil {
		query = query.Where("variant_id = ?", *variantID)
	} else {
		query = query.Where("variant_id IS NULL")
	}

	if err := query.First(&inventory).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return errors.New("inventory not found for product")
		}
		return errors.New("failed to check inventory")
	}

	if inventory.QuantityAvailable < quantity {
		return errors.New("insufficient inventory")
	}

	return nil
}

// reserveInventory reserves inventory for order items
func (s *OrderService) reserveInventory(tx *gorm.DB, orderItems []OrderItem) error {
	for _, item := range orderItems {
		var inventory models.Inventory
		query := tx.Where("product_id = ?", item.ProductID)

		if item.VariantID != nil {
			query = query.Where("variant_id = ?", *item.VariantID)
		} else {
			query = query.Where("variant_id IS NULL")
		}

		if err := query.First(&inventory).Error; err != nil {
			return fmt.Errorf("inventory not found for product %s", item.ProductID)
		}

		// Update inventory
		inventory.QuantityAvailable -= item.Quantity
		inventory.QuantityReserved += item.Quantity

		if err := tx.Save(&inventory).Error; err != nil {
			return fmt.Errorf("failed to reserve inventory for product %s", item.ProductID)
		}
	}

	return nil
}

// releaseInventory releases reserved inventory
func (s *OrderService) releaseInventory(tx *gorm.DB, orderItems []OrderItem) error {
	for _, item := range orderItems {
		var inventory models.Inventory
		query := tx.Where("product_id = ?", item.ProductID)

		if item.VariantID != nil {
			query = query.Where("variant_id = ?", *item.VariantID)
		} else {
			query = query.Where("variant_id IS NULL")
		}

		if err := query.First(&inventory).Error; err != nil {
			continue // Skip if inventory not found
		}

		// Release inventory
		inventory.QuantityAvailable += item.Quantity
		inventory.QuantityReserved -= item.Quantity

		if err := tx.Save(&inventory).Error; err != nil {
			return fmt.Errorf("failed to release inventory for product %s", item.ProductID)
		}
	}

	return nil
}
