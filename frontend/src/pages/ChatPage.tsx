import React from 'react';
import ChatInterface from '../components/chat/ChatInterface';
import { useAuth } from '../contexts/AuthContext';

const ChatPage: React.FC = () => {
  const { user } = useAuth();

  return (
    <div className="container mx-auto px-4 py-8">
      <div className="max-w-4xl mx-auto">
        <div className="mb-6">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">Shopping Assistant</h1>
          <p className="text-gray-600">
            Chat with our AI assistant to find products, get recommendations, and complete your purchase.
          </p>
        </div>

        <div className="h-[600px]">
          <ChatInterface 
            userId={user?.id}
            onCartUpdate={() => {
              // Trigger cart refresh if needed
              window.dispatchEvent(new CustomEvent('cart-updated'));
            }}
          />
        </div>

        <div className="mt-6 p-4 bg-blue-50 rounded-lg border border-blue-200">
          <h3 className="text-lg font-semibold text-blue-900 mb-2">How to use the chat:</h3>
          <ul className="text-blue-800 space-y-1">
            <li>• Ask about products: "Show me wireless headphones"</li>
            <li>• Get recommendations: "What's popular in electronics?"</li>
            <li>• Add to cart: "Add the blue t-shirt to my cart"</li>
            <li>• Check cart: "What's in my cart?"</li>
            <li>• Complete purchase: "I want to checkout"</li>
          </ul>
        </div>
      </div>
    </div>
  );
};

export default ChatPage;
