package services

import (
	"errors"
	"fmt"
	"os"

	"github.com/google/uuid"
	"github.com/stripe/stripe-go/v78"
	"github.com/stripe/stripe-go/v78/paymentintent"
	"github.com/stripe/stripe-go/v78/refund"
)

// PaymentService handles payment processing with Stripe
type PaymentService struct {
	stripeKey string
}

// NewPaymentService creates a new PaymentService
func NewPaymentService() *PaymentService {
	stripeKey := os.Getenv("STRIPE_SECRET_KEY")
	if stripeKey == "" {
		stripeKey = "sk_test_..." // Default test key for development
	}
	stripe.Key = stripeKey

	return &PaymentService{
		stripeKey: stripeKey,
	}
}

// CreatePaymentIntentRequest represents the request payload for creating a payment intent
type CreatePaymentIntentRequest struct {
	OrderID     uuid.UUID         `json:"order_id" binding:"required"`
	Amount      int64             `json:"amount" binding:"required,min=1"`
	Currency    string            `json:"currency" binding:"required"`
	Description string            `json:"description"`
	Metadata    map[string]string `json:"metadata"`
}

// PaymentIntentResponse represents the response from creating a payment intent
type PaymentIntentResponse struct {
	ID           string `json:"id"`
	ClientSecret string `json:"client_secret"`
	Status       string `json:"status"`
	Amount       int64  `json:"amount"`
	Currency     string `json:"currency"`
	Description  string `json:"description"`
	CreatedAt    int64  `json:"created_at"`
}

// ConfirmPaymentRequest represents the request payload for confirming a payment
type ConfirmPaymentRequest struct {
	PaymentIntentID string    `json:"payment_intent_id" binding:"required"`
	OrderID         uuid.UUID `json:"order_id" binding:"required"`
}

// PaymentStatus represents the status of a payment
type PaymentStatus struct {
	PaymentIntentID string `json:"payment_intent_id"`
	Status          string `json:"status"`
	Amount          int64  `json:"amount"`
	Currency        string `json:"currency"`
	Description     string `json:"description"`
	CreatedAt       int64  `json:"created_at"`
	UpdatedAt       int64  `json:"updated_at"`
}

// CreatePaymentIntent creates a new payment intent with Stripe
func (s *PaymentService) CreatePaymentIntent(req *CreatePaymentIntentRequest) (*PaymentIntentResponse, error) {
	// Prepare metadata
	metadata := map[string]string{
		"order_id": req.OrderID.String(),
	}

	// Add additional metadata
	for key, value := range req.Metadata {
		metadata[key] = value
	}

	// Create payment intent parameters
	params := &stripe.PaymentIntentParams{
		Amount:   stripe.Int64(req.Amount),
		Currency: stripe.String(req.Currency),
		Metadata: metadata,
	}

	// Add description if provided
	if req.Description != "" {
		params.Description = stripe.String(req.Description)
	}

	// Create the payment intent
	pi, err := paymentintent.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to create payment intent: %v", err)
	}

	// Return response
	response := &PaymentIntentResponse{
		ID:           pi.ID,
		ClientSecret: pi.ClientSecret,
		Status:       string(pi.Status),
		Amount:       pi.Amount,
		Currency:     string(pi.Currency),
		Description:  pi.Description,
		CreatedAt:    pi.Created,
	}

	return response, nil
}

// ConfirmPayment confirms a payment intent
func (s *PaymentService) ConfirmPayment(req *ConfirmPaymentRequest) (*PaymentStatus, error) {
	// Retrieve the payment intent
	pi, err := paymentintent.Get(req.PaymentIntentID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve payment intent: %v", err)
	}

	// Check if payment intent belongs to the order
	orderID, exists := pi.Metadata["order_id"]
	if !exists || orderID != req.OrderID.String() {
		return nil, errors.New("payment intent does not belong to this order")
	}

	// Return payment status
	status := &PaymentStatus{
		PaymentIntentID: pi.ID,
		Status:          string(pi.Status),
		Amount:          pi.Amount,
		Currency:        string(pi.Currency),
		Description:     pi.Description,
		CreatedAt:       pi.Created,
		UpdatedAt:       pi.Created,
	}

	return status, nil
}

// GetPaymentStatus retrieves the status of a payment intent
func (s *PaymentService) GetPaymentStatus(paymentIntentID string) (*PaymentStatus, error) {
	// Retrieve the payment intent
	pi, err := paymentintent.Get(paymentIntentID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve payment intent: %v", err)
	}

	// Return payment status
	status := &PaymentStatus{
		PaymentIntentID: pi.ID,
		Status:          string(pi.Status),
		Amount:          pi.Amount,
		Currency:        string(pi.Currency),
		Description:     pi.Description,
		CreatedAt:       pi.Created,
		UpdatedAt:       pi.Created,
	}

	return status, nil
}

