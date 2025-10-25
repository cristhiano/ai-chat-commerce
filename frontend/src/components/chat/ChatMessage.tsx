import React from 'react';
import type { ChatMessage, ProductSuggestion } from '../../types';
import ProductSuggestionCard from './ProductSuggestionCard';

interface ChatMessageProps {
  message: ChatMessage;
  onSuggestionClick?: (suggestion: ProductSuggestion) => void;
}

const ChatMessageComponent: React.FC<ChatMessageProps> = ({ 
  message, 
  onSuggestionClick 
}) => {
  const isUser = message.role === 'user';
  const isSystem = message.role === 'system';

  const formatTimestamp = (timestamp: string) => {
    try {
      const date = new Date(timestamp);
      return date.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' });
    } catch {
      return '';
    }
  };

  const renderMessageContent = () => {
    // Check if message contains product suggestions in metadata
    const suggestions = message.metadata?.suggestions as ProductSuggestion[] || [];
    const actions = message.metadata?.actions || [];

    return (
      <div className="space-y-3">
        {/* Main message content */}
        <div className="prose prose-sm max-w-none">
          {message.content.split('\n').map((line, index) => (
            <p key={index} className="mb-2 last:mb-0">
              {line}
            </p>
          ))}
        </div>

        {/* Product suggestions */}
        {suggestions.length > 0 && (
          <div className="mt-4">
            <h4 className="text-sm font-medium text-gray-700 mb-2">Suggested Products:</h4>
            <div className="grid grid-cols-1 sm:grid-cols-2 gap-2">
              {suggestions.map((suggestion, index) => (
                <ProductSuggestionCard
                  key={index}
                  suggestion={suggestion}
                  onClick={() => onSuggestionClick?.(suggestion)}
                  compact
                />
              ))}
            </div>
          </div>
        )}

        {/* Actions taken */}
        {actions.length > 0 && (
          <div className="mt-3 p-2 bg-blue-50 rounded-lg border border-blue-200">
            <div className="flex items-center space-x-2">
              <svg className="w-4 h-4 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M9 12l2 2 4-4m6 2a9 9 0 11-18 0 9 9 0 0118 0z" />
              </svg>
              <span className="text-sm font-medium text-blue-800">Actions taken:</span>
            </div>
            <div className="mt-1 space-y-1">
              {actions.map((action: any, index: number) => (
                <div key={index} className="text-sm text-blue-700">
                  {action.type === 'add_to_cart' && 'Added to cart'}
                  {action.type === 'remove_from_cart' && 'Removed from cart'}
                  {action.type === 'show_product' && 'Showing product details'}
                  {action.type === 'checkout' && 'Proceeding to checkout'}
                </div>
              ))}
            </div>
          </div>
        )}
      </div>
    );
  };

  if (isSystem) {
    return (
      <div className="flex justify-center">
        <div className="bg-gray-100 text-gray-600 text-sm px-3 py-1 rounded-full">
          {message.content}
        </div>
      </div>
    );
  }

  return (
    <div className={`flex ${isUser ? 'justify-end' : 'justify-start'} mb-4`}>
      <div className={`flex max-w-[80%] ${isUser ? 'flex-row-reverse' : 'flex-row'} items-start space-x-2`}>
        {/* Avatar */}
        <div className={`
          flex-shrink-0 w-8 h-8 rounded-full flex items-center justify-center text-sm font-medium
          ${isUser 
            ? 'bg-blue-600 text-white' 
            : 'bg-gray-200 text-gray-600'
          }
        `}>
          {isUser ? (
            <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M10 9a3 3 0 100-6 3 3 0 000 6zm-7 9a7 7 0 1114 0H3z" clipRule="evenodd" />
            </svg>
          ) : (
            <svg className="w-4 h-4" fill="currentColor" viewBox="0 0 20 20">
              <path fillRule="evenodd" d="M18 10c0 3.866-3.582 7-8 7a8.841 8.841 0 01-4.083-.98L3 20l1.86-5.12A7.963 7.963 0 012 10c0-3.866 3.582-7 8-7s8 3.134 8 7zM5.5 9a1.5 1.5 0 113 0 1.5 1.5 0 01-3 0zm9 0a1.5 1.5 0 113 0 1.5 1.5 0 01-3 0z" clipRule="evenodd" />
            </svg>
          )}
        </div>

        {/* Message bubble */}
        <div className={`
          px-4 py-3 rounded-lg shadow-sm
          ${isUser 
            ? 'bg-blue-600 text-white' 
            : 'bg-white text-gray-900 border border-gray-200'
          }
        `}>
          {renderMessageContent()}
          
          {/* Timestamp */}
          <div className={`
            mt-2 text-xs
            ${isUser ? 'text-blue-100' : 'text-gray-500'}
          `}>
            {formatTimestamp(message.timestamp)}
          </div>
        </div>
      </div>
    </div>
  );
};

export default ChatMessageComponent;
