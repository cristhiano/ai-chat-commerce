import React from 'react';
import { useCart } from '../../contexts/CartContext';
import CartActionButton from '../cart/CartActionButton';
import type { ProductSuggestion } from '../../types';

interface ProductSuggestionCardProps {
  suggestion: ProductSuggestion;
  onClick?: () => void;
  compact?: boolean;
  showAddToCart?: boolean;
}

const ProductSuggestionCard: React.FC<ProductSuggestionCardProps> = ({ 
  suggestion, 
  onClick, 
  compact = false,
  showAddToCart = false
}) => {
  const { product, reason, confidence } = suggestion;
  const { cart, addToCart, updateCartItem } = useCart();

  if (!product) {
    return null;
  }

  const handleClick = (e: React.MouseEvent) => {
    // Don't trigger onClick if clicking on add to cart button
    if ((e.target as HTMLElement).closest('button')) {
      return;
    }
    if (onClick && !showAddToCart) {
      onClick();
    }
  };

  // Format price with currency symbol
  const formatPrice = (price: number): string => {
    return `$${price.toFixed(2)}`;
  };

  // Check if product is out of stock
  const isOutOfStock = product.inventory && product.inventory.some(inv => inv.quantity_available === 0);

  const cardClasses = `
    bg-white rounded-lg shadow-sm hover:shadow-md transition-shadow
    border border-gray-200
    ${onClick && !showAddToCart ? 'cursor-pointer' : ''}
    ${isOutOfStock ? 'opacity-75' : ''}
  `;

  return (
    <div className={cardClasses} onClick={handleClick}>
      {/* Product Image Placeholder */}
      <div className="h-48 bg-gray-200 rounded-t-lg flex items-center justify-center border-b border-gray-300">
        <span className="text-4xl text-gray-400">🛍️</span>
      </div>

      {/* Product Info */}
      <div className="p-4">
        <h3 className="font-semibold text-gray-900 mb-2 line-clamp-2">
          {product.name}
        </h3>
        
        <p className="text-sm text-gray-600 mb-3 line-clamp-2">
          {product.description}
        </p>

        {/* Price and Category */}
        <div className="flex items-center justify-between mb-2">
          <span className="text-lg font-bold text-blue-600">
            {formatPrice(product.price)}
          </span>
          {product.category && (
            <span className="text-xs text-gray-500">
              {product.category.name}
            </span>
          )}
        </div>

        {/* Tags */}
        {product.tags && product.tags.length > 0 && (
          <div className="mb-3 flex flex-wrap gap-1">
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
              onUpdateQuantity={async (item) => {
                return await updateCartItem({
                  product_id: item.product_id,
                  variant_id: item.variant_id,
                  quantity: item.quantity,
                });
              }}
              className={isOutOfStock ? 'opacity-50 cursor-not-allowed' : ''}
            />
            {/* Show out of stock message */}
            {isOutOfStock && (
              <p className="mt-2 text-xs text-red-600 text-center">
                Out of stock
              </p>
            )}
          </div>
        )}
      </div>
    </div>
  );
};

export default ProductSuggestionCard;
