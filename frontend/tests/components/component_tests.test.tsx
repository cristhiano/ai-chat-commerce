import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import ProductList from '../../src/components/product/ProductList';
import ProductDetail from '../../src/components/product/ProductDetail';
import SearchBar from '../../src/components/search/SearchBar';
import { LoginForm, RegisterForm } from '../../src/components/auth/UserAuth';
import type { Product, ProductFilters } from '../../src/types';

// Mock API service
vi.mock('../../src/services/api', () => ({
  apiService: {
    getProducts: vi.fn(),
    getProduct: vi.fn(),
    searchProducts: vi.fn(),
  },
}));

// Mock contexts
vi.mock('../../src/contexts/CartContext', () => ({
  useCart: () => ({
    addToCart: vi.fn(),
    cart: {
      items: [],
      item_count: 0,
      total_amount: 0,
      subtotal: 0,
      tax_amount: 0,
      shipping_amount: 0,
    },
    isLoading: false,
  }),
}));

vi.mock('../../src/contexts/AuthContext', () => ({
  useAuth: () => ({
    user: null,
    isAuthenticated: false,
    login: vi.fn(),
    register: vi.fn(),
  }),
}));

// Test wrapper component
const TestWrapper = ({ children }: { children: React.ReactNode }) => {
  const queryClient = new QueryClient({
    defaultOptions: {
      queries: {
        retry: false,
      },
    },
  });

  return (
    <QueryClientProvider client={queryClient}>
      <BrowserRouter>
        {children}
      </BrowserRouter>
    </QueryClientProvider>
  );
};

// Mock product data
const mockProduct: Product = {
  id: '1',
  name: 'Test Product',
  description: 'A test product',
  price: 99.99,
  sku: 'TEST-001',
  status: 'active',
  category_id: 'cat-1',
  category: {
    id: 'cat-1',
    name: 'Electronics',
    slug: 'electronics',
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
  variants: [],
  images: [
    {
      id: 'img-1',
      product_id: '1',
      url: 'https://example.com/image.jpg',
      alt_text: 'Test Product Image',
      sort_order: 1,
      created_at: '2024-01-01T00:00:00Z',
      updated_at: '2024-01-01T00:00:00Z',
    },
  ],
  inventory: {
    id: 'inv-1',
    product_id: '1',
    quantity: 100,
    reserved_quantity: 0,
    created_at: '2024-01-01T00:00:00Z',
    updated_at: '2024-01-01T00:00:00Z',
  },
  tags: ['test', 'product'],
  created_at: '2024-01-01T00:00:00Z',
  updated_at: '2024-01-01T00:00:00Z',
};

const mockProductsResponse = {
  data: {
    products: [mockProduct],
    total: 1,
    page: 1,
    limit: 10,
    total_pages: 1,
    has_next: false,
    has_previous: false,
  },
};

describe('ProductList Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders product list correctly', async () => {
    const { apiService } = await import('../../src/services/api');
    vi.mocked(apiService.getProducts).mockResolvedValue(mockProductsResponse);

    render(
      <TestWrapper>
        <ProductList />
      </TestWrapper>
    );

    await waitFor(() => {
      expect(screen.getByText('Test Product')).toBeInTheDocument();
    });

    expect(screen.getByText('A test product')).toBeInTheDocument();
    expect(screen.getByText('$99.99')).toBeInTheDocument();
    expect(screen.getByText('SKU: TEST-001')).toBeInTheDocument();
  });

  it('displays loading state', () => {
    render(
      <TestWrapper>
        <ProductList />
      </TestWrapper>
    );

    expect(screen.getByRole('status')).toBeInTheDocument();
  });

  it('displays error state', async () => {
    const { apiService } = await import('../../src/services/api');
    vi.mocked(apiService.getProducts).mockRejectedValue(new Error('API Error'));

    render(
      <TestWrapper>
        <ProductList />
      </TestWrapper>
    );

    await waitFor(() => {
      expect(screen.getByText('Failed to load products')).toBeInTheDocument();
    });

    expect(screen.getByText('Try Again')).toBeInTheDocument();
  });

  it('displays empty state', async () => {
    const { apiService } = await import('../../src/services/api');
    vi.mocked(apiService.getProducts).mockResolvedValue({
      data: {
        products: [],
        total: 0,
        page: 1,
        limit: 10,
        total_pages: 0,
        has_next: false,
        has_previous: false,
      },
    });

    render(
      <TestWrapper>
        <ProductList />
      </TestWrapper>
    );

    await waitFor(() => {
      expect(screen.getByText('No products found')).toBeInTheDocument();
    });
  });

  it('handles filter changes', async () => {
    const { apiService } = await import('../../src/services/api');
    vi.mocked(apiService.getProducts).mockResolvedValue(mockProductsResponse);

    render(
      <TestWrapper>
        <ProductList showFilters={true} />
      </TestWrapper>
    );

    await waitFor(() => {
      expect(screen.getByText('Test Product')).toBeInTheDocument();
    });

    const searchInput = screen.getByPlaceholderText('Search products...');
    fireEvent.change(searchInput, { target: { value: 'test search' } });

    await waitFor(() => {
      expect(apiService.getProducts).toHaveBeenCalledWith(
        expect.objectContaining({
          search: 'test search',
        })
      );
    });
  });

  it('handles pagination', async () => {
    const { apiService } = await import('../../src/services/api');
    vi.mocked(apiService.getProducts).mockResolvedValue({
      data: {
        products: [mockProduct],
        total: 25,
        page: 1,
        limit: 10,
        total_pages: 3,
        has_next: true,
        has_previous: false,
      },
    });

    render(
      <TestWrapper>
        <ProductList />
      </TestWrapper>
    );

    await waitFor(() => {
      expect(screen.getByText('Test Product')).toBeInTheDocument();
    });

    const nextButton = screen.getByText('Next');
    fireEvent.click(nextButton);

    await waitFor(() => {
      expect(apiService.getProducts).toHaveBeenCalledWith(
        expect.objectContaining({
          page: 2,
        })
      );
    });
  });
});

