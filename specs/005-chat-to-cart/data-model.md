# Data Model: Product Cards with Add to Cart in Chat

## Overview
This document defines the data structures and entity relationships for displaying product cards with cart functionality in the chat interface.

## Core Entities

### Product

**Definition:** Product entity represents an item available for purchase that can be displayed in chat suggestions.

**Fields:**
```typescript
interface Product {
  id: string;                    // Unique product identifier
  name: string;                  // Product name
  description: string;           // Product description
  price: number;                 // Price in USD (no currency symbol)
  category?: {
    id: string;
    name: string;
  };                             // Product category information
  tags?: string[];               // Product tags (e.g., ["New", "Popular"])
  variants?: ProductVariant[];  // Product variants (size, color, etc.)
  inventory_count?: number;      // Available inventory
  image_url?: string;            // Product image URL (future)
}
```

**Validation Rules:**
- `id`: Required, non-empty string, UUID format
- `name`: Required, non-empty string, max 100 characters
- `description`: Optional, string, max 500 characters
- `price`: Required, positive number, decimal precision 2
- `category`: Optional, must include id and name if present
- `tags`: Optional array of strings, max 10 tags
- `variants`: Optional array of ProductVariant objects
- `inventory_count`: Optional, non-negative integer

**Relationships:**
- One-to-Many: Product → CartItems
- Many-to-One: Product → Category

---

### Product Variant

**Definition:** Represents different options for a product (size, color, material, etc.).

**Fields:**
```typescript
interface ProductVariant {
  id: string;                   // Unique variant identifier
  product_id: string;           // Parent product ID
  name: string;                 // Variant name (e.g., "Blue", "Large")
  value: string;                 // Variant value
  price_modifier?: number;       // Price adjustment
  inventory_count: number;       // Variant-specific inventory
  attributes: {
    type: string;                // e.g., "color", "size"
    [key: string]: any;
  };
}
```

**Validation Rules:**
- `id`: Required, non-empty string, UUID format
- `product_id`: Required, must reference existing product
- `name`: Required, non-empty string, max 50 characters
- `value`: Required, non-empty string
- `price_modifier`: Optional, number, can be negative
- `inventory_count`: Required, non-negative integer

**Relationships:**
- Many-to-One: ProductVariant → Product
- One-to-One: ProductVariant → CartItem

---

### Product Suggestion

**Definition:** Product recommendation shown in chat with context about why it's suggested.

**Fields:**
```typescript
interface ProductSuggestion {
  product: Product;             // Product entity
  reason?: string;              // Why this product was suggested
  confidence?: number;          // Confidence score (0-1)
  metadata?: {
    search_query?: string;      // Query that triggered this suggestion
    filters_applied?: string[]; // Filters used
    sort_method?: string;       // Sort method used
  };
}
```

**Validation Rules:**
- `product`: Required, valid Product object
- `reason`: Optional, string, max 200 characters
- `confidence`: Optional, number between 0 and 1
- `metadata`: Optional, object with string values

**Relationships:**
- One-to-One: ProductSuggestion → Product

---

### Cart

**Definition:** Shopping cart containing items user wants to purchase.

**Fields:**
```typescript
interface Cart {
  id: string;                   // Cart identifier (session-based)
  user_id?: string;             // User ID if authenticated
  items: CartItem[];            // Cart items
  subtotal: number;             // Subtotal before tax
  tax: number;                  // Tax amount
  total: number;                // Total amount
  created_at: string;           // ISO 8601 timestamp
  updated_at: string;           // ISO 8601 timestamp
  expires_at?: string;          // Cart expiration (for guest carts)
}
```

**Validation Rules:**
- `id`: Required, non-empty string
- `user_id`: Optional, string UUID
- `items`: Required, array of CartItem objects
- `subtotal`: Required, non-negative number
- `tax`: Required, non-negative number
- `total`: Required, non-negative number, equals subtotal + tax
- `created_at`: Required, valid ISO 8601 timestamp
- `updated_at`: Required, valid ISO 8601 timestamp

**State Transitions:**
- Empty → Active (first item added)
- Active → Updated (item added/removed)
- Active → Checkout (user proceeds to checkout)
- Active → Expired (guest cart timeout)

**Relationships:**
- One-to-Many: Cart → CartItem
- One-to-One: Cart → User (optional)

