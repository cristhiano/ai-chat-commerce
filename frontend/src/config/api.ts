// API configuration and constants

export const API_CONFIG = {
  BASE_URL: import.meta.env.VITE_API_BASE_URL || '',
  WS_URL: import.meta.env.VITE_WS_BASE_URL || 'ws://localhost:8080/ws',
  TIMEOUT: 10000, // 10 seconds
  RETRY_ATTEMPTS: 3,
  RETRY_DELAY: 1000, // 1 second
};

export const API_ENDPOINTS = {
  // Products
  PRODUCTS: '/api/v1/products',
  PRODUCT_BY_ID: (id: string) => `/api/v1/products/${id}`,
  PRODUCT_BY_SKU: (sku: string) => `/api/v1/products/sku/${sku}`,
  PRODUCT_SEARCH: '/api/v1/products/search',
  FEATURED_PRODUCTS: '/api/v1/products/featured',
  RELATED_PRODUCTS: (id: string) => `/api/v1/products/${id}/related`,

  // Categories
  CATEGORIES: '/api/v1/categories',
  CATEGORY_BY_ID: (id: string) => `/api/v1/categories/${id}`,
  CATEGORY_BY_SLUG: (slug: string) => `/api/v1/categories/slug/${slug}`,

  // Cart
  CART: '/api/v1/cart',
  CART_ADD: '/api/v1/cart/add',
  CART_UPDATE: '/api/v1/cart/update',
  CART_REMOVE: (productId: string) => `/api/v1/cart/remove/${productId}`,
  CART_CLEAR: '/api/v1/cart/clear',
  CART_CALCULATE: '/api/v1/cart/calculate',
  CART_COUNT: '/api/v1/cart/count',

  // Auth
  LOGIN: '/api/v1/auth/login',
  REGISTER: '/api/v1/auth/register',
  REFRESH: '/api/v1/auth/refresh',

  // Users
  USER_PROFILE: '/api/v1/users/profile',
  USER_ORDERS: '/api/v1/users/orders',

  // Orders
  ORDERS: '/api/v1/orders',
  ORDER_BY_ID: (id: string) => `/api/v1/orders/${id}`,

  // Admin
  ADMIN_PRODUCTS: '/api/v1/admin/products',
  ADMIN_INVENTORY: '/api/v1/admin/inventory',
  ADMIN_ORDERS: '/api/v1/admin/orders',

  // Health
  HEALTH: '/health',
};

export const HTTP_STATUS = {
  OK: 200,
  CREATED: 201,
  NO_CONTENT: 204,
  BAD_REQUEST: 400,
  UNAUTHORIZED: 401,
  FORBIDDEN: 403,
  NOT_FOUND: 404,
  CONFLICT: 409,
  UNPROCESSABLE_ENTITY: 422,
  INTERNAL_SERVER_ERROR: 500,
  SERVICE_UNAVAILABLE: 503,
};

export const ERROR_MESSAGES = {
  NETWORK_ERROR: 'Network error. Please check your connection.',
  SERVER_ERROR: 'Server error. Please try again later.',
  UNAUTHORIZED: 'You are not authorized to perform this action.',
  NOT_FOUND: 'The requested resource was not found.',
  VALIDATION_ERROR: 'Please check your input and try again.',
  TIMEOUT: 'Request timed out. Please try again.',
  UNKNOWN_ERROR: 'An unknown error occurred.',
};

export const PAGINATION_DEFAULTS = {
  PAGE: 1,
  LIMIT: 20,
  MAX_LIMIT: 100,
};

export const SORT_OPTIONS = {
  PRODUCTS: {
    NAME_ASC: 'name',
    NAME_DESC: '-name',
    PRICE_ASC: 'price',
    PRICE_DESC: '-price',
    CREATED_ASC: 'created_at',
    CREATED_DESC: '-created_at',
    POPULARITY: 'popularity',
  },
  ORDERS: {
    DATE_ASC: 'created_at',
    DATE_DESC: '-created_at',
    STATUS_ASC: 'status',
    STATUS_DESC: '-status',
    TOTAL_ASC: 'total_amount',
    TOTAL_DESC: '-total_amount',
  },
};

export const CURRENCY_FORMAT = {
  USD: {
    symbol: '$',
    code: 'USD',
    precision: 2,
  },
  EUR: {
    symbol: '€',
    code: 'EUR',
    precision: 2,
  },
  GBP: {
    symbol: '£',
    code: 'GBP',
    precision: 2,
  },
};

export const VALIDATION_RULES = {
  EMAIL: /^[^\s@]+@[^\s@]+\.[^\s@]+$/,
  PHONE: /^[\+]?[1-9][\d]{0,15}$/,
  PASSWORD: {
    minLength: 8,
    requireUppercase: true,
    requireLowercase: true,
    requireNumbers: true,
    requireSpecialChars: true,
  },
  PRODUCT_SKU: /^[A-Z0-9-_]+$/,
  SLUG: /^[a-z0-9-]+$/,
};

export const CACHE_KEYS = {
  PRODUCTS: 'products',
  CATEGORIES: 'categories',
  CART: 'cart',
  USER: 'user',
  ORDERS: 'orders',
  FEATURED_PRODUCTS: 'featured_products',
};

export const CACHE_DURATION = {
  PRODUCTS: 5 * 60 * 1000, // 5 minutes
  CATEGORIES: 30 * 60 * 1000, // 30 minutes
  CART: 0, // No cache for cart
  USER: 10 * 60 * 1000, // 10 minutes
  ORDERS: 5 * 60 * 1000, // 5 minutes
  FEATURED_PRODUCTS: 15 * 60 * 1000, // 15 minutes
};