describe('ProductDetail Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders product detail correctly', async () => {
    const { apiService } = await import('../../src/services/api');
    vi.mocked(apiService.getProduct).mockResolvedValue({ data: mockProduct });

    render(
      <TestWrapper>
        <ProductDetail />
      </TestWrapper>
    );

    await waitFor(() => {
      expect(screen.getByText('Test Product')).toBeInTheDocument();
    });

    expect(screen.getByText('A test product')).toBeInTheDocument();
    expect(screen.getByText('$99.99')).toBeInTheDocument();
    expect(screen.getByText('SKU: TEST-001')).toBeInTheDocument();
  });

  it('displays loading state', () => {
    render(
      <TestWrapper>
        <ProductDetail />
      </TestWrapper>
    );

    expect(screen.getByRole('status')).toBeInTheDocument();
  });

  it('displays error state', async () => {
    const { apiService } = await import('../../src/services/api');
    vi.mocked(apiService.getProduct).mockRejectedValue(new Error('Product not found'));

    render(
      <TestWrapper>
        <ProductDetail />
      </TestWrapper>
    );

    await waitFor(() => {
      expect(screen.getByText('Product not found')).toBeInTheDocument();
    });
  });

  it('handles quantity changes', async () => {
    const { apiService } = await import('../../src/services/api');
    vi.mocked(apiService.getProduct).mockResolvedValue({ data: mockProduct });

    render(
      <TestWrapper>
        <ProductDetail />
      </TestWrapper>
    );

    await waitFor(() => {
      expect(screen.getByText('Test Product')).toBeInTheDocument();
    });

    const incrementButton = screen.getByText('+');
    fireEvent.click(incrementButton);

    expect(screen.getByText('2')).toBeInTheDocument();
  });

  it('handles add to cart', async () => {
    const { apiService } = await import('../../src/services/api');
    const { useCart } = await import('../../src/contexts/CartContext');
    const mockAddToCart = vi.fn();
    
    vi.mocked(useCart).mockReturnValue({
      addToCart: mockAddToCart,
      cart: { items: [], item_count: 0, total_amount: 0, subtotal: 0, tax_amount: 0, shipping_amount: 0 },
      isLoading: false,
    });
    
    vi.mocked(apiService.getProduct).mockResolvedValue({ data: mockProduct });

    render(
      <TestWrapper>
        <ProductDetail />
      </TestWrapper>
    );

    await waitFor(() => {
      expect(screen.getByText('Test Product')).toBeInTheDocument();
    });

    const addToCartButton = screen.getByText('Add to Cart');
    fireEvent.click(addToCartButton);

    expect(mockAddToCart).toHaveBeenCalledWith({
      product_id: '1',
      quantity: 1,
    });
  });
});

