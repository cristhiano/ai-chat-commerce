-- Create materialized view for efficient search queries
CREATE MATERIALIZED VIEW IF NOT EXISTS product_search_view AS
SELECT 
    p.id,
    p.name,
    p.description,
    p.price,
    p.category_id,
    c.name as category_name,
    p.sku,
    p.status,
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
CREATE INDEX IF NOT EXISTS idx_product_search_view_gin ON product_search_view USING gin(search_vector);

-- Create function to refresh materialized view
CREATE OR REPLACE FUNCTION refresh_product_search_view()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY product_search_view;
END;
$$ LANGUAGE plpgsql;

-- Create trigger to automatically update search_vector
CREATE OR REPLACE FUNCTION product_search_vector_update() RETURNS trigger AS $$
BEGIN
    NEW.search_vector := 
        setweight(to_tsvector('english', COALESCE(NEW.name, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.description, '')), 'B');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Drop existing trigger if it exists
DROP TRIGGER IF EXISTS product_search_vector_trigger ON products;

-- Create trigger
CREATE TRIGGER product_search_vector_trigger
    BEFORE INSERT OR UPDATE ON products
    FOR EACH ROW
    EXECUTE FUNCTION product_search_vector_update();
