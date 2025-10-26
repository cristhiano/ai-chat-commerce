/**
 * Type Safety Tests: Product and ProductSuggestion Types
 * 
 * This test file validates that TypeScript types correctly handle
 * incomplete product data and enforce proper type checking.
 */

import { describe, it, expect } from 'vitest';
import type { Product, ProductSuggestion, Category, Inventory } from '../index';

describe('Product Type Safety', () => {
  describe('Product with optional fields', () => {
    it('should accept Product with all required fields only', () => {
      const minimalProduct: Product = {
        id: '123',
        name: 'Test Product',
        description: 'Test description',
        price: 99.99,
        category_id: 'cat-1',
        sku: 'TEST-001',
        status: 'active',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      };

      expect(minimalProduct).toBeDefined();
      expect(minimalProduct.id).toBe('123');
      expect(minimalProduct.tags).toBeUndefined();
      expect(minimalProduct.category).toBeUndefined();
      expect(minimalProduct.inventory).toBeUndefined();
    });

    it('should accept Product with optional tags', () => {
      const productWithTags: Product = {
        id: '123',
        name: 'Test Product',
        description: 'Test description',
        price: 99.99,
        category_id: 'cat-1',
        sku: 'TEST-001',
        status: 'active',
        tags: ['popular', 'bestseller'],
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      };

      expect(productWithTags.tags).toEqual(['popular', 'bestseller']);
      expect(productWithTags.tags?.length).toBe(2);
    });

    it('should accept Product with optional category', () => {
      const category: Category = {
        id: 'cat-1',
        name: 'Electronics',
        description: 'Electronic products',
        slug: 'electronics',
        sort_order: 1,
        is_active: true,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      };

      const productWithCategory: Product = {
        id: '123',
        name: 'Test Product',
        description: 'Test description',
        price: 99.99,
        category_id: 'cat-1',
        sku: 'TEST-001',
        status: 'active',
        category,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      };

      expect(productWithCategory.category).toBeDefined();
      expect(productWithCategory.category?.name).toBe('Electronics');
    });

    it('should accept Product with optional inventory', () => {
      const inventory: Inventory[] = [
        {
          id: 'inv-1',
          product_id: '123',
          warehouse_location: 'Warehouse A',
          quantity_available: 100,
          quantity_reserved: 10,
          low_stock_threshold: 20,
          reorder_point: 30,
        },
      ];

      const productWithInventory: Product = {
        id: '123',
        name: 'Test Product',
        description: 'Test description',
        price: 99.99,
        category_id: 'cat-1',
        sku: 'TEST-001',
        status: 'active',
        inventory,
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      };

      expect(productWithInventory.inventory).toBeDefined();
      expect(productWithInventory.inventory?.length).toBe(1);
      expect(productWithInventory.inventory?.[0].quantity_available).toBe(100);
    });

    it('should handle Product with all optional fields present', () => {
      const fullProduct: Product = {
        id: '123',
        name: 'Test Product',
        description: 'Test description',
        price: 99.99,
        category_id: 'cat-1',
        sku: 'TEST-001',
        status: 'active',
        metadata: { featured: true },
        tags: ['new', 'popular'],
        category: {
          id: 'cat-1',
          name: 'Electronics',
          description: 'Electronic products',
          slug: 'electronics',
          sort_order: 1,
          is_active: true,
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
        variants: [],
        inventory: [
          {
            id: 'inv-1',
            product_id: '123',
            warehouse_location: 'Warehouse A',
            quantity_available: 100,
            quantity_reserved: 10,
            low_stock_threshold: 20,
            reorder_point: 30,
          },
        ],
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      };

      expect(fullProduct).toBeDefined();
      expect(fullProduct.tags).toBeDefined();
      expect(fullProduct.category).toBeDefined();
      expect(fullProduct.inventory).toBeDefined();
      expect(fullProduct.metadata).toBeDefined();
    });
  });

  describe('ProductSuggestion with optional fields', () => {
    it('should accept ProductSuggestion with product only', () => {
      const suggestion: ProductSuggestion = {
        product: {
          id: '123',
          name: 'Test Product',
          description: 'Test description',
          price: 99.99,
          category_id: 'cat-1',
          sku: 'TEST-001',
          status: 'active',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
      };

      expect(suggestion.product).toBeDefined();
      expect(suggestion.reason).toBeUndefined();
      expect(suggestion.confidence).toBeUndefined();
      expect(suggestion.metadata).toBeUndefined();
    });

    it('should accept ProductSuggestion with optional reason', () => {
      const suggestion: ProductSuggestion = {
        product: {
          id: '123',
          name: 'Test Product',
          description: 'Test description',
          price: 99.99,
          category_id: 'cat-1',
          sku: 'TEST-001',
          status: 'active',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
        reason: 'Matches your search query',
      };

      expect(suggestion.reason).toBe('Matches your search query');
    });

    it('should accept ProductSuggestion with optional confidence', () => {
      const suggestion: ProductSuggestion = {
        product: {
          id: '123',
          name: 'Test Product',
          description: 'Test description',
          price: 99.99,
          category_id: 'cat-1',
          sku: 'TEST-001',
          status: 'active',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
        confidence: 0.85,
      };

      expect(suggestion.confidence).toBe(0.85);
    });

    it('should accept ProductSuggestion with optional metadata', () => {
      const suggestion: ProductSuggestion = {
        product: {
          id: '123',
          name: 'Test Product',
          description: 'Test description',
          price: 99.99,
          category_id: 'cat-1',
          sku: 'TEST-001',
          status: 'active',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
        metadata: {
          search_query: 'laptop',
          filters_applied: ['electronics', 'high-rating'],
          sort_method: 'price_asc',
        },
      };

      expect(suggestion.metadata).toBeDefined();
      expect(suggestion.metadata?.search_query).toBe('laptop');
      expect(suggestion.metadata?.filters_applied).toHaveLength(2);
    });

    it('should accept ProductSuggestion with all optional fields', () => {
      const suggestion: ProductSuggestion = {
        product: {
          id: '123',
          name: 'Test Product',
          description: 'Test description',
          price: 99.99,
          category_id: 'cat-1',
          sku: 'TEST-001',
          status: 'active',
          tags: ['new', 'popular'],
          category: {
            id: 'cat-1',
            name: 'Electronics',
            description: 'Electronic products',
            slug: 'electronics',
            sort_order: 1,
            is_active: true,
            created_at: '2024-01-01T00:00:00Z',
            updated_at: '2024-01-01T00:00:00Z',
          },
          inventory: [
            {
              id: 'inv-1',
              product_id: '123',
              warehouse_location: 'Warehouse A',
              quantity_available: 100,
              quantity_reserved: 10,
              low_stock_threshold: 20,
              reorder_point: 30,
            },
          ],
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
        reason: 'Best match for your search',
        confidence: 0.95,
        metadata: {
          search_query: 'laptop',
          filters_applied: ['electronics'],
          sort_method: 'relevance',
        },
      };

      expect(suggestion).toBeDefined();
      expect(suggestion.product.tags).toHaveLength(2);
      expect(suggestion.product.category).toBeDefined();
      expect(suggestion.product.inventory).toHaveLength(1);
      expect(suggestion.reason).toBeDefined();
      expect(suggestion.confidence).toBe(0.95);
      expect(suggestion.metadata).toBeDefined();
    });
  });

  describe('Safe access patterns', () => {
    it('should handle missing category gracefully', () => {
      const product: Product = {
        id: '123',
        name: 'Test Product',
        description: 'Test description',
        price: 99.99,
        category_id: 'cat-1',
        sku: 'TEST-001',
        status: 'active',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      };

      // Safe access with optional chaining
      const categoryName = product.category?.name;
      expect(categoryName).toBeUndefined();

      // Safe access with fallback
      const categoryNameWithFallback = product.category?.name ?? 'Uncategorized';
      expect(categoryNameWithFallback).toBe('Uncategorized');
    });

    it('should handle missing tags gracefully', () => {
      const product: Product = {
        id: '123',
        name: 'Test Product',
        description: 'Test description',
        price: 99.99,
        category_id: 'cat-1',
        sku: 'TEST-001',
        status: 'active',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      };

      // Safe access with optional chaining
      const firstTag = product.tags?.[0];
      expect(firstTag).toBeUndefined();

      // Safe length check
      const tagCount = product.tags?.length ?? 0;
      expect(tagCount).toBe(0);

      // Safe iteration
      const tags = product.tags ?? [];
      expect(tags).toHaveLength(0);
    });

    it('should handle missing inventory gracefully', () => {
      const product: Product = {
        id: '123',
        name: 'Test Product',
        description: 'Test description',
        price: 99.99,
        category_id: 'cat-1',
        sku: 'TEST-001',
        status: 'active',
        created_at: '2024-01-01T00:00:00Z',
        updated_at: '2024-01-01T00:00:00Z',
      };

      // Safe check for out of stock
      const hasInventory = product.inventory && product.inventory.length > 0;
      expect(hasInventory).toBeFalsy();

      // Safe check with default
      const isOutOfStock = product.inventory?.some((inv) => inv.quantity_available === 0) ?? false;
      expect(isOutOfStock).toBe(false);
    });

    it('should handle missing ProductSuggestion fields gracefully', () => {
      const suggestion: ProductSuggestion = {
        product: {
          id: '123',
          name: 'Test Product',
          description: 'Test description',
          price: 99.99,
          category_id: 'cat-1',
          sku: 'TEST-001',
          status: 'active',
          created_at: '2024-01-01T00:00:00Z',
          updated_at: '2024-01-01T00:00:00Z',
        },
      };

      // Safe access to optional fields
      const reason = suggestion.reason ?? 'No reason provided';
      expect(reason).toBe('No reason provided');

      const confidence = suggestion.confidence ?? 0;
      expect(confidence).toBe(0);

      const searchQuery = suggestion.metadata?.search_query ?? '';
      expect(searchQuery).toBe('');
    });
  });
});

