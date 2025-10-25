package services

import (
	"chat-ecommerce-backend/internal/models"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sashabaranov/go-openai"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// ChatService handles chat-based shopping interactions
type ChatService struct {
	db             *gorm.DB
	openaiClient   *openai.Client
	productService *ProductService
	cartService    *ShoppingCartService
}

// NewChatService creates a new ChatService
func NewChatService(db *gorm.DB, productService *ProductService, cartService *ShoppingCartService) *ChatService {
	client := openai.NewClient(os.Getenv("OPENAI_API_KEY"))

	return &ChatService{
		db:             db,
		openaiClient:   client,
		productService: productService,
		cartService:    cartService,
	}
}

// ChatMessageService represents a message in the chat conversation for service layer
type ChatMessageService struct {
	ID        uuid.UUID              `json:"id"`
	SessionID string                 `json:"session_id"`
	UserID    *uuid.UUID             `json:"user_id,omitempty"`
	Role      string                 `json:"role"` // "user", "assistant", "system"
	Content   string                 `json:"content"`
	Metadata  map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt string                 `json:"created_at"`
}

// ChatSession represents a chat session
type ChatSession struct {
	ID        string                 `json:"id"`
	UserID    *uuid.UUID             `json:"user_id,omitempty"`
	Context   map[string]interface{} `json:"context"`
	CreatedAt string                 `json:"created_at"`
	UpdatedAt string                 `json:"updated_at"`
}

// ChatResponse represents the response from the chat service
type ChatResponse struct {
	Message     string                 `json:"message"`
	Actions     []ChatAction           `json:"actions,omitempty"`
	Suggestions []ProductSuggestion    `json:"suggestions,omitempty"`
	Context     map[string]interface{} `json:"context,omitempty"`
	Error       string                 `json:"error,omitempty"`
}

// ChatAction represents an action to be taken based on the chat
type ChatAction struct {
	Type    string                 `json:"type"` // "add_to_cart", "remove_from_cart", "show_product", "checkout", etc.
	Payload map[string]interface{} `json:"payload"`
}

// ProductSuggestion represents a product suggestion
type ProductSuggestion struct {
	Product    *models.Product `json:"product"`
	Reason     string          `json:"reason"`
	Confidence float64         `json:"confidence"`
}

// ProcessMessage processes a user message and returns a chat response
func (s *ChatService) ProcessMessage(sessionID string, userID *uuid.UUID, message string) (*ChatResponse, error) {
	// Get conversation history
	history, err := s.GetConversationHistory(sessionID, 10)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversation history: %v", err)
	}

	// Get current cart state
	cart, err := s.cartService.GetCart(sessionID, userID)
	if err != nil {
		log.Printf("Warning: failed to get cart: %v", err)
		cart = nil
	}

	// Get available products for context
	products, err := s.productService.GetProducts(ProductFilters{
		Status: "active",
		Limit:  20,
	})
	if err != nil {
		log.Printf("Warning: failed to get products: %v", err)
		products = nil
	}

	// Build system prompt
	systemPrompt := s.buildSystemPrompt(cart, products)

	// Prepare messages for OpenAI
	messages := []openai.ChatCompletionMessage{
		{
			Role:    openai.ChatMessageRoleSystem,
			Content: systemPrompt,
		},
	}

	// Add conversation history
	for _, msg := range history {
		role := openai.ChatMessageRoleUser
		if msg.Role == "assistant" {
			role = openai.ChatMessageRoleAssistant
		}
		messages = append(messages, openai.ChatCompletionMessage{
			Role:    role,
			Content: msg.Content,
		})
	}

	// Add current user message
	messages = append(messages, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: message,
	})

	// Call OpenAI API
	response, err := s.openaiClient.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:       openai.GPT4,
			Messages:    messages,
			MaxTokens:   500,
			Temperature: 0.7,
		},
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get OpenAI response: %v", err)
	}

	assistantMessage := response.Choices[0].Message.Content

	// Parse the response for actions
	actions, suggestions, err := s.parseResponse(assistantMessage, products)
	if err != nil {
		log.Printf("Warning: failed to parse response: %v", err)
	}

	// Execute actions
	for _, action := range actions {
		err := s.executeAction(action, userID, sessionID)
		if err != nil {
			log.Printf("Warning: failed to execute action %s: %v", action.Type, err)
		}
	}

	// Save messages to database
	err = s.saveMessage(sessionID, userID, "user", message, nil)
	if err != nil {
		log.Printf("Warning: failed to save user message: %v", err)
	}

	err = s.saveMessage(sessionID, userID, "assistant", assistantMessage, map[string]interface{}{
		"actions":     actions,
		"suggestions": suggestions,
	})
	if err != nil {
		log.Printf("Warning: failed to save assistant message: %v", err)
	}

	return &ChatResponse{
		Message:     assistantMessage,
		Actions:     actions,
		Suggestions: suggestions,
		Context: map[string]interface{}{
			"session_id": sessionID,
			"user_id":    userID,
		},
	}, nil
}