---

### Cart Item

**Definition:** Single item in shopping cart with quantity and variant information.

**Fields:**
```typescript
interface CartItem {
  id: string;                   // Unique cart item identifier
  product_id: string;           // Product identifier
  variant_id?: string;          // Variant identifier (if applicable)
  quantity: number;             // Quantity in cart
  unit_price: number;           // Price per unit
  subtotal: number;             // quantity * unit_price
  product_name: string;         // Cached product name
  product_image?: string;       // Cached product image
  created_at: string;           // ISO 8601 timestamp
  updated_at: string;           // ISO 8601 timestamp
}
```

**Validation Rules:**
- `id`: Required, non-empty string, UUID format
- `product_id`: Required, must reference existing product
- `variant_id`: Optional, must reference existing variant if provided
- `quantity`: Required, positive integer, max 99
- `unit_price`: Required, non-negative number
- `subtotal`: Required, positive number, equals quantity * unit_price

**Business Rules:**
- Quantity must not exceed inventory_count for product/variant
- If product has variants, variant_id must be provided
- Subtotal calculated automatically: quantity * unit_price
- Can remove item by setting quantity to 0 (client-side action)

**Relationships:**
- Many-to-One: CartItem → Cart
- One-to-One: CartItem → Product
- One-to-One: CartItem → ProductVariant (optional)

---

## API Request/Response Models

### Add to Cart Request

```typescript
interface AddToCartRequest {
  product_id: string;          // Product to add
  variant_id?: string;         // Specific variant (if applicable)
  quantity: number;            // Quantity to add (default: 1)
  session_id?: string;         // Session ID for guest users
}
```

**Validation Rules:**
- `product_id`: Required, must exist in database
- `variant_id`: Optional, must exist and belong to product_id if provided
- `quantity`: Required, positive integer, max 99
- `session_id`: Optional, used for guest cart

**Response:**
```typescript
interface CartResponse {
  success: boolean;
  cart: Cart;
  message?: string;            // Success/error message
}
```

---

### Update Cart Item Request

```typescript
interface UpdateCartItemRequest {
  item_id: string;             // Cart item ID to update
  quantity: number;            // New quantity (0 removes item)
}
```

**Validation Rules:**
- `item_id`: Required, must exist in user's cart
- `quantity`: Required, non-negative integer (0 = remove)
- If quantity > 0, must not exceed inventory

**Response:**
```typescript
interface CartResponse {
  success: boolean;
  cart: Cart;
  message?: string;
}
```

---

### Get Cart Request

```typescript
interface GetCartRequest {
  session_id?: string;        // Session ID for guest users
  user_id?: string;           // User ID if authenticated
}
```

**Validation Rules:**
- One of session_id or user_id must be provided
- If both provided, user_id takes precedence

**Response:**
```typescript
interface CartResponse {
  success: boolean;
  cart: Cart | null;          // null if cart is empty
  message?: string;
}
```

---

## UI State Models

### Product Card State

```typescript
interface ProductCardState {
  productId: string;
  isInCart: boolean;
  cartQuantity: number;
  isLoading: boolean;
  error?: string;
  isQuantityEditorOpen: boolean;
}
```

**State Transitions:**
- `idle` → `loading` (user clicks Add to Cart)
- `loading` → `success` (API succeeds)
- `loading` → `error` (API fails)
- `success` → `idle` (after timeout)
- `idle` → `quantity-editor` (clicks on in-cart item)
- `quantity-editor` → `idle` (closes editor)

---

### Cart Action Button State

```typescript
type ButtonState = 
  | 'idle'           // Ready to add
  | 'loading'        // Adding to cart
  | 'success'        // Successfully added
  | 'error'          // Failed to add
  | 'in-cart';       // Already in cart

interface CartActionButtonState {
  state: ButtonState;
  currentQuantity?: number;
  error?: string;
}
```

**State Display:**
- `idle`: "Add to Cart" button, blue
- `loading`: "Adding..." with spinner
- `success`: "Added!" with checkmark, green, auto-transitions to in-cart
- `error`: "Error" with retry, red
- `in-cart`: "In cart (Qty)" with cart icon, gray

---

## Component Props Interfaces

### ProductSuggestionCard Props

