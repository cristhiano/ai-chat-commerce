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

	// Get available products for context with full details including category
	productList, err := s.productService.GetProducts(ProductFilters{
		Status: "active",
		Page:   1,
		Limit:  20,
	})
	if err != nil {
		log.Printf("Warning: failed to get products: %v", err)
		productList = nil
	}

	var products *ProductListResponse
	if productList != nil {
		products = productList
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

	// Parse the response for actions and generate suggestions based on USER's message
	var actions []ChatAction
	var suggestions []ProductSuggestion
	if products != nil && products.Products != nil {
		// Parse AI response for actions (add_to_cart, etc.)
		actions, _, err = s.parseResponse(assistantMessage, products)
		if err != nil {
			log.Printf("Warning: failed to parse response: %v", err)
		}

		// Generate suggestions based on the USER's original message (not AI's response)
		suggestions = s.generateRelevantSuggestions(message, products.Products)
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

IMPORTANT: When users ask for product recommendations or search for products:
- DO NOT list product names, prices, or detailed descriptions in your text response
- Instead, give a brief, friendly response like "I found some great options for you!" or "Here are some recommendations based on your request"
- The actual products will be shown as visual cards separately
- Keep your text response short and conversational

When users ask to add items to cart, respond with a JSON action like:
{"type": "add_to_cart", "payload": {"product_id": "product-id", "quantity": 1}}

When users ask to remove items, respond with:
{"type": "remove_from_cart", "payload": {"product_id": "product-id"}}

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
	if products != nil && products.Products != nil {
		suggestions = s.generateRelevantSuggestions(message, products.Products)
	}

	return actions, suggestions, nil
}

// generateRelevantSuggestions generates product suggestions based on message content and intent
func (s *ChatService) generateRelevantSuggestions(message string, products []models.Product) []ProductSuggestion {
	var suggestions []ProductSuggestion
	messageLower := strings.ToLower(message)

	// Ensure products have all necessary fields loaded
	for i := range products {
		// Load category relationship if not already loaded (check by empty ID)
		if products[i].Category.ID == uuid.Nil && products[i].CategoryID != uuid.Nil {
			var category models.Category
			if err := s.db.First(&category, products[i].CategoryID).Error; err == nil {
				products[i].Category = category
			}
		}
	}

	// Define intent keywords to avoid suggesting products when user is clearly not looking for them
	// Keep this list minimal - only truly negative scenarios
	negativeIntents := []string{
		"not interested", "don't want", "no thanks", "not looking",
		"remove", "delete", "cancel",
		"stop", "quit", "exit", "bye", "goodbye",
	}

	// Check if user has strong negative intent (explicitly not looking for products)
	hasNegativeIntent := false
	for _, intent := range negativeIntents {
		if strings.Contains(messageLower, intent) {
			hasNegativeIntent = true
			break
		}
	}

	// If user has negative intent, don't suggest products
	if hasNegativeIntent {
		return suggestions
	}

	// Define positive intent keywords that indicate user wants product suggestions
	positiveIntents := []string{
		"show", "find", "search", "look for", "looking for",
		"want", "need", "interested", "like",
		"buy", "get", "purchase", "checkout", "cart",
		"suggest", "recommend", "recommendation", "advice",
		"what", "which", "any", "do you have", "got",
		"best", "good", "great", "popular", "top",
		"cheap", "expensive", "affordable", "budget",
		"electronics", "clothing", "books", "garden", "home",
		"product", "products", "item", "items",
	}

	// Check if user has positive intent
	hasPositiveIntent := false
	for _, intent := range positiveIntents {
		if strings.Contains(messageLower, intent) {
			hasPositiveIntent = true
			break
		}
	}

	// If no positive intent detected, don't suggest products
	if !hasPositiveIntent {
		return suggestions
	}

	// Generate suggestions based on semantic matching
	for _, product := range products {
		confidence := s.calculateRelevanceScore(messageLower, product)

		// Debug logging to see what's matching
		if confidence > 0.2 {
			log.Printf("[DEBUG] Product: %s, Category: %s, Confidence: %.2f",
				product.Name, product.Category.Name, confidence)
		}

		// Only suggest products with reasonable relevance (balanced threshold)
		if confidence >= 0.4 {
			suggestions = append(suggestions, ProductSuggestion{
				Product:    &product,
				Reason:     s.generateReason(messageLower, product),
				Confidence: confidence,
			})
		}
	}

	// Log final suggestions
	log.Printf("[DEBUG] Message: '%s' - Generated %d suggestions", message, len(suggestions))

	// Sort by confidence (descending)
	for i := 0; i < len(suggestions)-1; i++ {
		for j := i + 1; j < len(suggestions); j++ {
			if suggestions[i].Confidence < suggestions[j].Confidence {
				suggestions[i], suggestions[j] = suggestions[j], suggestions[i]
			}
		}
	}

	// If we have no suggestions but user clearly wants products, show top products
	if len(suggestions) == 0 && hasPositiveIntent {
		log.Printf("[DEBUG] No high-confidence matches, showing top products")
		// Show up to 3 products with highest scores (even if below threshold)
		type scoredProduct struct {
			product    models.Product
			confidence float64
		}
		var allScored []scoredProduct

		for _, product := range products {
			confidence := s.calculateRelevanceScore(messageLower, product)
			if confidence > 0.1 { // Very low bar for fallback
				allScored = append(allScored, scoredProduct{product, confidence})
			}
		}

		// Sort by confidence
		for i := 0; i < len(allScored)-1; i++ {
			for j := i + 1; j < len(allScored); j++ {
				if allScored[i].confidence < allScored[j].confidence {
					allScored[i], allScored[j] = allScored[j], allScored[i]
				}
			}
		}

		// Take top 3 as fallback suggestions
		limit := 3
		if len(allScored) < limit {
			limit = len(allScored)
		}

		for i := 0; i < limit; i++ {
			suggestions = append(suggestions, ProductSuggestion{
				Product:    &allScored[i].product,
				Reason:     "Related to your search",
				Confidence: allScored[i].confidence,
			})
		}
	} else if len(suggestions) > 3 {
		// Limit to top 3 suggestions
		suggestions = suggestions[:3]
	}

	return suggestions
}

// calculateRelevanceScore calculates how relevant a product is to the user's message
func (s *ChatService) calculateRelevanceScore(message string, product models.Product) float64 {
	score := 0.0

	// Stop words to ignore in keyword matching
	stopWords := map[string]bool{
		"a": true, "an": true, "the": true, "and": true, "or": true, "but": true,
		"in": true, "on": true, "at": true, "to": true, "for": true, "of": true,
		"with": true, "by": true, "from": true, "up": true, "about": true,
		"me": true, "you": true, "i": true, "my": true, "your": true, "is": true,
		"are": true, "was": true, "were": true, "be": true, "been": true, "being": true,
		"have": true, "has": true, "had": true, "do": true, "does": true, "did": true,
		"will": true, "would": true, "could": true, "should": true, "can": true, "may": true,
	}

	// Direct product name match (highest score)
	productNameLower := strings.ToLower(product.Name)
	if strings.Contains(message, productNameLower) {
		score += 0.9
	}

	// Category match - check if category is loaded
	if product.Category.ID != uuid.Nil {
		categoryNameLower := strings.ToLower(product.Category.Name)
		if strings.Contains(message, categoryNameLower) {
			score += 0.7
		}
	}

	// Keyword matching in product name and description (with stop word filtering)
	descriptionLower := strings.ToLower(product.Description)
	keywords := strings.Fields(message)
	matchedKeywords := 0
	totalMeaningfulKeywords := 0

	for _, keyword := range keywords {
		keywordLower := strings.ToLower(keyword)
		// Skip stop words and very short words
		if stopWords[keywordLower] || len(keywordLower) < 3 {
			continue
		}

		totalMeaningfulKeywords++

		// Check for partial word matches (e.g., "book" matches "books")
		keywordStem := keywordLower
		if len(keywordLower) > 4 && strings.HasSuffix(keywordLower, "s") {
			keywordStem = keywordLower[:len(keywordLower)-1]
		}

		// Higher score for product name match
		if strings.Contains(productNameLower, keywordLower) || strings.Contains(productNameLower, keywordStem) {
			score += 0.25
			matchedKeywords++
		} else if strings.Contains(descriptionLower, keywordLower) || strings.Contains(descriptionLower, keywordStem) {
			score += 0.08
			matchedKeywords++
		}
	}

	// Bonus for matching multiple keywords (indicates better relevance)
	if matchedKeywords >= 2 {
		score += 0.15
	}

	// Strong bonus if most keywords matched
	if totalMeaningfulKeywords > 0 && float64(matchedKeywords)/float64(totalMeaningfulKeywords) >= 0.5 {
		score += 0.2
	}

	// Specific product type keywords - only add if they match product name or description
	productTypeKeywords := map[string]float64{
		"headphone": 0.6, "headphones": 0.6, "audio": 0.5, "speaker": 0.6,
		"shirt": 0.6, "t-shirt": 0.6, "clothing": 0.5, "apparel": 0.5,
		"book": 0.7, "books": 0.7, "novel": 0.6, "guide": 0.6, "read": 0.5, "reading": 0.5,
		"garden": 0.6, "tools": 0.6, "outdoor": 0.5, "plant": 0.5,
	}

	// Semantic/contextual keywords that map to product categories
	// These help match user intent to product types
	contextualKeywords := map[string]map[string]float64{
		// Entertainment/boredom → books, electronics, headphones
		"bored": {
			"book": 0.4, "books": 0.4, "headphone": 0.4, "headphones": 0.4,
			"wireless": 0.3, "audio": 0.3, "entertainment": 0.3,
		},
		"boredom": {
			"book": 0.4, "books": 0.4, "headphone": 0.4, "headphones": 0.4,
			"wireless": 0.3, "audio": 0.3, "entertainment": 0.3,
		},
		"entertainment": {
			"book": 0.4, "books": 0.4, "headphone": 0.4, "headphones": 0.4,
			"wireless": 0.3, "audio": 0.3,
		},
		"entertain": {
			"book": 0.4, "books": 0.4, "headphone": 0.4, "headphones": 0.4,
			"wireless": 0.3, "audio": 0.3,
		},
		// Commute/travel → books, headphones, portable electronics
		"commute": {
			"book": 0.4, "books": 0.4, "headphone": 0.4, "headphones": 0.4,
			"wireless": 0.3, "portable": 0.3, "travel": 0.3,
		},
		"commuting": {
			"book": 0.4, "books": 0.4, "headphone": 0.4, "headphones": 0.4,
			"wireless": 0.3, "portable": 0.3, "travel": 0.3,
		},
		"travel": {
			"book": 0.4, "books": 0.4, "headphone": 0.4, "headphones": 0.4,
			"wireless": 0.3, "portable": 0.3,
		},
		"traveling": {
			"book": 0.4, "books": 0.4, "headphone": 0.4, "headphones": 0.4,
			"wireless": 0.3, "portable": 0.3,
		},
		// Music/audio → headphones, speakers
		"music": {
			"headphone": 0.5, "headphones": 0.5, "speaker": 0.4, "audio": 0.4, "wireless": 0.3,
		},
		"listen": {
			"headphone": 0.5, "headphones": 0.5, "audio": 0.4, "book": 0.3, "books": 0.3,
		},
		"listening": {
			"headphone": 0.5, "headphones": 0.5, "audio": 0.4, "book": 0.3, "books": 0.3,
		},
		// Reading → books
		"read": {
			"book": 0.6, "books": 0.6, "novel": 0.4, "guide": 0.3,
		},
		"reading": {
			"book": 0.6, "books": 0.6, "novel": 0.4, "guide": 0.3,
		},
		// Fitness/exercise → clothing, outdoor
		"exercise": {
			"shirt": 0.4, "clothing": 0.4, "outdoor": 0.3,
		},
		"workout": {
			"shirt": 0.4, "clothing": 0.4, "outdoor": 0.3,
		},
		"fitness": {
			"shirt": 0.4, "clothing": 0.4, "outdoor": 0.3,
		},
		// Fashion/style → clothing
		"wear": {
			"shirt": 0.4, "clothing": 0.4, "apparel": 0.3,
		},
		"wearing": {
			"shirt": 0.4, "clothing": 0.4, "apparel": 0.3,
		},
		"fashion": {
			"shirt": 0.4, "clothing": 0.4, "apparel": 0.4,
		},
	}

	for keyword, weight := range productTypeKeywords {
		if strings.Contains(message, keyword) {
			// Only add score if the product actually relates to this keyword
			if strings.Contains(productNameLower, keyword) ||
				strings.Contains(descriptionLower, keyword) ||
				strings.Contains(strings.ToLower(product.Category.Name), keyword) {
				score += weight
			}
		}
	}

	// Apply contextual keyword matching
	// This helps with semantic understanding (e.g., "bored while commuting" → books/headphones)
	for contextKeyword, productKeywords := range contextualKeywords {
		if strings.Contains(message, contextKeyword) {
			// Check if this product matches any of the related product keywords
			for productKeyword, weight := range productKeywords {
				if strings.Contains(productNameLower, productKeyword) ||
					strings.Contains(descriptionLower, productKeyword) ||
					strings.Contains(strings.ToLower(product.Category.Name), productKeyword) {
					score += weight
					// Only apply the highest matching weight per context keyword
					break
				}
			}
		}
	}

	// Cap the score at 1.0
	if score > 1.0 {
		score = 1.0
	}

	return score
}

// generateReason generates a human-readable reason for the suggestion
func (s *ChatService) generateReason(message string, product models.Product) string {
	productNameLower := strings.ToLower(product.Name)
	categoryNameLower := strings.ToLower(product.Category.Name)

	if strings.Contains(message, productNameLower) {
		return "Directly mentioned"
	}

	if strings.Contains(message, categoryNameLower) {
		return "Matches your category interest"
	}

	// Check for contextual matches
	if strings.Contains(message, "bored") || strings.Contains(message, "boredom") {
		if strings.Contains(productNameLower, "book") || strings.Contains(categoryNameLower, "book") {
			return "Great for entertainment"
		}
		if strings.Contains(productNameLower, "headphone") || strings.Contains(productNameLower, "audio") {
			return "Great for entertainment"
		}
	}

	if strings.Contains(message, "commut") || strings.Contains(message, "travel") {
		if strings.Contains(productNameLower, "book") || strings.Contains(categoryNameLower, "book") {
			return "Perfect for travel"
		}
		if strings.Contains(productNameLower, "headphone") || strings.Contains(productNameLower, "wireless") {
			return "Perfect for commuting"
		}
	}

	if strings.Contains(message, "music") || strings.Contains(message, "listen") {
		if strings.Contains(productNameLower, "headphone") || strings.Contains(productNameLower, "audio") {
			return "Great for listening"
		}
	}

	if strings.Contains(message, "read") {
		if strings.Contains(productNameLower, "book") || strings.Contains(categoryNameLower, "book") {
			return "Great for reading"
		}
	}

	// Check for specific keywords
	if strings.Contains(message, "cheap") || strings.Contains(message, "budget") {
		return "Good value option"
	}

	if strings.Contains(message, "best") || strings.Contains(message, "recommend") {
		return "Recommended product"
	}

	return "Related to your search"
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
