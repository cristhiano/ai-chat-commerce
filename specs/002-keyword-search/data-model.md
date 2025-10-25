# Data Model: Product Keyword Search

**Date:** 2024-12-19  
**Project:** Product Keyword Search  
**Technology Stack:** Golang, PostgreSQL, GORM  
**Version:** 1.0

## Database Schema Overview

The search feature extends the existing product catalog data model with search-specific entities, full-text search indexes, and cache tables for optimal performance.

## Search-Specific Entities

### SearchCache
**Purpose:** Cache frequently accessed search results to improve performance

```go
type SearchCache struct {
    ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    Query       string    `gorm:"size:500;not null;index"`
    Filters     datatypes.JSON `gorm:"type:jsonb"`
    Results     datatypes.JSON `gorm:"type:jsonb;not null"`
    ResultCount int       `gorm:"not null"`
    CreatedAt   time.Time
    ExpiresAt   time.Time `gorm:"not null;index"`
    
    // Index for efficient expiration cleanup
    // CREATE INDEX idx_search_cache_expires ON search_cache(expires_at);
}
```

**Validation Rules:**
- Query must be non-empty
- Results must be valid JSON array
- ExpiresAt must be in the future

### SearchSuggestion
**Purpose:** Store autocomplete suggestions for common queries

```go
type SearchSuggestion struct {
    ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    Query       string    `gorm:"size:200;not null;index"`
    Suggestions pq.StringArray `gorm:"type:text[];not null"`
    Occurrences int       `gorm:"default:1;index"`
    LastUsed    time.Time
    CreatedAt   time.Time
    
    // Index for efficient query lookups
    // CREATE UNIQUE INDEX idx_search_suggestion_query ON search_suggestions(query);
}
```

**Validation Rules:**
- Query must be between 2-200 characters
- Suggestions array must not be empty
- Occurrences must be positive

### SearchAnalytics
**Purpose:** Track search queries for analytics and improvement

```go
type SearchAnalytics struct {
    ID          uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    Query       string    `gorm:"size:500;not null;index"`
    Filters     datatypes.JSON `gorm:"type:jsonb"`
    ResultCount int       `gorm:"not null"`
    SelectedProducts pq.StringArray `gorm:"type:uuid[]"`
    ResponseTime     int       `gorm:"not null"` // milliseconds
    CacheHit         bool      `gorm:"default:false"`
    UserID           *uuid.UUID `gorm:"type:uuid;index"`
    SessionID        string    `gorm:"size:100;index"`
    CreatedAt        time.Time
    
    // Index for time-based queries
    // CREATE INDEX idx_search_analytics_created ON search_analytics(created_at);
}
```

**Validation Rules:**
- Query must be non-empty
- ResponseTime must be positive
- At least UserID or SessionID must be provided

## Search-Specific Fields on Existing Entities

### Product (Extended for Search)
**Purpose:** Add search-specific fields and indexes to existing product entity

**Additional Search Fields:**
```go
// These fields can be added to the existing Product struct
SearchVector tsvector `gorm:"type:tsvector"` // For full-text search
SearchWeight float64  `gorm:"default:0"`      // For relevance ranking
Popularity   int      `gorm:"default:0"`      // Product popularity score
```

**Search-Specific Indexes:**
```sql
-- Full-text search indexes
CREATE INDEX idx_product_name_gin ON products USING gin(to_tsvector('english', name));
CREATE INDEX idx_product_description_gin ON products USING gin(to_tsvector('english', description));
CREATE INDEX idx_product_tags_gin ON products USING gin(tags);

-- Additional performance indexes
CREATE INDEX idx_product_category_status ON products(category_id, status);
CREATE INDEX idx_product_price_range ON products(price) WHERE status = 'active';

-- Composite index for common filter combinations
CREATE INDEX idx_product_search_filters ON products(category_id, price, status) WHERE status = 'active';
```

**Full-Text Search Configuration:**
```sql
-- Create search vector column (can be populated via trigger or application)
ALTER TABLE products ADD COLUMN IF NOT EXISTS search_vector tsvector;

-- Trigger to automatically update search_vector
CREATE OR REPLACE FUNCTION product_search_vector_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vector := 
        setweight(to_tsvector('english', COALESCE(NEW.name, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.description, '')), 'B') ||
        setweight(to_tsvector('english', COALESCE(array_to_string(NEW.tags, ' '), '')), 'C');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER product_search_vector_trigger
    BEFORE INSERT OR UPDATE ON products
    FOR EACH ROW
    EXECUTE FUNCTION product_search_vector_update();
```

