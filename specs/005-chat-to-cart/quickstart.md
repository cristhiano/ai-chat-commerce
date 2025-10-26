# Quickstart: Product Cards with Add to Cart in Chat

## Overview

This guide will help you quickly set up and implement enhanced product cards with cart functionality in the chat interface. The product cards will match the visual design of the product listings and enable users to add items to cart directly from chat.

**Estimated Time:** 2-3 hours  
**Prerequisites:** Basic React/TypeScript knowledge, understanding of existing CartContext

---

## Step 1: Understand Current Implementation

### Current Components

**ProductSuggestionCard** (`frontend/src/components/chat/ProductSuggestionCard.tsx`)
- Currently displays products horizontally with minimal styling
- Has basic cart integration but not visually appealing
- Missing: full product details, proper card layout, quantity management

**CartActionButton** (`frontend/src/components/cart/CartActionButton.tsx`)
- Already exists with basic state management
- Shows: idle, loading, success, error, in-cart states
- Missing: quantity editor for in-cart items

### Design Goal

Match the product card design from `ProductsPage` which includes:
- Image placeholder
- Product name
- Description (2 lines max)
- Price and category
- Tags (first 2)
- Add to Cart button

---

## Step 2: Update ProductSuggestionCard Component

### File Location
`frontend/src/components/chat/ProductSuggestionCard.tsx`

### Changes Required

1. **Update Layout**: Change from horizontal to vertical layout matching ProductsPage

```tsx
// OLD: Horizontal layout
<div className="flex space-x-3">
  <div className="flex-shrink-0 w-16 h-16">
    {/* Image */}
  </div>
  <div className="flex-1">
    {/* Product info */}
  </div>
</div>

// NEW: Vertical layout matching ProductsPage
<div className="bg-white rounded-lg shadow-sm hover:shadow-md transition-shadow">
  {/* Image section */}
  <div className="h-48 bg-gray-200 rounded-t-lg flex items-center justify-center">
    <span className="text-4xl text-gray-400">🛍️</span>
  </div>
  
  {/* Info section */}
  <div className="p-4">
    <h3 className="font-semibold text-gray-900 mb-2 line-clamp-2">
      {product.name}
    </h3>
    <p className="text-sm text-gray-600 mb-3 line-clamp-2">
      {product.description}
    </p>
    
    {/* Price and category */}
    <div className="flex items-center justify-between">
      <span className="text-lg font-bold text-blue-600">
        ${product.price.toFixed(2)}
      </span>
      {product.category && (
        <span className="text-xs text-gray-500">
          {product.category.name}
        </span>
      )}
    </div>
    
    {/* Tags */}
    {product.tags && product.tags.length > 0 && (
      <div className="mt-2 flex flex-wrap gap-1">
        {product.tags.slice(0, 2).map((tag) => (
          <span key={tag} className="text-xs bg-gray-100 text-gray-600 px-2 py-1 rounded">
            {tag}
          </span>
        ))}
      </div>
    )}
    
    {/* Add to Cart Button */}
    {showAddToCart && (
      <div className="mt-3">
        <CartActionButton
          productId={product.id}
          currentCart={cart}
          onAddToCart={addToCart}
        />
      </div>
    )}
  </div>
</div>
```

2. **Remove Compact Mode**: Consolidate to single full-card layout
3. **Update Click Behavior**: Prevent card click from interfering with button clicks

```tsx
const handleCardClick = (e: React.MouseEvent) => {
  // Don't trigger if clicking on button
  if ((e.target as HTMLElement).closest('button')) {
    return;
  }
  // Only handle click if not showing add to cart
  if (onClick && !showAddToCart) {
    onClick();
  }
};
```

---

## Step 3: Enhance CartActionButton

### File Location
`frontend/src/components/cart/CartActionButton.tsx`

### Changes Required

Add quantity editor functionality:

