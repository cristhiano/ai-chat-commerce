-- Optimize search performance with proper indexing
-- This migration adds additional indexes for search optimization

-- Add composite indexes for common search patterns
CREATE INDEX IF NOT EXISTS idx_products_search_composite 
ON products (status, category_id, price) 
WHERE status = 'active';

-- Add index for availability filtering
CREATE INDEX IF NOT EXISTS idx_products_availability 
ON products (status, id) 
WHERE status = 'active';

-- Add index for price range queries
CREATE INDEX IF NOT EXISTS idx_products_price_range 
ON products (price, status) 
WHERE status = 'active';

-- Add index for category and price combination
CREATE INDEX IF NOT EXISTS idx_products_category_price 
ON products (category_id, price, status) 
WHERE status = 'active';

-- Add index for search vector with status filter
CREATE INDEX IF NOT EXISTS idx_products_search_vector_active 
ON products USING GIN (search_vector) 
WHERE status = 'active';

-- Add index for popularity-based sorting
CREATE INDEX IF NOT EXISTS idx_products_popularity_active 
ON products (popularity DESC, created_at DESC) 
WHERE status = 'active';

-- Add index for SKU searches
CREATE INDEX IF NOT EXISTS idx_products_sku_active 
ON products (sku, status) 
WHERE status = 'active';

-- Add index for metadata searches (brand, etc.)
CREATE INDEX IF NOT EXISTS idx_products_metadata_brand 
ON products USING GIN (metadata) 
WHERE status = 'active' AND metadata ? 'brand';

-- Add index for search analytics queries
CREATE INDEX IF NOT EXISTS idx_search_analytics_time_query 
ON search_analytics (created_at DESC, query);

-- Add index for search analytics user tracking
CREATE INDEX IF NOT EXISTS idx_search_analytics_user_time 
ON search_analytics (user_id, created_at DESC) 
WHERE user_id IS NOT NULL;

-- Add index for search cache expiration
CREATE INDEX IF NOT EXISTS idx_search_cache_expiration 
ON search_cache (expires_at, created_at);

-- Add index for search suggestions
CREATE INDEX IF NOT EXISTS idx_search_suggestions_popular 
ON search_suggestions (occurrences DESC, last_used DESC);

-- Optimize materialized view refresh
CREATE INDEX IF NOT EXISTS idx_productsearchview_category_price 
ON ProductSearchView (category_name, price);

-- Add index for inventory availability
CREATE INDEX IF NOT EXISTS idx_inventory_product_available 
ON inventory (product_id, stock_quantity) 
WHERE stock_quantity > 0;

-- Add partial index for active inventory
CREATE INDEX IF NOT EXISTS idx_inventory_active 
ON inventory (product_id, stock_quantity) 
WHERE stock_quantity > 0;

-- Add index for product variants
CREATE INDEX IF NOT EXISTS idx_product_variants_active 
ON product_variants (product_id, status) 
WHERE status = 'active';

-- Add index for product images
CREATE INDEX IF NOT EXISTS idx_product_images_primary 
ON product_images (product_id, is_primary) 
WHERE is_primary = true;

-- Add index for categories hierarchy
CREATE INDEX IF NOT EXISTS idx_categories_parent 
ON categories (parent_id, name) 
WHERE parent_id IS NOT NULL;

-- Add index for categories with product count
CREATE INDEX IF NOT EXISTS idx_categories_active 
ON categories (id, name) 
WHERE status = 'active';

-- Add index for search clicks tracking
CREATE INDEX IF NOT EXISTS idx_search_clicks_query_time 
ON search_clicks (query, created_at DESC);

-- Add index for search sessions
CREATE INDEX IF NOT EXISTS idx_search_sessions_user_time 
ON search_sessions (user_id, started_at DESC) 
WHERE user_id IS NOT NULL;

-- Add index for query refinements
CREATE INDEX IF NOT EXISTS idx_query_refinements_pattern 
ON query_refinements (original_query, refined_query);

-- Add index for filter usage
CREATE INDEX IF NOT EXISTS idx_filter_usage_query_time 
ON filter_usage (query, used_at DESC);

-- Add index for search abandonment
CREATE INDEX IF NOT EXISTS idx_search_abandonment_query 
ON search_abandonment (query, abandoned_at DESC);

-- Update table statistics for better query planning
ANALYZE products;
ANALYZE categories;
ANALYZE inventory;
ANALYZE search_analytics;
ANALYZE search_cache;
ANALYZE search_suggestions;
ANALYZE ProductSearchView;

-- Create function to refresh materialized view efficiently
CREATE OR REPLACE FUNCTION refresh_search_view_efficiently()
RETURNS void AS $$
BEGIN
    -- Only refresh if there are significant changes
    IF EXISTS (
        SELECT 1 FROM products 
        WHERE updated_at > (
            SELECT COALESCE(MAX(last_refresh), '1970-01-01'::timestamp) 
            FROM search_view_refresh_log
        )
    ) THEN
        REFRESH MATERIALIZED VIEW CONCURRENTLY ProductSearchView;
        
        -- Log the refresh
        INSERT INTO search_view_refresh_log (last_refresh) 
        VALUES (NOW()) 
        ON CONFLICT (id) DO UPDATE SET last_refresh = NOW();
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Create table to track materialized view refresh
CREATE TABLE IF NOT EXISTS search_view_refresh_log (
    id SERIAL PRIMARY KEY,
    last_refresh TIMESTAMP DEFAULT NOW()
);

-- Insert initial record
INSERT INTO search_view_refresh_log (last_refresh) 
VALUES (NOW()) 
ON CONFLICT (id) DO NOTHING;

-- Create function to get search performance stats
CREATE OR REPLACE FUNCTION get_search_performance_stats()
RETURNS TABLE (
    total_searches BIGINT,
    avg_response_time NUMERIC,
    cache_hit_rate NUMERIC,
    no_results_rate NUMERIC
) AS $$
BEGIN
    RETURN QUERY
    SELECT 
        COUNT(*) as total_searches,
        AVG(response_time) as avg_response_time,
        (COUNT(*) FILTER (WHERE cache_hit = true) * 100.0 / COUNT(*)) as cache_hit_rate,
        (COUNT(*) FILTER (WHERE result_count = 0) * 100.0 / COUNT(*)) as no_results_rate
    FROM search_analytics
    WHERE created_at > NOW() - INTERVAL '24 hours';
END;
$$ LANGUAGE plpgsql;
