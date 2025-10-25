package services

import (
	"chat-ecommerce-backend/internal/models"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/lib/pq"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// CreateProductRequest represents the request payload for creating a product
type CreateProductRequest struct {
	Name        string         `json:"name" binding:"required"`
	Description string         `json:"description" binding:"required"`
	Price       float64        `json:"price" binding:"required,min=0"`
	CategoryID  uuid.UUID      `json:"category_id" binding:"required"`
	SKU         string         `json:"sku" binding:"required"`
	Status      string         `json:"status"`
	Metadata    datatypes.JSON `json:"metadata"`
	Tags        pq.StringArray `json:"tags"`
	Images      datatypes.JSON `json:"images"`
}

// Product represents a product in the catalog (alias for models.Product)
type Product = models.Product

// ProductService handles product-related business logic
type ProductService struct {
	db *gorm.DB
}

// NewProductService creates a new ProductService
func NewProductService(db *gorm.DB) *ProductService {
	return &ProductService{
		db: db,
	}
}

// ProductFilters represents search and filter parameters
type ProductFilters struct {
	Search     string    `json:"search"`
	CategoryID uuid.UUID `json:"category_id"`
	MinPrice   float64   `json:"min_price"`
	MaxPrice   float64   `json:"max_price"`
	Status     string    `json:"status"`
	Tags       []string  `json:"tags"`
	Page       int       `json:"page"`
	Limit      int       `json:"limit"`
	SortBy     string    `json:"sort_by"`
	SortOrder  string    `json:"sort_order"`
}

// ProductListResponse represents paginated product list response
type ProductListResponse struct {
	Products    []models.Product `json:"products"`
	Total       int64            `json:"total"`
	Page        int              `json:"page"`
	Limit       int              `json:"limit"`
	TotalPages  int              `json:"total_pages"`
	HasNext     bool             `json:"has_next"`
	HasPrevious bool             `json:"has_previous"`
}

// GetProducts retrieves products with filtering and pagination
func (s *ProductService) GetProducts(filters ProductFilters) (*ProductListResponse, error) {
	var products []models.Product
	var total int64

	query := s.db.Model(&models.Product{})

	// Apply filters
	if filters.Search != "" {
		searchTerm := "%" + strings.ToLower(filters.Search) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", searchTerm, searchTerm)
	}

	if filters.CategoryID != uuid.Nil {
		query = query.Where("category_id = ?", filters.CategoryID)
	}

	if filters.MinPrice > 0 {
		query = query.Where("price >= ?", filters.MinPrice)
	}

	if filters.MaxPrice > 0 {
		query = query.Where("price <= ?", filters.MaxPrice)
	}

	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}

	if len(filters.Tags) > 0 {
		query = query.Where("tags && ?", filters.Tags)
	}

	// Count total records
	if err := query.Count(&total).Error; err != nil {
		return nil, fmt.Errorf("failed to count products: %w", err)
	}

	// Apply pagination
	offset := (filters.Page - 1) * filters.Limit
	if offset < 0 {
		offset = 0
	}

	// Apply sorting
	sortBy := filters.SortBy
	if sortBy == "" {
		sortBy = "created_at"
	}
	sortOrder := filters.SortOrder
	if sortOrder == "" {
		sortOrder = "desc"
	}
	orderClause := fmt.Sprintf("%s %s", sortBy, sortOrder)
	query = query.Order(orderClause)

	// Execute query with pagination
	if err := query.Offset(offset).Limit(filters.Limit).
		Preload("Category").
		Preload("Variants").
		Preload("Inventory").
		Find(&products).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch products: %w", err)
	}

	// Calculate pagination info
	totalPages := int((total + int64(filters.Limit) - 1) / int64(filters.Limit))
	hasNext := filters.Page < totalPages
	hasPrevious := filters.Page > 1

	return &ProductListResponse{
		Products:    products,
		Total:       total,
		Page:        filters.Page,
		Limit:       filters.Limit,
		TotalPages:  totalPages,
		HasNext:     hasNext,
		HasPrevious: hasPrevious,
	}, nil
}

// GetProductByID retrieves a single product by ID
func (s *ProductService) GetProductByID(id uuid.UUID) (*models.Product, error) {
	var product models.Product

	if err := s.db.Where("id = ?", id).
		Preload("Category").
		Preload("Variants").
		Preload("Inventory").
		First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("product not found")
		}
		return nil, fmt.Errorf("failed to fetch product: %w", err)
	}

	return &product, nil
}

// GetProductBySKU retrieves a product by SKU
func (s *ProductService) GetProductBySKU(sku string) (*models.Product, error) {
	var product models.Product

	if err := s.db.Where("sku = ?", sku).
		Preload("Category").
		Preload("Variants").
		Preload("Inventory").
		First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("product not found")
		}
		return nil, fmt.Errorf("failed to fetch product: %w", err)
	}

	return &product, nil
}

