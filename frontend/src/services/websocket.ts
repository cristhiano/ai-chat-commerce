import type { ChatMessage, ChatAction, ProductSuggestion } from '../types';

export interface WebSocketMessage {
  type: 'message' | 'typing' | 'suggestions' | 'actions' | 'error';
  data: any;
  sessionId: string;
  userId?: string;
}

export interface WebSocketServiceOptions {
  sessionId: string;
  userId?: string;
  onMessage?: (message: ChatMessage) => void;
  onTyping?: (isTyping: boolean) => void;
  onSuggestions?: (suggestions: ProductSuggestion[]) => void;
  onActions?: (actions: ChatAction[]) => void;
  onError?: (error: string) => void;
  onConnect?: () => void;
  onDisconnect?: () => void;
}

export class WebSocketService {
  private ws: WebSocket | null = null;
  private options: WebSocketServiceOptions;
  private reconnectAttempts = 0;
  private maxReconnectAttempts = 5;
  private reconnectDelay = 3000;
  private isConnecting = false;

  constructor(options: WebSocketServiceOptions) {
    this.options = options;
  }

  connect(): Promise<void> {
    return new Promise((resolve, reject) => {
      if (this.isConnecting || (this.ws && this.ws.readyState === WebSocket.OPEN)) {
        resolve();
        return;
      }

      this.isConnecting = true;

      try {
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const wsUrl = `${protocol}//${window.location.host}/api/v1/chat/ws?session_id=${this.options.sessionId}`;
        
        this.ws = new WebSocket(wsUrl);

        this.ws.onopen = () => {
          this.isConnecting = false;
          this.reconnectAttempts = 0;
          this.options.onConnect?.();
          resolve();
        };

        this.ws.onmessage = (event) => {
          try {
            const data: WebSocketMessage = JSON.parse(event.data);
            this.handleMessage(data);
          } catch (error) {
            console.error('Failed to parse WebSocket message:', error);
            this.options.onError?.('Failed to parse message');
          }
        };

        this.ws.onclose = (event) => {
          this.isConnecting = false;
          this.options.onDisconnect?.();
          
          // Attempt to reconnect if not a clean close
          if (event.code !== 1000 && this.reconnectAttempts < this.maxReconnectAttempts) {
            this.scheduleReconnect();
          }
        };

        this.ws.onerror = (error) => {
          this.isConnecting = false;
          console.error('WebSocket error:', error);
          this.options.onError?.('Connection error');
          reject(error);
        };

      } catch (error) {
        this.isConnecting = false;
        reject(error);
      }
    });
  }

  disconnect(): void {
    if (this.ws) {
      this.ws.close(1000, 'User disconnected');
      this.ws = null;
    }
  }

  sendMessage(content: string): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      this.options.onError?.('Not connected to chat service');
      return;
    }

    const message = {
      type: 'message',
      data: {
        content,
        session_id: this.options.sessionId,
        user_id: this.options.userId,
      },
    };

    this.ws.send(JSON.stringify(message));
  }

  sendTypingIndicator(isTyping: boolean): void {
    if (!this.ws || this.ws.readyState !== WebSocket.OPEN) {
      return;
    }

    const message = {
      type: 'typing',
      data: {
        is_typing: isTyping,
      },
      sessionId: this.options.sessionId,
    };

    this.ws.send(JSON.stringify(message));
  }

  private handleMessage(data: WebSocketMessage): void {
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
        this.options.onMessage?.(message);
        break;

      case 'typing':
        this.options.onTyping?.(data.data.is_typing);
        break;

      case 'suggestions':
        this.options.onSuggestions?.(data.data);
        break;

      case 'actions':
        this.options.onActions?.(data.data);
        break;

      case 'error':
        this.options.onError?.(data.data.message);
        break;

      default:
        console.log('Unknown message type:', data.type);
    }
  }

  private scheduleReconnect(): void {
    this.reconnectAttempts++;
    const delay = this.reconnectDelay * Math.pow(2, this.reconnectAttempts - 1);
    
    setTimeout(() => {
      if (this.reconnectAttempts <= this.maxReconnectAttempts) {
        console.log(`Attempting to reconnect (${this.reconnectAttempts}/${this.maxReconnectAttempts})`);
        this.connect().catch(error => {
          console.error('Reconnection failed:', error);
        });
      }
    }, delay);
  }

  get isConnected(): boolean {
    return this.ws?.readyState === WebSocket.OPEN;
  }

  get connectionState(): number {
    return this.ws?.readyState ?? WebSocket.CLOSED;
  }
}

// Singleton instance for global use
let wsServiceInstance: WebSocketService | null = null;

export const createWebSocketService = (options: WebSocketServiceOptions): WebSocketService => {
  if (wsServiceInstance) {
    wsServiceInstance.disconnect();
  }
  
  wsServiceInstance = new WebSocketService(options);
  return wsServiceInstance;
};

export const getWebSocketService = (): WebSocketService | null => {
  return wsServiceInstance;
};

export const disconnectWebSocketService = (): void => {
  if (wsServiceInstance) {
    wsServiceInstance.disconnect();
    wsServiceInstance = null;
  }
};
