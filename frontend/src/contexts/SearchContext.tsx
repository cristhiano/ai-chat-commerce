import React, { createContext, useContext, useReducer, ReactNode } from 'react';
import { ProductResult, PaginationInfo } from '../../services/searchApi';

// Types
export interface SearchFilters {
  category_id?: string;
  price_min?: number;
  price_max?: number;
  availability?: string;
}

export interface SearchState {
  query: string;
  filters: SearchFilters;
  results: ProductResult[];
  pagination: PaginationInfo | null;
  loading: boolean;
  error: string | null;
  totalResults: number;
  responseTime: number;
  sortBy: string;
  page: number;
  pageSize: number;
  filterOptions: any;
}

export interface SearchContextType {
  state: SearchState;
  dispatch: React.Dispatch<SearchAction>;
}

// Action types
export type SearchAction =
  | { type: 'SET_QUERY'; payload: string }
  | { type: 'SET_FILTERS'; payload: SearchFilters }
  | { type: 'UPDATE_FILTER'; payload: { key: keyof SearchFilters; value: any } }
  | { type: 'CLEAR_FILTER'; payload: keyof SearchFilters }
  | { type: 'CLEAR_ALL_FILTERS' }
  | { type: 'SET_SORT_BY'; payload: string }
  | { type: 'SET_PAGE'; payload: number }
  | { type: 'SET_PAGE_SIZE'; payload: number }
  | { type: 'SET_LOADING'; payload: boolean }
  | { type: 'SET_ERROR'; payload: string | null }
  | { type: 'SET_RESULTS'; payload: { results: ProductResult[]; pagination: PaginationInfo; totalResults: number; responseTime: number } }
  | { type: 'SET_FILTER_OPTIONS'; payload: any }
  | { type: 'CLEAR_RESULTS' }
  | { type: 'RESET_SEARCH' };

// Initial state
const initialState: SearchState = {
  query: '',
  filters: {},
  results: [],
  pagination: null,
  loading: false,
  error: null,
  totalResults: 0,
  responseTime: 0,
  sortBy: 'relevance',
  page: 1,
  pageSize: 20,
  filterOptions: null,
};

// Reducer
function searchReducer(state: SearchState, action: SearchAction): SearchState {
  switch (action.type) {
    case 'SET_QUERY':
      return {
        ...state,
        query: action.payload,
        page: 1, // Reset to first page when query changes
      };

    case 'SET_FILTERS':
      return {
        ...state,
        filters: action.payload,
        page: 1, // Reset to first page when filters change
      };

    case 'UPDATE_FILTER':
      return {
        ...state,
        filters: {
          ...state.filters,
          [action.payload.key]: action.payload.value,
        },
        page: 1, // Reset to first page when filters change
      };

    case 'CLEAR_FILTER':
      const newFilters = { ...state.filters };
      delete newFilters[action.payload];
      return {
        ...state,
        filters: newFilters,
        page: 1, // Reset to first page when filters change
      };

    case 'CLEAR_ALL_FILTERS':
      return {
        ...state,
        filters: {},
        page: 1, // Reset to first page when filters change
      };

    case 'SET_SORT_BY':
      return {
        ...state,
        sortBy: action.payload,
        page: 1, // Reset to first page when sort changes
      };

    case 'SET_PAGE':
      return {
        ...state,
        page: action.payload,
      };

    case 'SET_PAGE_SIZE':
      return {
        ...state,
        pageSize: action.payload,
        page: 1, // Reset to first page when page size changes
      };

    case 'SET_LOADING':
      return {
        ...state,
        loading: action.payload,
        error: action.payload ? null : state.error, // Clear error when starting to load
      };

    case 'SET_ERROR':
      return {
        ...state,
        error: action.payload,
        loading: false,
      };

    case 'SET_FILTER_OPTIONS':
      return {
        ...state,
        filterOptions: action.payload,
      };

    case 'CLEAR_RESULTS':
      return {
        ...state,
        results: [],
        pagination: null,
        totalResults: 0,
        responseTime: 0,
        error: null,
      };

    case 'RESET_SEARCH':
      return {
        ...initialState,
        pageSize: state.pageSize, // Keep current page size
        filterOptions: state.filterOptions, // Keep filter options
      };

    default:
      return state;
  }
}

// Context
const SearchContext = createContext<SearchContextType | undefined>(undefined);

// Provider component
interface SearchProviderProps {
  children: ReactNode;
}

export const SearchProvider: React.FC<SearchProviderProps> = ({ children }) => {
  const [state, dispatch] = useReducer(searchReducer, initialState);

  return (
    <SearchContext.Provider value={{ state, dispatch }}>
      {children}
    </SearchContext.Provider>
  );
};

// Hook to use search context
export const useSearch = (): SearchContextType => {
  const context = useContext(SearchContext);
  if (context === undefined) {
    throw new Error('useSearch must be used within a SearchProvider');
  }
  return context;
};

// Helper hooks for specific actions
export const useSearchActions = () => {
  const { dispatch } = useSearch();

  return {
    setQuery: (query: string) => dispatch({ type: 'SET_QUERY', payload: query }),
    setFilters: (filters: SearchFilters) => dispatch({ type: 'SET_FILTERS', payload: filters }),
    updateFilter: (key: keyof SearchFilters, value: any) => 
      dispatch({ type: 'UPDATE_FILTER', payload: { key, value } }),
    clearFilter: (key: keyof SearchFilters) => 
      dispatch({ type: 'CLEAR_FILTER', payload: key }),
    clearAllFilters: () => dispatch({ type: 'CLEAR_ALL_FILTERS' }),
    setSortBy: (sortBy: string) => dispatch({ type: 'SET_SORT_BY', payload: sortBy }),
    setPage: (page: number) => dispatch({ type: 'SET_PAGE', payload: page }),
    setPageSize: (pageSize: number) => dispatch({ type: 'SET_PAGE_SIZE', payload: pageSize }),
    setLoading: (loading: boolean) => dispatch({ type: 'SET_LOADING', payload: loading }),
    setError: (error: string | null) => dispatch({ type: 'SET_ERROR', payload: error }),
    setResults: (results: ProductResult[], pagination: PaginationInfo, totalResults: number, responseTime: number) =>
      dispatch({ type: 'SET_RESULTS', payload: { results, pagination, totalResults, responseTime } }),
    setFilterOptions: (options: any) => dispatch({ type: 'SET_FILTER_OPTIONS', payload: options }),
    clearResults: () => dispatch({ type: 'CLEAR_RESULTS' }),
    resetSearch: () => dispatch({ type: 'RESET_SEARCH' }),
  };
};

// Hook to get search state
export const useSearchState = () => {
  const { state } = useSearch();
  return state;
};
