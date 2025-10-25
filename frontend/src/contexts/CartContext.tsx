import React, { createContext, useContext, useReducer, useEffect } from 'react';
import type { CartResponse, AddToCartRequest, UpdateCartItemRequest } from '../types';
import { apiService } from '../services/api';

interface CartState {
  cart: CartResponse | null;
  loading: boolean;
  error: string | null;
}

interface CartContextType extends CartState {
  addToCart: (item: AddToCartRequest) => Promise<boolean>;
  updateCartItem: (item: UpdateCartItemRequest) => Promise<boolean>;
  removeFromCart: (productId: string, variantId?: string) => Promise<boolean>;
  clearCart: () => Promise<boolean>;
  fetchCart: () => Promise<void>;
  clearError: () => void;
}

type CartAction =
  | { type: 'CART_START' }
  | { type: 'CART_SUCCESS'; payload: CartResponse }
  | { type: 'CART_FAILURE'; payload: string }
  | { type: 'CART_CLEAR' }
  | { type: 'CLEAR_ERROR' };

const initialState: CartState = {
  cart: null,
  loading: false,
  error: null,
};

const cartReducer = (state: CartState, action: CartAction): CartState => {
  switch (action.type) {
    case 'CART_START':
      return {
        ...state,
        loading: true,
        error: null,
      };
    case 'CART_SUCCESS':
      return {
        ...state,
        cart: action.payload,
        loading: false,
        error: null,
      };
    case 'CART_FAILURE':
      return {
        ...state,
        loading: false,
        error: action.payload,
      };
    case 'CART_CLEAR':
      return {
        ...state,
        cart: null,
        loading: false,
        error: null,
      };
    case 'CLEAR_ERROR':
      return {
        ...state,
        error: null,
      };
    default:
      return state;
  }
};

const CartContext = createContext<CartContextType | undefined>(undefined);

export const useCart = () => {
  const context = useContext(CartContext);
  if (context === undefined) {
    throw new Error('useCart must be used within a CartProvider');
  }
  return context;
};

interface CartProviderProps {
  children: React.ReactNode;
}

export const CartProvider: React.FC<CartProviderProps> = ({ children }) => {
  const [state, dispatch] = useReducer(cartReducer, initialState);

  const fetchCart = async () => {
    dispatch({ type: 'CART_START' });
    
    try {
      const response = await apiService.getCart();
      
      if (response.error) {
        dispatch({ type: 'CART_FAILURE', payload: response.error });
      } else if (response.data) {
        dispatch({ type: 'CART_SUCCESS', payload: response.data });
      }
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to fetch cart';
      dispatch({ type: 'CART_FAILURE', payload: errorMessage });
    }
  };

  const addToCart = async (item: AddToCartRequest): Promise<boolean> => {
    dispatch({ type: 'CART_START' });
    
    try {
      const response = await apiService.addToCart(item);
      
      if (response.error) {
        dispatch({ type: 'CART_FAILURE', payload: response.error });
        return false;
      }
      
      // Refresh cart after successful add
      await fetchCart();
      return true;
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to add to cart';
      dispatch({ type: 'CART_FAILURE', payload: errorMessage });
      return false;
    }
  };

  const updateCartItem = async (item: UpdateCartItemRequest): Promise<boolean> => {
    dispatch({ type: 'CART_START' });
    
    try {
      const response = await apiService.updateCartItem(item);
      
      if (response.error) {
        dispatch({ type: 'CART_FAILURE', payload: response.error });
        return false;
      }
      
      // Refresh cart after successful update
      await fetchCart();
      return true;
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to update cart item';
      dispatch({ type: 'CART_FAILURE', payload: errorMessage });
      return false;
    }
  };

  const removeFromCart = async (productId: string, variantId?: string): Promise<boolean> => {
    dispatch({ type: 'CART_START' });
    
    try {
      const response = await apiService.removeFromCart(productId, variantId);
      
      if (response.error) {
        dispatch({ type: 'CART_FAILURE', payload: response.error });
        return false;
      }
      
      // Refresh cart after successful removal
      await fetchCart();
      return true;
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to remove from cart';
      dispatch({ type: 'CART_FAILURE', payload: errorMessage });
      return false;
    }
  };

  const clearCart = async (): Promise<boolean> => {
    dispatch({ type: 'CART_START' });
    
    try {
      const response = await apiService.clearCart();
      
      if (response.error) {
        dispatch({ type: 'CART_FAILURE', payload: response.error });
        return false;
      }
      
      dispatch({ type: 'CART_CLEAR' });
      return true;
    } catch (error) {
      const errorMessage = error instanceof Error ? error.message : 'Failed to clear cart';
      dispatch({ type: 'CART_FAILURE', payload: errorMessage });
      return false;
    }
  };

  const clearError = () => {
    dispatch({ type: 'CLEAR_ERROR' });
  };

  // Fetch cart on mount
  useEffect(() => {
    fetchCart();
  }, []);

  const value: CartContextType = {
    ...state,
    addToCart,
    updateCartItem,
    removeFromCart,
    clearCart,
    fetchCart,
    clearError,
  };

  return (
    <CartContext.Provider value={value}>
      {children}
    </CartContext.Provider>
  );
};
