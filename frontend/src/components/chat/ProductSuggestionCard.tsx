import React from 'react';
import type { ProductSuggestion } from '../../types';

interface ProductSuggestionCardProps {
  suggestion: ProductSuggestion;
  onClick?: () => void;
  compact?: boolean;
}

const ProductSuggestionCard: React.FC<ProductSuggestionCardProps> = ({ 
  suggestion, 
  onClick, 
  compact = false 
}) => {
  const { product, reason, confidence } = suggestion;

  if (!product) {
    return null;
  }

  const handleClick = () => {
    if (onClick) {
      onClick();
    }
  };

  const cardClasses = `
    bg-white border border-gray-200 rounded-lg shadow-sm hover:shadow-md
    transition-shadow duration-200 cursor-pointer
    ${compact ? 'p-3' : 'p-4'}
    ${onClick ? 'hover:border-blue-300' : ''}
  `;

  return (
    <div className={cardClasses} onClick={handleClick}>
      <div className="flex space-x-3">
        {/* Product Image Placeholder */}
        <div className="flex-shrink-0">
          <div className={`
            bg-gray-200 rounded-lg flex items-center justify-center
            ${compact ? 'w-12 h-12' : 'w-16 h-16'}
          `}>
            <svg className="w-6 h-6 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M4 16l4.586-4.586a2 2 0 012.828 0L16 16m-2-2l1.586-1.586a2 2 0 012.828 0L20 14m-6-6h.01M6 20h12a2 2 0 002-2V6a2 2 0 00-2-2H6a2 2 0 00-2 2v12a2 2 0 002 2z" />
            </svg>
          </div>
        </div>

        {/* Product Info */}
        <div className="flex-1 min-w-0">
          <h3 className={`
            font-medium text-gray-900 truncate
            ${compact ? 'text-sm' : 'text-base'}
          `}>
            {product.name}
          </h3>
          
          <p className={`
            text-gray-600 mt-1
            ${compact ? 'text-xs line-clamp-2' : 'text-sm line-clamp-3'}
          `}>
            {product.description}
          </p>

          <div className="flex items-center justify-between mt-2">
            <div className="flex items-center space-x-2">
              <span className={`
                font-semibold text-blue-600
                ${compact ? 'text-sm' : 'text-base'}
              `}>
                ${product.price.toFixed(2)}
              </span>
              
              {confidence && (
                <span className={`
                  px-2 py-1 rounded-full text-xs font-medium
                  ${confidence > 0.8 
                    ? 'bg-green-100 text-green-800' 
                    : confidence > 0.6 
                    ? 'bg-yellow-100 text-yellow-800'
                    : 'bg-gray-100 text-gray-800'
                  }
                `}>
                  {Math.round(confidence * 100)}% match
                </span>
              )}
            </div>

            {reason && (
              <span className="text-xs text-gray-500 truncate max-w-20">
                {reason}
              </span>
            )}
          </div>

          {/* Category */}
          {product.category && (
            <div className="mt-1">
              <span className="inline-flex items-center px-2 py-1 rounded-full text-xs font-medium bg-gray-100 text-gray-800">
                {product.category.name}
              </span>
            </div>
          )}
        </div>

        {/* Click indicator */}
        {onClick && (
          <div className="flex-shrink-0 flex items-center">
            <svg className="w-4 h-4 text-gray-400" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 5l7 7-7 7" />
            </svg>
          </div>
        )}
      </div>
    </div>
  );
};

export default ProductSuggestionCard;
