import axios from 'axios';

const API_BASE_URL = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080/api';

// Types
export interface ProductResult {
  id: string;
  name: string;
  description: string;
  price: number;
  category: string;
  sku: string;
  availability: string;
  relevance_score: number;
  image_url?: string;
}

export interface PaginationInfo {
  current_page: number;
  page_size: number;
  total_pages: number;
  total_results: number;
  has_next: boolean;
  has_previous: boolean;
}

export interface SearchResponse {
  results: ProductResult[];
  pagination: PaginationInfo;
  query: string;
  total_results: number;
  response_time_ms: number;
}

export interface SearchRequest {
  query: string;
  filters?: {
    category_id?: string;
    price_min?: number;
    price_max?: number;
    availability?: string;
  };
  sort_by?: string;
  page?: number;
  page_size?: number;
}

export interface SuggestionsResponse {
  suggestions: string[];
}

export interface FilterOptions {
  price_ranges: Array<{
    min: number;
    max: number;
    label: string;
  }>;
  categories: Array<{
    id: string;
    name: string;
    product_count: number;
  }>;
  availability_options: string[];
}

export interface AnalyticsRequest {
  query: string;
  filters?: Record<string, any>;
  result_count: number;
  selected_products?: string;
  response_time_ms: number;
  cache_hit: boolean;
  session_id: string;
  user_id?: string;
}

// API Client
class SearchApiClient {
  private baseURL: string;

  constructor(baseURL: string = API_BASE_URL) {
    this.baseURL = baseURL;
  }

  // Search products
  async searchProducts(request: SearchRequest): Promise<SearchResponse> {
    try {
      const response = await axios.post<SearchResponse>(`${this.baseURL}/search`, request);
      return response.data;
    } catch (error) {
      if (axios.isAxiosError(error)) {
        throw new Error(error.response?.data?.error?.message || 'Search failed');
      }
      throw new Error('Search failed');
    }
  }

  // Get autocomplete suggestions
  async getSuggestions(query: string, limit: number = 10): Promise<string[]> {
    try {
      const response = await axios.get<SuggestionsResponse>(`${this.baseURL}/search/suggestions`, {
        params: { q: query, limit }
      });
      return response.data.suggestions;
    } catch (error) {
      if (axios.isAxiosError(error)) {
        throw new Error(error.response?.data?.error?.message || 'Failed to get suggestions');
      }
      throw new Error('Failed to get suggestions');
    }
  }

  // Get filter options
  async getFilterOptions(categoryId?: string): Promise<FilterOptions> {
    try {
      const response = await axios.get<FilterOptions>(`${this.baseURL}/search/filters`, {
        params: categoryId ? { category_id: categoryId } : {}
      });
      return response.data;
    } catch (error) {
      if (axios.isAxiosError(error)) {
        throw new Error(error.response?.data?.error?.message || 'Failed to get filter options');
      }
      throw new Error('Failed to get filter options');
    }
  }

  // Log analytics
  async logAnalytics(analytics: AnalyticsRequest): Promise<void> {
    try {
      await axios.post(`${this.baseURL}/search/analytics`, analytics);
    } catch (error) {
      // Analytics failures should not break the user experience
      console.warn('Failed to log analytics:', error);
    }
  }
}

// Create singleton instance
export const searchApi = new SearchApiClient();

// Utility functions
export const formatPrice = (price: number): string => {
  return new Intl.NumberFormat('en-US', {
    style: 'currency',
    currency: 'USD'
  }).format(price);
};

export const getAvailabilityText = (availability: string): string => {
  switch (availability) {
    case 'in_stock':
      return 'In Stock';
    case 'low_stock':
      return 'Low Stock';
    case 'out_of_stock':
      return 'Out of Stock';
    default:
      return 'Unknown';
  }
};

export const getAvailabilityColor = (availability: string): string => {
  switch (availability) {
    case 'in_stock':
      return 'text-green-600 bg-green-100';
    case 'low_stock':
      return 'text-yellow-600 bg-yellow-100';
    case 'out_of_stock':
      return 'text-red-600 bg-red-100';
    default:
      return 'text-gray-600 bg-gray-100';
  }
};

export default searchApi;