```typescript
interface ProductSuggestionCardProps {
  suggestion: ProductSuggestion;
  onClick?: () => void;              // Optional card click handler
  compact?: boolean;                  // Compact layout mode
  showAddToCart?: boolean;           // Show add to cart button
  onAddToCart?: (productId: string, variantId?: string, quantity?: number) => void;
  onQuantityChange?: (productId: string, quantity: number) => void;
  className?: string;
}
```

### CartActionButton Props

```typescript
interface CartActionButtonProps {
  productId: string;
  variantId?: string;
  currentCart: Cart | null;
  onAddToCart: (item: AddToCartRequest) => Promise<boolean>;
  onUpdateQuantity?: (item: { product_id: string; variant_id?: string; quantity: number }) => Promise<boolean>;
  className?: string;
  showQuantityEditor?: boolean;
}
```

### QuantityEditor Props

```typescript
interface QuantityEditorProps {
  productId: string;
  variantId?: string;
  currentQuantity: number;
  maxQuantity?: number;
  onUpdate: (quantity: number) => void;
  onRemove: () => void;
  className?: string;
}
```

---

## Data Flow

### Add to Cart Flow

1. User clicks "Add to Cart" button
2. `CartActionButton` creates `AddToCartRequest`
3. Button state changes to `loading`
4. Request sent to `CartContext.addToCart()`
5. CartContext calls backend API `POST /api/v1/cart/add`
6. Backend validates and adds item to cart
7. Backend returns updated `CartResponse`
8. CartContext updates global cart state
9. Button state changes to `success`
10. After 1s, button state changes to `in-cart`
11. Cart badge updates
12. Notification shown (optional)

### Update Quantity Flow

1. User clicks on item already in cart
2. Quantity editor opens (inline)
3. User modifies quantity
4. User clicks "Apply"
5. Request sent to `CartContext.updateQuantity()`
6. CartContext calls backend API `PUT /api/v1/cart/update`
7. Backend validates and updates cart
8. Backend returns updated `CartResponse`
9. CartContext updates global cart state
10. Quantity editor shows updated quantity
11. Cart badge updates

### Remove from Cart Flow

1. User sets quantity to 0 in quantity editor
2. Or user clicks "Remove" button
3. Request sent to `CartContext.updateQuantity(quantity: 0)`
4. Backend removes item from cart
5. Backend returns updated `CartResponse`
6. CartContext updates global cart state
7. Button state changes to `idle`
8. Cart badge updates

---

## Database Schema (Reference)

### products Table
```sql
CREATE TABLE products (
  id UUID PRIMARY KEY,
  name VARCHAR(100) NOT NULL,
  description TEXT,
  price DECIMAL(10,2) NOT NULL,
  category_id UUID,
  inventory_count INTEGER DEFAULT 0,
  image_url VARCHAR(500),
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);
```

### product_variants Table
```sql
CREATE TABLE product_variants (
  id UUID PRIMARY KEY,
  product_id UUID NOT NULL REFERENCES products(id),
  name VARCHAR(50) NOT NULL,
  value VARCHAR(100) NOT NULL,
  price_modifier DECIMAL(10,2) DEFAULT 0,
  inventory_count INTEGER DEFAULT 0,
  attributes JSONB,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);
```

### cart_items Table
```sql
CREATE TABLE cart_items (
  id UUID PRIMARY KEY,
  cart_id UUID NOT NULL REFERENCES carts(id),
  product_id UUID NOT NULL REFERENCES products(id),
  variant_id UUID REFERENCES product_variants(id),
  quantity INTEGER NOT NULL CHECK (quantity > 0),
  unit_price DECIMAL(10,2) NOT NULL,
  subtotal DECIMAL(10,2) NOT NULL,
  created_at TIMESTAMP DEFAULT NOW(),
  updated_at TIMESTAMP DEFAULT NOW()
);
```

---

## Validation Logic

### Client-Side Validation

