# Quick Start Guide: Shopping Cart Page

## Overview

This guide provides a quick reference for developers implementing the Shopping Cart Page feature. It covers the essential concepts, key components, and common tasks.

## Key Concepts

### Cart Management Flow

1. **Cart Display:** CartPage fetches cart data from CartContext
2. **Quantity Updates:** User modifies quantity, optimistic update occurs, API call made, state refreshes
3. **Item Removal:** User confirms deletion, API call made, item removed from state
4. **Empty State:** No items in cart, show helpful message with browse button

### Data Flow Pattern

```
Component → CartContext → API Service → Backend → Database
     ↑                                                        ↓
     ←←←←←←←←←←←←←←←← State Update ←←←←←←←←←←←←←←←←←←←←←←←←←←
```

## Essential Components

### CartPage
Main container component that orchestrates the cart view.

**Key Responsibilities:**
- Integrate with CartContext for state management
- Render appropriate view based on cart state (empty, loading, populated, error)
- Handle navigation to product catalog and checkout

### CartItem
Displays individual cart item with product details and quantity controls.

**Key Props:**
- `item`: CartItem object with product data
- `onQuantityChange`: Callback for quantity updates
- `onRemove`: Callback for item removal

**Key Features:**
- Product image, name, SKU
- Quantity increment/decrement controls
- Remove button
- Variant information display

### CartSummary
Displays cart totals and action buttons.

**Key Sections:**
- Totals (subtotal, tax, shipping, total)
- Action buttons (Continue Shopping, Proceed to Checkout, Clear Cart)

## Common Tasks

### Adding a New Cart Operation

1. **Define the UI interaction** in CartItem or CartSummary
2. **Call CartContext method** (updateCartItem, removeFromCart, etc.)
3. **Handle optimistic update** (immediate UI feedback)
4. **Handle API response** (success or error)
5. **Refresh cart state** after successful operation

Example:
```typescript
const handleQuantityChange = async (newQuantity: number) => {
  // Optimistic update
  setLocalQuantity(newQuantity);
  
  // API call
  const success = await updateCartItem({ productId, quantity: newQuantity });
  
  // Handle result
  if (success) {
    await fetchCart(); // Refresh cart state
  } else {
    // Revert optimistic update
    setLocalQuantity(previousQuantity);
  }
};
```

### Handling Empty Cart State

When `cart.items.length === 0`:
- Render EmptyCart component
- Display helpful message
- Show "Browse Products" button
- Hide cart items and summary

### Implementing Quantity Controls

1. **Decrement button:**
   - Validates minimum quantity (1)
   - Updates cart through API
   - Shows loading state during update

2. **Increment button:**
   - Validates maximum quantity (99)
   - Checks inventory availability
   - Updates cart through API
   - Shows loading state during update

3. **Validation:**
   - Client-side validation before API call
   - Server-side validation for security
   - Display appropriate error messages

## API Usage

### Fetching Cart Data

```typescript
const { cart, loading, error, fetchCart } = useCart();

// Fetch on component mount
useEffect(() => {
  fetchCart();
}, []);
```

### Updating Item Quantity

```typescript
const { updateCartItem } = useCart();

await updateCartItem({
  product_id: productId,
  variant_id: variantId,
  quantity: newQuantity
});
```

### Removing an Item

```typescript
const { removeFromCart } = useCart();

await removeFromCart(productId, variantId);
```

### Clearing Cart

```typescript
const { clearCart } = useCart();

// Show confirmation dialog first
const confirmed = window.confirm('Clear all items from cart?');
if (confirmed) {
  await clearCart();
}
```

## Testing Patterns

### Unit Test: CartItem Rendering

```typescript
test('renders cart item with correct details', () => {
  const item = {
    product_id: '123',
    quantity: 2,
    unit_price: 19.99,
    total_price: 39.98,
    product_name: 'Test Product',
    sku: 'TEST-001'
  };
  
  render(<CartItem item={item} />);
  
  expect(screen.getByText('Test Product')).toBeInTheDocument();
  expect(screen.getByText('SKU: TEST-001')).toBeInTheDocument();
  expect(screen.getByText('$19.99')).toBeInTheDocument();
  expect(screen.getByText('2')).toBeInTheDocument();
});
```

### Integration Test: Quantity Update

```typescript
test('updates cart item quantity', async () => {
  const updateCartItem = jest.fn();
  
  render(
    <CartContext.Provider value={{ updateCartItem, ...otherContext }}>
      <CartItem item={mockItem} />
    </CartContext.Provider>
  );
  
  await userEvent.click(screen.getByLabelText('Increase quantity'));
  
  expect(updateCartItem).toHaveBeenCalledWith({
    product_id: mockItem.product_id,
    quantity: mockItem.quantity + 1
  });
});
```

## Accessibility Checklist

- [ ] All buttons have keyboard support (Enter/Space)
- [ ] Quantity controls have ARIA labels
- [ ] Error messages announced to screen readers
- [ ] Focus management on state changes
- [ ] Color contrast meets WCAG 2.1 AA standards
- [ ] Interactive elements have focus indicators

## Performance Optimization

### Debounce Quantity Updates

```typescript
const debouncedUpdateCart = useDebounce(
  async (quantity) => {
    await updateCartItem({ product_id, quantity });
  },
  300
);
```

### Optimistic UI Updates

```typescript
const handleQuantityChange = (newQuantity) => {
  // Optimistic update
  setOptimisticQuantity(newQuantity);
  
  // Actual update
  updateCartItem({ quantity: newQuantity }).then(
    () => setOptimisticQuantity(null), // Clear optimistic state
    () => setOptimisticQuantity(oldQuantity) // Revert on error
  );
};
```

## Error Handling

### Network Errors

```typescript
try {
  await updateCartItem(item);
} catch (error) {
  setError('Failed to update cart. Please try again.');
  // Revert optimistic update
}
```

### Validation Errors

```typescript
if (quantity > maxQuantity) {
  setError(`Maximum quantity is ${maxQuantity}`);
  return;
}

if (quantity < 1) {
  setError('Quantity must be at least 1');
  return;
}
```

## Common Pitfalls

1. **Forgetting to fetch cart on mount** - Always fetch cart data when CartPage mounts
2. **Not handling loading states** - Show loading spinners during API calls
3. **Missing error handling** - Gracefully handle API failures
4. **Confirmation dialogs** - Always confirm destructive actions (remove, clear)
5. **Accessibility** - Ensure keyboard navigation works for all interactions
6. **Optimistic updates** - Provide immediate feedback, then confirm with API

## Next Steps

After implementing the basic cart page:
1. Add smooth animations for quantity changes
2. Implement mobile-optimized sticky checkout button
3. Add accessibility features (ARIA labels, keyboard nav)
4. Optimize performance (React.memo, debouncing)
5. Write comprehensive tests
6. Cross-browser testing
7. Mobile device testing

