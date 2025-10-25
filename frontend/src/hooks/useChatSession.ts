import { useState, useEffect, useCallback, useRef } from 'react';
import type { ChatMessage, ChatSession, ProductSuggestion } from '../types';
import { WebSocketService, createWebSocketService, disconnectWebSocketService } from '../services/websocket';
import fetchService from '../utils/fetch';

export interface UseChatSessionOptions {
  sessionId?: string;
  userId?: string;
  autoConnect?: boolean;
  onCartUpdate?: () => void;
}

export interface UseChatSessionReturn {
  // Connection state
  isConnected: boolean;
  isConnecting: boolean;
  error: string | null;
  
  // Chat state
  messages: ChatMessage[];
  suggestions: ProductSuggestion[];
  isTyping: boolean;
  
  // Actions
  sendMessage: (content: string) => void;
  connect: () => Promise<void>;
  disconnect: () => void;
  clearMessages: () => void;
  loadHistory: () => Promise<void>;
  
  // Session info
  sessionId: string;
}

export const useChatSession = (options: UseChatSessionOptions = {}): UseChatSessionReturn => {
  const {
    sessionId: providedSessionId,
    userId,
    autoConnect = true,
    onCartUpdate,
  } = options;

  // Generate session ID if not provided
  const sessionId = providedSessionId || `session_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;

  // State
  const [isConnected, setIsConnected] = useState(false);
  const [isConnecting, setIsConnecting] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [messages, setMessages] = useState<ChatMessage[]>([]);
  const [suggestions, setSuggestions] = useState<ProductSuggestion[]>([]);
  const [isTyping, setIsTyping] = useState(false);

  // Refs
  const wsServiceRef = useRef<WebSocketService | null>(null);

  // Initialize WebSocket service
  useEffect(() => {
    const wsService = createWebSocketService({
      sessionId,
      userId,
      onMessage: (message) => {
        setMessages(prev => [...prev, message]);
      },
      onTyping: (typing) => {
        setIsTyping(typing);
      },
      onSuggestions: (newSuggestions) => {
        setSuggestions(newSuggestions);
      },
      onActions: (actions) => {
        // Handle cart-related actions
        const cartActions = actions.filter(action => 
          action.type === 'add_to_cart' || action.type === 'remove_from_cart'
        );
        
        if (cartActions.length > 0 && onCartUpdate) {
          onCartUpdate();
        }
      },
      onError: (errorMessage) => {
        setError(errorMessage);
      },
      onConnect: () => {
        setIsConnected(true);
        setIsConnecting(false);
        setError(null);
      },
      onDisconnect: () => {
        setIsConnected(false);
        setIsConnecting(false);
      },
    });

    wsServiceRef.current = wsService;

    // Auto-connect if enabled
    if (autoConnect) {
      connect();
    }

    // Load chat history
    loadHistory();

    return () => {
      wsService.disconnect();
      disconnectWebSocketService();
    };
  }, [sessionId, userId, autoConnect]);

  const connect = useCallback(async () => {
    if (!wsServiceRef.current || isConnecting) {
      return;
    }

    setIsConnecting(true);
    setError(null);

    try {
      await wsServiceRef.current.connect();
    } catch (err) {
      setIsConnecting(false);
      setError('Failed to connect to chat service');
      console.error('Connection error:', err);
    }
  }, [isConnecting]);

  const disconnect = useCallback(() => {
    if (wsServiceRef.current) {
      wsServiceRef.current.disconnect();
    }
  }, []);

  const sendMessage = useCallback((content: string) => {
    if (!wsServiceRef.current || !isConnected) {
      setError('Not connected to chat service');
      return;
    }

    wsServiceRef.current.sendMessage(content);
  }, [isConnected]);

  const clearMessages = useCallback(() => {
    setMessages([]);
    setSuggestions([]);
  }, []);

  const loadHistory = useCallback(async () => {
    try {
      const result = await fetchService.get(`/api/v1/chat/history/${sessionId}`);
      if (result.data && result.data.success && result.data.data.messages) {
        setMessages(result.data.data.messages);
      } else if (result.error) {
        console.error('Failed to load chat history:', result.error);
      }
    } catch (err) {
      console.error('Failed to load chat history:', err);
    }
  }, [sessionId]);

  return {
    isConnected,
    isConnecting,
    error,
    messages,
    suggestions,
    isTyping,
    sendMessage,
    connect,
    disconnect,
    clearMessages,
    loadHistory,
    sessionId,
  };
};

// Hook for managing chat sessions
export const useChatSessions = () => {
  const [sessions, setSessions] = useState<ChatSession[]>([]);
  const [activeSessionId, setActiveSessionId] = useState<string | null>(null);

  const createSession = useCallback(async (userId?: string): Promise<string> => {
    try {
      const result = await fetchService.post('/api/v1/chat/session', { user_id: userId });
      
      if (result.data && result.data.success) {
          const sessionId = result.data.data.id;
          setActiveSessionId(sessionId);
          return sessionId;
        }
      } else if (result.error) {
        console.error('Failed to create session:', result.error);
      }
    } catch (error) {
      console.error('Failed to create chat session:', error);
    }

    // Fallback: generate session ID locally
    const sessionId = `session_${Date.now()}_${Math.random().toString(36).substr(2, 9)}`;
    setActiveSessionId(sessionId);
    return sessionId;
  }, []);

  const getSession = useCallback(async (sessionId: string): Promise<ChatSession | null> => {
    try {
      const result = await fetchService.get(`/api/v1/chat/session/${sessionId}`);
      if (result.data && result.data.success) {
        return result.data.data;
      } else if (result.error) {
        console.error('Failed to get chat session:', result.error);
      }
    } catch (error) {
      console.error('Failed to get chat session:', error);
    }
    return null;
  }, []);

  const loadSessions = useCallback(async (userId?: string) => {
    try {
      const url = userId ? `/api/v1/chat/sessions?user_id=${userId}` : '/api/v1/chat/sessions';
      const result = await fetchService.get(url);
      
      if (result.data && result.data.success) {
        setSessions(result.data.data);
      } else if (result.error) {
        console.error('Failed to load chat sessions:', result.error);
      }
    } catch (error) {
      console.error('Failed to load chat sessions:', error);
    }
  }, []);

  return {
    sessions,
    activeSessionId,
    createSession,
    getSession,
    loadSessions,
    setActiveSessionId,
  };
};