### Category (Extended for Search)
**Purpose:** Add search-specific fields to existing category entity

**Additional Search Fields:**
```go
// These fields can be added to the existing Category struct
SearchKeywords pq.StringArray `gorm:"type:text[]"` // Alternative search terms
```

## Search View for Read Optimization

### ProductSearchView
**Purpose:** Materialized view for efficient search queries combining multiple tables

```sql
CREATE MATERIALIZED VIEW product_search_view AS
SELECT 
    p.id,
    p.name,
    p.description,
    p.price,
    p.category_id,
    c.name as category_name,
    p.sku,
    p.status,
    p.tags,
    p.search_vector,
    p.popularity,
    p.created_at,
    COALESCE(SUM(iv.quantity_available), 0) as total_available
FROM products p
LEFT JOIN categories c ON p.category_id = c.id
LEFT JOIN inventory iv ON p.id = iv.product_id
WHERE p.status = 'active'
GROUP BY p.id, c.id;

-- Create index on materialized view
CREATE INDEX idx_product_search_view_gin ON product_search_view USING gin(search_vector);

-- Refresh function (should be called periodically or after significant updates)
CREATE OR REPLACE FUNCTION refresh_product_search_view()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY product_search_view;
END;
$$ LANGUAGE plpgsql;
```

## Database Indexes Summary

### Performance Indexes
```sql
-- Products table
CREATE INDEX idx_products_name_gin ON products USING gin(to_tsvector('english', name));
CREATE INDEX idx_products_description_gin ON products USING gin(to_tsvector('english', description));
CREATE INDEX idx_products_tags_gin ON products USING gin(tags);
CREATE INDEX idx_products_category_status ON products(category_id, status);
CREATE INDEX idx_products_price ON products(price);
CREATE INDEX idx_products_status ON products(status);

-- Search cache
CREATE INDEX idx_search_cache_query ON search_cache(query);
CREATE INDEX idx_search_cache_expires ON search_cache(expires_at);

-- Search suggestions
CREATE INDEX idx_search_suggestion_query ON search_suggestions(query);

-- Search analytics
CREATE INDEX idx_search_analytics_query ON search_analytics(query);
CREATE INDEX idx_search_analytics_created ON search_analytics(created_at);
CREATE INDEX idx_search_analytics_user ON search_analytics(user_id) WHERE user_id IS NOT NULL;
```

## State Transitions and Validation

### Search Cache Lifecycle
1. **Created**: Cache entry created with TTL (default: 5 minutes)
2. **Valid**: Cache entry within TTL period
3. **Expired**: Cache entry past TTL, eligible for cleanup
4. **Deleted**: Cache entry removed by cleanup job

### Search Suggestion Lifecycle
1. **Created**: New suggestion added from user query
2. **Updated**: Occurrence count incremented on reuse
3. **Promoted**: High-occurrence suggestions prioritized
4. **Archived**: Low-occurrence suggestions removed periodically

## Data Integrity Rules

### Search Cache
- Automatic expiration via background job running every minute
- Query string must be unique per filter combination
- JSON validation for Results and Filters fields

### Search Suggestions
- Query must be unique (enforced by unique index)
- Suggestions array must contain at least 1 item
- Automatic cleanup of suggestions with occurrence count < 2 (after 30 days)

### Search Analytics
- Retention period: 90 days (can be configured)
- Automatic archival of old records via scheduled job
- Privacy: UserID can be NULL for anonymous users

## Migration Strategy

### Phase 1: Add Search Indexes
```sql
-- Add full-text search indexes to existing products table
CREATE INDEX CONCURRENTLY idx_products_name_gin ON products USING gin(to_tsvector('english', name));
CREATE INDEX CONCURRENTLY idx_products_description_gin ON products USING gin(to_tsvector('english', description));
```

### Phase 2: Create Search Tables
```sql
-- Create search cache table
CREATE TABLE search_cache (...);

-- Create search suggestions table  
CREATE TABLE search_suggestions (...);

-- Create search analytics table
CREATE TABLE search_analytics (...);
```

### Phase 3: Add Materialized View
```sql
-- Create materialized view for optimized searches
CREATE MATERIALIZED VIEW product_search_view AS ...;
```

### Phase 4: Populate Search Data
```sql
-- Update existing products with search_vector
UPDATE products SET search_vector = ...;

-- Create trigger for automatic updates
CREATE TRIGGER product_search_vector_trigger ...;
```
