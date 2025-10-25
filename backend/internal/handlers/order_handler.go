package handlers

import (
	"chat-ecommerce-backend/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// OrderHandler handles order-related HTTP requests
type OrderHandler struct {
	orderService *services.OrderService
}

// NewOrderHandler creates a new OrderHandler
func NewOrderHandler(orderService *services.OrderService) *OrderHandler {
	return &OrderHandler{
		orderService: orderService,
	}
}

// CreateOrder handles POST /api/v1/orders
func (h *OrderHandler) CreateOrder(c *gin.Context) {
	var req services.CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Get user ID from context if authenticated
	if userID, exists := c.Get("user_id"); exists {
		req.UserID = userID.(uuid.UUID)
	}

	// Get session ID from context or generate one
	if sessionID, exists := c.Get("session_id"); exists {
		req.SessionID = sessionID.(string)
	} else {
		req.SessionID = uuid.New().String()
	}

	order, err := h.orderService.CreateOrder(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"order": order})
}

// GetOrder handles GET /api/v1/orders/:id
func (h *OrderHandler) GetOrder(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}

	order, err := h.orderService.GetOrderByID(orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Check if user can access this order
	if userID, exists := c.Get("user_id"); exists {
		if order.UserID != userID.(uuid.UUID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
	} else {
		// For anonymous users, check session ID
		if sessionID, exists := c.Get("session_id"); exists {
			if order.SessionID != sessionID.(string) {
				c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
				return
			}
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"order": order})
}

// GetOrderByNumber handles GET /api/v1/orders/number/:number
func (h *OrderHandler) GetOrderByNumber(c *gin.Context) {
	orderNumber := c.Param("number")
	if orderNumber == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order number is required"})
		return
	}

	order, err := h.orderService.GetOrderByNumber(orderNumber)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Check if user can access this order
	if userID, exists := c.Get("user_id"); exists {
		if order.UserID != userID.(uuid.UUID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
	} else {
		// For anonymous users, check session ID
		if sessionID, exists := c.Get("session_id"); exists {
			if order.SessionID != sessionID.(string) {
				c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
				return
			}
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{"order": order})
}

// GetUserOrders handles GET /api/v1/user/orders
func (h *OrderHandler) GetUserOrders(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
		return
	}

	// Parse pagination parameters
	page := 1
	limit := 10

	if pageStr := c.Query("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil && l > 0 && l <= 100 {
			limit = l
		}
	}

	orders, total, err := h.orderService.GetUserOrders(userID.(uuid.UUID), page, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	totalPages := (total + int64(limit) - 1) / int64(limit)

	c.JSON(http.StatusOK, gin.H{
		"orders":       orders,
		"total":        total,
		"page":         page,
		"limit":        limit,
		"total_pages":  totalPages,
		"has_next":     page < int(totalPages),
		"has_previous": page > 1,
	})
}

// UpdateOrderStatus handles PUT /api/v1/orders/:id/status (Admin only)
func (h *OrderHandler) UpdateOrderStatus(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}

	var req services.UpdateOrderStatusRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.orderService.UpdateOrderStatus(orderID, &req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"order": order})
}

// UpdatePaymentStatus handles PUT /api/v1/orders/:id/payment-status
func (h *OrderHandler) UpdatePaymentStatus(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}

	var req struct {
		PaymentStatus   string `json:"payment_status" binding:"required"`
		PaymentIntentID string `json:"payment_intent_id"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := h.orderService.UpdatePaymentStatus(orderID, req.PaymentStatus, req.PaymentIntentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"order": order})
}

// CancelOrder handles DELETE /api/v1/orders/:id
func (h *OrderHandler) CancelOrder(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}

	order, err := h.orderService.CancelOrder(orderID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"order": order})
}

// GetOrderSummary handles GET /api/v1/orders/:id/summary
func (h *OrderHandler) GetOrderSummary(c *gin.Context) {
	orderIDStr := c.Param("id")
	orderID, err := uuid.Parse(orderIDStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid order ID"})
		return
	}

	order, err := h.orderService.GetOrderByID(orderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	// Check access permissions
	if userID, exists := c.Get("user_id"); exists {
		if order.UserID != userID.(uuid.UUID) {
			c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
			return
		}
	} else {
		if sessionID, exists := c.Get("session_id"); exists {
			if order.SessionID != sessionID.(string) {
				c.JSON(http.StatusForbidden, gin.H{"error": "access denied"})
				return
			}
		} else {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "authentication required"})
			return
		}
	}

	// Return summary without sensitive information
	summary := gin.H{
		"order_number":    order.OrderNumber,
		"status":          order.Status,
		"payment_status":  order.PaymentStatus,
		"subtotal":        order.Subtotal,
		"tax_amount":      order.TaxAmount,
		"shipping_amount": order.ShippingAmount,
		"total_amount":    order.TotalAmount,
		"currency":        order.Currency,
		"created_at":      order.CreatedAt,
		"updated_at":      order.UpdatedAt,
		"item_count":      len(order.Items),
	}

	c.JSON(http.StatusOK, gin.H{"summary": summary})
}
