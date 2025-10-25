# Data Model: Chat-Based Ecommerce Application

**Date:** 2024-12-19  
**Project:** Chat-Based Ecommerce Application  
**Technology Stack:** Golang, PostgreSQL, GORM  
**Version:** 1.0

## Database Schema Overview

The application uses PostgreSQL with GORM ORM for type-safe database operations. All entities follow Go naming conventions and include proper indexing for performance.

## Entity Definitions

### Products
**Purpose:** Core product catalog with comprehensive attributes for chat recommendations and traditional browsing

```go
type Product struct {
    ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    Name        string    `gorm:"size:255;not null;index"`
    Description string    `gorm:"type:text;not null"`
    Price       float64   `gorm:"type:decimal(10,2);not null;index"`
    CategoryID  uuid.UUID `gorm:"type:uuid;not null;index"`
    SKU         string    `gorm:"size:100;uniqueIndex;not null"`
    Status      string    `gorm:"size:20;default:'active';index"`
    Metadata    datatypes.JSON `gorm:"type:jsonb"`
    Tags        pq.StringArray `gorm:"type:text[]"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
    
    // Relationships
    Category    Category         `gorm:"foreignKey:CategoryID"`
    Variants    []ProductVariant `gorm:"foreignKey:ProductID"`
    Images      []ProductImage   `gorm:"foreignKey:ProductID"`
    Inventory   []Inventory      `gorm:"foreignKey:ProductID"`
    OrderItems  []OrderItem      `gorm:"foreignKey:ProductID"`
}
```

**Validation Rules:**
- Price must be positive
- SKU must be unique across all products
- Name must be between 1-255 characters
- Status must be one of: active, inactive, discontinued

### ProductVariants
**Purpose:** Handle product variations like size, color, material

```go
type ProductVariant struct {
    ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    ProductID     uuid.UUID `gorm:"type:uuid;not null;index"`
    VariantName   string    `gorm:"size:50;not null"`
    VariantValue  string    `gorm:"size:100;not null"`
    PriceModifier float64   `gorm:"type:decimal(10,2);default:0"`
    SKUSuffix     string    `gorm:"size:20"`
    IsDefault     bool      `gorm:"default:false"`
    CreatedAt     time.Time
    
    // Relationships
    Product Product `gorm:"foreignKey:ProductID"`
}
```

**Validation Rules:**
- Product must exist
- Variant name must be non-empty
- Only one default variant per product
- SKU suffix must be unique within product

### Categories
**Purpose:** Product categorization for browsing and recommendations

```go
type Category struct {
    ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    Name        string    `gorm:"size:100;not null;index"`
    Description string    `gorm:"type:text"`
    ParentID    *uuid.UUID `gorm:"type:uuid;index"`
    Slug        string    `gorm:"size:100;uniqueIndex;not null"`
    SortOrder   int       `gorm:"default:0"`
    IsActive    bool      `gorm:"default:true"`
    CreatedAt   time.Time
    
    // Relationships
    Parent   *Category  `gorm:"foreignKey:ParentID"`
    Children []Category `gorm:"foreignKey:ParentID"`
    Products []Product  `gorm:"foreignKey:CategoryID"`
}
```

**Validation Rules:**
- Name must be unique within parent category
- Slug must be unique globally
- Cannot be parent of itself

### Inventory
**Purpose:** Track stock levels and manage reservations

```go
type Inventory struct {
    ID                  uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    ProductID           uuid.UUID `gorm:"type:uuid;not null;index"`
    VariantID           *uuid.UUID `gorm:"type:uuid;index"`
    WarehouseLocation   string    `gorm:"size:50;not null"`
    QuantityAvailable   int       `gorm:"not null;default:0"`
    QuantityReserved    int       `gorm:"not null;default:0"`
    LowStockThreshold   int       `gorm:"default:10"`
    ReorderPoint        int       `gorm:"default:5"`
    LastRestocked       *time.Time
    CreatedAt           time.Time
    UpdatedAt           time.Time
    
    // Relationships
    Product     Product                `gorm:"foreignKey:ProductID"`
    Variant     *ProductVariant        `gorm:"foreignKey:VariantID"`
    Reservations []InventoryReservation `gorm:"foreignKey:InventoryID"`
}
```

**Validation Rules:**
- Quantity available cannot be negative
- Quantity reserved cannot exceed quantity available
- Low stock threshold must be positive
- Reorder point must be less than low stock threshold

### InventoryReservations
**Purpose:** Handle checkout conflicts and timeout-based reservations

```go
type InventoryReservation struct {
    ID              uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    InventoryID     uuid.UUID `gorm:"type:uuid;not null;index"`
    SessionID       string    `gorm:"size:100;not null;index"`
    UserID          *uuid.UUID `gorm:"type:uuid;index"`
    QuantityReserved int       `gorm:"not null"`
    ExpiresAt       time.Time `gorm:"not null;index"`
    Status          string    `gorm:"size:20;default:'active';index"`
    CreatedAt       time.Time
    
    // Relationships
    Inventory Inventory `gorm:"foreignKey:InventoryID"`
    User      *User     `gorm:"foreignKey:UserID"`
}
```

**Validation Rules:**
- Quantity reserved must be positive
- Expires at must be in the future
- Cannot exceed available inventory
- Session ID required for anonymous users

### Users
**Purpose:** Customer accounts and authentication

```go
type User struct {
    ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    Email         string    `gorm:"size:255;uniqueIndex;not null"`
    PasswordHash  string    `gorm:"size:255;not null"`
    FirstName     string    `gorm:"size:50;not null"`
    LastName      string    `gorm:"size:50;not null"`
    Phone         string    `gorm:"size:20"`
    DateOfBirth   *time.Time
    Preferences   datatypes.JSON `gorm:"type:jsonb"`
    EmailVerified bool      `gorm:"default:false"`
    Status        string    `gorm:"size:20;default:'active';index"`
    CreatedAt     time.Time
    UpdatedAt     time.Time
    
    // Relationships
    ChatSessions []ChatSession `gorm:"foreignKey:UserID"`
    ShoppingCarts []ShoppingCart `gorm:"foreignKey:UserID"`
    Orders       []Order      `gorm:"foreignKey:UserID"`
    Reservations []InventoryReservation `gorm:"foreignKey:UserID"`
}
```

**Validation Rules:**
- Email must be valid format and unique
- Password must meet security requirements
- First and last name required
- Status must be one of: active, inactive, suspended

### ChatSessions
**Purpose:** Maintain conversation context and shopping state

```go
type ChatSession struct {
    ID                uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    SessionID         string    `gorm:"size:100;uniqueIndex;not null"`
    UserID            *uuid.UUID `gorm:"type:uuid;index"`
    ConversationHistory datatypes.JSON `gorm:"type:jsonb"`
    Context           datatypes.JSON `gorm:"type:jsonb"`
    CartState         datatypes.JSON `gorm:"type:jsonb"`
    Preferences       datatypes.JSON `gorm:"type:jsonb"`
    Status            string    `gorm:"size:20;default:'active';index"`
    LastActivity      time.Time `gorm:"index"`
    CreatedAt         time.Time
    ExpiresAt         time.Time `gorm:"index"`
    
    // Relationships
    User User `gorm:"foreignKey:UserID"`
}
```

**Validation Rules:**
- Session ID required
- Last activity must be updated on each interaction
- Expires at must be 24 hours from creation
- Context must be valid JSON structure

### ShoppingCarts
**Purpose:** Unified cart state across chat and traditional interfaces

```go
type ShoppingCart struct {
    ID            uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    SessionID     string    `gorm:"size:100;not null;index"`
    UserID        *uuid.UUID `gorm:"type:uuid;index"`
    Items         datatypes.JSON `gorm:"type:jsonb"`
    Subtotal      float64   `gorm:"type:decimal(10,2);default:0"`
    TaxAmount     float64   `gorm:"type:decimal(10,2);default:0"`
    ShippingAmount float64   `gorm:"type:decimal(10,2);default:0"`
    TotalAmount   float64   `gorm:"type:decimal(10,2);default:0"`
    Currency      string    `gorm:"size:3;default:'USD'"`
    CreatedAt     time.Time
    UpdatedAt     time.Time
    
    // Relationships
    User User `gorm:"foreignKey:UserID"`
}
```

**Validation Rules:**
- Session ID required
- Items must be valid JSON array
- Calculated amounts must be non-negative
- Currency must be valid ISO code

### Orders
**Purpose:** Completed purchase transactions

```go
type Order struct {
    ID              uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    OrderNumber     string    `gorm:"size:50;uniqueIndex;not null"`
    UserID          uuid.UUID `gorm:"type:uuid;not null;index"`
    SessionID       string    `gorm:"size:100;not null"`
    Status          string    `gorm:"size:20;default:'pending';index"`
    Subtotal        float64   `gorm:"type:decimal(10,2);not null"`
    TaxAmount       float64   `gorm:"type:decimal(10,2);not null"`
    ShippingAmount  float64   `gorm:"type:decimal(10,2);not null"`
    TotalAmount     float64   `gorm:"type:decimal(10,2);not null"`
    Currency        string    `gorm:"size:3;not null"`
    PaymentStatus   string    `gorm:"size:20;default:'pending';index"`
    ShippingAddress datatypes.JSON `gorm:"type:jsonb;not null"`
    BillingAddress  datatypes.JSON `gorm:"type:jsonb;not null"`
    PaymentIntentID string    `gorm:"size:100"`
    CreatedAt       time.Time
    UpdatedAt       time.Time
    
    // Relationships
    User  User        `gorm:"foreignKey:UserID"`
    Items []OrderItem `gorm:"foreignKey:OrderID"`
}
```

**Validation Rules:**
- Order number must be unique
- User ID required for order processing
- All amounts must be non-negative
- Status transitions must follow defined workflow
- Addresses must contain required fields

### OrderItems
**Purpose:** Individual items within an order

```go
type OrderItem struct {
    ID              uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    OrderID         uuid.UUID `gorm:"type:uuid;not null;index"`
    ProductID       uuid.UUID `gorm:"type:uuid;not null;index"`
    VariantID       *uuid.UUID `gorm:"type:uuid;index"`
    Quantity        int       `gorm:"not null"`
    UnitPrice       float64   `gorm:"type:decimal(10,2);not null"`
    TotalPrice      float64   `gorm:"type:decimal(10,2);not null"`
    ProductSnapshot datatypes.JSON `gorm:"type:jsonb"`
    CreatedAt       time.Time
    
    // Relationships
    Order   Order         `gorm:"foreignKey:OrderID"`
    Product Product       `gorm:"foreignKey:ProductID"`
    Variant *ProductVariant `gorm:"foreignKey:VariantID"`
}
```

**Validation Rules:**
- Quantity must be positive
- Unit price must be non-negative
- Total price must equal quantity × unit price
- Product snapshot must be valid JSON

## Database Indexes

### Primary Indexes
```sql
-- Products
CREATE INDEX idx_products_name ON products(name);
CREATE INDEX idx_products_category_id ON products(category_id);
CREATE INDEX idx_products_status ON products(status);
CREATE INDEX idx_products_sku ON products(sku);