1. **Add new props**:
```tsx
interface CartActionButtonProps {
  // ... existing props
  showQuantityEditor?: boolean;
  onOpenQuantityEditor?: () => void;
}
```

2. **Detect in-cart state**:
```tsx
const isInCart = currentCart?.items.find(
  item => item.product_id === productId && item.variant_id === variantId
)?.quantity > 0;

// Show "In cart (Qty)" when in cart
if (isInCart) {
  return {
    state: 'in-cart',
    label: `In cart (${currentQuantity})`
  };
}
```

3. **Open quantity editor on click** (when already in cart):
```tsx
const handleInCartClick = () => {
  if (onOpenQuantityEditor) {
    onOpenQuantityEditor();
  }
};
```

---

## Step 4: Create QuantityEditor Component

### File Location
`frontend/src/components/cart/QuantityEditor.tsx`

### Create New Component

```tsx
import React, { useState } from 'react';

interface QuantityEditorProps {
  productId: string;
  variantId?: string;
  currentQuantity: number;
  maxQuantity?: number;
  onUpdate: (quantity: number) => void;
  onRemove: () => void;
  onClose: () => void;
}

const QuantityEditor: React.FC<QuantityEditorProps> = ({
  currentQuantity,
  maxQuantity = 99,
  onUpdate,
  onRemove,
  onClose,
}) => {
  const [quantity, setQuantity] = useState(currentQuantity);

  const handleIncrement = () => {
    if (quantity < maxQuantity) {
      setQuantity(quantity + 1);
    }
  };

  const handleDecrement = () => {
    if (quantity > 1) {
      setQuantity(quantity - 1);
    }
  };

  const handleApply = () => {
    if (quantity === 0) {
      onRemove();
    } else {
      onUpdate(quantity);
    }
    onClose();
  };

  return (
    <div className="mt-2 p-3 bg-gray-50 rounded-lg border border-gray-200">
      <div className="flex items-center justify-between mb-3">
        <span className="text-sm font-medium text-gray-700">Quantity</span>
        <button
          onClick={onClose}
          className="text-gray-400 hover:text-gray-600"
          aria-label="Close editor"
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <div className="flex items-center space-x-3 mb-3">
        <button
          onClick={handleDecrement}
          disabled={quantity <= 1}
          className="w-8 h-8 rounded-full border border-gray-300 flex items-center justify-center hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed"
          aria-label="Decrease quantity"
        >
          <span className="text-gray-600">−</span>
        </button>

        <input
          type="number"
          value={quantity}
          onChange={(e) => setQuantity(Math.max(1, Math.min(maxQuantity, parseInt(e.target.value) || 1)))}
          min={1}
          max={maxQuantity}
          className="w-16 text-center border border-gray-300 rounded px-2 py-1"
          aria-label="Quantity"
        />

        <button
          onClick={handleIncrement}
          disabled={quantity >= maxQuantity}
          className="w-8 h-8 rounded-full border border-gray-300 flex items-center justify-center hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed"
          aria-label="Increase quantity"
        >
          <span className="text-gray-600">+</span>
        </button>
      </div>

      <div className="flex space-x-2">
        <button
          onClick={handleApply}
          className="flex-1 bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 transition-colors"
        >
          Apply
        </button>
        <button
          onClick={onRemove}
          className="px-4 py-2 text-red-600 border border-red-600 rounded-md hover:bg-red-50 transition-colors"
        >
          Remove
        </button>
      </div>
    </div>
  );
};

export default QuantityEditor;
```

---

## Step 5: Update ChatInterface to Pass Cart Context

### File Location
`frontend/src/components/chat/ChatInterface.tsx`

### Changes Required

Ensure cart context is passed to ProductSuggestionCard:

