// Core types for the ecommerce application

export interface Product {
  id: string;
  name: string;
  description: string;
  price: number;
  category_id: string;
  sku: string;
  status: string;
  metadata?: Record<string, any>;
  tags: string[];
  created_at: string;
  updated_at: string;
  category?: Category;
  variants?: ProductVariant[];
  inventory?: Inventory[];
}

export interface ProductVariant {
  id: string;
  product_id: string;
  variant_name: string;
  variant_value: string;
  price_modifier: number;
  sku_suffix: string;
  is_default: boolean;
  created_at: string;
}

export interface ProductImage {
  id: string;
  product_id: string;
  url: string;
  alt_text: string;
  is_primary: boolean;
  sort_order: number;
  created_at: string;
}

export interface Category {
  id: string;
  name: string;
  description: string;
  parent_id?: string;
  slug: string;
  sort_order: number;
  is_active: boolean;
  created_at: string;
  updated_at: string;
  parent?: Category;
  children?: Category[];
  products?: Product[];
}

export interface Inventory {
  id: string;
  product_id: string;
  variant_id?: string;
  warehouse_location: string;
  quantity_available: number;
  quantity_reserved: number;
  low_stock_threshold: number;
  reorder_point: number;
  last_restocked?: string;
  product?: Product;
  variant?: ProductVariant;
}

export interface CartItem {
  product_id: string;
  variant_id?: string;
  quantity: number;
  unit_price: number;
  total_price: number;
  product_name: string;
  sku: string;
}

export interface CartResponse {
  items: CartItem[];
  subtotal: number;
  tax_amount: number;
  shipping_amount: number;
  total_amount: number;
  currency: string;
  item_count: number;
}

export interface AddToCartRequest {
  product_id: string;
  variant_id?: string;
  quantity: number;
}

export interface UpdateCartItemRequest {
  product_id: string;
  variant_id?: string;
  quantity: number;
}

export interface User {
  id: string;
  email: string;
  first_name: string;
  last_name: string;
  phone?: string;
  date_of_birth?: string;
  preferences?: Record<string, any>;
  email_verified: boolean;
  status: string;
  created_at: string;
  updated_at: string;
}

export interface Order {
  id: string;
  order_number: string;
  user_id: string;
  session_id: string;
  status: string;
  subtotal: number;
  tax_amount: number;
  shipping_amount: number;
  total_amount: number;
  currency: string;
  payment_status: string;
  shipping_address: Record<string, any>;
  billing_address: Record<string, any>;
  payment_intent_id?: string;
  created_at: string;
  updated_at: string;
  user?: User;
  items?: OrderItem[];
}

export interface OrderItem {
  id: string;
  order_id: string;
  product_id: string;
  variant_id?: string;
  quantity: number;
  unit_price: number;
  total_price: number;
  product_snapshot?: Record<string, any>;
  created_at: string;
  order?: Order;
  product?: Product;
  variant?: ProductVariant;
}

export interface ProductFilters {
  search?: string;
  category_id?: string;
  min_price?: number;
  max_price?: number;
  status?: string;
  tags?: string[];
  page?: number;
  limit?: number;
  sort_by?: string;
  sort_order?: string;
}

export interface ProductListResponse {
  products: Product[];
  total: number;
  page: number;
  limit: number;
  total_pages: number;
  has_next: boolean;
  has_previous: boolean;
}

export interface ApiResponse<T> {
  data?: T;
  error?: string;
  message?: string;
}

export interface PaginationParams {
  page: number;
  limit: number;
}

export interface SearchParams {
  q: string;
  limit?: number;
}

// Auth types
export interface LoginRequest {
  email: string;
  password: string;
}

export interface RegisterRequest {
  email: string;
  password: string;
  first_name: string;
  last_name: string;
  phone?: string;
}

export interface AuthResponse {
  user: User;
  access_token: string;
  refresh_token: string;
  expires_in: number;
}

export interface RefreshTokenRequest {
  refresh_token: string;
}

// Chat types
export interface ChatMessage {
  id: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  timestamp: string;
  metadata?: Record<string, any>;
}

export interface ChatSession {
  id: string;
  session_id: string;
  user_id?: string;
  conversation_history: ChatMessage[];
  context: Record<string, any>;
  cart_state: CartItem[];
  preferences: Record<string, any>;
  status: string;
  last_activity: string;
  expires_at: string;
  created_at: string;
  updated_at: string;
}

// Error types
export interface ApiError {
  error: string;
  message?: string;
  details?: Record<string, any>;
}

// Form types
export interface ContactForm {
  name: string;
  email: string;
  subject: string;
  message: string;
}

export interface NewsletterForm {
  email: string;
}

// UI types
export interface LoadingState {
  isLoading: boolean;
  error?: string;
}

export interface ModalProps {
  isOpen: boolean;
  onClose: () => void;
  title?: string;
  children: React.ReactNode;
}

// Chat types
export interface ChatMessage {
  id: string;
  sessionId: string;
  userId?: string;
  role: 'user' | 'assistant' | 'system';
  content: string;
  metadata?: Record<string, any>;
  timestamp: string;
}

export interface ChatSession {
  id: string;
  userId?: string;
  context: Record<string, any>;
  createdAt: string;
  updatedAt: string;
}

export interface ChatAction {
  type: string;
  payload: Record<string, any>;
}

export interface ProductSuggestion {
  product: Product;
  reason: string;
  confidence: number;
}

export interface ChatResponse {
  message: string;
  actions?: ChatAction[];
  suggestions?: ProductSuggestion[];
  context?: Record<string, any>;
  error?: string;
}

export interface NotificationProps {
  type: 'success' | 'error' | 'warning' | 'info';
  message: string;
  duration?: number;
  onClose?: () => void;
}
