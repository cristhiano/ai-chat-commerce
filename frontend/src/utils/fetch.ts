// Centralized fetch utility with proper headers
import { API_CONFIG } from '../config/api';

interface FetchOptions extends RequestInit {
  includeSession?: boolean;
  includeAuth?: boolean;
}

class FetchService {
  private getSessionId(): string {
    let sessionId = localStorage.getItem('session_id');
    if (!sessionId) {
      sessionId = `session_${Math.random().toString(36).substr(2, 9)}_${Date.now()}`;
      localStorage.setItem('session_id', sessionId);
    }
    return sessionId;
  }

  private getAuthToken(): string | null {
    return localStorage.getItem('auth_token');
  }

  async fetch<T = any>(
    endpoint: string, 
    options: FetchOptions = {}
  ): Promise<{ data?: T; error?: string; message?: string }> {
    const url = `${API_CONFIG.BASE_URL}${endpoint}`;
    
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      ...(options.headers as Record<string, string>),
    };

    // Include session ID if requested (default: true)
    if (options.includeSession !== false) {
      headers['X-Session-ID'] = this.getSessionId();
    }

    // Include auth token if requested and available
    if (options.includeAuth !== false) {
      const token = this.getAuthToken();
      if (token) {
        headers.Authorization = `Bearer ${token}`;
      }
    }

    try {
      const response = await fetch(url, {
        ...options,
        headers,
      });

      const data = await response.json();

      if (!response.ok) {
        return {
          error: data.error || 'An error occurred',
          message: data.message,
        };
      }

      return { data };
    } catch (error) {
      return {
        error: error instanceof Error ? error.message : 'Network error',
      };
    }
  }

  // Convenience methods for common HTTP verbs
  async get<T = any>(endpoint: string, options: FetchOptions = {}): Promise<{ data?: T; error?: string; message?: string }> {
    return this.fetch<T>(endpoint, { ...options, method: 'GET' });
  }

  async post<T = any>(endpoint: string, body?: any, options: FetchOptions = {}): Promise<{ data?: T; error?: string; message?: string }> {
    return this.fetch<T>(endpoint, {
      ...options,
      method: 'POST',
      body: body ? JSON.stringify(body) : undefined,
    });
  }

  async put<T = any>(endpoint: string, body?: any, options: FetchOptions = {}): Promise<{ data?: T; error?: string; message?: string }> {
    return this.fetch<T>(endpoint, {
      ...options,
      method: 'PUT',
      body: body ? JSON.stringify(body) : undefined,
    });
  }

  async delete<T = any>(endpoint: string, options: FetchOptions = {}): Promise<{ data?: T; error?: string; message?: string }> {
    return this.fetch<T>(endpoint, { ...options, method: 'DELETE' });
  }
}

// Export singleton instance
export const fetchService = new FetchService();
export default fetchService;
