import React from 'react';
import ChatInterface from '../components/chat/ChatInterface';
import { useAuth } from '../contexts/AuthContext';

const ChatPage: React.FC = () => {
  const { user } = useAuth();

  return (
    <div className="container mx-auto px-4 py-8 h-screen flex flex-col">
      <div className="max-w-4xl mx-auto flex flex-col flex-1 w-full">
        <div className="mb-6 flex-shrink-0">
          <h1 className="text-3xl font-bold text-gray-900 mb-2">Shopping Assistant</h1>
          <p className="text-gray-600">
            Chat with our AI assistant to find products, get recommendations, and complete your purchase.
          </p>
        </div>

        <div className="flex-1 min-h-0">
          <ChatInterface 
            userId={user?.id}
            onCartUpdate={() => {
              // Trigger cart refresh if needed
              window.dispatchEvent(new CustomEvent('cart-updated'));
            }}
          />
        </div>
      </div>
    </div>
  );
};

export default ChatPage;
