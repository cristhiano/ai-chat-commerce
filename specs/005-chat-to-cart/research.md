# Research: Product Cards with Add to Cart in Chat

## Overview
This document consolidates research findings and design decisions for implementing product cards with cart functionality in the chat interface.

## Design Decisions

### 1. Product Card Design Standardization

**Decision:** ProductSuggestionCard components in chat should match the visual design of product cards on ProductsPage.

**Rationale:** 
- User Experience Consistency: Users expect similar products to look the same across the application
- Visual Familiarity: Consistent card design reduces cognitive load
- Design System Compliance: Matches existing patterns in ProductsPage

**Alternatives Considered:**
- Keep simplified chat-only design: **Rejected** - causes inconsistent UX
- Create hybrid design: **Rejected** - adds complexity without benefit
- Full product detail cards: **Rejected** - too large for chat context

**Implementation:**
- Use same card dimensions and layout as ProductsPage
- Include: image placeholder, name, description, price, category, tags (first 2)
- Add cart action button below product info
- Responsive design for chat panel width

---

### 2. Product Image Handling

**Decision:** Use image placeholder (emoji/icon) consistent with ProductsPage until product images are implemented.

**Rationale:**
- Currently ProductsPage uses emoji placeholder (🛍️)
- Maintains visual consistency
- No additional API calls required
- Can be upgraded later when image API is ready

**Alternatives Considered:**
- Load images from product API: **Rejected** - API doesn't return image URLs yet
- Use default product images: **Rejected** - no asset server configured
- Fetch images separately: **Rejected** - adds latency and complexity

**Implementation:**
- Use same `w-full h-48 bg-gray-200` container with centered emoji
- Can replace with `<img>` when product.image_url becomes available

---

### 3. Product Name Click Behavior

**Decision:** Product name in chat should NOT navigate to product detail page; clicking card opens quantity editor if in cart, or shows product info if not.

**Rationale:**
- Chat context: Users want to stay in conversation
- Primary action: Add to cart, not navigation
- UX principle: Avoid breaking chat flow with page navigation

**Alternatives Considered:**
- Navigate to product page: **Rejected** - disrupts chat flow
- Show product details in modal: **Considered** - possible future enhancement
- Open in new tab: **Rejected** - poor mobile UX

**Implementation:**
- Make entire card clickable only for quantity editor (if already in cart)
- If not in cart, card click does nothing (user must use Add to Cart button)
- Prevent event bubbling from Add to Cart button click

---

### 4. Product Variants in Chat

**Decision:** For products with variants, show "Select Variant" button instead of Add to Cart, which opens variant selector modal.

**Rationale:**
- Simplified chat UX: Users shouldn't need to see all variants immediately
- Clear call to action: "Select Variant" is more explicit than disabled Add to Cart
- Modal pattern: Keeps variant selection contained to chat context

**Alternatives Considered:**
- Show all variants in card: **Rejected** - too many options in chat
- Disable Add to Cart: **Rejected** - not clear why disabled
- Always use first variant: **Rejected** - wrong product could be added

**Implementation:**
- Check if product has variants (product.variants array)
- If variants exist: Show "Select Variant" button
- On click: Open modal with variant options (size, color, etc.)
- After selection: Proceed with Add to Cart flow

---

### 5. Quantity Editor Display Pattern

**Decision:** Use inline quantity editor within card (not modal) when user clicks on item already in cart.

**Rationale:**
- Context preservation: Keep in card context
- Quick edit: Inline is faster than modal
- Visual simplicity: Matches e-commerce patterns (Amazon, etc.)

**Alternatives Considered:**
- Modal popup: **Rejected** - breaks visual flow, harder on mobile
- New page: **Rejected** - too disruptive
- Side panel: **Rejected** - not suitable for chat layout

**Implementation:**
- When item is in cart: Button shows "In cart (Qty)"
- On button click: Expand inline editor below button
- Show: decrease (-), quantity display, increase (+), Remove button
- Apply button or auto-apply on change
- Smooth height transition animation

