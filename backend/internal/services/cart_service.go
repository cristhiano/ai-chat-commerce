package services

import (
	"chat-ecommerce-backend/internal/models"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ShoppingCartService handles shopping cart business logic
type ShoppingCartService struct {
	db *gorm.DB
}

// NewShoppingCartService creates a new ShoppingCartService
func NewShoppingCartService(db *gorm.DB) *ShoppingCartService {
	return &ShoppingCartService{
		db: db,
	}
}

// CartItem represents an item in the shopping cart
type CartItem struct {
	ProductID   uuid.UUID  `json:"product_id"`
	VariantID   *uuid.UUID `json:"variant_id,omitempty"`
	Quantity    int        `json:"quantity"`
	UnitPrice   float64    `json:"unit_price"`
	TotalPrice  float64    `json:"total_price"`
	ProductName string     `json:"product_name"`
	SKU         string     `json:"sku"`
}

// AddToCartRequest represents the request to add an item to cart
type AddToCartRequest struct {
	ProductID uuid.UUID  `json:"product_id" binding:"required"`
	VariantID *uuid.UUID `json:"variant_id,omitempty"`
	Quantity  int        `json:"quantity" binding:"required,min=1"`
}

// UpdateCartItemRequest represents the request to update a cart item
type UpdateCartItemRequest struct {
	ProductID uuid.UUID  `json:"product_id" binding:"required"`
	VariantID *uuid.UUID `json:"variant_id,omitempty"`
	Quantity  int        `json:"quantity" binding:"required,min=0"`
}

// CartResponse represents the cart response
type CartResponse struct {
	Items          []CartItem `json:"items"`
	Subtotal       float64    `json:"subtotal"`
	TaxAmount      float64    `json:"tax_amount"`
	ShippingAmount float64    `json:"shipping_amount"`
	TotalAmount    float64    `json:"total_amount"`
	Currency       string     `json:"currency"`
	ItemCount      int        `json:"item_count"`
}

// GetCart retrieves the shopping cart for a user or session
func (s *ShoppingCartService) GetCart(sessionID string, userID *uuid.UUID) (*CartResponse, error) {
	var cart models.ShoppingCart

	query := s.db.Where("session_id = ?", sessionID)
	if userID != nil {
		query = query.Or("user_id = ?", *userID)
	}

	if err := query.First(&cart).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Return empty cart
			return &CartResponse{
				Items:          []CartItem{},
				Subtotal:       0,
				TaxAmount:      0,
				ShippingAmount: 0,
				TotalAmount:    0,
				Currency:       "USD",
				ItemCount:      0,
			}, nil
		}
		return nil, fmt.Errorf("failed to fetch cart: %w", err)
	}

	// Parse cart items from JSON
	var items []CartItem
	if cart.Items != nil {
		if err := json.Unmarshal(cart.Items, &items); err != nil {
			return nil, fmt.Errorf("failed to parse cart items: %w", err)
		}
	}

	// Calculate item count
	itemCount := 0
	for _, item := range items {
		itemCount += item.Quantity
	}

	return &CartResponse{
		Items:          items,
		Subtotal:       cart.Subtotal,
		TaxAmount:      cart.TaxAmount,
		ShippingAmount: cart.ShippingAmount,
		TotalAmount:    cart.TotalAmount,
		Currency:       cart.Currency,
		ItemCount:      itemCount,
	}, nil
}

