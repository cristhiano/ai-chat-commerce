import React, { useState } from 'react';
import { ChevronDown, X, Filter } from 'lucide-react';

interface FilterOption {
  id: string;
  name: string;
  product_count?: number;
}

interface PriceRange {
  min: number;
  max: number;
  label: string;
}

interface SearchFiltersProps {
  filters: {
    category_id?: string;
    price_min?: number;
    price_max?: number;
    availability?: string;
    brand?: string;
  };
  onFiltersChange: (filters: any) => void;
  filterOptions?: {
    categories: FilterOption[];
    price_ranges: PriceRange[];
    availability_options: string[];
    brands: FilterOption[];
  };
  className?: string;
}

export const SearchFilters: React.FC<SearchFiltersProps> = ({
  filters,
  onFiltersChange,
  filterOptions,
  className = ""
}) => {
  const [isExpanded, setIsExpanded] = useState(false);

  const handleCategoryChange = (categoryId: string) => {
    onFiltersChange({
      ...filters,
      category_id: categoryId || undefined,
    });
  };

  const handlePriceRangeChange = (min: number, max: number) => {
    onFiltersChange({
      ...filters,
      price_min: min > 0 ? min : undefined,
      price_max: max > 0 ? max : undefined,
    });
  };

  const handleAvailabilityChange = (availability: string) => {
    onFiltersChange({
      ...filters,
      availability: availability === 'all' ? undefined : availability,
    });
  };

  const handleBrandChange = (brand: string) => {
    onFiltersChange({
      ...filters,
      brand: brand || undefined,
    });
  };

  const clearAllFilters = () => {
    onFiltersChange({});
  };

  const hasActiveFilters = Object.values(filters).some(value => 
    value !== undefined && value !== null && value !== ''
  );

  const getActiveFilterCount = () => {
    return Object.values(filters).filter(value => 
      value !== undefined && value !== null && value !== ''
    ).length;
  };

  return (
    <div className={`bg-white border border-gray-200 rounded-lg ${className}`}>
      {/* Filter Header */}
      <div 
        className="flex items-center justify-between p-4 cursor-pointer hover:bg-gray-50"
        onClick={() => setIsExpanded(!isExpanded)}
      >
        <div className="flex items-center space-x-2">
          <Filter className="h-5 w-5 text-gray-500" />
          <span className="font-medium text-gray-900">Filters</span>
          {hasActiveFilters && (
            <span className="bg-blue-100 text-blue-800 text-xs font-medium px-2 py-1 rounded-full">
              {getActiveFilterCount()}
            </span>
          )}
        </div>
        
        <div className="flex items-center space-x-2">
          {hasActiveFilters && (
            <button
              onClick={(e) => {
                e.stopPropagation();
                clearAllFilters();
              }}
              className="text-sm text-gray-500 hover:text-gray-700"
            >
              Clear all
            </button>
          )}
          <ChevronDown 
            className={`h-4 w-4 text-gray-500 transition-transform duration-200 ${
              isExpanded ? 'rotate-180' : ''
            }`} 
          />
        </div>
      </div>

      {/* Filter Content */}
      {isExpanded && (
        <div className="border-t border-gray-200 p-4 space-y-6">
          {/* Category Filter */}
          {filterOptions?.categories && filterOptions.categories.length > 0 && (
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Category
              </label>
              <select
                value={filters.category_id || ''}
                onChange={(e) => handleCategoryChange(e.target.value)}
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              >
                <option value="">All Categories</option>
                {filterOptions.categories.map((category) => (
                  <option key={category.id} value={category.id}>
                    {category.name} {category.product_count && `(${category.product_count})`}
                  </option>
                ))}
              </select>
            </div>
          )}

          {/* Price Range Filter */}
          {filterOptions?.price_ranges && filterOptions.price_ranges.length > 0 && (
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Price Range
              </label>
              <div className="space-y-2">
                {filterOptions.price_ranges.map((range, index) => (
                  <label key={index} className="flex items-center">
                    <input
                      type="radio"
                      name="price-range"
                      value={index}
                      checked={
                        (filters.price_min === range.min && filters.price_max === range.max) ||
                        (!filters.price_min && !filters.price_max && range.min === 0 && range.max === 0)
                      }
                      onChange={() => handlePriceRangeChange(range.min, range.max)}
                      className="mr-2 text-blue-600 focus:ring-blue-500"
                    />
                    <span className="text-sm text-gray-700">{range.label}</span>
                  </label>
                ))}
              </div>
            </div>
          )}

          {/* Availability Filter */}
          {filterOptions?.availability_options && filterOptions.availability_options.length > 0 && (
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Availability
              </label>
              <div className="space-y-2">
                {filterOptions.availability_options.map((option) => (
                  <label key={option} className="flex items-center">
                    <input
                      type="radio"
                      name="availability"
                      value={option}
                      checked={filters.availability === option || (!filters.availability && option === 'all')}
                      onChange={() => handleAvailabilityChange(option)}
                      className="mr-2 text-blue-600 focus:ring-blue-500"
                    />
                    <span className="text-sm text-gray-700 capitalize">
                      {option.replace('_', ' ')}
                    </span>
                  </label>
                ))}
              </div>
            </div>
          )}

          {/* Brand Filter */}
          {filterOptions?.brands && filterOptions.brands.length > 0 && (
            <div>
              <label className="block text-sm font-medium text-gray-700 mb-2">
                Brand
              </label>
              <select
                value={filters.brand || ''}
                onChange={(e) => handleBrandChange(e.target.value)}
                className="w-full border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
              >
                <option value="">All Brands</option>
                {filterOptions.brands.map((brand) => (
                  <option key={brand.id} value={brand.id}>
                    {brand.name} {brand.product_count && `(${brand.product_count})`}
                  </option>
                ))}
              </select>
            </div>
          )}

          {/* Custom Price Range */}
          <div>
            <label className="block text-sm font-medium text-gray-700 mb-2">
              Custom Price Range
            </label>
            <div className="flex items-center space-x-2">
              <input
                type="number"
                placeholder="Min"
                value={filters.price_min || ''}
                onChange={(e) => {
                  const value = e.target.value ? parseFloat(e.target.value) : undefined;
                  onFiltersChange({
                    ...filters,
                    price_min: value,
                  });
                }}
                className="flex-1 border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                min="0"
                step="0.01"
              />
              <span className="text-gray-500">to</span>
              <input
                type="number"
                placeholder="Max"
                value={filters.price_max || ''}
                onChange={(e) => {
                  const value = e.target.value ? parseFloat(e.target.value) : undefined;
                  onFiltersChange({
                    ...filters,
                    price_max: value,
                  });
                }}
                className="flex-1 border border-gray-300 rounded-md px-3 py-2 text-sm focus:outline-none focus:ring-2 focus:ring-blue-500 focus:border-blue-500"
                min="0"
                step="0.01"
              />
            </div>
          </div>
        </div>
      )}
    </div>
  );
};