---

### 6. Animation Style for State Transitions

**Decision:** Use smooth CSS transitions (200ms) with Tailwind classes for all state changes.

**Rationale:**
- Performance: CSS transitions are GPU-accelerated
- Consistency: Matches existing transition patterns
- Accessibility: Prefers-reduced-motion media query support

**Alternatives Considered:**
- No animations: **Rejected** - jarring UX
- JavaScript-based animations: **Rejected** - more complex, worse performance
- Spring animations: **Rejected** - overkill for button states

**Implementation:**
- Button state transitions: `transition-all duration-200`
- Quantity editor expand: `transition-height duration-200`
- Success checkmark: `animate-pulse` for 1 second
- Error shake: `animate-bounce` once

---

### 7. Cart Context Integration

**Decision:** Use existing CartContext methods (addToCart, updateQuantity) without modification.

**Rationale:**
- Already implemented and tested
- Handles guest and authenticated users
- Provides WebSocket synchronization
- No breaking changes required

**Implementation:**
```typescript
const { cart, addToCart, updateQuantity } = useCart();
```

---

### 8. Notification Pattern

**Decision:** Use existing NotificationContext for success/error messages.

**Rationale:**
- Consistent with rest of application
- Centralized notification management
- Auto-dismiss for success messages

**Implementation:**
```typescript
const { showNotification } = useNotification();
showNotification('Added to cart!', 'success');
```

---

## Technical Research

### Existing Patterns Analysis

**ProductsPage Product Cards:**
- Dimensions: Full width in grid (1-4 columns responsive)
- Image: 48 height (h-48), centered emoji placeholder
- Layout: Vertical stack (image, info, price/category)
- Interactions: Entire card is Link to product page
- Typography: font-semibold name, text-sm description

**ProductSuggestionCard (Current):**
- Layout: Horizontal (image left, info right)
- Compact mode: Smaller dimensions
- Click behavior: Sends message to chat
- Missing: Add to cart, full product display

**Gap Analysis:**
- Need to match vertical layout of ProductsPage
- Need to add cart functionality
- Need to add quantity management
- Need to improve visual design

### Cart API Integration

**Endpoints Used:**
- `POST /api/v1/cart/add` - Add item (returns CartResponse)
- `PUT /api/v1/cart/update` - Update quantity
- `GET /api/v1/cart/` - Get current cart state

**Request Format:**
```typescript
interface AddToCartRequest {
  product_id: string;
  variant_id?: string;
  quantity: number;
}
```

**Response Format:**
```typescript
interface CartResponse {
  items: CartItem[];
  total: number;
  subtotal: number;
}
```

---

## Component Architecture

### Enhanced ProductSuggestionCard

**Structure:**
```tsx
<div className="bg-white rounded-lg shadow-sm hover:shadow-md transition-shadow">
  {/* Image Section */}
  <div className="h-48 bg-gray-200 rounded-t-lg">
    <Image or Placeholder />
  </div>
  
  {/* Info Section */}
  <div className="p-4">
    <h3 className="font-semibold">Product Name</h3>
    <p className="text-sm text-gray-600 line-clamp-2">Description</p>
    
    {/* Price and Category */}
    <div className="flex items-center justify-between mt-2">
      <span className="text-lg font-bold text-blue-600">$Price</span>
      <span className="text-xs text-gray-500">Category</span>
    </div>
    
    {/* Tags */}
    <div className="mt-2">
      <Tags />
    </div>
    
    {/* Cart Action Button */}
    <div className="mt-3">
      <CartActionButton />
    </div>
  </div>
</div>
```

### CartActionButton Enhancement

**New Props:**
- `onEditQuantity: () => void` - Opens quantity editor
- `showQuantityEditor: boolean` - Controls inline editor visibility

