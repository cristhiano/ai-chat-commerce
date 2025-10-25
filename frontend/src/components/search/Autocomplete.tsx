import React, { useState, useEffect, useRef } from 'react';
import { Search, X, ChevronDown } from 'lucide-react';
import { useDebounce } from '../../hooks/useDebounce';
import { searchApi } from '../../services/searchApi';

interface AutocompleteProps {
  onSuggestionSelect: (suggestion: string) => void;
  onQueryChange: (query: string) => void;
  placeholder?: string;
  className?: string;
  disabled?: boolean;
  initialValue?: string;
}

export const Autocomplete: React.FC<AutocompleteProps> = ({
  onSuggestionSelect,
  onQueryChange,
  placeholder = "Search products...",
  className = "",
  disabled = false,
  initialValue = ""
}) => {
  const [query, setQuery] = useState(initialValue);
  const [suggestions, setSuggestions] = useState<string[]>([]);
  const [isOpen, setIsOpen] = useState(false);
  const [isLoading, setIsLoading] = useState(false);
  const [selectedIndex, setSelectedIndex] = useState(-1);
  
  const inputRef = useRef<HTMLInputElement>(null);
  const suggestionsRef = useRef<HTMLDivElement>(null);
  
  // Debounce the query to avoid excessive API calls
  const debouncedQuery = useDebounce(query, 300);

  // Fetch suggestions when query changes
  useEffect(() => {
    if (debouncedQuery.length >= 2) {
      fetchSuggestions(debouncedQuery);
    } else {
      setSuggestions([]);
      setIsOpen(false);
    }
  }, [debouncedQuery]);

  // Notify parent of query changes
  useEffect(() => {
    onQueryChange(query);
  }, [query, onQueryChange]);

  const fetchSuggestions = async (searchQuery: string) => {
    try {
      setIsLoading(true);
      const suggestions = await searchApi.getSuggestions(searchQuery, 10);
      setSuggestions(suggestions);
      setIsOpen(suggestions.length > 0);
      setSelectedIndex(-1);
    } catch (error) {
      console.error('Failed to fetch suggestions:', error);
      setSuggestions([]);
      setIsOpen(false);
    } finally {
      setIsLoading(false);
    }
  };

  const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    const value = e.target.value;
    setQuery(value);
    setIsOpen(value.length >= 2);
  };

  const handleInputFocus = () => {
    if (suggestions.length > 0) {
      setIsOpen(true);
    }
  };

  const handleInputBlur = () => {
    // Delay closing to allow clicking on suggestions
    setTimeout(() => {
      setIsOpen(false);
      setSelectedIndex(-1);
    }, 200);
  };

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (!isOpen || suggestions.length === 0) {
      if (e.key === 'Enter') {
        onSuggestionSelect(query);
        setIsOpen(false);
      }
      return;
    }

    switch (e.key) {
      case 'ArrowDown':
        e.preventDefault();
        setSelectedIndex(prev => 
          prev < suggestions.length - 1 ? prev + 1 : prev
        );
        break;
      case 'ArrowUp':
        e.preventDefault();
        setSelectedIndex(prev => prev > 0 ? prev - 1 : -1);
        break;
      case 'Enter':
        e.preventDefault();
        if (selectedIndex >= 0) {
          handleSuggestionSelect(suggestions[selectedIndex]);
        } else {
          onSuggestionSelect(query);
        }
        setIsOpen(false);
        break;
      case 'Escape':
        setIsOpen(false);
        setSelectedIndex(-1);
        inputRef.current?.blur();
        break;
    }
  };

  const handleSuggestionSelect = (suggestion: string) => {
    setQuery(suggestion);
    setIsOpen(false);
    setSelectedIndex(-1);
    onSuggestionSelect(suggestion);
  };

  const handleClear = () => {
    setQuery("");
    setSuggestions([]);
    setIsOpen(false);
    setSelectedIndex(-1);
    inputRef.current?.focus();
  };

  return (
    <div className={`relative ${className}`}>
      <div className="relative">
        <div className="absolute inset-y-0 left-0 pl-3 flex items-center pointer-events-none">
          <Search className="h-5 w-5 text-gray-400" />
        </div>
        
        <input
          ref={inputRef}
          type="text"
          value={query}
          onChange={handleInputChange}
          onFocus={handleInputFocus}
          onBlur={handleInputBlur}
          onKeyDown={handleKeyDown}
          placeholder={placeholder}
          disabled={disabled}
          className={`
            block w-full pl-10 pr-10 py-3 border border-gray-300 rounded-lg
            focus:ring-2 focus:ring-blue-500 focus:border-blue-500
            disabled:bg-gray-50 disabled:text-gray-500 disabled:cursor-not-allowed
            transition-colors duration-200
            ${isOpen ? 'rounded-b-none' : ''}
          `}
          aria-label="Search products with autocomplete"
          aria-expanded={isOpen}
          aria-haspopup="listbox"
          aria-activedescendant={selectedIndex >= 0 ? `suggestion-${selectedIndex}` : undefined}
        />
        
        <div className="absolute inset-y-0 right-0 flex items-center">
          {isLoading && (
            <div className="animate-spin rounded-full h-4 w-4 border-b-2 border-blue-600 mr-2"></div>
          )}
          
          {query && (
            <button
              type="button"
              onClick={handleClear}
              className="mr-2 p-1 text-gray-400 hover:text-gray-600"
              aria-label="Clear search"
            >
              <X className="h-4 w-4" />
            </button>
          )}
          
          <ChevronDown 
            className={`h-4 w-4 text-gray-400 transition-transform duration-200 ${
              isOpen ? 'rotate-180' : ''
            }`} 
          />
        </div>
      </div>

      {/* Suggestions Dropdown */}
      {isOpen && suggestions.length > 0 && (
        <div
          ref={suggestionsRef}
          className="absolute z-10 w-full bg-white border border-gray-300 border-t-0 rounded-b-lg shadow-lg max-h-60 overflow-y-auto"
          role="listbox"
        >
          {suggestions.map((suggestion, index) => (
            <div
              key={index}
              id={`suggestion-${index}`}
              className={`
                px-4 py-2 cursor-pointer transition-colors duration-150
                ${index === selectedIndex 
                  ? 'bg-blue-50 text-blue-900' 
                  : 'text-gray-900 hover:bg-gray-50'
                }
              `}
              onClick={() => handleSuggestionSelect(suggestion)}
              role="option"
              aria-selected={index === selectedIndex}
            >
              <div className="flex items-center">
                <Search className="h-4 w-4 text-gray-400 mr-2" />
                <span className="truncate">{suggestion}</span>
              </div>
            </div>
          ))}
        </div>
      )}

      {/* No suggestions message */}
      {isOpen && suggestions.length === 0 && !isLoading && query.length >= 2 && (
        <div className="absolute z-10 w-full bg-white border border-gray-300 border-t-0 rounded-b-lg shadow-lg">
          <div className="px-4 py-2 text-gray-500 text-sm">
            No suggestions found
          </div>
        </div>
      )}
    </div>
  );
};
