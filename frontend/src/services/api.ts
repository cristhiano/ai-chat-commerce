import type { 
  Product,
  ProductListResponse,
  ProductFilters,
  Category,
  CartResponse,
  AddToCartRequest,
  UpdateCartItemRequest,
  User,
  Order,
  ApiResponse,
  SearchParams,
  AuthResponse,
  LoginRequest,
  RegisterRequest
} from '../types';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

class ApiService {
  private baseURL: string;
  private sessionId: string;

  constructor() {
    this.baseURL = API_BASE_URL;
    this.sessionId = this.getOrCreateSessionId();
  }

  private getOrCreateSessionId(): string {
    let sessionId = localStorage.getItem('session_id');
    if (!sessionId) {
      sessionId = this.generateSessionId();
      localStorage.setItem('session_id', sessionId);
    }
    return sessionId;
  }

  private generateSessionId(): string {
    return 'session_' + Math.random().toString(36).substr(2, 9) + '_' + Date.now();
  }

  private getAuthToken(): string | null {
    return localStorage.getItem('access_token');
  }

  private async request<T>(
    endpoint: string, 
    options: RequestInit = {}
  ): Promise<ApiResponse<T>> {
    const url = `${this.baseURL}${endpoint}`;
    const token = this.getAuthToken();
    
    const headers: Record<string, string> = {
      'Content-Type': 'application/json',
      'X-Session-ID': this.sessionId,
      ...(options.headers as Record<string, string>),
    };

    if (token) {
      headers.Authorization = `Bearer ${token}`;
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

  // Product API methods
  async getProducts(filters: ProductFilters = {}): Promise<ApiResponse<ProductListResponse>> {
    const params = new URLSearchParams();
    
    Object.entries(filters).forEach(([key, value]) => {
      if (value !== undefined && value !== null && value !== '') {
        if (Array.isArray(value)) {
          params.append(key, value.join(','));
        } else {
          params.append(key, String(value));
        }
      }
    });

    const queryString = params.toString();
    const endpoint = `/api/v1/products/${queryString ? `?${queryString}` : ''}`;
    
    return this.request<ProductListResponse>(endpoint);
  }

  async getProduct(id: string): Promise<ApiResponse<Product>> {
    return this.request<Product>(`/api/v1/products/${id}`);
  }

  async getProductBySKU(sku: string): Promise<ApiResponse<Product>> {
    return this.request<Product>(`/api/v1/products/sku/${sku}`);
  }

  async searchProducts(params: SearchParams): Promise<ApiResponse<Product[]>> {
    const queryParams = new URLSearchParams();
    queryParams.append('q', params.q);
    if (params.limit) {
      queryParams.append('limit', String(params.limit));
    }
    
    return this.request<Product[]>(`/api/v1/products/search?${queryParams.toString()}`);
  }

  async getFeaturedProducts(limit: number = 10): Promise<ApiResponse<Product[]>> {
    return this.request<Product[]>(`/api/v1/products/featured?limit=${limit}`);
  }

  async getRelatedProducts(productId: string, limit: number = 5): Promise<ApiResponse<Product[]>> {
    return this.request<Product[]>(`/api/v1/products/${productId}/related?limit=${limit}`);
  }

  // Category API methods
  async getCategories(): Promise<ApiResponse<Category[]>> {
    const response = await this.request<{ categories: Category[] }>('/api/v1/categories/');
    if (response.data) {
      return { data: response.data.categories };
    }
    return { error: response.error || 'Failed to fetch categories' };
  }

  async getCategory(id: string): Promise<ApiResponse<Category>> {
    return this.request<Category>(`/api/v1/categories/${id}`);
  }

  async getCategoryBySlug(slug: string): Promise<ApiResponse<Category>> {
    return this.request<Category>(`/api/v1/categories/slug/${slug}`);
  }

  // Cart API methods
  async getCart(): Promise<ApiResponse<CartResponse>> {
    return this.request<CartResponse>('/api/v1/cart/');
  }

  async addToCart(item: AddToCartRequest): Promise<ApiResponse<void>> {
    return this.request<void>('/api/v1/cart/add', {
      method: 'POST',
      body: JSON.stringify(item),
    });
  }

  async updateCartItem(item: UpdateCartItemRequest): Promise<ApiResponse<void>> {
    return this.request<void>('/api/v1/cart/update', {
      method: 'PUT',
      body: JSON.stringify(item),
    });
  }

  async removeFromCart(productId: string, variantId?: string): Promise<ApiResponse<void>> {
    const params = new URLSearchParams();
    if (variantId) {
      params.append('variant_id', variantId);
    }
    
    const queryString = params.toString();
    return this.request<void>(`/api/v1/cart/remove/${productId}${queryString ? `?${queryString}` : ''}`, {
      method: 'DELETE',
    });
  }

  async clearCart(): Promise<ApiResponse<void>> {
    return this.request<void>('/api/v1/cart/clear', {
      method: 'DELETE',
    });
  }

  async calculateCartTotals(): Promise<ApiResponse<CartResponse>> {
    return this.request<CartResponse>('/api/v1/cart/calculate', {
      method: 'POST',
    });
  }

  async getCartItemCount(): Promise<ApiResponse<{ item_count: number }>> {
    return this.request<{ item_count: number }>('/api/v1/cart/count');
  }

  // Auth API methods
  async login(credentials: LoginRequest): Promise<ApiResponse<AuthResponse>> {
    const response = await this.request<AuthResponse>('/api/v1/auth/login', {
      method: 'POST',
      body: JSON.stringify(credentials),
    });

    if (response.data) {
      localStorage.setItem('access_token', response.data.access_token);
      localStorage.setItem('refresh_token', response.data.refresh_token);
    }

    return response;
  }

  async register(userData: RegisterRequest): Promise<ApiResponse<AuthResponse>> {
    const response = await this.request<AuthResponse>('/api/v1/auth/register', {
      method: 'POST',
      body: JSON.stringify(userData),
    });

    if (response.data) {
      localStorage.setItem('access_token', response.data.access_token);
      localStorage.setItem('refresh_token', response.data.refresh_token);
    }

    return response;
  }

  async logout(): Promise<void> {
    localStorage.removeItem('access_token');
    localStorage.removeItem('refresh_token');
  }

  async refreshToken(): Promise<ApiResponse<AuthResponse>> {
    const refreshToken = localStorage.getItem('refresh_token');
    if (!refreshToken) {
      return { error: 'No refresh token available' };
    }

    const response = await this.request<AuthResponse>('/api/v1/auth/refresh', {
      method: 'POST',
      body: JSON.stringify({ refresh_token: refreshToken }),
    });

    if (response.data) {
      localStorage.setItem('access_token', response.data.access_token);
      localStorage.setItem('refresh_token', response.data.refresh_token);
    }

    return response;
  }

  // User API methods
  async getUserProfile(): Promise<ApiResponse<User>> {
    return this.request<User>('/api/v1/users/profile');
  }

  async updateUserProfile(userData: Partial<User>): Promise<ApiResponse<User>> {
    return this.request<User>('/api/v1/users/profile', {
      method: 'PUT',
      body: JSON.stringify(userData),
    });
  }

  async getUserOrders(): Promise<ApiResponse<Order[]>> {
    const response = await this.request<{ orders: Order[] }>('/api/v1/users/orders');
    if (response.data) {
      return { data: response.data.orders };
    }
    return { error: response.error || 'Failed to fetch orders' };
  }

  // Order API methods
  async createOrder(orderData: any): Promise<ApiResponse<Order>> {
    return this.request<Order>('/api/v1/orders', {
      method: 'POST',
      body: JSON.stringify(orderData),
    });
  }

  async getOrder(id: string): Promise<ApiResponse<Order>> {
    return this.request<Order>(`/api/v1/orders/${id}`);
  }

  // Health check
  async healthCheck(): Promise<ApiResponse<{ status: string; message: string }>> {
    return this.request<{ status: string; message: string }>('/health');
  }
}

// Create and export a singleton instance
export const apiService = new ApiService();
export default apiService;