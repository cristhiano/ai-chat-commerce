import React, { useState, useEffect, useRef } from 'react';
import type { ChatMessage, ProductSuggestion, ChatAction } from '../../types';
import ChatInput from './ChatInput';
import ChatMessageComponent from './ChatMessage';
import fetchService from '../../utils/fetch';

interface ChatInterfaceProps {
  sessionId?: string;
  userId?: string;
  onCartUpdate?: () => void;
}

const ChatInterface: React.FC<ChatInterfaceProps> = ({ 
  sessionId, 
  userId, 
  onCartUpdate 
}) => {
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [isConnected, setIsConnected] = useState(false);
  const [isTyping, setIsTyping] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const connectionAttempted = useRef<boolean>(false);
  const reconnecting = useRef<boolean>(false);
  const messagesContainerRef = useRef<HTMLDivElement | null>(null);

  // Generate session ID if not provided - use useMemo to prevent regeneration
  const currentSessionId = React.useMemo(() => {
    return sessionId || `session_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
  }, [sessionId]);

  useEffect(() => {
    // Prevent multiple connection attempts due to React StrictMode
    if (connectionAttempted.current) {
      return;
    }
    connectionAttempted.current = true;

    connectWebSocket();
    loadChatHistory();

    return () => {
      connectionAttempted.current = false;
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, [currentSessionId]);

  // Auto-scroll to bottom when new messages arrive
  useEffect(() => {
    scrollToBottom();
  }, [messages, isTyping]);

  const scrollToBottom = () => {
    if (messagesContainerRef.current) {
      messagesContainerRef.current.scrollTop = messagesContainerRef.current.scrollHeight;
    }
  };

  const connectWebSocket = () => {
    // Close existing connection if any
    if (wsRef.current) {
      wsRef.current.close();
    }

    const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${protocol}//localhost:8080/api/v1/chat/ws?session_id=${currentSessionId}`;
    
    try {
      const ws = new WebSocket(wsUrl);
      wsRef.current = ws;

      ws.onopen = () => {
        setIsConnected(true);
        setError(null);
        console.log('WebSocket connected');
      };

      ws.onmessage = (event) => {
        try {
          const data = JSON.parse(event.data);
          handleWebSocketMessage(data);
        } catch (err) {
          console.error('Failed to parse WebSocket message:', err);
        }
      };

      ws.onclose = () => {
        setIsConnected(false);
        console.log('WebSocket disconnected');
        // Only attempt to reconnect if not already reconnecting and connection was not intentionally closed
        if (!reconnecting.current && wsRef.current && wsRef.current.readyState === WebSocket.CLOSED) {
          reconnecting.current = true;
          setTimeout(() => {
            if (!isConnected && connectionAttempted.current) {
              console.log('Attempting to reconnect...');
              connectWebSocket();
            }
            reconnecting.current = false;
          }, 3000);
        }
      };

      ws.onerror = (error) => {
        console.error('WebSocket error:', error);
        setError('Connection error. Please refresh the page.');
      };
    } catch (err) {
      console.error('Failed to create WebSocket connection:', err);
      setError('Failed to connect to chat service.');
    }
  };

  const handleWebSocketMessage = (data: any) => {
    switch (data.type) {
      case 'message':
        const message: ChatMessage = {
          id: data.data.id,
          sessionId: data.data.session_id,
          userId: data.data.user_id,
          role: data.data.role,
          content: data.data.content,
          metadata: data.data.metadata,
          timestamp: data.data.timestamp,
        };
        // Prevent duplicate messages by checking if message already exists
        setMessages(prev => {
          const messageExists = prev.some(msg => msg.id === message.id);
          if (messageExists) {
            return prev;
          }
          return [...prev, message];
        });
        break;

      case 'typing':
        setIsTyping(data.data.is_typing);
        break;

      case 'suggestions':
        // Suggestions are now handled in message metadata, no need to store separately
        console.log('Received suggestions:', data.data);
        break;

      case 'actions':
        handleChatActions(data.data);
        break;

      case 'error':
        setError(data.data.message);
        break;

      default:
        console.log('Unknown message type:', data.type);
    }
  };

  const handleChatActions = (actions: ChatAction[]) => {
    actions.forEach(action => {
      switch (action.type) {
        case 'add_to_cart':
        case 'remove_from_cart':
          // Trigger cart update callback
          if (onCartUpdate) {
            onCartUpdate();
          }
          break;
        default:
          console.log('Unhandled action:', action);
      }
    });
  };

  const loadChatHistory = async () => {
    try {
      const result = await fetchService.get(`/api/v1/chat/history/${currentSessionId}`);
      if (result.data && result.data.success && result.data.data.messages) {
        setMessages(result.data.data.messages);
      } else if (result.error) {
        console.error('Failed to load chat history:', result.error);
      }
    } catch (err) {
      console.error('Failed to load chat history:', err);
    }
  };

  const sendMessage = (content: string) => {
    if (!wsRef.current || wsRef.current.readyState !== WebSocket.OPEN) {
      setError('Not connected to chat service');
      return;
    }

    // Immediately add user message to UI for instant feedback
    const userMessage: ChatMessage = {
      id: `user-${Date.now()}`,
      sessionId: currentSessionId,
      userId: userId,
      role: 'user',
      content,
      timestamp: new Date().toISOString(),
    };
    
    setMessages(prev => [...prev, userMessage]);

    // Send to server
    const message = {
      type: 'message',
      data: {
        content,
        session_id: currentSessionId,
        user_id: userId,
      },
    };

    wsRef.current.send(JSON.stringify(message));
  };

  const handleSuggestionClick = (suggestion: ProductSuggestion) => {
    if (suggestion.product) {
      sendMessage(`Tell me more about ${suggestion.product.name}`);
    }
  };

  return (
    <div className="flex flex-col h-full bg-white border border-gray-200 rounded-lg shadow-sm">
      {/* Header */}
      <div className="flex items-center justify-between p-4 border-b border-gray-200 bg-gray-50">
        <div className="flex items-center space-x-2">
          <div className="w-3 h-3 rounded-full bg-green-500"></div>
          <h3 className="text-lg font-semibold text-gray-900">Shopping Assistant</h3>
        </div>
        <div className="flex items-center space-x-2">
          <div className={`w-2 h-2 rounded-full ${isConnected ? 'bg-green-500' : 'bg-red-500'}`}></div>
          <span className="text-sm text-gray-600">
            {isConnected ? 'Connected' : 'Disconnected'}
          </span>
        </div>
      </div>

      {/* Error Message */}
      {error && (
        <div className="p-3 bg-red-50 border-b border-red-200">
          <div className="flex items-center">
            <div className="flex-shrink-0">
              <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
              </svg>
            </div>
            <div className="ml-3">
              <p className="text-sm text-red-800">{error}</p>
            </div>
          </div>
        </div>
      )}

      {/* Messages */}
      <div ref={messagesContainerRef} className="flex-1 overflow-y-auto p-4 space-y-4">
        {messages.length === 0 ? (
          <div className="text-center text-gray-500 py-8">
            <div className="mb-4">
              <svg className="mx-auto h-12 w-12 text-gray-400" fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M8 12h.01M12 12h.01M16 12h.01M21 12c0 4.418-4.03 8-9 8a9.863 9.863 0 01-4.255-.949L3 20l1.395-3.72C3.512 15.042 3 13.574 3 12c0-4.418 4.03-8 9-8s9 3.582 9 8z" />
              </svg>
            </div>
            <p className="text-lg font-medium">Welcome to your shopping assistant!</p>
            <p className="text-sm">Ask me about products, add items to your cart, or get recommendations.</p>
          </div>
        ) : (
          messages.map((message) => (
            <ChatMessageComponent
              key={message.id}
              message={message}
              onSuggestionClick={handleSuggestionClick}
            />
          ))
        )}

        {/* Typing Indicator */}
        {isTyping && (
          <div className="flex items-center space-x-2 text-gray-500">
            <div className="flex space-x-1">
              <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce"></div>
              <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0.1s' }}></div>
              <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0.2s' }}></div>
            </div>
            <span className="text-sm">Assistant is typing...</span>
          </div>
        )}
      </div>

      {/* Input */}
      <div className="border-t border-gray-200 p-4">
        <ChatInput
          onSendMessage={sendMessage}
          disabled={!isConnected}
          placeholder={isConnected ? "Ask me about products..." : "Connecting..."}
        />
      </div>
    </div>
  );
};

export default ChatInterface;
