# Data Model: Shopping Cart Page

## Overview

This document defines the data structures and relationships for the Shopping Cart Page feature. The data model leverages existing backend cart infrastructure through the CartContext API.

## Entities

### ShoppingCart

The shopping cart entity is already defined in the backend (`backend/internal/models/models.go`). The frontend interfaces with this through the CartContext.

**Attributes:**
- `ID` (UUID): Unique cart identifier
- `SessionID` (string): Session-based cart tracking
- `UserID` (UUID, optional): User identifier when authenticated
- `Items` (JSON): Array of cart items stored as JSONB
- `Subtotal` (decimal): Sum of all item prices
- `TaxAmount` (decimal): Calculated tax amount
- `ShippingAmount` (decimal): Calculated shipping cost
- `TotalAmount` (decimal): Total cart value
- `Currency` (string): Currency code (default: USD)
- `CreatedAt` (timestamp): Cart creation time
- `UpdatedAt` (timestamp): Last update time

### CartItem (from CartService)

Individual items within the shopping cart, stored in the Items JSONB field.

**Attributes:**
- `ProductID` (UUID): Reference to product
- `VariantID` (UUID, optional): Product variant identifier
- `Quantity` (integer): Item quantity (min: 1, max: 99)
- `UnitPrice` (decimal): Price per unit
- `TotalPrice` (decimal): Quantity × UnitPrice
- `ProductName` (string): Display name for product
- `SKU` (string): Product SKU for tracking

**Validation Rules:**
- Quantity must be between 1 and 99
- UnitPrice must be positive
- TotalPrice must equal Quantity × UnitPrice
- ProductName and SKU are required

### CartResponse

The response structure returned by the CartContext API.

**Attributes:**
- `Items` (array of CartItem): All items in cart
- `Subtotal` (decimal): Sum of item prices
- `TaxAmount` (decimal): Calculated tax
- `ShippingAmount` (decimal): Shipping cost
- `TotalAmount` (decimal): Final total
- `Currency` (string): Currency code
- `ItemCount` (integer): Total number of distinct items

**State Scenarios:**
- **Empty Cart:** Items array is empty, all amounts are 0
- **Populated Cart:** Items array contains one or more items
- **Loading State:** No data available yet
- **Error State:** Error message present, items may be stale

## Relationships

### ShoppingCart → User
- Optional relationship when user is authenticated
- Allows cart persistence across sessions
- Guest users have session-only carts

### ShoppingCart → CartItem (One-to-Many)
- Shopping cart contains multiple items
- Items stored as JSON array in database
- Items managed through API endpoints

### CartItem → Product
- CartItem references Product by ProductID
- Product details fetched for display
- Inventory validation against Product

### CartItem → ProductVariant
- Optional relationship when item has variant
- Variant details (size, color, etc.) displayed
- Variant-specific pricing applied

## State Transitions

### Cart Lifecycle

1. **Empty State:**
   - No items in cart
   - User sees "Your cart is empty" message
   - Browse products button available

2. **Populated State:**
   - One or more items in cart
   - Items displayed with details
   - Totals calculated and displayed

3. **Loading State:**
   - Fetching cart data from API
   - Loading spinner displayed
   - Previous cart state may persist

4. **Error State:**
   - API call failed
   - Error message displayed
   - User can retry operation

### Item Operations

**Add Item:**
- POST /api/v1/cart/add
- Item added to Items array
- Totals recalculated
- Optimistic UI update

**Update Quantity:**
- PUT /api/v1/cart/update
- Quantity changed for specific item
- Totals recalculated
- Optimistic UI update

**Remove Item:**
- DELETE /api/v1/cart/remove/:product_id
- Item removed from Items array
- Totals recalculated
- Confirmation dialog shown

**Clear Cart:**
- DELETE /api/v1/cart/clear
- All items removed
- Totals reset to zero
- Confirmation dialog shown

## Data Flow

### Fetch Cart Flow

```
CartPage Component
  ↓
CartContext.fetchCart()
  ↓
API GET /api/v1/cart/
  ↓
Backend CartService.GetCart()
  ↓
Database SELECT query
  ↓
CartResponse with items and totals
  ↓
State update in CartContext
  ↓
CartPage re-renders with data
```

### Update Quantity Flow

```
User clicks increment button
  ↓
Optimistic UI update (immediate feedback)
  ↓
CartContext.updateCartItem()
  ↓
API PUT /api/v1/cart/update
  ↓
Backend CartService.UpdateCartItem()
  ↓
Database UPDATE query
  ↓
CartResponse with updated totals
  ↓
State update in CartContext
  ↓
CartPage re-renders with fresh data
```

### Remove Item Flow

```
User clicks remove button
  ↓
Confirmation dialog shown
  ↓
User confirms deletion
  ↓
CartContext.removeFromCart()
  ↓
API DELETE /api/v1/cart/remove/:product_id
  ↓
Backend CartService.RemoveFromCart()
  ↓
Database UPDATE (remove item from array)
  ↓
CartResponse with updated totals
  ↓
State update in CartContext
  ↓
CartPage re-renders without removed item
```

## Constraints and Validation

### Client-Side Validation
- Quantity must be between 1 and 99
- Quantity must be a positive integer
- Remove action requires confirmation
- Clear cart requires confirmation

### Server-Side Validation
- Item must exist in cart before update/remove
- Inventory available check on quantity increase
- Quantity constraints enforced
- Totals recalculated server-side

### Data Consistency
- Cart totals always match sum of item totals
- Item quantities never exceed available inventory
- Cart persists across page refreshes via CartContext
- Session-based cart for guest users

## Performance Considerations

### Optimization Strategies
- Items stored as JSON array (denormalized for performance)
- Lazy loading of cart data only when page accessed
- Optimistic UI updates for immediate feedback
- Debounce quantity changes to reduce API calls
- React.memo for CartItem to prevent unnecessary re-renders

### Caching Strategy
- CartContext caches cart state in memory
- No local storage caching (session-based only)
- Cart refetched on page load
- Auto-refresh after modifications

## Security Considerations

### Input Validation
- All quantity values validated client and server-side
- SQL injection prevented through GORM parameterized queries
- XSS prevented through React's automatic escaping
- CSRF protection via session-based authentication

### Authorization
- Cart visibility: Own cart only (session-based)
- User cart: Authenticated users can access their cart
- Guest cart: Session-based cart for unauthenticated users

