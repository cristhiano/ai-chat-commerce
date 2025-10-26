import React, { useState } from 'react';
import type { AddToCartRequest, CartResponse } from '../../types';
import QuantityEditor from './QuantityEditor';
import { useNotification } from '../../contexts/NotificationContext';

interface CartActionButtonProps {
  productId: string;
  variantId?: string;
  currentCart: CartResponse | null;
  onAddToCart: (item: AddToCartRequest) => Promise<boolean>;
  onUpdateQuantity?: (item: { product_id: string; variant_id?: string; quantity: number }) => Promise<boolean>;
  className?: string;
}

type ButtonState = 'idle' | 'loading' | 'success' | 'error' | 'in-cart';

const CartActionButton: React.FC<CartActionButtonProps> = ({
  productId,
  variantId,
  currentCart,
  onAddToCart,
  onUpdateQuantity,
  className = '',
}) => {
  const [error, setError] = useState<string | undefined>();
  const { addNotification, error: showError } = useNotification();

  // Find current quantity in cart for this product
  const currentQuantity = currentCart?.items.find(
    item => item.product_id === productId && item.variant_id === variantId
  )?.quantity;

  // Set initial state based on cart
  const isInCart = currentQuantity !== undefined && currentQuantity > 0;
  const initialState: ButtonState = isInCart ? 'in-cart' : 'idle';

  const [currentState, setCurrentState] = useState<ButtonState>(initialState);
  const [showQuantityEditor, setShowQuantityEditor] = useState(false);

  // Sync state with cart changes
  React.useEffect(() => {
    if (isInCart && currentState !== 'in-cart') {
      setCurrentState('in-cart');
    } else if (!isInCart && currentState === 'in-cart' && currentQuantity === 0) {
      setCurrentState('idle');
    }
  }, [isInCart, currentState, currentQuantity]);

  const handleClick = async () => {
    // If already in cart, open quantity editor
    if (currentState === 'in-cart') {
      setShowQuantityEditor(true);
      return;
    }

    // Add to cart
    setCurrentState('loading');
    setError(undefined);

    try {
      const success = await onAddToCart({
        product_id: productId,
        variant_id: variantId,
        quantity: 1,
      });

      if (success) {
        setCurrentState('success');
        addNotification({ type: 'success', message: 'Added to cart!', duration: 3000 });
        // Auto-transition to in-cart after 1 second
        setTimeout(() => {
          setCurrentState('in-cart');
        }, 1000);
      } else {
        setCurrentState('error');
        const errorMsg = 'Failed to add item to cart';
        setError(errorMsg);
        showError(errorMsg);
      }
    } catch (err) {
      setCurrentState('error');
      const errorMsg = err instanceof Error ? err.message : 'Failed to add item to cart';
      setError(errorMsg);
      showError(errorMsg);
    }
  };

  const handleUpdateQuantity = async (quantity: number): Promise<boolean> => {
    if (!onUpdateQuantity) {
      return false;
    }

    try {
      const successResult = await onUpdateQuantity({
        product_id: productId,
        variant_id: variantId,
        quantity,
      });

      if (successResult) {
        addNotification({ type: 'success', message: 'Cart updated!', duration: 3000 });
      }
      return successResult;
    } catch (err) {
      console.error('Failed to update quantity:', err);
      showError('Failed to update cart');
      return false;
    }
  };

  const handleRemove = async () => {
    if (!onUpdateQuantity) {
      return;
    }

    try {
      await onUpdateQuantity({
        product_id: productId,
        variant_id: variantId,
        quantity: 0,
      });
      setShowQuantityEditor(false);
      addNotification({ type: 'success', message: 'Item removed from cart', duration: 3000 });
    } catch (err) {
      console.error('Failed to remove item:', err);
      showError('Failed to remove item');
    }
  };

  const handleCloseEditor = () => {
    setShowQuantityEditor(false);
  };

  const handleRetry = () => {
    handleClick();
  };

  const buttonClasses = `
    inline-flex items-center justify-center rounded-md font-medium
    transition-all duration-200
    focus:outline-none focus:ring-2 focus:ring-offset-2
    disabled:opacity-50 disabled:cursor-not-allowed
    ${className}
  `;

  const renderButtonContent = () => {
    switch (currentState) {
      case 'idle':
        return (
          <>
            <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z" />
            </svg>
            Add to Cart
          </>
        );

      case 'loading':
        return (
          <>
            <svg className="animate-spin -ml-1 mr-3 h-5 w-5 text-white" xmlns="http://www.w3.org/2000/svg" fill="none" viewBox="0 0 24 24">
              <circle className="opacity-25" cx="12" cy="12" r="10" stroke="currentColor" strokeWidth="4"></circle>
              <path className="opacity-75" fill="currentColor" d="M4 12a8 8 0 018-8V0C5.373 0 0 5.373 0 12h4zm2 5.291A7.962 7.962 0 014 12H0c0 3.042 1.135 5.824 3 7.938l3-2.647z"></path>
            </svg>
            Adding...
          </>
        );

      case 'success':
        return (
          <>
            <svg className="w-5 h-5 mr-2 text-white" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M5 13l4 4L19 7" />
            </svg>
            Added!
          </>
        );

      case 'error':
        return (
          <>
            <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 8v4m0 4h.01M21 12a9 9 0 11-18 0 9 9 0 0118 0z" />
            </svg>
            Error
          </>
        );

      case 'in-cart':
        return (
          <>
            <svg className="w-5 h-5 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M3 3h2l.4 2M7 13h10l4-8H5.4M7 13L5.4 5M7 13l-2.293 2.293c-.63.63-.184 1.707.707 1.707H17m0 0a2 2 0 100 4 2 2 0 000-4zm-8 2a2 2 0 11-4 0 2 2 0 014 0z" />
            </svg>
            In cart ({currentQuantity})
          </>
        );

      default:
        return 'Add to Cart';
    }
  };

  const getButtonVariant = () => {
    switch (currentState) {
      case 'idle':
        return 'bg-blue-600 text-white hover:bg-blue-700 focus:ring-blue-500';
      case 'loading':
        return 'bg-blue-600 text-white cursor-wait';
      case 'success':
        return 'bg-green-600 text-white';
      case 'error':
        return 'bg-red-600 text-white hover:bg-red-700 focus:ring-red-500';
      case 'in-cart':
        return 'bg-gray-600 text-white hover:bg-gray-700 focus:ring-gray-500';
      default:
        return 'bg-blue-600 text-white hover:bg-blue-700 focus:ring-blue-500';
    }
  };

  return (
    <div className="w-full">
      <button
        onClick={currentState === 'error' ? handleRetry : handleClick}
        disabled={currentState === 'loading'}
        className={`${buttonClasses} ${getButtonVariant()} w-full py-2 px-4 text-sm`}
        aria-label={
          currentState === 'idle' ? 'Add item to cart' :
          currentState === 'loading' ? 'Adding to cart' :
          currentState === 'success' ? 'Added to cart successfully' :
          currentState === 'error' ? 'Error adding to cart, click to retry' :
          `Item in cart, quantity ${currentQuantity}`
        }
        aria-live="polite"
        aria-atomic="true"
      >
        {renderButtonContent()}
      </button>
      {currentState === 'error' && error && (
        <p className="mt-1 text-xs text-red-600" role="alert">
          {error}
        </p>
      )}
      
      {/* Quantity Editor */}
      {showQuantityEditor && isInCart && (
        <div className="transition-all duration-200 ease-in-out animate-in slide-in-from-top-1">
          <QuantityEditor
            productId={productId}
            variantId={variantId}
            currentQuantity={currentQuantity || 1}
            maxQuantity={99}
            onUpdate={handleUpdateQuantity}
            onRemove={handleRemove}
            onClose={handleCloseEditor}
          />
        </div>
      )}
    </div>
  );
};

export default CartActionButton;