**New Features:**
- Detects if item is in cart
- Shows current quantity
- Opens inline quantity editor
- Apply/Remove actions

### Quantity Editor Component

**New Component:** `QuantityEditor`
- Props: productId, variantId, currentQuantity
- Features: -/+, display, Remove
- On apply: calls updateQuantity or removes item
- Auto-collapse after action

---

## Open Questions (Resolved)

✅ **Q:** Should product images be loaded from product API?
**A:** Use placeholder matching ProductsPage (emoji) until image URLs available

✅ **Q:** What should happen when user clicks product name?
**A:** Card click opens quantity editor if in cart, otherwise no action (use Add to Cart button)

✅ **Q:** How to handle product variants in chat?
**A:** Show "Select Variant" button, open modal for selection

✅ **Q:** Should quantity editor be modal or inline?
**A:** Inline editor within card for context and speed

✅ **Q:** What animation style for state transitions?
**A:** CSS transitions (200ms) with Tailwind classes

---

## Dependencies

### Required Components (Available)
- ✅ CartContext
- ✅ NotificationContext
- ✅ ProductSuggestionCard (needs enhancement)
- ✅ CartActionButton (needs enhancement)

### New Components (To Create)
- 📦 QuantityEditor component
- 📦 VariantSelector modal

### External Libraries (Already in Use)
- React 18
- TypeScript
- Tailwind CSS
- @tanstack/react-query

---

## Testing Strategy

### Unit Tests
- CartActionButton: All states (idle, loading, success, error, in-cart)
- ProductSuggestionCard: Rendering, click handlers, cart integration
- QuantityEditor: Increment, decrement, remove, apply

### Integration Tests
- Cart operations: Add, update quantity, remove
- State synchronization: Cart badge updates
- Error handling: Network failures, out of stock

### E2E Tests
- Complete flow: Product suggestion → Add to Cart → Success → View cart
- Quantity management: Add item, increase quantity, remove item
- Variant selection: Select variant → Add to cart
- Error scenarios: Out of stock, network error

---

## Performance Considerations

### Optimizations
- Memoize ProductSuggestionCard with React.memo
- Debounce rapid button clicks (200ms)
- Optimistic UI updates for perceived speed
- Lazy load variant selector modal

### Monitoring
- Cart operation latency (target: <500ms)
- Button click response time (target: <100ms)
- Error rate (alert if >2%)

---

## Accessibility

### Requirements (WCAG 2.1 AA)
- Keyboard navigation for all buttons
- ARIA labels for state changes
- Screen reader announcements for cart updates
- Focus management in quantity editor
- High contrast for all text

### Implementation
- Button aria-label describes current state
- aria-live="polite" for dynamic content
- Focus trap in quantity editor
- Skip links for keyboard users

---

## Security Considerations

### Cart Operations
- Server validates inventory availability
- Server validates product exists and is available
- Rate limiting prevents abuse
- Session-based cart for guest users

### Data Validation
- Client validates quantity (positive number)
- Server validates variant exists and is available
- Server validates user owns cart items

---

## Migration Plan

### Existing ProductSuggestionCard Updates
1. Change layout from horizontal to vertical (match ProductsPage)
2. Add image placeholder section
3. Add category and tags display
4. Integrate CartActionButton
5. Add quantity editor support

### Backward Compatibility
- Keep `compact` prop for backward compatibility
- Keep `onClick` prop for backward compatibility
- `showAddToCart` prop controls new functionality

---

## Success Criteria

✅ Product cards match ProductsPage design
✅ Users can add to cart with single click
✅ Success feedback appears within 500ms
✅ Quantity management works smoothly
✅ Error states handled gracefully
✅ Mobile responsive design
✅ Accessibility compliant
✅ Code coverage ≥80%

---

## Next Steps

1. Generate data-model.md with product and cart entities
2. Create API contracts for cart operations
3. Write quickstart guide for implementation
4. Begin Phase 1: Component implementation