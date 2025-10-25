package handlers

import (
	"chat-ecommerce-backend/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// CartHandler handles shopping cart HTTP requests
type CartHandler struct {
	cartService *services.ShoppingCartService
}

// NewCartHandler creates a new CartHandler
func NewCartHandler(cartService *services.ShoppingCartService) *CartHandler {
	return &CartHandler{
		cartService: cartService,
	}
}

// GetCart handles GET /api/v1/cart
func (h *CartHandler) GetCart(c *gin.Context) {
	// Get session ID from header or generate one
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = c.GetString("session_id")
		if sessionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID is required"})
			return
		}
	}

	// Get user ID from context (set by auth middleware)
	var userID *uuid.UUID
	if userIDStr, exists := c.Get("user_id"); exists {
		if id, ok := userIDStr.(uuid.UUID); ok {
			userID = &id
		}
	}

	cart, err := h.cartService.GetCart(sessionID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, cart)
}

// AddToCart handles POST /api/v1/cart/add
func (h *CartHandler) AddToCart(c *gin.Context) {
	// Get session ID from header or generate one
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = c.GetString("session_id")
		if sessionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID is required"})
			return
		}
	}

	// Get user ID from context (set by auth middleware)
	var userID *uuid.UUID
	if userIDStr, exists := c.Get("user_id"); exists {
		if id, ok := userIDStr.(uuid.UUID); ok {
			userID = &id
		}
	}

	var req services.AddToCartRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.cartService.AddToCart(sessionID, userID, req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item added to cart successfully"})
}

// UpdateCartItem handles PUT /api/v1/cart/update
func (h *CartHandler) UpdateCartItem(c *gin.Context) {
	// Get session ID from header or generate one
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = c.GetString("session_id")
		if sessionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID is required"})
			return
		}
	}

	// Get user ID from context (set by auth middleware)
	var userID *uuid.UUID
	if userIDStr, exists := c.Get("user_id"); exists {
		if id, ok := userIDStr.(uuid.UUID); ok {
			userID = &id
		}
	}

	var req services.UpdateCartItemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.cartService.UpdateCartItem(sessionID, userID, req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cart item updated successfully"})
}

// RemoveFromCart handles DELETE /api/v1/cart/remove/:product_id
func (h *CartHandler) RemoveFromCart(c *gin.Context) {
	// Get session ID from header or generate one
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = c.GetString("session_id")
		if sessionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID is required"})
			return
		}
	}

	// Get user ID from context (set by auth middleware)
	var userID *uuid.UUID
	if userIDStr, exists := c.Get("user_id"); exists {
		if id, ok := userIDStr.(uuid.UUID); ok {
			userID = &id
		}
	}

	productIDStr := c.Param("product_id")
	productID, err := uuid.Parse(productIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	// Get variant ID from query parameter if provided
	var variantID *uuid.UUID
	if variantIDStr := c.Query("variant_id"); variantIDStr != "" {
		if id, err := uuid.Parse(variantIDStr); err == nil {
			variantID = &id
		}
	}

	if err := h.cartService.RemoveFromCart(sessionID, userID, productID, variantID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Item removed from cart successfully"})
}

// ClearCart handles DELETE /api/v1/cart/clear
func (h *CartHandler) ClearCart(c *gin.Context) {
	// Get session ID from header or generate one
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = c.GetString("session_id")
		if sessionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID is required"})
			return
		}
	}

	// Get user ID from context (set by auth middleware)
	var userID *uuid.UUID
	if userIDStr, exists := c.Get("user_id"); exists {
		if id, ok := userIDStr.(uuid.UUID); ok {
			userID = &id
		}
	}

	if err := h.cartService.ClearCart(sessionID, userID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "Cart cleared successfully"})
}

// CalculateTotals handles POST /api/v1/cart/calculate
func (h *CartHandler) CalculateTotals(c *gin.Context) {
	// Get session ID from header or generate one
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = c.GetString("session_id")
		if sessionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID is required"})
			return
		}
	}

	// Get user ID from context (set by auth middleware)
	var userID *uuid.UUID
	if userIDStr, exists := c.Get("user_id"); exists {
		if id, ok := userIDStr.(uuid.UUID); ok {
			userID = &id
		}
	}

	cart, err := h.cartService.GetCart(sessionID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Calculate totals with tax and shipping
	totals, err := h.cartService.CalculateCartTotals(cart)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, totals)
}

// GetCartItemCount handles GET /api/v1/cart/count
func (h *CartHandler) GetCartItemCount(c *gin.Context) {
	// Get session ID from header or generate one
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID = c.GetString("session_id")
		if sessionID == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Session ID is required"})
			return
		}
	}

	// Get user ID from context (set by auth middleware)
	var userID *uuid.UUID
	if userIDStr, exists := c.Get("user_id"); exists {
		if id, ok := userIDStr.(uuid.UUID); ok {
			userID = &id
		}
	}

	cart, err := h.cartService.GetCart(sessionID, userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"item_count": cart.ItemCount})
}
