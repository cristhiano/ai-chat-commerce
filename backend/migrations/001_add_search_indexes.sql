-- Add search-specific fields to products table
ALTER TABLE products ADD COLUMN IF NOT EXISTS search_vector tsvector;
ALTER TABLE products ADD COLUMN IF NOT EXISTS search_weight float8 DEFAULT 0;
ALTER TABLE products ADD COLUMN IF NOT EXISTS popularity integer DEFAULT 0;

-- Create full-text search indexes
CREATE INDEX IF NOT EXISTS idx_products_name_gin ON products USING gin(to_tsvector('english', name));
CREATE INDEX IF NOT EXISTS idx_products_description_gin ON products USING gin(to_tsvector('english', description));
CREATE INDEX IF NOT EXISTS idx_products_search_vector_gin ON products USING gin(search_vector);

-- Additional performance indexes
CREATE INDEX IF NOT EXISTS idx_products_category_status ON products(category_id, status);
CREATE INDEX IF NOT EXISTS idx_products_price_range ON products(price) WHERE status = 'active';
CREATE INDEX IF NOT EXISTS idx_products_search_filters ON products(category_id, price, status) WHERE status = 'active';

-- Create search cache table
CREATE TABLE IF NOT EXISTS search_cache (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    query varchar(500) NOT NULL,
    filters jsonb,
    results jsonb NOT NULL,
    result_count integer NOT NULL,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP,
    expires_at timestamp NOT NULL
);

-- Create indexes for search cache
CREATE INDEX IF NOT EXISTS idx_search_cache_query ON search_cache(query);
CREATE INDEX IF NOT EXISTS idx_search_cache_expires ON search_cache(expires_at);

-- Create search suggestions table
CREATE TABLE IF NOT EXISTS search_suggestions (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    query varchar(200) NOT NULL,
    suggestions text NOT NULL,
    occurrences integer DEFAULT 1,
    last_used timestamp DEFAULT CURRENT_TIMESTAMP,
    created_at timestamp DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for search suggestions
CREATE UNIQUE INDEX IF NOT EXISTS idx_search_suggestion_query ON search_suggestions(query);

-- Create search analytics table
CREATE TABLE IF NOT EXISTS search_analytics (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    query varchar(500) NOT NULL,
    filters jsonb,
    result_count integer NOT NULL,
    selected_products text,
    response_time integer NOT NULL,
    cache_hit boolean DEFAULT false,
    user_id uuid,
    session_id varchar(100),
    created_at timestamp DEFAULT CURRENT_TIMESTAMP
);

-- Create indexes for search analytics
CREATE INDEX IF NOT EXISTS idx_search_analytics_query ON search_analytics(query);
CREATE INDEX IF NOT EXISTS idx_search_analytics_created ON search_analytics(created_at);
CREATE INDEX IF NOT EXISTS idx_search_analytics_user ON search_analytics(user_id) WHERE user_id IS NOT NULL;