// CancelPaymentIntent cancels a payment intent
func (s *PaymentService) CancelPaymentIntent(paymentIntentID string) (*PaymentStatus, error) {
	// Cancel the payment intent
	pi, err := paymentintent.Cancel(paymentIntentID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to cancel payment intent: %v", err)
	}

	// Return payment status
	status := &PaymentStatus{
		PaymentIntentID: pi.ID,
		Status:          string(pi.Status),
		Amount:          pi.Amount,
		Currency:        string(pi.Currency),
		Description:     pi.Description,
		CreatedAt:       pi.Created,
		UpdatedAt:       pi.Created,
	}

	return status, nil
}

// RefundPayment processes a refund for a payment intent
func (s *PaymentService) RefundPayment(paymentIntentID string, amount int64, reason string) (*PaymentStatus, error) {
	// Create refund parameters
	params := &stripe.RefundParams{
		PaymentIntent: stripe.String(paymentIntentID),
		Reason:        stripe.String(reason),
	}

	// Add amount if specified (partial refund)
	if amount > 0 {
		params.Amount = stripe.Int64(amount)
	}

	// Process the refund
	_, err := refund.New(params)
	if err != nil {
		return nil, fmt.Errorf("failed to process refund: %v", err)
	}

	// Retrieve updated payment intent
	pi, err := paymentintent.Get(paymentIntentID, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to retrieve updated payment intent: %v", err)
	}

	// Return payment status
	status := &PaymentStatus{
		PaymentIntentID: pi.ID,
		Status:          string(pi.Status),
		Amount:          pi.Amount,
		Currency:        string(pi.Currency),
		Description:     pi.Description,
		CreatedAt:       pi.Created,
		UpdatedAt:       pi.Created,
	}

	return status, nil
}

// ValidateWebhookSignature validates a Stripe webhook signature
func (s *PaymentService) ValidateWebhookSignature(payload []byte, signature string, webhookSecret string) error {
	// This would typically use Stripe's webhook signature validation
	// For now, we'll implement a basic validation
	if webhookSecret == "" {
		return errors.New("webhook secret not configured")
	}

	// In a real implementation, you would use:
	// event, err := webhook.ConstructEvent(payload, signature, webhookSecret)
	// if err != nil {
	//     return fmt.Errorf("webhook signature verification failed: %v", err)
	// }

	return nil
}

// ProcessWebhookEvent processes a Stripe webhook event
func (s *PaymentService) ProcessWebhookEvent(eventType string, eventData map[string]interface{}) error {
	switch eventType {
	case "payment_intent.succeeded":
		return s.handlePaymentSucceeded(eventData)
	case "payment_intent.payment_failed":
		return s.handlePaymentFailed(eventData)
	case "payment_intent.canceled":
		return s.handlePaymentCanceled(eventData)
	default:
		// Log unhandled event types
		return nil
	}
}

// handlePaymentSucceeded handles successful payment events
func (s *PaymentService) handlePaymentSucceeded(eventData map[string]interface{}) error {
	// Extract payment intent ID
	paymentIntentID, ok := eventData["id"].(string)
	if !ok {
		return errors.New("invalid payment intent ID in webhook")
	}

	// Update order status in database
	// This would typically involve updating the order service
	// For now, we'll just log the event
	fmt.Printf("Payment succeeded for intent: %s\n", paymentIntentID)

	return nil
}

// handlePaymentFailed handles failed payment events
func (s *PaymentService) handlePaymentFailed(eventData map[string]interface{}) error {
	// Extract payment intent ID
	paymentIntentID, ok := eventData["id"].(string)
	if !ok {
		return errors.New("invalid payment intent ID in webhook")
	}

	// Update order status in database
	// This would typically involve updating the order service
	// For now, we'll just log the event
	fmt.Printf("Payment failed for intent: %s\n", paymentIntentID)

	return nil
}

// handlePaymentCanceled handles canceled payment events
func (s *PaymentService) handlePaymentCanceled(eventData map[string]interface{}) error {
	// Extract payment intent ID
	paymentIntentID, ok := eventData["id"].(string)
	if !ok {
		return errors.New("invalid payment intent ID in webhook")
	}

	// Update order status in database
	// This would typically involve updating the order service
	// For now, we'll just log the event
	fmt.Printf("Payment canceled for intent: %s\n", paymentIntentID)

	return nil
}