-- Inventory
CREATE INDEX idx_inventory_product_id ON inventory(product_id);
CREATE INDEX idx_inventory_warehouse_location ON inventory(warehouse_location);

-- Chat Sessions
CREATE INDEX idx_chat_sessions_session_id ON chat_sessions(session_id);
CREATE INDEX idx_chat_sessions_last_activity ON chat_sessions(last_activity);
CREATE INDEX idx_chat_sessions_expires_at ON chat_sessions(expires_at);

-- Orders
CREATE INDEX idx_orders_order_number ON orders(order_number);
CREATE INDEX idx_orders_user_id ON orders(user_id);
CREATE INDEX idx_orders_status ON orders(status);
CREATE INDEX idx_orders_created_at ON orders(created_at);
```

### Composite Indexes
```sql
-- Inventory variants
CREATE INDEX idx_inventory_product_variant ON inventory(product_id, variant_id);

-- Chat sessions with activity
CREATE INDEX idx_chat_sessions_session_activity ON chat_sessions(session_id, last_activity);

-- User orders
CREATE INDEX idx_orders_user_created ON orders(user_id, created_at);
```

### JSON Indexes
```sql
-- Products metadata
CREATE INDEX idx_products_metadata ON products USING GIN (metadata);

-- Chat conversation history
CREATE INDEX idx_chat_sessions_history ON chat_sessions USING GIN (conversation_history);