// buildSystemPrompt builds the system prompt for OpenAI
func (s *ChatService) buildSystemPrompt(cart *CartResponse, products *ProductListResponse) string {
	prompt := `You are a helpful shopping assistant for an e-commerce store. Your role is to help users find products, manage their cart, and complete purchases through natural conversation.

Available product categories:
- Electronics: Electronic devices and gadgets
- Clothing: Fashion and apparel  
- Books: Books and literature
- Home & Garden: Home improvement and garden supplies

Current cart status:`

	if cart != nil {
		prompt += fmt.Sprintf(`
- Items in cart: %d
- Total amount: $%.2f
- Items:`, cart.ItemCount, cart.TotalAmount)

		for _, item := range cart.Items {
			prompt += fmt.Sprintf("\n  - %s (Qty: %d, Price: $%.2f)", item.ProductName, item.Quantity, item.UnitPrice)
		}
	} else {
		prompt += "\n- Cart is empty"
	}

	prompt += `

Available products:`

	if products != nil {
		for _, product := range products.Products {
			prompt += fmt.Sprintf("\n- %s: %s (Price: $%.2f, SKU: %s)", product.Name, product.Description, product.Price, product.SKU)
		}
	}

	prompt += `

You can help users with:
1. Product search and recommendations
2. Adding/removing items from cart
3. Checking cart contents
4. Providing product information
5. Assisting with checkout process

When users ask to add items to cart, respond with a JSON action like:
{"type": "add_to_cart", "payload": {"product_id": "product-id", "quantity": 1}}

When users ask to remove items, respond with:
{"type": "remove_from_cart", "payload": {"product_id": "product-id"}}

When showing products, include relevant details and suggest adding to cart if appropriate.

Be friendly, helpful, and conversational. Always confirm actions taken and provide next steps.`

	return prompt
}

// parseResponse parses the assistant's response for actions and suggestions
func (s *ChatService) parseResponse(message string, products *ProductListResponse) ([]ChatAction, []ProductSuggestion, error) {
	var actions []ChatAction
	var suggestions []ProductSuggestion

	// Look for JSON actions in the message
	lines := strings.Split(message, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "{") && strings.HasSuffix(line, "}") {
			var action ChatAction
			err := json.Unmarshal([]byte(line), &action)
			if err == nil {
				actions = append(actions, action)
			}
		}
	}

	// Generate product suggestions based on message content
	if products != nil {
		messageLower := strings.ToLower(message)
		for _, product := range products.Products {
			productNameLower := strings.ToLower(product.Name)

			// Simple keyword matching for suggestions
			if strings.Contains(messageLower, productNameLower) ||
				strings.Contains(messageLower, strings.ToLower(product.Category.Name)) {
				suggestions = append(suggestions, ProductSuggestion{
					Product:    &product,
					Reason:     "Mentioned in conversation",
					Confidence: 0.8,
				})
			}
		}
	}

	return actions, suggestions, nil
}

// executeAction executes a chat action
func (s *ChatService) executeAction(action ChatAction, userID *uuid.UUID, sessionID string) error {
	switch action.Type {
	case "add_to_cart":
		productIDStr, ok := action.Payload["product_id"].(string)
		if !ok {
			return fmt.Errorf("missing product_id in add_to_cart action")
		}

		productID, err := uuid.Parse(productIDStr)
		if err != nil {
			return fmt.Errorf("invalid product_id: %v", err)
		}

		quantity := 1
		if q, ok := action.Payload["quantity"].(float64); ok {
			quantity = int(q)
		}

		err = s.cartService.AddToCart(sessionID, userID, AddToCartRequest{
			ProductID: productID,
			Quantity:  quantity,
		})
		return err

	case "remove_from_cart":
		productIDStr, ok := action.Payload["product_id"].(string)
		if !ok {
			return fmt.Errorf("missing product_id in remove_from_cart action")
		}

		productID, err := uuid.Parse(productIDStr)
		if err != nil {
			return fmt.Errorf("invalid product_id: %v", err)
		}

		return s.cartService.RemoveFromCart(sessionID, userID, productID, nil)

	default:
		return fmt.Errorf("unknown action type: %s", action.Type)
	}
}