```typescript
function validateAddToCart(product: Product, variantId?: string, quantity: number): ValidationResult {
  // Product validation
  if (!product) {
    return { valid: false, error: 'Product not found' };
  }
  
  // Inventory validation
  if (product.inventory_count !== undefined && product.inventory_count === 0) {
    return { valid: false, error: 'Product is out of stock' };
  }
  
  // Variant validation
  if (variantId) {
    const variant = product.variants?.find(v => v.id === variantId);
    if (!variant) {
      return { valid: false, error: 'Invalid variant selected' };
    }
    if (variant.inventory_count === 0) {
      return { valid: false, error: 'Selected variant is out of stock' };
    }
  }
  
  // Quantity validation
  if (quantity <= 0) {
    return { valid: false, error: 'Quantity must be positive' };
  }
  if (quantity > 99) {
    return { valid: false, error: 'Maximum quantity is 99' };
  }
  
  return { valid: true };
}
```

### Server-Side Validation

Server must validate:
- Product exists and is available
- Variant exists and belongs to product (if provided)
- Inventory is sufficient
- User has permission to add to cart
- Quantity is within limits (1-99)
- Session is valid (for guest carts)

---

## Error Handling

### Client-Side Errors

```typescript
enum CartError {
  NETWORK_ERROR = 'network_error',
  OUT_OF_STOCK = 'out_of_stock',
  INVALID_PRODUCT = 'invalid_product',
  INVALID_VARIANT = 'invalid_variant',
  QUANTITY_EXCEEDED = 'quantity_exceeded',
  UNAUTHORIZED = 'unauthorized',
  SERVER_ERROR = 'server_error'
}

interface CartErrorResponse {
  error: CartError;
  message: string;
  details?: any;
}
```

### Error Display

- Network errors: "Connection failed. Please try again."
- Out of stock: "Sorry, this item is out of stock."
- Invalid product: "Product not found. Please refresh."
- Quantity exceeded: "Maximum quantity is 99."
- Server error: "Something went wrong. Please try again."

---

## State Management

### Global Cart State (CartContext)

```typescript
interface CartContextState {
  cart: Cart | null;
  isLoading: boolean;
  error: string | null;
  lastUpdated: string;
}

interface CartContextActions {
  addToCart: (request: AddToCartRequest) => Promise<boolean>;
  updateQuantity: (itemId: string, quantity: number) => Promise<boolean>;
  removeFromCart: (itemId: string) => Promise<boolean>;
  getCart: () => Promise<void>;
  clearCart: () => void;
}
```

### Local Component State

Each ProductSuggestionCard manages its own UI state:
- Loading state for Add to Cart button
- Quantity editor visibility
- Error message display

---

## Performance Considerations

### Data Caching

- Cache cart state in local storage (guest users)
- Cache product data in React Query cache
- Memoize expensive calculations (cart totals)

### Optimistic Updates

- Update UI optimistically on Add to Cart
- Revert on failure
- Show loading state immediately

### Debouncing

- Debounce rapid quantity changes (300ms)
- Debounce multiple Add to Cart clicks (200ms)

---

## Security Considerations

### Data Protection

- Never expose inventory counts to unauthorized users
- Server validates all cart operations
- Rate limit cart API calls
- Sanitize all user inputs

### Session Management

- Guest carts expire after 7 days
- Authenticated carts persist indefinitely
- WebSocket syncs cart across tabs

---

## Testing Data

### Mock Data Examples

```typescript
const mockProduct: Product = {
  id: '123e4567-e89b-12d3-a456-426614174000',
  name: 'Wireless Headphones',
  description: 'High-quality wireless headphones with noise cancellation',
  price: 199.99,
  category: {
    id: 'cat-001',
    name: 'Electronics'
  },
  tags: ['New', 'Popular'],
  inventory_count: 15,
  variants: [
    {
      id: 'var-001',
      product_id: '123e4567-e89b-12d3-a456-426614174000',
      name: 'Black',
      value: 'Black',
      inventory_count: 10,
      attributes: { type: 'color' }
    }
  ]
};

const mockCart: Cart = {
  id: 'cart-123',
  items: [
    {
      id: 'item-001',
      product_id: '123e4567-e89b-12d3-a456-426614174000',
      quantity: 2,
      unit_price: 199.99,
      subtotal: 399.98,
      product_name: 'Wireless Headphones',
      created_at: '2024-12-26T10:00:00Z',
      updated_at: '2024-12-26T10:00:00Z'
    }
  ],
  subtotal: 399.98,
  tax: 39.99,
  total: 439.97,
  created_at: '2024-12-26T10:00:00Z',
  updated_at: '2024-12-26T10:00:00Z'
};
```