import React, { useState } from 'react';
import { Link } from 'react-router-dom';
import { useCart } from '../../contexts/CartContext';
import { formatCurrency } from '../../utils';

const ShoppingCart: React.FC = () => {
  const { cart, updateCartItem, removeFromCart, clearCart, isLoading } = useCart();
  const [updatingItems, setUpdatingItems] = useState<Set<string>>(new Set());

  // Handle quantity update
  const handleQuantityUpdate = async (productId: string, variantId: string | undefined, newQuantity: number) => {
    if (newQuantity < 1) {
      await handleRemoveItem(productId, variantId);
      return;
    }

    setUpdatingItems(prev => new Set(prev).add(productId));
    try {
      await updateCartItem({
        product_id: productId,
        variant_id: variantId,
        quantity: newQuantity,
      });
    } catch (error) {
      console.error('Failed to update cart item:', error);
    } finally {
      setUpdatingItems(prev => {
        const newSet = new Set(prev);
        newSet.delete(productId);
        return newSet;
      });
    }
  };

  // Handle remove item
  const handleRemoveItem = async (productId: string, variantId: string | undefined) => {
    setUpdatingItems(prev => new Set(prev).add(productId));
    try {
      await removeFromCart(productId, variantId);
    } catch (error) {
      console.error('Failed to remove cart item:', error);
    } finally {
      setUpdatingItems(prev => {
        const newSet = new Set(prev);
        newSet.delete(productId);
        return newSet;
      });
    }
  };

  // Handle clear cart
  const handleClearCart = async () => {
    if (window.confirm('Are you sure you want to clear your cart?')) {
      try {
        await clearCart();
      } catch (error) {
        console.error('Failed to clear cart:', error);
      }
    }
  };

  // Loading state
  if (isLoading) {
    return (
      <div className="flex justify-center items-center py-12">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  // Empty cart state
  if (!cart || cart.items.length === 0) {
    return (
      <div className="text-center py-12">
        <div className="text-gray-600 text-lg mb-4">Your cart is empty</div>
        <p className="text-gray-500 mb-6">Add some products to get started!</p>
        <Link
          to="/products"
          className="bg-blue-600 text-white px-6 py-3 rounded-lg hover:bg-blue-700 transition-colors"
        >
          Continue Shopping
        </Link>
      </div>
    );
  }

  return (
    <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
      <div className="flex items-center justify-between mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Shopping Cart</h1>
        <button
          onClick={handleClearCart}
          className="text-red-600 hover:text-red-800 text-sm"
        >
          Clear Cart
        </button>
      </div>

      <div className="grid grid-cols-1 lg:grid-cols-3 gap-8">
        {/* Cart Items */}
        <div className="lg:col-span-2">
          <div className="bg-white rounded-lg shadow-sm border">
            <div className="px-6 py-4 border-b border-gray-200">
              <h2 className="text-lg font-semibold text-gray-900">
                Cart Items ({cart.item_count})
              </h2>
            </div>
            
            <div className="divide-y divide-gray-200">
              {cart.items.map((item) => {
                const itemKey = `${item.product_id}-${item.variant_id || 'default'}`;
                const isUpdating = updatingItems.has(itemKey);
                
                return (
                  <div key={itemKey} className="p-6">
                    <div className="flex items-center space-x-4">
                      {/* Product Image Placeholder */}
                      <div className="w-20 h-20 bg-gray-200 rounded-lg flex-shrink-0">
                        <img
                          src="/placeholder-product.jpg"
                          alt={item.product_name}
                          className="w-full h-full object-cover rounded-lg"
                        />
                      </div>

                      {/* Product Info */}
                      <div className="flex-1 min-w-0">
                        <h3 className="text-lg font-medium text-gray-900 truncate">
                          {item.product_name}
                        </h3>
                        <p className="text-sm text-gray-500">
                          SKU: {item.sku}
                        </p>
                        {item.variant_id && (
                          <p className="text-sm text-gray-500">
                            Variant: {item.variant_id}
                          </p>
                        )}
                        <p className="text-lg font-semibold text-blue-600">
                          {formatCurrency(item.unit_price)}
                        </p>
                      </div>

                      {/* Quantity Controls */}
                      <div className="flex items-center space-x-3">
                        <button
                          onClick={() => handleQuantityUpdate(
                            item.product_id, 
                            item.variant_id, 
                            item.quantity - 1
                          )}
                          disabled={isUpdating}
                          className="w-8 h-8 rounded-full border border-gray-300 flex items-center justify-center hover:bg-gray-50 disabled:opacity-50"
                        >
                          -
                        </button>
                        
                        <span className={`text-lg font-medium w-12 text-center ${isUpdating ? 'opacity-50' : ''}`}>
                          {item.quantity}
                        </span>
                        
                        <button
                          onClick={() => handleQuantityUpdate(
                            item.product_id, 
                            item.variant_id, 
                            item.quantity + 1
                          )}
                          disabled={isUpdating}
                          className="w-8 h-8 rounded-full border border-gray-300 flex items-center justify-center hover:bg-gray-50 disabled:opacity-50"
                        >
                          +
                        </button>
                      </div>

                      {/* Item Total */}
                      <div className="text-right">
                        <p className="text-lg font-semibold text-gray-900">
                          {formatCurrency(item.total_price)}
                        </p>
                        <button
                          onClick={() => handleRemoveItem(item.product_id, item.variant_id)}
                          disabled={isUpdating}
                          className="text-red-600 hover:text-red-800 text-sm disabled:opacity-50"
                        >
                          Remove
                        </button>
                      </div>
                    </div>
                  </div>
                );
              })}
            </div>
          </div>
        </div>

        {/* Order Summary */}
        <div className="lg:col-span-1">
          <div className="bg-white rounded-lg shadow-sm border sticky top-8">
            <div className="px-6 py-4 border-b border-gray-200">
              <h2 className="text-lg font-semibold text-gray-900">
                Order Summary
              </h2>
            </div>
            
            <div className="px-6 py-4 space-y-4">
              {/* Subtotal */}
              <div className="flex justify-between">
                <span className="text-gray-600">Subtotal</span>
                <span className="font-medium">{formatCurrency(cart.subtotal)}</span>
              </div>

              {/* Tax */}
              <div className="flex justify-between">
                <span className="text-gray-600">Tax</span>
                <span className="font-medium">{formatCurrency(cart.tax_amount)}</span>
              </div>

              {/* Shipping */}
              <div className="flex justify-between">
                <span className="text-gray-600">Shipping</span>
                <span className="font-medium">{formatCurrency(cart.shipping_amount)}</span>
              </div>

              {/* Total */}
              <div className="border-t border-gray-200 pt-4">
                <div className="flex justify-between">
                  <span className="text-lg font-semibold text-gray-900">Total</span>
                  <span className="text-lg font-semibold text-blue-600">
                    {formatCurrency(cart.total_amount)}
                  </span>
                </div>
              </div>

              {/* Checkout Button */}
              <Link
                to="/checkout"
                className="w-full bg-blue-600 text-white py-3 px-6 rounded-lg font-semibold hover:bg-blue-700 transition-colors text-center block"
              >
                Proceed to Checkout
              </Link>

              {/* Continue Shopping */}
              <Link
                to="/products"
                className="w-full bg-gray-100 text-gray-700 py-3 px-6 rounded-lg font-semibold hover:bg-gray-200 transition-colors text-center block"
              >
                Continue Shopping
              </Link>
            </div>
          </div>

          {/* Shipping Info */}
          <div className="mt-6 bg-blue-50 rounded-lg p-4">
            <h3 className="text-sm font-semibold text-blue-900 mb-2">
              Free Shipping
            </h3>
            <p className="text-sm text-blue-700">
              Orders over $50 qualify for free shipping. Your order qualifies!
            </p>
          </div>

          {/* Security Info */}
          <div className="mt-4 bg-green-50 rounded-lg p-4">
            <h3 className="text-sm font-semibold text-green-900 mb-2">
              Secure Checkout
            </h3>
            <p className="text-sm text-green-700">
              Your payment information is encrypted and secure.
            </p>
          </div>
        </div>
      </div>
    </div>
  );
};

export default ShoppingCart;