// GetConversationHistory retrieves conversation history for a session
func (s *ChatService) GetConversationHistory(sessionID string, limit int) ([]ChatMessageService, error) {
	var dbMessages []models.ChatMessage

	err := s.db.Where("session_id = ?", sessionID).
		Order("created_at DESC").
		Limit(limit).
		Find(&dbMessages).Error

	if err != nil {
		return nil, err
	}

	// Convert to service layer messages
	var messages []ChatMessageService
	for _, msg := range dbMessages {
		var metadata map[string]interface{}
		if msg.Metadata != nil {
			json.Unmarshal(msg.Metadata, &metadata)
		}

		messages = append(messages, ChatMessageService{
			ID:        msg.ID,
			SessionID: msg.SessionID,
			UserID:    msg.UserID,
			Role:      msg.Role,
			Content:   msg.Content,
			Metadata:  metadata,
			CreatedAt: msg.CreatedAt.Format("2006-01-02T15:04:05Z"),
		})
	}

	// Reverse to get chronological order
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return messages, nil
}

// saveMessage saves a message to the database
func (s *ChatService) saveMessage(sessionID string, userID *uuid.UUID, role, content string, metadata map[string]interface{}) error {
	// First, get the chat session to get its ID
	var chatSession models.ChatSession
	err := s.db.Where("session_id = ?", sessionID).First(&chatSession).Error
	if err != nil {
		return fmt.Errorf("failed to find chat session: %v", err)
	}

	var metadataJSON datatypes.JSON
	if metadata != nil {
		metadataBytes, _ := json.Marshal(metadata)
		metadataJSON = metadataBytes
	}

	message := models.ChatMessage{
		ID:            uuid.New(),
		ChatSessionID: chatSession.ID,
		SessionID:     sessionID,
		UserID:        userID,
		Role:          role,
		Content:       content,
		Metadata:      metadataJSON,
		CreatedAt:     time.Now(),
	}

	return s.db.Create(&message).Error
}

// GetChatSession retrieves or creates a chat session
func (s *ChatService) GetChatSession(sessionID string, userID *uuid.UUID) (*ChatSession, error) {
	var dbSession models.ChatSession

	err := s.db.Where("session_id = ?", sessionID).First(&dbSession).Error
	if err == gorm.ErrRecordNotFound {
		// Create new session
		contextJSON, _ := json.Marshal(make(map[string]interface{}))
		dbSession = models.ChatSession{
			ID:        uuid.New(),
			SessionID: sessionID,
			UserID:    userID,
			Context:   contextJSON,
			CreatedAt: time.Now(),
			ExpiresAt: time.Now().Add(24 * time.Hour),
		}
		err = s.db.Create(&dbSession).Error
		if err != nil {
			return nil, err
		}
	}

	// Convert to service layer session
	var context map[string]interface{}
	if dbSession.Context != nil {
		json.Unmarshal(dbSession.Context, &context)
	}

	session := &ChatSession{
		ID:        dbSession.SessionID,
		UserID:    dbSession.UserID,
		Context:   context,
		CreatedAt: dbSession.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt: dbSession.LastActivity.Format("2006-01-02T15:04:05Z"),
	}

	return session, nil
}

// SearchProducts searches for products based on natural language query
func (s *ChatService) SearchProducts(query string, limit int) ([]ProductSuggestion, error) {
	products, err := s.productService.SearchProducts(query, limit)
	if err != nil {
		return nil, err
	}

	var suggestions []ProductSuggestion
	for _, product := range products {
		suggestions = append(suggestions, ProductSuggestion{
			Product:    &product,
			Reason:     "Search result",
			Confidence: 0.9,
		})
	}

	return suggestions, nil
}

// GetProductRecommendations gets product recommendations based on context
func (s *ChatService) GetProductRecommendations(sessionID string, userID *uuid.UUID, limit int) ([]ProductSuggestion, error) {
	// Get featured products as base recommendations
	products, err := s.productService.GetFeaturedProducts(limit)
	if err != nil {
		return nil, err
	}

	var suggestions []ProductSuggestion
	for _, product := range products {
		suggestions = append(suggestions, ProductSuggestion{
			Product:    &product,
			Reason:     "Featured product",
			Confidence: 0.7,
		})
	}

	return suggestions, nil
}