```tsx
import { useCart } from '../../contexts/CartContext';

const ChatInterface: React.FC = () => {
  const { cart, addToCart } = useCart();
  
  // ... existing code
  
  return (
    <div>
      {/* ... chat messages */}
      
      {/* Product suggestions */}
      {suggestions.map((suggestion) => (
        <ProductSuggestionCard
          key={suggestion.product.id}
          suggestion={suggestion}
          showAddToCart={true}
          onAddToCart={addToCart}
        />
      ))}
    </div>
  );
};
```

---

## Step 6: Testing Your Implementation

### Manual Testing Steps

1. **Start the application**:
```bash
cd frontend
npm run dev
```

2. **Test Add to Cart**:
   - Open chat interface
   - Ask for product recommendations
   - Click "Add to Cart" on a product
   - Verify button shows "Adding..." then "Added!"
   - Verify cart badge updates

3. **Test Quantity Management**:
   - Click "Add to Cart" on same product again
   - Verify quantity editor opens
   - Change quantity and click "Apply"
   - Verify cart updates

4. **Test Error Handling**:
   - Test with out of stock product
   - Test with invalid product
   - Verify error messages appear

### Automated Testing

Run the test suite:
```bash
npm test
```

Key tests to add:
- `ProductSuggestionCard.test.tsx` - Rendering, layout, button interaction
- `CartActionButton.test.tsx` - State transitions, quantity management
- `QuantityEditor.test.tsx` - Increment, decrement, remove

---

## Step 7: Styling Adjustments

### Responsive Design

Ensure cards work on mobile:
```tsx
className="bg-white rounded-lg shadow-sm hover:shadow-md transition-shadow
  w-full max-w-sm mx-auto"
```

### Animation

Add smooth transitions:
```tsx
className="transition-all duration-200"
```

---

## Step 8: Accessibility

Add ARIA labels:
```tsx
<button
  aria-label="Add product to cart"
  aria-live="polite"
  aria-atomic="true"
>
  {buttonContent}
</button>
```

Keyboard navigation:
- Tab through buttons
- Enter to activate
- Escape to close quantity editor

---

## Common Issues and Solutions

### Issue: Button state doesn't update after adding to cart

**Solution**: Ensure CartContext updates are propogated. Check that cart state updates in CartContext trigger component re-renders.

```tsx
// Use effect to listen to cart changes
useEffect(() => {
  const item = cart?.items.find(
    item => item.product_id === productId && item.variant_id === variantId
  );
  setCurrentQuantity(item?.quantity || 0);
}, [cart, productId, variantId]);
```

### Issue: Card click interferes with button click

**Solution**: Use event.stopPropagation() on button click:

```tsx
const handleAddToCart = (e: React.MouseEvent) => {
  e.stopPropagation();
  // ... add to cart logic
};
```

### Issue: Quantity editor doesn't close

**Solution**: Pass onClose handler and manage state:

```tsx
const [isQuantityEditorOpen, setIsQuantityEditorOpen] = useState(false);

const handleCloseEditor = () => {
  setIsQuantityEditorOpen(false);
};
```

---

## Deployment

After implementing:

1. **Build for production**:
```bash
cd frontend
npm run build
```

2. **Test production build**:
```bash
npm run preview
```

3. **Verify**:
   - Cards render correctly
   - Cart operations work
   - No console errors
   - Performance is acceptable

---

## Next Steps

- Add product images when image API is available
- Implement variant selector modal
- Add product detail view from chat
- Enhance error messages
- Add analytics tracking

---

## Resources

- [React Context API](https://react.dev/reference/react/useContext)
- [TypeScript Interfaces](https://www.typescriptlang.org/docs/handbook/interfaces.html)
- [Tailwind CSS](https://tailwindcss.com/docs)
- [Accessibility Guidelines](https://www.w3.org/WAI/WCAG21/quickref/)

---

## Support

If you encounter issues:
1. Check console for errors
2. Verify CartContext is properly configured
3. Check network tab for API errors
4. Review test coverage

---

Good luck implementing enhanced product cards in chat! 🎉