describe('SearchBar Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders search bar correctly', () => {
    render(
      <TestWrapper>
        <SearchBar />
      </TestWrapper>
    );

    expect(screen.getByPlaceholderText('Search products...')).toBeInTheDocument();
    expect(screen.getByText('Search')).toBeInTheDocument();
    expect(screen.getByText('Filters')).toBeInTheDocument();
  });

  it('handles search input changes', () => {
    const mockOnSearch = vi.fn();
    
    render(
      <TestWrapper>
        <SearchBar onSearch={mockOnSearch} />
      </TestWrapper>
    );

    const searchInput = screen.getByPlaceholderText('Search products...');
    fireEvent.change(searchInput, { target: { value: 'test search' } });

    expect(searchInput).toHaveValue('test search');
  });

  it('handles search form submission', () => {
    const mockOnSearch = vi.fn();
    
    render(
      <TestWrapper>
        <SearchBar onSearch={mockOnSearch} />
      </TestWrapper>
    );

    const searchInput = screen.getByPlaceholderText('Search products...');
    const searchButton = screen.getByText('Search');

    fireEvent.change(searchInput, { target: { value: 'test search' } });
    fireEvent.click(searchButton);

    expect(mockOnSearch).toHaveBeenCalledWith('test search');
  });

  it('toggles advanced filters', () => {
    render(
      <TestWrapper>
        <SearchBar showFilters={true} />
      </TestWrapper>
    );

    const filtersButton = screen.getByText('Filters');
    fireEvent.click(filtersButton);

    expect(screen.getByText('Price Range')).toBeInTheDocument();
    expect(screen.getByText('Category')).toBeInTheDocument();
    expect(screen.getByText('Status')).toBeInTheDocument();
    expect(screen.getByText('Sort By')).toBeInTheDocument();
  });

  it('handles filter changes', () => {
    const mockOnFiltersChange = vi.fn();
    
    render(
      <TestWrapper>
        <SearchBar onFiltersChange={mockOnFiltersChange} showFilters={true} />
      </TestWrapper>
    );

    const filtersButton = screen.getByText('Filters');
    fireEvent.click(filtersButton);

    const minPriceInput = screen.getByPlaceholderText('Min');
    fireEvent.change(minPriceInput, { target: { value: '10' } });

    expect(mockOnFiltersChange).toHaveBeenCalledWith(
      expect.objectContaining({
        min_price: 10,
      })
    );
  });

  it('clears all filters', () => {
    const mockOnFiltersChange = vi.fn();
    
    render(
      <TestWrapper>
        <SearchBar onFiltersChange={mockOnFiltersChange} showFilters={true} />
      </TestWrapper>
    );

    const filtersButton = screen.getByText('Filters');
    fireEvent.click(filtersButton);

    const clearButton = screen.getByText('Clear all filters');
    fireEvent.click(clearButton);

    expect(mockOnFiltersChange).toHaveBeenCalledWith({
      search: '',
      min_price: undefined,
      max_price: undefined,
      category_id: undefined,
      status: 'active',
      sort_by: 'created_at',
      sort_order: 'desc',
    });
  });
});

describe('LoginForm Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders login form correctly', () => {
    render(
      <TestWrapper>
        <LoginForm />
      </TestWrapper>
    );

    expect(screen.getByText('Sign In')).toBeInTheDocument();
    expect(screen.getByLabelText('Email Address')).toBeInTheDocument();
    expect(screen.getByLabelText('Password')).toBeInTheDocument();
    expect(screen.getByText('Sign In')).toBeInTheDocument();
    expect(screen.getByText("Don't have an account?")).toBeInTheDocument();
  });

  it('handles form input changes', () => {
    render(
      <TestWrapper>
        <LoginForm />
      </TestWrapper>
    );

    const emailInput = screen.getByLabelText('Email Address');
    const passwordInput = screen.getByLabelText('Password');

    fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });

    expect(emailInput).toHaveValue('test@example.com');
    expect(passwordInput).toHaveValue('password123');
  });

  it('handles form submission', async () => {
    const { useAuth } = await import('../../src/contexts/AuthContext');
    const mockLogin = vi.fn();
    
    vi.mocked(useAuth).mockReturnValue({
      user: null,
      isAuthenticated: false,
      login: mockLogin,
      register: vi.fn(),
    });

    render(
      <TestWrapper>
        <LoginForm />
      </TestWrapper>
    );

    const emailInput = screen.getByLabelText('Email Address');
    const passwordInput = screen.getByLabelText('Password');
    const submitButton = screen.getByText('Sign In');

    fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    fireEvent.click(submitButton);

    expect(mockLogin).toHaveBeenCalledWith('test@example.com', 'password123');
  });

  it('displays validation errors', async () => {
    const { useAuth } = await import('../../src/contexts/AuthContext');
    const mockLogin = vi.fn().mockRejectedValue(new Error('Invalid credentials'));
    
    vi.mocked(useAuth).mockReturnValue({
      user: null,
      isAuthenticated: false,
      login: mockLogin,
      register: vi.fn(),
    });

    render(
      <TestWrapper>
        <LoginForm />
      </TestWrapper>
    );

    const emailInput = screen.getByLabelText('Email Address');
    const passwordInput = screen.getByLabelText('Password');
    const submitButton = screen.getByText('Sign In');

    fireEvent.change(emailInput, { target: { value: 'test@example.com' } });
    fireEvent.change(passwordInput, { target: { value: 'wrongpassword' } });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText('Invalid credentials')).toBeInTheDocument();
    });
  });
});

