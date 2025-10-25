import React, { useState, useEffect } from 'react';
import { Link } from 'react-router-dom';
import { useQuery } from '@tanstack/react-query';
import { apiService } from '../../services/api';
import type { Product, ProductFilters } from '../../types';
import { formatCurrency } from '../../utils';

interface ProductListProps {
  filters?: ProductFilters;
  showFilters?: boolean;
  onFiltersChange?: (filters: ProductFilters) => void;
}

const ProductList: React.FC<ProductListProps> = ({ 
  filters = {}, 
  showFilters = true, 
  onFiltersChange 
}) => {
  const [currentFilters, setCurrentFilters] = useState<ProductFilters>(filters);
  const [currentPage, setCurrentPage] = useState(1);

  // Fetch products with current filters
  const { data: productsData, isLoading, error } = useQuery({
    queryKey: ['products', currentFilters, currentPage],
    queryFn: () => apiService.getProducts({
      ...currentFilters,
      page: currentPage,
      limit: 12
    }),
    staleTime: 5 * 60 * 1000, // 5 minutes
  });

  const products = productsData?.data?.products || [];
  const pagination = productsData?.data ? {
    total: productsData.data.total,
    page: productsData.data.page,
    limit: productsData.data.limit,
    totalPages: productsData.data.total_pages,
    hasNext: productsData.data.has_next,
    hasPrevious: productsData.data.has_previous,
  } : null;

  // Handle filter changes
  const handleFilterChange = (newFilters: Partial<ProductFilters>) => {
    const updatedFilters = { ...currentFilters, ...newFilters };
    setCurrentFilters(updatedFilters);
    setCurrentPage(1); // Reset to first page when filters change
    onFiltersChange?.(updatedFilters);
  };

  // Handle pagination
  const handlePageChange = (page: number) => {
    setCurrentPage(page);
    window.scrollTo({ top: 0, behavior: 'smooth' });
  };

  // Loading state
  if (isLoading) {
    return (
      <div className="flex justify-center items-center py-12">
        <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-blue-600"></div>
      </div>
    );
  }

  // Error state
  if (error) {
    return (
      <div className="text-center py-12">
        <div className="text-red-600 text-lg mb-4">Failed to load products</div>
        <button 
          onClick={() => window.location.reload()}
          className="bg-blue-600 text-white px-4 py-2 rounded hover:bg-blue-700"
        >
          Try Again
        </button>
      </div>
    );
  }

  // Empty state
  if (products.length === 0) {
    return (
      <div className="text-center py-12">
        <div className="text-gray-600 text-lg mb-4">No products found</div>
        <p className="text-gray-500">Try adjusting your search or filters</p>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {/* Filters */}
      {showFilters && (
        <div className="bg-white p-4 rounded-lg shadow-sm border">
          <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
            {/* Search */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Search
              </label>
              <input
                type="text"
                placeholder="Search products..."
                value={currentFilters.search || ''}
                onChange={(e) => handleFilterChange({ search: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              />
            </div>

            {/* Price Range */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Price Range
              </label>
              <div className="flex space-x-2">
                <input
                  type="number"
                  placeholder="Min"
                  value={currentFilters.min_price || ''}
                  onChange={(e) => handleFilterChange({ 
                    min_price: e.target.value ? parseFloat(e.target.value) : undefined 
                  })}
                  className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
                <input
                  type="number"
                  placeholder="Max"
                  value={currentFilters.max_price || ''}
                  onChange={(e) => handleFilterChange({ 
                    max_price: e.target.value ? parseFloat(e.target.value) : undefined 
                  })}
                  className="flex-1 px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
                />
              </div>
            </div>

            {/* Sort */}
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-1">
                Sort By
              </label>
              <select
                value={currentFilters.sort_by || 'created_at'}
                onChange={(e) => handleFilterChange({ sort_by: e.target.value })}
                className="w-full px-3 py-2 border border-gray-300 rounded-md focus:outline-none focus:ring-2 focus:ring-blue-500"
              >
                <option value="created_at">Newest</option>
                <option value="name">Name</option>
                <option value="price">Price</option>
                <option value="popularity">Popularity</option>
              </select>
            </div>
          </div>
        </div>
      )}

      {/* Products Grid */}
      <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
        {products.map((product) => (
          <ProductCard key={product.id} product={product} />
        ))}
      </div>

      {/* Pagination */}
      {pagination && pagination.totalPages > 1 && (
        <div className="flex justify-center items-center space-x-2 py-8">
          <button
            onClick={() => handlePageChange(currentPage - 1)}
            disabled={!pagination.hasPrevious}
            className="px-4 py-2 border border-gray-300 rounded-md disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50"
          >
            Previous
          </button>
          
          <div className="flex space-x-1">
            {Array.from({ length: Math.min(5, pagination.totalPages) }, (_, i) => {
              const page = i + 1;
              return (
                <button
                  key={page}
                  onClick={() => handlePageChange(page)}
                  className={`px-3 py-2 rounded-md ${
                    page === currentPage
                      ? 'bg-blue-600 text-white'
                      : 'border border-gray-300 hover:bg-gray-50'
                  }`}
                >
                  {page}
                </button>
              );
            })}
          </div>

          <button
            onClick={() => handlePageChange(currentPage + 1)}
            disabled={!pagination.hasNext}
            className="px-4 py-2 border border-gray-300 rounded-md disabled:opacity-50 disabled:cursor-not-allowed hover:bg-gray-50"
          >
            Next
          </button>
        </div>
      )}
    </div>
  );
};

// Product Card Component
interface ProductCardProps {
  product: Product;
}

const ProductCard: React.FC<ProductCardProps> = ({ product }) => {
  return (
    <div className="bg-white rounded-lg shadow-sm border hover:shadow-md transition-shadow">
      <Link to={`/products/${product.id}`} className="block">
        {/* Product Image */}
        <div className="aspect-w-16 aspect-h-12 bg-gray-200 rounded-t-lg overflow-hidden">
          <img
            src={product.images?.[0]?.url || '/placeholder-product.jpg'}
            alt={product.name}
            className="w-full h-48 object-cover"
          />
        </div>

        {/* Product Info */}
        <div className="p-4">
          <h3 className="font-semibold text-lg text-gray-900 mb-2 line-clamp-2">
            {product.name}
          </h3>
          
          <p className="text-gray-600 text-sm mb-3 line-clamp-2">
            {product.description}
          </p>

          <div className="flex items-center justify-between">
            <span className="text-2xl font-bold text-blue-600">
              {formatCurrency(product.price)}
            </span>
            
            <span className="text-sm text-gray-500">
              SKU: {product.sku}
            </span>
          </div>

          {/* Status Badge */}
          <div className="mt-3">
            <span className={`inline-block px-2 py-1 text-xs rounded-full ${
              product.status === 'active' 
                ? 'bg-green-100 text-green-800' 
                : 'bg-gray-100 text-gray-800'
            }`}>
              {product.status}
            </span>
          </div>
        </div>
      </Link>
    </div>
  );
};

export default ProductList;
