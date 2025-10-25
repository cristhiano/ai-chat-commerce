package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

// Product represents a product in the catalog
type Product struct {
	ID          uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string         `gorm:"size:255;not null;index" json:"name"`
	Description string         `gorm:"type:text;not null" json:"description"`
	Price       float64        `gorm:"type:decimal(10,2);not null;index" json:"price"`
	CategoryID  uuid.UUID      `gorm:"type:uuid;not null;index" json:"category_id"`
	SKU         string         `gorm:"size:100;uniqueIndex;not null" json:"sku"`
	Status      string         `gorm:"size:20;default:'active';index" json:"status"`
	Metadata    datatypes.JSON `gorm:"type:jsonb" json:"metadata"`
	// Tags        pq.StringArray `gorm:"type:text[]" json:"tags"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relationships
	Category   Category         `gorm:"foreignKey:CategoryID" json:"category"`
	Variants   []ProductVariant `gorm:"foreignKey:ProductID" json:"variants"`
	Images     []ProductImage   `gorm:"foreignKey:ProductID" json:"images"`
	Inventory  []Inventory      `gorm:"foreignKey:ProductID" json:"inventory"`
	OrderItems []OrderItem      `gorm:"foreignKey:ProductID" json:"order_items"`
}

// ProductVariant represents product variations like size, color, material
type ProductVariant struct {
	ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProductID     uuid.UUID `gorm:"type:uuid;not null;index" json:"product_id"`
	VariantName   string    `gorm:"size:50;not null" json:"variant_name"`
	VariantValue  string    `gorm:"size:100;not null" json:"variant_value"`
	PriceModifier float64   `gorm:"type:decimal(10,2);default:0" json:"price_modifier"`
	SKUSuffix     string    `gorm:"size:20" json:"sku_suffix"`
	IsDefault     bool      `gorm:"default:false" json:"is_default"`
	CreatedAt     time.Time `json:"created_at"`

	// Relationships
	Product Product `gorm:"foreignKey:ProductID" json:"product"`
}

// ProductImage represents product images
type ProductImage struct {
	ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProductID uuid.UUID `gorm:"type:uuid;not null;index" json:"product_id"`
	URL       string    `gorm:"size:500;not null" json:"url"`
	AltText   string    `gorm:"size:255" json:"alt_text"`
	IsPrimary bool      `gorm:"default:false" json:"is_primary"`
	SortOrder int       `gorm:"default:0" json:"sort_order"`
	CreatedAt time.Time `json:"created_at"`

	// Relationships
	Product Product `gorm:"foreignKey:ProductID" json:"product"`
}

// Category represents product categories
type Category struct {
	ID          uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Name        string     `gorm:"size:100;not null;index" json:"name"`
	Description string     `gorm:"type:text" json:"description"`
	ParentID    *uuid.UUID `gorm:"type:uuid;index" json:"parent_id"`
	Slug        string     `gorm:"size:100;uniqueIndex;not null" json:"slug"`
	SortOrder   int        `gorm:"default:0" json:"sort_order"`
	IsActive    bool       `gorm:"default:true" json:"is_active"`
	CreatedAt   time.Time  `json:"created_at"`

	// Relationships
	Parent   *Category  `gorm:"foreignKey:ParentID" json:"parent"`
	Children []Category `gorm:"foreignKey:ParentID" json:"children"`
	Products []Product  `gorm:"foreignKey:CategoryID" json:"products"`
}

// Inventory represents stock levels and warehouse information
type Inventory struct {
	ID                uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProductID         uuid.UUID  `gorm:"type:uuid;not null;index" json:"product_id"`
	VariantID         *uuid.UUID `gorm:"type:uuid;index" json:"variant_id"`
	WarehouseLocation string     `gorm:"size:50;not null" json:"warehouse_location"`
	QuantityAvailable int        `gorm:"not null;default:0" json:"quantity_available"`
	QuantityReserved  int        `gorm:"not null;default:0" json:"quantity_reserved"`
	LowStockThreshold int        `gorm:"default:10" json:"low_stock_threshold"`
	ReorderPoint      int        `gorm:"default:5" json:"reorder_point"`
	LastRestocked     *time.Time `json:"last_restocked"`
	CreatedAt         time.Time  `json:"created_at"`
	UpdatedAt         time.Time  `json:"updated_at"`

	// Relationships
	Product      Product                `gorm:"foreignKey:ProductID" json:"product"`
	Variant      *ProductVariant        `gorm:"foreignKey:VariantID" json:"variant"`
	Reservations []InventoryReservation `gorm:"foreignKey:InventoryID" json:"reservations"`
}

