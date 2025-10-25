package handlers

import (
	"chat-ecommerce-backend/internal/services"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// PaymentHandler handles payment-related HTTP requests
type PaymentHandler struct {
	paymentService *services.PaymentService
	orderService   *services.OrderService
}

// NewPaymentHandler creates a new PaymentHandler
func NewPaymentHandler(paymentService *services.PaymentService, orderService *services.OrderService) *PaymentHandler {
	return &PaymentHandler{
		paymentService: paymentService,
		orderService:   orderService,
	}
}

// CreatePaymentIntent handles POST /api/v1/payments/create-intent
func (h *PaymentHandler) CreatePaymentIntent(c *gin.Context) {
	var req services.CreatePaymentIntentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify order exists and belongs to user
	order, err := h.orderService.GetOrderByID(req.OrderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	// Check if user can access this order
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

	// Verify amount matches order total
	expectedAmount := int64(order.TotalAmount * 100) // Convert to cents
	if req.Amount != expectedAmount {
		c.JSON(http.StatusBadRequest, gin.H{"error": "amount does not match order total"})
		return
	}

	// Create payment intent
	response, err := h.paymentService.CreatePaymentIntent(&req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Update order with payment intent ID
	_, err = h.orderService.UpdatePaymentStatus(req.OrderID, "processing", response.ID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update order payment status"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{"payment_intent": response})
}

// ConfirmPayment handles POST /api/v1/payments/confirm
func (h *PaymentHandler) ConfirmPayment(c *gin.Context) {
	var req services.ConfirmPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Verify order exists and belongs to user
	order, err := h.orderService.GetOrderByID(req.OrderID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}

	// Check if user can access this order
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

	// Confirm payment
	status, err := h.paymentService.ConfirmPayment(&req)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update order payment status based on payment status
	var orderPaymentStatus string
	switch status.Status {
	case "succeeded":
		orderPaymentStatus = "paid"
	case "requires_payment_method", "requires_confirmation":
		orderPaymentStatus = "pending"
	case "canceled":
		orderPaymentStatus = "canceled"
	default:
		orderPaymentStatus = "failed"
	}

	_, err = h.orderService.UpdatePaymentStatus(req.OrderID, orderPaymentStatus, req.PaymentIntentID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update order payment status"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payment_status": status})
}

// GetPaymentStatus handles GET /api/v1/payments/:payment_intent_id/status
func (h *PaymentHandler) GetPaymentStatus(c *gin.Context) {
	paymentIntentID := c.Param("payment_intent_id")
	if paymentIntentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment intent ID is required"})
		return
	}

	status, err := h.paymentService.GetPaymentStatus(paymentIntentID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payment_status": status})
}

// CancelPayment handles POST /api/v1/payments/:payment_intent_id/cancel
func (h *PaymentHandler) CancelPayment(c *gin.Context) {
	paymentIntentID := c.Param("payment_intent_id")
	if paymentIntentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment intent ID is required"})
		return
	}

	status, err := h.paymentService.CancelPaymentIntent(paymentIntentID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payment_status": status})
}

// RefundPayment handles POST /api/v1/payments/:payment_intent_id/refund
func (h *PaymentHandler) RefundPayment(c *gin.Context) {
	paymentIntentID := c.Param("payment_intent_id")
	if paymentIntentID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "payment intent ID is required"})
		return
	}

	var req struct {
		Amount int64  `json:"amount"`
		Reason string `json:"reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Default reason if not provided
	if req.Reason == "" {
		req.Reason = "requested_by_customer"
	}

	status, err := h.paymentService.RefundPayment(paymentIntentID, req.Amount, req.Reason)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"payment_status": status})
}

// HandleWebhook handles POST /api/v1/payments/webhook
func (h *PaymentHandler) HandleWebhook(c *gin.Context) {
	// Get the raw body
	body, err := c.GetRawData()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read request body"})
		return
	}

	// Get the signature header
	signature := c.GetHeader("Stripe-Signature")
	if signature == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing stripe signature"})
		return
	}

	// Get webhook secret from environment
	webhookSecret := c.GetHeader("X-Webhook-Secret")
	if webhookSecret == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing webhook secret"})
		return
	}

	// Validate webhook signature
	err = h.paymentService.ValidateWebhookSignature(body, signature, webhookSecret)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Parse webhook event (simplified - in real implementation, use Stripe's webhook parsing)
	var eventData map[string]interface{}
	if err := c.ShouldBindJSON(&eventData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid webhook data"})
		return
	}

	// Extract event type
	eventType, ok := eventData["type"].(string)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid event type"})
		return
	}

	// Process webhook event
	err = h.paymentService.ProcessWebhookEvent(eventType, eventData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"status": "webhook processed"})
}

// GetPaymentMethods handles GET /api/v1/payments/methods
func (h *PaymentHandler) GetPaymentMethods(c *gin.Context) {
	// Return available payment methods
	methods := []gin.H{
		{
			"id":          "card",
			"name":        "Credit/Debit Card",
			"description": "Pay with Visa, Mastercard, American Express",
			"enabled":     true,
		},
		{
			"id":          "apple_pay",
			"name":        "Apple Pay",
			"description": "Pay with Apple Pay",
			"enabled":     false, // Disabled for now
		},
		{
			"id":          "google_pay",
			"name":        "Google Pay",
			"description": "Pay with Google Pay",
			"enabled":     false, // Disabled for now
		},
	}

	c.JSON(http.StatusOK, gin.H{"payment_methods": methods})
}