-- User preferences
CREATE INDEX idx_users_preferences ON users USING GIN (preferences);
```

## State Transitions

### Order Status Workflow
```go
const (
    OrderStatusPending    = "pending"
    OrderStatusConfirmed = "confirmed"
    OrderStatusProcessing = "processing"
    OrderStatusShipped   = "shipped"
    OrderStatusDelivered = "delivered"
    OrderStatusCancelled = "cancelled"
)

// Valid transitions
var OrderStatusTransitions = map[string][]string{
    OrderStatusPending:    {OrderStatusConfirmed, OrderStatusCancelled},
    OrderStatusConfirmed: {OrderStatusProcessing, OrderStatusCancelled},
    OrderStatusProcessing: {OrderStatusShipped, OrderStatusCancelled},
    OrderStatusShipped:   {OrderStatusDelivered, OrderStatusCancelled},
    OrderStatusDelivered: {},
    OrderStatusCancelled: {},
}
```

### Inventory Reservation Workflow
```go
const (
    ReservationStatusActive    = "active"
    ReservationStatusCompleted = "completed"
    ReservationStatusExpired   = "expired"
    ReservationStatusCancelled = "cancelled"
)
```

### Chat Session Workflow
```go
const (
    ChatSessionStatusActive    = "active"
    ChatSessionStatusCompleted = "completed"
    ChatSessionStatusAbandoned = "abandoned"
)
```

## GORM Configuration

### Database Connection
```go
func ConnectDatabase() (*gorm.DB, error) {
    dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable",
        os.Getenv("DB_HOST"),
        os.Getenv("DB_USER"),
        os.Getenv("DB_PASSWORD"),
        os.Getenv("DB_NAME"),
        os.Getenv("DB_PORT"),
    )
    
    db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
        NamingStrategy: schema.NamingStrategy{
            TablePrefix:   "",
            SingularTable: false,
        },
        Logger: logger.Default.LogMode(logger.Info),
    })
    
    if err != nil {
        return nil, err
    }
    
    // Auto-migrate all models
    err = db.AutoMigrate(
        &Product{},
        &ProductVariant{},
        &Category{},
        &Inventory{},
        &InventoryReservation{},
        &User{},
        &ChatSession{},
        &ShoppingCart{},
        &Order{},
        &OrderItem{},
    )
    
    return db, err
}
```

### Custom Validators
```go
func (p *Product) Validate() error {
    if p.Price < 0 {
        return errors.New("price must be positive")
    }
    if len(p.Name) == 0 || len(p.Name) > 255 {
        return errors.New("name must be between 1-255 characters")
    }
    if p.Status != "active" && p.Status != "inactive" && p.Status != "discontinued" {
        return errors.New("invalid status")
    }
    return nil
}
```

This data model provides a solid foundation for the chat-based ecommerce application with proper relationships, validation, and performance optimization through strategic indexing.