// InventoryAlert represents inventory alerts
type InventoryAlert struct {
	ID              uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	ProductID       uuid.UUID  `gorm:"type:uuid;not null;index" json:"product_id"`
	VariantID       *uuid.UUID `gorm:"type:uuid;index" json:"variant_id"`
	CurrentQuantity int        `gorm:"not null" json:"current_quantity"`
	Threshold       int        `gorm:"not null" json:"threshold"`
	Location        string     `gorm:"size:50" json:"location"`
	AlertType       string     `gorm:"size:20;not null" json:"alert_type"` // "low_stock", "out_of_stock", "overstock"
	IsRead          bool       `gorm:"default:false" json:"is_read"`
	CreatedAt       time.Time  `json:"created_at"`

	// Relationships
	Product Product         `gorm:"foreignKey:ProductID" json:"product"`
	Variant *ProductVariant `gorm:"foreignKey:VariantID" json:"variant"`
}

// InventoryReservation represents temporary inventory reservations
type InventoryReservation struct {
	ID               uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	InventoryID      uuid.UUID  `gorm:"type:uuid;not null;index" json:"inventory_id"`
	SessionID        string     `gorm:"size:100;not null;index" json:"session_id"`
	UserID           *uuid.UUID `gorm:"type:uuid;index" json:"user_id"`
	QuantityReserved int        `gorm:"not null" json:"quantity_reserved"`
	ExpiresAt        time.Time  `gorm:"not null;index" json:"expires_at"`
	Status           string     `gorm:"size:20;default:'active';index" json:"status"`
	CreatedAt        time.Time  `json:"created_at"`

	// Relationships
	Inventory Inventory `gorm:"foreignKey:InventoryID" json:"inventory"`
	User      *User     `gorm:"foreignKey:UserID" json:"user"`
}

// User represents customer accounts
type User struct {
	ID            uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	Email         string         `gorm:"size:255;uniqueIndex;not null" json:"email"`
	PasswordHash  string         `gorm:"size:255;not null" json:"-"`
	FirstName     string         `gorm:"size:50;not null" json:"first_name"`
	LastName      string         `gorm:"size:50;not null" json:"last_name"`
	Phone         string         `gorm:"size:20" json:"phone"`
	DateOfBirth   *time.Time     `json:"date_of_birth"`
	Preferences   datatypes.JSON `gorm:"type:jsonb" json:"preferences"`
	EmailVerified bool           `gorm:"default:false" json:"email_verified"`
	Status        string         `gorm:"size:20;default:'active';index" json:"status"`
	CreatedAt     time.Time      `json:"created_at"`
	UpdatedAt     time.Time      `json:"updated_at"`

	// Relationships
	ChatSessions  []ChatSession          `gorm:"foreignKey:UserID" json:"chat_sessions"`
	ShoppingCarts []ShoppingCart         `gorm:"foreignKey:UserID" json:"shopping_carts"`
	Orders        []Order                `gorm:"foreignKey:UserID" json:"orders"`
	Reservations  []InventoryReservation `gorm:"foreignKey:UserID" json:"reservations"`
}

// ChatSession represents chat conversation sessions
type ChatSession struct {
	ID                  uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SessionID           string         `gorm:"size:100;uniqueIndex;not null" json:"session_id"`
	UserID              *uuid.UUID     `gorm:"type:uuid;index" json:"user_id"`
	ConversationHistory datatypes.JSON `gorm:"type:jsonb" json:"conversation_history"`
	Context             datatypes.JSON `gorm:"type:jsonb" json:"context"`
	CartState           datatypes.JSON `gorm:"type:jsonb" json:"cart_state"`
	Preferences         datatypes.JSON `gorm:"type:jsonb" json:"preferences"`
	Status              string         `gorm:"size:20;default:'active';index" json:"status"`
	LastActivity        time.Time      `gorm:"index" json:"last_activity"`
	CreatedAt           time.Time      `json:"created_at"`
	ExpiresAt           time.Time      `gorm:"index" json:"expires_at"`

	// Relationships
	User     User          `gorm:"foreignKey:UserID" json:"user"`
	Messages []ChatMessage `gorm:"foreignKey:SessionID;references:SessionID" json:"messages"`
}

// ChatMessage represents a message in a chat conversation
type ChatMessage struct {
	ID        uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SessionID string         `gorm:"size:100;not null;index" json:"session_id"`
	UserID    *uuid.UUID     `gorm:"type:uuid;index" json:"user_id"`
	Role      string         `gorm:"size:20;not null" json:"role"` // "user", "assistant", "system"
	Content   string         `gorm:"type:text;not null" json:"content"`
	Metadata  datatypes.JSON `gorm:"type:jsonb" json:"metadata"`
	CreatedAt time.Time      `json:"created_at"`

	// Relationships
	User    User        `gorm:"foreignKey:UserID" json:"user"`
	Session ChatSession `gorm:"foreignKey:SessionID;references:SessionID" json:"session"`
}