// AddToCart adds an item to the shopping cart
func (s *ShoppingCartService) AddToCart(sessionID string, userID *uuid.UUID, req AddToCartRequest) error {
	// Get or create cart
	cart, err := s.getOrCreateCart(sessionID, userID)
	if err != nil {
		return err
	}

	// Get product details
	var product models.Product
	if err := s.db.Where("id = ? AND status = ?", req.ProductID, "active").First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("product not found or inactive")
		}
		return fmt.Errorf("failed to fetch product: %w", err)
	}

	// Check inventory if variant is specified
	if req.VariantID != nil {
		var inventory models.Inventory
		if err := s.db.Where("product_id = ? AND variant_id = ?", req.ProductID, *req.VariantID).First(&inventory).Error; err != nil {
			return fmt.Errorf("inventory not found for variant")
		}
		if inventory.QuantityAvailable < req.Quantity {
			return fmt.Errorf("insufficient inventory")
		}
	} else {
		// Check base product inventory
		var inventory models.Inventory
		if err := s.db.Where("product_id = ? AND variant_id IS NULL", req.ProductID).First(&inventory).Error; err == nil {
			if inventory.QuantityAvailable < req.Quantity {
				return fmt.Errorf("insufficient inventory")
			}
		}
	}

	// Parse existing cart items
	var items []CartItem
	if cart.Items != nil {
		if err := json.Unmarshal(cart.Items, &items); err != nil {
			return fmt.Errorf("failed to parse cart items: %w", err)
		}
	}

	// Calculate unit price
	unitPrice := product.Price
	if req.VariantID != nil {
		var variant models.ProductVariant
		if err := s.db.Where("id = ?", *req.VariantID).First(&variant).Error; err == nil {
			unitPrice += variant.PriceModifier
		}
	}

	// Check if item already exists in cart
	itemFound := false
	for i, item := range items {
		if item.ProductID == req.ProductID &&
			((req.VariantID == nil && item.VariantID == nil) ||
				(req.VariantID != nil && item.VariantID != nil && *item.VariantID == *req.VariantID)) {
			// Update existing item
			items[i].Quantity += req.Quantity
			items[i].TotalPrice = float64(items[i].Quantity) * items[i].UnitPrice
			itemFound = true
			break
		}
	}

	// Add new item if not found
	if !itemFound {
		newItem := CartItem{
			ProductID:   req.ProductID,
			VariantID:   req.VariantID,
			Quantity:    req.Quantity,
			UnitPrice:   unitPrice,
			TotalPrice:  float64(req.Quantity) * unitPrice,
			ProductName: product.Name,
			SKU:         product.SKU,
		}
		items = append(items, newItem)
	}

	// Calculate totals
	subtotal := 0.0
	for _, item := range items {
		subtotal += item.TotalPrice
	}

	// Convert items to JSON
	itemsJSON, err := json.Marshal(items)
	if err != nil {
		return fmt.Errorf("failed to marshal cart items: %w", err)
	}

	// Update cart
	updates := map[string]interface{}{
		"items":           itemsJSON,
		"subtotal":        subtotal,
		"tax_amount":      0, // TODO: Calculate tax
		"shipping_amount": 0, // TODO: Calculate shipping
		"total_amount":    subtotal,
		"updated_at":      time.Now(),
	}

	if userID != nil {
		updates["user_id"] = *userID
	}

	if err := s.db.Model(cart).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update cart: %w", err)
	}

	return nil
}

// UpdateCartItem updates the quantity of an item in the cart
func (s *ShoppingCartService) UpdateCartItem(sessionID string, userID *uuid.UUID, req UpdateCartItemRequest) error {
	// Get cart
	cart, err := s.getOrCreateCart(sessionID, userID)
	if err != nil {
		return err
	}

	// Parse existing cart items
	var items []CartItem
	if cart.Items != nil {
		if err := json.Unmarshal(cart.Items, &items); err != nil {
			return fmt.Errorf("failed to parse cart items: %w", err)
		}
	}

	// Find and update item
	itemFound := false
	for i, item := range items {
		if item.ProductID == req.ProductID &&
			((req.VariantID == nil && item.VariantID == nil) ||
				(req.VariantID != nil && item.VariantID != nil && *item.VariantID == *req.VariantID)) {
			if req.Quantity == 0 {
				// Remove item
				items = append(items[:i], items[i+1:]...)
			} else {
				// Update quantity
				items[i].Quantity = req.Quantity
				items[i].TotalPrice = float64(req.Quantity) * items[i].UnitPrice
			}
			itemFound = true
			break
		}
	}

	if !itemFound {
		return fmt.Errorf("item not found in cart")
	}

	// Calculate totals
	subtotal := 0.0
	for _, item := range items {
		subtotal += item.TotalPrice
	}

	// Convert items to JSON
	itemsJSON, err := json.Marshal(items)
	if err != nil {
		return fmt.Errorf("failed to marshal cart items: %w", err)
	}

	// Update cart
	updates := map[string]interface{}{
		"items":           itemsJSON,
		"subtotal":        subtotal,
		"tax_amount":      0, // TODO: Calculate tax
		"shipping_amount": 0, // TODO: Calculate shipping
		"total_amount":    subtotal,
		"updated_at":      time.Now(),
	}

	if err := s.db.Model(cart).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update cart: %w", err)
	}

	return nil
}