// CreateProduct creates a new product
func (s *ProductService) CreateProduct(product *models.Product) error {
	// Validate required fields
	if product.Name == "" {
		return fmt.Errorf("product name is required")
	}
	if product.Description == "" {
		return fmt.Errorf("product description is required")
	}
	if product.Price <= 0 {
		return fmt.Errorf("product price must be greater than 0")
	}
	if product.SKU == "" {
		return fmt.Errorf("product SKU is required")
	}

	// Check if SKU already exists
	var existingProduct models.Product
	if err := s.db.Where("sku = ?", product.SKU).First(&existingProduct).Error; err == nil {
		return fmt.Errorf("product with SKU %s already exists", product.SKU)
	}

	// Set default values
	if product.Status == "" {
		product.Status = "active"
	}

	// Generate UUID if not provided
	if product.ID == uuid.Nil {
		product.ID = uuid.New()
	}

	if err := s.db.Create(product).Error; err != nil {
		return fmt.Errorf("failed to create product: %w", err)
	}

	return nil
}

// UpdateProduct updates an existing product
func (s *ProductService) UpdateProduct(id uuid.UUID, updates map[string]interface{}) error {
	// Validate that product exists
	var product models.Product
	if err := s.db.Where("id = ?", id).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("product not found")
		}
		return fmt.Errorf("failed to fetch product: %w", err)
	}

	// Validate updates
	if name, ok := updates["name"].(string); ok && name == "" {
		return fmt.Errorf("product name cannot be empty")
	}
	if price, ok := updates["price"].(float64); ok && price <= 0 {
		return fmt.Errorf("product price must be greater than 0")
	}
	if sku, ok := updates["sku"].(string); ok && sku != "" {
		// Check if SKU already exists for a different product
		var existingProduct models.Product
		if err := s.db.Where("sku = ? AND id != ?", sku, id).First(&existingProduct).Error; err == nil {
			return fmt.Errorf("product with SKU %s already exists", sku)
		}
	}

	if err := s.db.Model(&product).Updates(updates).Error; err != nil {
		return fmt.Errorf("failed to update product: %w", err)
	}

	return nil
}

// DeleteProduct soft deletes a product
func (s *ProductService) DeleteProduct(id uuid.UUID) error {
	// Check if product exists
	var product models.Product
	if err := s.db.Where("id = ?", id).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("product not found")
		}
		return fmt.Errorf("failed to fetch product: %w", err)
	}

	// Soft delete by setting status to inactive
	if err := s.db.Model(&product).Update("status", "inactive").Error; err != nil {
		return fmt.Errorf("failed to delete product: %w", err)
	}

	return nil
}

// GetCategories retrieves all active categories
func (s *ProductService) GetCategories() ([]models.Category, error) {
	var categories []models.Category

	if err := s.db.Where("is_active = ?", true).
		Order("sort_order ASC, name ASC").
		Find(&categories).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch categories: %w", err)
	}

	return categories, nil
}

// GetCategoryByID retrieves a category by ID
func (s *ProductService) GetCategoryByID(id uuid.UUID) (*models.Category, error) {
	var category models.Category

	if err := s.db.Where("id = ? AND is_active = ?", id, true).
		Preload("Products").
		First(&category).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to fetch category: %w", err)
	}

	return &category, nil
}

// GetCategoryBySlug retrieves a category by slug
func (s *ProductService) GetCategoryBySlug(slug string) (*models.Category, error) {
	var category models.Category

	if err := s.db.Where("slug = ? AND is_active = ?", slug, true).
		Preload("Products").
		First(&category).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, fmt.Errorf("category not found")
		}
		return nil, fmt.Errorf("failed to fetch category: %w", err)
	}

	return &category, nil
}

// SearchProducts performs full-text search on products
func (s *ProductService) SearchProducts(query string, limit int) ([]models.Product, error) {
	var products []models.Product

	if query == "" {
		return products, nil
	}

	searchTerm := "%" + strings.ToLower(query) + "%"

	if err := s.db.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ? OR ? = ANY(tags)",
		searchTerm, searchTerm, strings.ToLower(query)).
		Where("status = ?", "active").
		Limit(limit).
		Preload("Category").
		Preload("Variants").
		Find(&products).Error; err != nil {
		return nil, fmt.Errorf("failed to search products: %w", err)
	}

	return products, nil
}

// GetFeaturedProducts retrieves featured products
func (s *ProductService) GetFeaturedProducts(limit int) ([]models.Product, error) {
	var products []models.Product

	if err := s.db.Where("status = ?", "active").
		Order("created_at DESC").
		Limit(limit).
		Preload("Category").
		Preload("Variants").
		Find(&products).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch featured products: %w", err)
	}

	return products, nil
}

// GetRelatedProducts retrieves products related to the given product
func (s *ProductService) GetRelatedProducts(productID uuid.UUID, limit int) ([]models.Product, error) {
	var product models.Product
	if err := s.db.Where("id = ?", productID).First(&product).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch product: %w", err)
	}

	var relatedProducts []models.Product

	// Find products in the same category
	if err := s.db.Where("category_id = ? AND id != ? AND status = ?",
		product.CategoryID, productID, "active").
		Limit(limit).
		Preload("Category").
		Preload("Variants").
		Find(&relatedProducts).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch related products: %w", err)
	}

	return relatedProducts, nil
}