// ShoppingCart represents unified cart state
type ShoppingCart struct {
	ID             uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	SessionID      string         `gorm:"size:100;not null;index" json:"session_id"`
	UserID         *uuid.UUID     `gorm:"type:uuid;index" json:"user_id"`
	Items          datatypes.JSON `gorm:"type:jsonb" json:"items"`
	Subtotal       float64        `gorm:"type:decimal(10,2);default:0" json:"subtotal"`
	TaxAmount      float64        `gorm:"type:decimal(10,2);default:0" json:"tax_amount"`
	ShippingAmount float64        `gorm:"type:decimal(10,2);default:0" json:"shipping_amount"`
	TotalAmount    float64        `gorm:"type:decimal(10,2);default:0" json:"total_amount"`
	Currency       string         `gorm:"size:3;default:'USD'" json:"currency"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`

	// Relationships
	User User `gorm:"foreignKey:UserID" json:"user"`
}

// Order represents completed purchase transactions
type Order struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OrderNumber     string         `gorm:"size:50;uniqueIndex;not null" json:"order_number"`
	UserID          uuid.UUID      `gorm:"type:uuid;not null;index" json:"user_id"`
	SessionID       string         `gorm:"size:100;not null" json:"session_id"`
	Status          string         `gorm:"size:20;default:'pending';index" json:"status"`
	Subtotal        float64        `gorm:"type:decimal(10,2);not null" json:"subtotal"`
	TaxAmount       float64        `gorm:"type:decimal(10,2);not null" json:"tax_amount"`
	ShippingAmount  float64        `gorm:"type:decimal(10,2);not null" json:"shipping_amount"`
	TotalAmount     float64        `gorm:"type:decimal(10,2);not null" json:"total_amount"`
	Currency        string         `gorm:"size:3;not null" json:"currency"`
	PaymentStatus   string         `gorm:"size:20;default:'pending';index" json:"payment_status"`
	ShippingAddress datatypes.JSON `gorm:"type:jsonb;not null" json:"shipping_address"`
	BillingAddress  datatypes.JSON `gorm:"type:jsonb;not null" json:"billing_address"`
	PaymentIntentID string         `gorm:"size:100" json:"payment_intent_id"`
	CreatedAt       time.Time      `json:"created_at"`
	UpdatedAt       time.Time      `json:"updated_at"`

	// Relationships
	User  User        `gorm:"foreignKey:UserID" json:"user"`
	Items []OrderItem `gorm:"foreignKey:OrderID" json:"items"`
}

// OrderItem represents individual items within an order
type OrderItem struct {
	ID              uuid.UUID      `gorm:"type:uuid;primary_key;default:gen_random_uuid()" json:"id"`
	OrderID         uuid.UUID      `gorm:"type:uuid;not null;index" json:"order_id"`
	ProductID       uuid.UUID      `gorm:"type:uuid;not null;index" json:"product_id"`
	VariantID       *uuid.UUID     `gorm:"type:uuid;index" json:"variant_id"`
	Quantity        int            `gorm:"not null" json:"quantity"`
	UnitPrice       float64        `gorm:"type:decimal(10,2);not null" json:"unit_price"`
	TotalPrice      float64        `gorm:"type:decimal(10,2);not null" json:"total_price"`
	ProductSnapshot datatypes.JSON `gorm:"type:jsonb" json:"product_snapshot"`
	CreatedAt       time.Time      `json:"created_at"`

	// Relationships
	Order   Order           `gorm:"foreignKey:OrderID" json:"order"`
	Product Product         `gorm:"foreignKey:ProductID" json:"product"`
	Variant *ProductVariant `gorm:"foreignKey:VariantID" json:"variant"`
}

// TableName methods for custom table names
func (Product) TableName() string {
	return "products"
}

func (ProductVariant) TableName() string {
	return "product_variants"
}

func (ProductImage) TableName() string {
	return "product_images"
}

func (Category) TableName() string {
	return "categories"
}

func (Inventory) TableName() string {
	return "inventory"
}

func (InventoryAlert) TableName() string {
	return "inventory_alerts"
}

func (InventoryReservation) TableName() string {
	return "inventory_reservations"
}

func (User) TableName() string {
	return "users"
}

func (ChatSession) TableName() string {
	return "chat_sessions"
}

func (ChatMessage) TableName() string {
	return "chat_messages"
}

func (ShoppingCart) TableName() string {
	return "shopping_carts"
}

func (Order) TableName() string {
	return "orders"
}

func (OrderItem) TableName() string {
	return "order_items"
}
