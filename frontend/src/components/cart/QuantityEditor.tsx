import React, { useState } from 'react';

interface QuantityEditorProps {
  productId: string;
  variantId?: string;
  currentQuantity: number;
  maxQuantity?: number;
  maxAvailableQuantity?: number; // Inventory-based limit
  onUpdate: (quantity: number) => Promise<boolean>;
  onRemove: () => void;
  onClose: () => void;
  className?: string;
}

const QuantityEditor: React.FC<QuantityEditorProps> = ({
  currentQuantity,
  maxQuantity = 99,
  maxAvailableQuantity,
  onUpdate,
  onRemove,
  onClose,
  className = '',
}) => {
  const [quantity, setQuantity] = useState(currentQuantity);
  const [isUpdating, setIsUpdating] = useState(false);

  // Calculate the actual maximum allowed quantity
  const actualMaxQuantity = maxAvailableQuantity !== undefined 
    ? Math.min(maxQuantity, maxAvailableQuantity)
    : maxQuantity;

  const handleIncrement = () => {
    if (quantity < actualMaxQuantity) {
      setQuantity(quantity + 1);
    }
  };

  const handleDecrement = () => {
    if (quantity > 1) {
      setQuantity(quantity - 1);
    }
  };

  const handleApply = async () => {
    if (quantity === 0) {
      onRemove();
      return;
    }

    setIsUpdating(true);
    try {
      await onUpdate(quantity);
    } finally {
      setIsUpdating(false);
      onClose();
    }
  };

  const handleRemove = () => {
    onRemove();
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter') {
      handleApply();
    } else if (e.key === 'Escape') {
      onClose();
    }
  };

  return (
    <div className={`mt-3 p-3 bg-gray-50 rounded-lg border border-gray-200 ${className}`}>
      <div className="flex items-center justify-between mb-3">
        <span className="text-sm font-medium text-gray-700">Quantity</span>
        <button
          onClick={onClose}
          className="text-gray-400 hover:text-gray-600 transition-colors"
          aria-label="Close editor"
          type="button"
        >
          <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M6 18L18 6M6 6l12 12" />
          </svg>
        </button>
      </div>

      <div className="flex items-center space-x-3 mb-3">
        <button
          onClick={handleDecrement}
          disabled={quantity <= 1 || isUpdating}
          className="w-11 h-11 rounded-full border border-gray-300 flex items-center justify-center hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          aria-label="Decrease quantity"
          type="button"
        >
          <span className="text-gray-600 font-bold">−</span>
        </button>

        <input
          type="number"
          value={quantity}
          onChange={(e) => setQuantity(Math.max(1, Math.min(actualMaxQuantity, parseInt(e.target.value) || 1)))}
          onKeyDown={handleKeyDown}
          min={1}
          max={actualMaxQuantity}
          className="w-16 text-center border border-gray-300 rounded px-2 py-1 focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          aria-label="Quantity"
          disabled={isUpdating}
        />

        <button
          onClick={handleIncrement}
          disabled={quantity >= actualMaxQuantity || isUpdating}
          className="w-11 h-11 rounded-full border border-gray-300 flex items-center justify-center hover:bg-gray-100 disabled:opacity-50 disabled:cursor-not-allowed transition-colors"
          aria-label="Increase quantity"
          type="button"
        >
          <span className="text-gray-600 font-bold">+</span>
        </button>
      </div>

      <div className="flex space-x-2">
        <button
          onClick={handleApply}
          disabled={isUpdating}
          className="flex-1 bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          type="button"
        >
          {isUpdating ? 'Updating...' : 'Apply'}
        </button>
        <button
          onClick={handleRemove}
          disabled={isUpdating}
          className="px-4 py-2 text-red-600 border border-red-600 rounded-md hover:bg-red-50 transition-colors disabled:opacity-50 disabled:cursor-not-allowed"
          type="button"
        >
          Remove
        </button>
      </div>
    </div>
  );
};

export default QuantityEditor;

