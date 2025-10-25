import React from 'react';
import { SearchInput } from '../components/search/SearchInput';
import { SearchResults } from '../components/search/SearchResults';
import { SearchPagination } from '../components/search/SearchPagination';
import { useSearch, useSearchActions } from '../contexts/SearchContext';
import { searchApi, ProductResult } from '../services/searchApi';
import { useEffect } from 'react';

export const SearchPage: React.FC = () => {
  const { state } = useSearch();
  const actions = useSearchActions();

  // Perform search when query or filters change
  useEffect(() => {
    if (state.query.trim()) {
      performSearch();
    } else {
      actions.clearResults();
    }
  }, [state.query, state.filters, state.sortBy, state.page, state.pageSize]);

  const performSearch = async () => {
    try {
      actions.setLoading(true);
      
      const response = await searchApi.searchProducts({
        query: state.query,
        filters: state.filters,
        sort_by: state.sortBy,
        page: state.page,
        page_size: state.pageSize,
      });

      actions.setResults(
        response.results,
        response.pagination,
        response.total_results,
        response.response_time_ms
      );

      // Log analytics
      await searchApi.logAnalytics({
        query: state.query,
        filters: state.filters,
        result_count: response.total_results,
        response_time_ms: response.response_time_ms,
        cache_hit: false, // This would be determined by the API response
        session_id: 'session-' + Date.now(), // Simple session ID
      });

    } catch (error) {
      actions.setError(error instanceof Error ? error.message : 'Search failed');
    }
  };

  const handleSearch = (query: string) => {
    actions.setQuery(query);
  };

  const handleProductClick = (product: ProductResult) => {
    // Navigate to product detail page
    console.log('Navigate to product:', product.id);
    // This would typically use React Router
    // navigate(`/products/${product.id}`);
  };

  const handleAddToCart = (product: ProductResult) => {
    // Add to cart logic
    console.log('Add to cart:', product.id);
    // This would typically dispatch to cart context
  };

  const handlePageChange = (page: number) => {
    actions.setPage(page);
  };

  return (
    <div className="min-h-screen bg-gray-50">
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Header */}
        <div className="mb-8">
          <h1 className="text-3xl font-bold text-gray-900">Search Products</h1>
          <p className="mt-2 text-gray-600">
            Find the perfect products for your needs
          </p>
        </div>

        {/* Search Input */}
        <div className="mb-8">
          <SearchInput
            onSearch={handleSearch}
            placeholder="Search for products..."
            initialValue={state.query}
            className="max-w-2xl"
          />
        </div>

        {/* Search Results Info */}
        {state.results.length > 0 && (
          <div className="mb-6">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-700">
                  Found {state.totalResults} result{state.totalResults !== 1 ? 's' : ''}
                  {state.query && ` for "${state.query}"`}
                  {state.responseTime > 0 && ` in ${state.responseTime}ms`}
                </p>
              </div>
              
              {/* Sort Options */}
              <div className="flex items-center space-x-2">
                <label htmlFor="sort-select" className="text-sm font-medium text-gray-700">
                  Sort by:
                </label>
                <select
                  id="sort-select"
                  value={state.sortBy}
                  onChange={(e) => actions.setSortBy(e.target.value)}
                  className="border border-gray-300 rounded-md px-3 py-1 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                >
                  <option value="relevance">Relevance</option>
                  <option value="price_asc">Price: Low to High</option>
                  <option value="price_desc">Price: High to Low</option>
                  <option value="popularity">Popularity</option>
                  <option value="newest">Newest</option>
                </select>
              </div>
            </div>
          </div>
        )}

        {/* Error Message */}
        {state.error && (
          <div className="mb-6 bg-red-50 border border-red-200 rounded-md p-4">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-red-400" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M10 18a8 8 0 100-16 8 8 0 000 16zM8.707 7.293a1 1 0 00-1.414 1.414L8.586 10l-1.293 1.293a1 1 0 101.414 1.414L10 11.414l1.293 1.293a1 1 0 001.414-1.414L11.414 10l1.293-1.293a1 1 0 00-1.414-1.414L10 8.586 8.707 7.293z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <h3 className="text-sm font-medium text-red-800">
                  Search Error
                </h3>
                <div className="mt-2 text-sm text-red-700">
                  <p>{state.error}</p>
                </div>
              </div>
            </div>
          </div>
        )}

        {/* Search Results */}
        <SearchResults
          results={state.results}
          loading={state.loading}
          onProductClick={handleProductClick}
          onAddToCart={handleAddToCart}
        />

        {/* Pagination */}
        {state.pagination && state.pagination.total_pages > 1 && (
          <div className="mt-8">
            <SearchPagination
              pagination={state.pagination}
              onPageChange={handlePageChange}
            />
          </div>
        )}

        {/* Empty State */}
        {!state.loading && state.results.length === 0 && state.query && !state.error && (
          <div className="text-center py-12">
            <div className="mx-auto h-12 w-12 text-gray-400">
              <svg fill="none" viewBox="0 0 24 24" stroke="currentColor">
                <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z" />
              </svg>
            </div>
            <h3 className="mt-2 text-sm font-medium text-gray-900">No products found</h3>
            <p className="mt-1 text-sm text-gray-500">
              Try adjusting your search terms or browse our categories.
            </p>
          </div>
        )}
      </div>
    </div>
  );
};