// RemoveFromCart removes an item from the cart
func (s *ShoppingCartService) RemoveFromCart(sessionID string, userID *uuid.UUID, productID uuid.UUID, variantID *uuid.UUID) error {
	req := UpdateCartItemRequest{
		ProductID: productID,
		VariantID: variantID,
		Quantity:  0, // Setting to 0 removes the item
	}
	return s.UpdateCartItem(sessionID, userID, req)
}

// ClearCart clears all items from the cart
func (s *ShoppingCartService) ClearCart(sessionID string, userID *uuid.UUID) error {
	// Get cart
	cart, err := s.getOrCreateCart(sessionID, userID)
	if err != nil {
		return err
	}

	// Clear items
	updates := map[string]interface{}{
		"items":           datatypes.JSON("[]"),
		"subtotal":        0,
		"tax_amount":      0,
		"shipping_amount": 0,
		"total_amount":    0,
		"updated_at":      time.Now(),
	}

	if err := s.db.Model(cart).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to clear cart: %w", err)
	}

	return nil
}

// getOrCreateCart gets an existing cart or creates a new one
func (s *ShoppingCartService) getOrCreateCart(sessionID string, userID *uuid.UUID) (*models.ShoppingCart, error) {
	var cart models.ShoppingCart

	query := s.db.Where("session_id = ?", sessionID)
	if userID != nil {
		query = query.Or("user_id = ?", *userID)
	}

	if err := query.First(&cart).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			// Create new cart
			cart = models.ShoppingCart{
				SessionID:      sessionID,
				UserID:         userID,
				Items:          datatypes.JSON("[]"),
				Subtotal:       0,
				TaxAmount:      0,
				ShippingAmount: 0,
				TotalAmount:    0,
				Currency:       "USD",
			}

			if err := s.db.Create(&cart).Error; err != nil {
				return nil, fmt.Errorf("failed to create cart: %w", err)
			}
		} else {
			return nil, fmt.Errorf("failed to fetch cart: %w", err)
		}
	}

	return &cart, nil
}

// CalculateCartTotals calculates tax and shipping for the cart
func (s *ShoppingCartService) CalculateCartTotals(cart *CartResponse) (*CartResponse, error) {
	// TODO: Implement tax calculation based on location
	// TODO: Implement shipping calculation based on weight/distance

	// For now, just return the cart with basic calculations
	taxAmount := cart.Subtotal * 0.08 // 8% tax
	shippingAmount := 0.0
	if cart.Subtotal > 0 && cart.Subtotal < 50 {
		shippingAmount = 5.99 // Standard shipping
	}

	totalAmount := cart.Subtotal + taxAmount + shippingAmount

	return &CartResponse{
		Items:          cart.Items,
		Subtotal:       cart.Subtotal,
		TaxAmount:      taxAmount,
		ShippingAmount: shippingAmount,
		TotalAmount:    totalAmount,
		Currency:       cart.Currency,
		ItemCount:      cart.ItemCount,
	}, nil
}
