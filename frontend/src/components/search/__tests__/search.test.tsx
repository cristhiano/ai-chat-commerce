import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, fireEvent, waitFor } from '@testing-library/react';
import { BrowserRouter } from 'react-router-dom';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import SearchInput from '../SearchInput';
import SearchResults from '../SearchResults';
import SearchPagination from '../SearchPagination';
import Autocomplete from '../Autocomplete';
import SearchFilters from '../SearchFilters';

// Mock API service
vi.mock('../../../services/searchApi', () => ({
  searchProducts: vi.fn(),
  getSuggestions: vi.fn(),
  getFilterOptions: vi.fn(),
}));

// Mock SearchContext
const mockSearchContext = {
  query: '',
  results: [],
  filters: {},
  isLoading: false,
  error: null,
  setQuery: vi.fn(),
  setFilters: vi.fn(),
  performSearch: vi.fn(),
  clearResults: vi.fn(),
};

vi.mock('../../../contexts/SearchContext', () => ({
  useSearch: () => mockSearchContext,
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

describe('SearchInput', () => {
  it('renders search input field', () => {
    render(
      <TestWrapper>
        <SearchInput />
      </TestWrapper>
    );

    const input = screen.getByPlaceholderText(/search/i);
    expect(input).toBeInTheDocument();
  });

  it('calls setQuery when user types', () => {
    const setQueryMock = vi.fn();
    mockSearchContext.setQuery = setQueryMock;

    render(
      <TestWrapper>
        <SearchInput />
      </TestWrapper>
    );

    const input = screen.getByPlaceholderText(/search/i);
    fireEvent.change(input, { target: { value: 'laptop' } });

    expect(setQueryMock).toHaveBeenCalledWith('laptop');
  });

  it('handles search submission', async () => {
    const performSearchMock = vi.fn();
    mockSearchContext.performSearch = performSearchMock;

    render(
      <TestWrapper>
        <SearchInput />
      </TestWrapper>
    );

    const input = screen.getByPlaceholderText(/search/i);
    const form = input.closest('form');

    fireEvent.submit(form!);

    await waitFor(() => {
      expect(performSearchMock).toHaveBeenCalled();
    });
  });
});

describe('SearchResults', () => {
  const mockResults = [
    {
      id: '1',
      name: 'Gaming Laptop',
      description: 'High-performance gaming laptop',
      price: 1299.99,
      category: 'Electronics',
      sku: 'LAP001',
      availability: 'in_stock',
      relevance_score: 0.95,
    },
    {
      id: '2',
      name: 'Wireless Mouse',
      description: 'Ergonomic wireless mouse',
      price: 29.99,
      category: 'Electronics',
      sku: 'MOU001',
      availability: 'in_stock',
      relevance_score: 0.80,
    },
  ];

  beforeEach(() => {
    mockSearchContext.results = mockResults;
    mockSearchContext.isLoading = false;
  });

  it('renders search results', () => {
    render(
      <TestWrapper>
        <SearchResults />
      </TestWrapper>
    );

    expect(screen.getByText('Gaming Laptop')).toBeInTheDocument();
    expect(screen.getByText('Wireless Mouse')).toBeInTheDocument();
  });

  it('displays loading state', () => {
    mockSearchContext.isLoading = true;

    render(
      <TestWrapper>
        <SearchResults />
      </TestWrapper>
    );

    expect(screen.getByText(/loading/i)).toBeInTheDocument();
  });

  it('displays empty state when no results', () => {
    mockSearchContext.results = [];

    render(
      <TestWrapper>
        <SearchResults />
      </TestWrapper>
    );

    expect(screen.getByText(/no results found/i)).toBeInTheDocument();
  });

  it('displays error message', () => {
    mockSearchContext.error = 'Search failed';

    render(
      <TestWrapper>
        <SearchResults />
      </TestWrapper>
    );

    expect(screen.getByText(/search failed/i)).toBeInTheDocument();
  });
});

describe('SearchPagination', () => {
  const mockPagination = {
    current_page: 1,
    page_size: 20,
    total_pages: 5,
    total_results: 100,
    has_next: true,
    has_previous: false,
  };

  beforeEach(() => {
    mockSearchContext.pagination = mockPagination;
  });

  it('renders pagination controls', () => {
    render(
      <TestWrapper>
        <SearchPagination />
      </TestWrapper>
    );

    expect(screen.getByText(/page/i)).toBeInTheDocument();
  });

  it('displays current page and total pages', () => {
    render(
      <TestWrapper>
        <SearchPagination />
      </TestWrapper>
    );

    expect(screen.getByText(/1 of 5/i)).toBeInTheDocument();
  });

  it('handles next page click', () => {
    const performSearchMock = vi.fn();
    mockSearchContext.performSearch = performSearchMock;

    render(
      <TestWrapper>
        <SearchPagination />
      </TestWrapper>
    );

    const nextButton = screen.getByRole('button', { name: /next/i });
    fireEvent.click(nextButton);

    expect(performSearchMock).toHaveBeenCalled();
  });

  it('handles previous page click', () => {
    const performSearchMock = vi.fn();
    mockSearchContext.performSearch = performSearchMock;
    mockSearchContext.pagination.has_previous = true;
    mockSearchContext.pagination.current_page = 2;

    render(
      <TestWrapper>
        <SearchPagination />
      </TestWrapper>
    );

    const prevButton = screen.getByRole('button', { name: /previous/i });
    fireEvent.click(prevButton);

    expect(performSearchMock).toHaveBeenCalled();
  });
});

describe('Autocomplete', () => {
  beforeEach(() => {
    vi.mock('../../../services/searchApi', () => ({
      getSuggestions: vi.fn().mockResolvedValue(['laptop', 'laptop bag', 'laptop stand']),
    }));
  });

  it('renders autocomplete suggestions', async () => {
    render(
      <TestWrapper>
        <Autocomplete query="lap" />
      </TestWrapper>
    );

    await waitFor(() => {
      expect(screen.getByText(/laptop/i)).toBeInTheDocument();
    });
  });

  it('handles suggestion click', () => {
    const setQueryMock = vi.fn();
    mockSearchContext.setQuery = setQueryMock;

    render(
      <TestWrapper>
        <Autocomplete query="lap" />
      </TestWrapper>
    );

    const suggestion = screen.getByText(/laptop bag/i);
    fireEvent.click(suggestion);

    expect(setQueryMock).toHaveBeenCalled();
  });
});

describe('SearchFilters', () => {
  it('renders filter options', () => {
    render(
      <TestWrapper>
        <SearchFilters />
      </TestWrapper>
    );

    expect(screen.getByText(/price/i)).toBeInTheDocument();
    expect(screen.getByText(/category/i)).toBeInTheDocument();
  });

  it('handles filter change', () => {
    const setFiltersMock = vi.fn();
    mockSearchContext.setFilters = setFiltersMock;

    render(
      <TestWrapper>
        <SearchFilters />
      </TestWrapper>
    );

    const priceInput = screen.getByPlaceholderText(/min price/i);
    fireEvent.change(priceInput, { target: { value: '100' } });

    expect(setFiltersMock).toHaveBeenCalled();
  });
});