describe('RegisterForm Component', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders register form correctly', () => {
    render(
      <TestWrapper>
        <RegisterForm />
      </TestWrapper>
    );

    expect(screen.getByText('Create Account')).toBeInTheDocument();
    expect(screen.getByLabelText('First Name *')).toBeInTheDocument();
    expect(screen.getByLabelText('Last Name *')).toBeInTheDocument();
    expect(screen.getByLabelText('Email Address *')).toBeInTheDocument();
    expect(screen.getByLabelText('Phone Number')).toBeInTheDocument();
    expect(screen.getByLabelText('Password *')).toBeInTheDocument();
    expect(screen.getByLabelText('Confirm Password *')).toBeInTheDocument();
    expect(screen.getByText('Create Account')).toBeInTheDocument();
  });

  it('handles form input changes', () => {
    render(
      <TestWrapper>
        <RegisterForm />
      </TestWrapper>
    );

    const firstNameInput = screen.getByLabelText('First Name *');
    const lastNameInput = screen.getByLabelText('Last Name *');
    const emailInput = screen.getByLabelText('Email Address *');

    fireEvent.change(firstNameInput, { target: { value: 'John' } });
    fireEvent.change(lastNameInput, { target: { value: 'Doe' } });
    fireEvent.change(emailInput, { target: { value: 'john@example.com' } });

    expect(firstNameInput).toHaveValue('John');
    expect(lastNameInput).toHaveValue('Doe');
    expect(emailInput).toHaveValue('john@example.com');
  });

  it('validates password confirmation', async () => {
    render(
      <TestWrapper>
        <RegisterForm />
      </TestWrapper>
    );

    const passwordInput = screen.getByLabelText('Password *');
    const confirmPasswordInput = screen.getByLabelText('Confirm Password *');
    const submitButton = screen.getByText('Create Account');

    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    fireEvent.change(confirmPasswordInput, { target: { value: 'different123' } });
    fireEvent.click(submitButton);

    await waitFor(() => {
      expect(screen.getByText('Passwords do not match')).toBeInTheDocument();
    });
  });

  it('handles form submission', async () => {
    const { useAuth } = await import('../../src/contexts/AuthContext');
    const mockRegister = vi.fn();
    
    vi.mocked(useAuth).mockReturnValue({
      user: null,
      isAuthenticated: false,
      login: vi.fn(),
      register: mockRegister,
    });

    render(
      <TestWrapper>
        <RegisterForm />
      </TestWrapper>
    );

    const firstNameInput = screen.getByLabelText('First Name *');
    const lastNameInput = screen.getByLabelText('Last Name *');
    const emailInput = screen.getByLabelText('Email Address *');
    const passwordInput = screen.getByLabelText('Password *');
    const confirmPasswordInput = screen.getByLabelText('Confirm Password *');
    const submitButton = screen.getByText('Create Account');

    fireEvent.change(firstNameInput, { target: { value: 'John' } });
    fireEvent.change(lastNameInput, { target: { value: 'Doe' } });
    fireEvent.change(emailInput, { target: { value: 'john@example.com' } });
    fireEvent.change(passwordInput, { target: { value: 'password123' } });
    fireEvent.change(confirmPasswordInput, { target: { value: 'password123' } });
    fireEvent.click(submitButton);

    expect(mockRegister).toHaveBeenCalledWith({
      first_name: 'John',
      last_name: 'Doe',
      email: 'john@example.com',
      password: 'password123',
      phone: undefined,
    });
  });
});
