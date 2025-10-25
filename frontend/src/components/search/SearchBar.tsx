import React, { useState, useEffect } from 'react';
import { useNavigate, useLocation } from 'react-router-dom';
import type { ProductFilters } from '../../types';

interface SearchBarProps {
  onSearch?: (query: string) => void;
  onFiltersChange?: (filters: ProductFilters) => void;
  placeholder?: string;
  showFilters?: boolean;
  className?: string;
}

const SearchBar: React.FC<SearchBarProps> = ({
  onSearch,
  onFiltersChange,
  placeholder = "Search products...",
  showFilters = true,
  className = ""
}) => {
  const navigate = useNavigate();
  const location = useLocation();
  
  const [query, setQuery] = useState('');
  const [showAdvancedFilters, setShowAdvancedFilters] = useState(false);
  const [filters, setFilters] = useState<ProductFilters>({
    search: '',
    min_price: undefined,
    max_price: undefined,
    category_id: undefined,
    status: 'active',
    sort_by: 'created_at',
    sort_order: 'desc'
  });

  // Initialize filters from URL params
  useEffect(() => {
    const urlParams = new URLSearchParams(location.search);
    const urlFilters: ProductFilters = {
      search: urlParams.get('q') || '',
      min_price: urlParams.get('min_price') ? parseFloat(urlParams.get('min_price')!) : undefined,
      max_price: urlParams.get('max_price') ? parseFloat(urlParams.get('max_price')!) : undefined,
      category_id: urlParams.get('category_id') || undefined,
      status: urlParams.get('status') || 'active',
      sort_by: urlParams.get('sort_by') || 'created_at',
      sort_order: urlParams.get('sort_order') || 'desc'
    };
    
    setFilters(urlFilters);
    setQuery(urlFilters.search || '');
  }, [location.search]);

  // Handle search input change
  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setQuery(value);
    
    // Update filters
    const newFilters = { ...filters, search: value };
    setFilters(newFilters);
    onFiltersChange?.(newFilters);
  };

  // Handle search submission
  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    
    // Update URL with search params
    const urlParams = new URLSearchParams();
    if (query) urlParams.set('q', query);
    if (filters.min_price) urlParams.set('min_price', filters.min_price.toString());
    if (filters.max_price) urlParams.set('max_price', filters.max_price.toString());
    if (filters.category_id) urlParams.set('category_id', filters.category_id);
    if (filters.status) urlParams.set('status', filters.status);
    if (filters.sort_by) urlParams.set('sort_by', filters.sort_by);
    if (filters.sort_order) urlParams.set('sort_order', filters.sort_order);

    const searchUrl = `/search?${urlParams.toString()}`;
    navigate(searchUrl);
    
    onSearch?.(query);
  };

  // Handle filter changes
  const handleFilterChange = (key: keyof ProductFilters, value: any) => {
    const newFilters = { ...filters, [key]: value };
    setFilters(newFilters);
    onFiltersChange?.(newFilters);
  };

  // Clear all filters
  const clearFilters = () => {
    const clearedFilters: ProductFilters = {
      search: '',
      min_price: undefined,
      max_price: undefined,
      category_id: undefined,
      status: 'active',
      sort_by: 'created_at',
      sort_order: 'desc'
    };
    setFilters(clearedFilters);
    setQuery('');
    onFiltersChange?.(clearedFilters);
    navigate('/search');
  };

  return (
    <div className={`bg-white rounded-lg shadow-sm border p-4 ${className}`}>
      {/* Main Search Bar */}
      <form onSubmit={handleSearch} className="flex space-x-2">
        <div className="flex-1 relative">
          <input
            type="text"
            value={query}
            onChange={handleInputChange}
            placeholder={placeholder}
            className="w-full px-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-transparent"
          />
          <div className="absolute inset-y-0 right-0 flex items-center pr-3">
            <svg
              className="h-5 w-5 text-gray-400"
              fill="none"
              stroke="currentColor"
              viewBox="0 0 24 24"
            >
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M21 21l-6-6m2-5a7 7 0 11-14 0 7 7 0 0114 0z"
              />
            </svg>
          </div>
        </div>
        
        <button
          type="submit"
          className="bg-blue-600 text-white px-6 py-2 rounded-lg hover:bg-blue-700 transition-colors"
        >
          Search
        </button>
        
        {showFilters && (
          <button
            type="button"
            onClick={() => setShowAdvancedFilters(!showAdvancedFilters)}
            className="bg-gray-100 text-gray-700 px-4 py-2 rounded-lg hover:bg-gray-200 transition-colors"
          >
            Filters
          </button>
        )}
      </form>

      {/* Advanced Filters */}
      {showFilters && showAdvancedFilters && (
        <div className="mt-4 pt-4 border-t border-gray-200">
          <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-4">
            {/* Price Range */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Price Range
              </label>
              <div className="flex space-x-2">
                <input
                  type="number"
                  placeholder="Min"
                  value={filters.min_price || ''}
                  onChange={(e) => handleFilterChange('min_price', e.target.value ? parseFloat(e.target.value) : undefined)}
                  className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
                <input
                  type="number"
                  placeholder="Max"
                  value={filters.max_price || ''}
                  onChange={(e) => handleFilterChange('max_price', e.target.value ? parseFloat(e.target.value) : undefined)}
                  className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
            </div>

            {/* Category */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Category
              </label>
              <select
                value={filters.category_id || ''}
                onChange={(e) => handleFilterChange('category_id', e.target.value || undefined)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="">All Categories</option>
                {/* Categories would be loaded from API */}
                <option value="electronics">Electronics</option>
                <option value="clothing">Clothing</option>
                <option value="books">Books</option>
                <option value="home">Home & Garden</option>
              </select>
            </div>

            {/* Status */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Status
              </label>
              <select
                value={filters.status || 'active'}
                onChange={(e) => handleFilterChange('status', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="active">Active</option>
                <option value="inactive">Inactive</option>
                <option value="discontinued">Discontinued</option>
              </select>
            </div>

            {/* Sort */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Sort By
              </label>
              <select
                value={filters.sort_by || 'created_at'}
                onChange={(e) => handleFilterChange('sort_by', e.target.value)}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="created_at">Newest</option>
                <option value="name">Name</option>
                <option value="price">Price</option>
                <option value="popularity">Popularity</option>
              </select>
            </div>
          </div>

          {/* Filter Actions */}
          <div className="mt-4 flex justify-between items-center">
            <button
              onClick={clearFilters}
              className="text-gray-500 hover:text-gray-700 text-sm"
            >
              Clear all filters
            </button>
            
            <div className="flex space-x-2">
              <button
                onClick={() => setShowAdvancedFilters(false)}
                className="px-4 py-2 text-gray-600 hover:text-gray-800"
              >
                Close
              </button>
              <button
                onClick={handleSearch}
                className="bg-blue-600 text-white px-4 py-2 rounded-md hover:bg-blue-700"
              >
                Apply Filters
              </button>
            </div>
          </div>
        </div>
      )}

      {/* Active Filters Display */}
      {(filters.min_price || filters.max_price || filters.category_id || filters.status !== 'active') && (
        <div className="mt-3 flex flex-wrap gap-2">
          <span className="text-sm text-gray-600">Active filters:</span>
          
          {filters.min_price && (
            <span className="inline-flex items-center px-2 py-1 bg-blue-100 text-blue-800 text-xs rounded-full">
              Min: ${filters.min_price}
              <button
                onClick={() => handleFilterChange('min_price', undefined)}
                className="ml-1 text-blue-600 hover:text-blue-800"
              >
                ×
              </button>
            </span>
          )}
          
          {filters.max_price && (
            <span className="inline-flex items-center px-2 py-1 bg-blue-100 text-blue-800 text-xs rounded-full">
              Max: ${filters.max_price}
              <button
                onClick={() => handleFilterChange('max_price', undefined)}
                className="ml-1 text-blue-600 hover:text-blue-800"
              >
                ×
              </button>
            </span>
          )}
          
          {filters.category_id && (
            <span className="inline-flex items-center px-2 py-1 bg-blue-100 text-blue-800 text-xs rounded-full">
              Category: {filters.category_id}
              <button
                onClick={() => handleFilterChange('category_id', undefined)}
                className="ml-1 text-blue-600 hover:text-blue-800"
              >
                ×
              </button>
            </span>
          )}
          
          {filters.status !== 'active' && (
            <span className="inline-flex items-center px-2 py-1 bg-blue-100 text-blue-800 text-xs rounded-full">
              Status: {filters.status}
              <button
                onClick={() => handleFilterChange('status', 'active')}
                className="ml-1 text-blue-600 hover:text-blue-800"
              >
                ×
              </button>
            </span>
          )}
        </div>
      )}
    </div>
  );
};

export default SearchBar;
