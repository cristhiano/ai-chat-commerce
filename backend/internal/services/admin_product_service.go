package services

import (
	"chat-ecommerce-backend/internal/models"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
	"gorm.io/gorm"
)

// AdminProductService handles admin-specific product operations
type AdminProductService struct {
	db *gorm.DB
}

// NewAdminProductService creates a new AdminProductService
func NewAdminProductService(db *gorm.DB) *AdminProductService {
	return &AdminProductService{
		db: db,
	}
}

// AdminProductRequest represents the request payload for admin product operations
type AdminProductRequest struct {
	Name        string                  `json:"name" binding:"required"`
	Description string                  `json:"description" binding:"required"`
	Price       float64                 `json:"price" binding:"required,min=0"`
	CategoryID  uuid.UUID               `json:"category_id" binding:"required"`
	SKU         string                  `json:"sku" binding:"required"`
	Status      string                  `json:"status"`
	Metadata    map[string]interface{}  `json:"metadata"`
	Tags        []string                `json:"tags"`
	Images      []ProductImageRequest   `json:"images"`
	Variants    []ProductVariantRequest `json:"variants"`
	Inventory   []InventoryRequest      `json:"inventory"`
}

// ProductImageRequest represents a product image request
type ProductImageRequest struct {
	URL       string `json:"url" binding:"required"`
	AltText   string `json:"alt_text"`
	IsPrimary bool   `json:"is_primary"`
	SortOrder int    `json:"sort_order"`
}

// ProductVariantRequest represents a product variant request
type ProductVariantRequest struct {
	VariantName   string  `json:"variant_name" binding:"required"`
	VariantValue  string  `json:"variant_value" binding:"required"`
	PriceModifier float64 `json:"price_modifier"`
	SKUSuffix     string  `json:"sku_suffix"`
	IsDefault     bool    `json:"is_default"`
}

// InventoryRequest represents an inventory request
type InventoryRequest struct {
	VariantID *uuid.UUID `json:"variant_id"`
	Quantity  int        `json:"quantity" binding:"required,min=0"`
	Location  string     `json:"location"`
	Reserved  int        `json:"reserved"`
}

// AdminProductResponse represents the response for admin product operations
type AdminProductResponse struct {
	Product   *models.Product         `json:"product"`
	Variants  []models.ProductVariant `json:"variants"`
	Images    []models.ProductImage   `json:"images"`
	Inventory []models.Inventory      `json:"inventory"`
}

// BulkImportRequest represents a bulk import request
type BulkImportRequest struct {
	Products       []AdminProductRequest `json:"products" binding:"required"`
	UpdateExisting bool                  `json:"update_existing"`
}

// BulkImportResponse represents the response for bulk import
type BulkImportResponse struct {
	TotalProcessed int               `json:"total_processed"`
	Created        int               `json:"created"`
	Updated        int               `json:"updated"`
	Errors         []BulkImportError `json:"errors"`
}

// BulkImportError represents an error in bulk import
type BulkImportError struct {
	Index int    `json:"index"`
	SKU   string `json:"sku"`
	Error string `json:"error"`
}

// CreateProduct creates a new product with all related data
func (s *AdminProductService) CreateProduct(req AdminProductRequest) (*AdminProductResponse, error) {
	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Convert metadata to JSON
	var metadataJSON datatypes.JSON
	if req.Metadata != nil {
		metadataBytes, err := json.Marshal(req.Metadata)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to marshal metadata: %v", err)
		}
		metadataJSON = datatypes.JSON(metadataBytes)
	}

	// Create product
	product := &models.Product{
		Name:        req.Name,
		Description: req.Description,
		Price:       req.Price,
		CategoryID:  req.CategoryID,
		SKU:         req.SKU,
		Status:      req.Status,
		Metadata:    metadataJSON,
		// Tags:        pq.StringArray(req.Tags),
	}

	if err := tx.Create(product).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to create product: %v", err)
	}

	// Create variants
	var variants []models.ProductVariant
	for _, variantReq := range req.Variants {
		variant := models.ProductVariant{
			ProductID:     product.ID,
			VariantName:   variantReq.VariantName,
			VariantValue:  variantReq.VariantValue,
			PriceModifier: variantReq.PriceModifier,
			SKUSuffix:     variantReq.SKUSuffix,
			IsDefault:     variantReq.IsDefault,
		}
		if err := tx.Create(&variant).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create variant: %v", err)
		}
		variants = append(variants, variant)
	}

	// Create images
	var images []models.ProductImage
	for i, imageReq := range req.Images {
		image := models.ProductImage{
			ProductID: product.ID,
			URL:       imageReq.URL,
			AltText:   imageReq.AltText,
			IsPrimary: imageReq.IsPrimary,
			SortOrder: imageReq.SortOrder,
		}
		if imageReq.SortOrder == 0 {
			image.SortOrder = i + 1
		}
		if err := tx.Create(&image).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create image: %json", err)
		}
		images = append(images, image)
	}

	// Create inventory
	var inventory []models.Inventory
	for _, invReq := range req.Inventory {
		inventoryItem := models.Inventory{
			ProductID:         product.ID,
			VariantID:         invReq.VariantID,
			QuantityAvailable: invReq.Quantity,
			WarehouseLocation: invReq.Location,
			QuantityReserved:  invReq.Reserved,
		}
		if err := tx.Create(&inventoryItem).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create inventory: %v", err)
		}
		inventory = append(inventory, inventoryItem)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return &AdminProductResponse{
		Product:   product,
		Variants:  variants,
		Images:    images,
		Inventory: inventory,
	}, nil
}

// UpdateProduct updates an existing product
func (s *AdminProductService) UpdateProduct(id uuid.UUID, req AdminProductRequest) (*AdminProductResponse, error) {
	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Find existing product
	var product models.Product
	if err := tx.Preload("Variants").Preload("Images").Preload("Inventory").First(&product, id).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("product not found: %v", err)
	}

	// Convert metadata to JSON
	var metadataJSON datatypes.JSON
	if req.Metadata != nil {
		metadataBytes, err := json.Marshal(req.Metadata)
		if err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to marshal metadata: %v", err)
		}
		metadataJSON = datatypes.JSON(metadataBytes)
	}

	// Update product fields
	product.Name = req.Name
	product.Description = req.Description
	product.Price = req.Price
	product.CategoryID = req.CategoryID
	product.SKU = req.SKU
	product.Status = req.Status
	product.Metadata = metadataJSON
	// product.Tags = pq.StringArray(req.Tags)

	if err := tx.Save(&product).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to update product: %v", err)
	}

	// Update variants (delete existing and create new)
	if err := tx.Where("product_id = ?", product.ID).Delete(&models.ProductVariant{}).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to delete existing variants: %v", err)
	}

	var variants []models.ProductVariant
	for _, variantReq := range req.Variants {
		variant := models.ProductVariant{
			ProductID:     product.ID,
			VariantName:   variantReq.VariantName,
			VariantValue:  variantReq.VariantValue,
			PriceModifier: variantReq.PriceModifier,
			SKUSuffix:     variantReq.SKUSuffix,
			IsDefault:     variantReq.IsDefault,
		}
		if err := tx.Create(&variant).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create variant: %v", err)
		}
		variants = append(variants, variant)
	}

	// Update images (delete existing and create new)
	if err := tx.Where("product_id = ?", product.ID).Delete(&models.ProductImage{}).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to delete existing images: %v", err)
	}

	var images []models.ProductImage
	for i, imageReq := range req.Images {
		image := models.ProductImage{
			ProductID: product.ID,
			URL:       imageReq.URL,
			AltText:   imageReq.AltText,
			IsPrimary: imageReq.IsPrimary,
			SortOrder: imageReq.SortOrder,
		}
		if imageReq.SortOrder == 0 {
			image.SortOrder = i + 1
		}
		if err := tx.Create(&image).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create image: %v", err)
		}
		images = append(images, image)
	}

	// Update inventory (delete existing and create new)
	if err := tx.Where("product_id = ?", product.ID).Delete(&models.Inventory{}).Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to delete existing inventory: %v", err)
	}

	var inventory []models.Inventory
	for _, invReq := range req.Inventory {
		inventoryItem := models.Inventory{
			ProductID:         product.ID,
			VariantID:         invReq.VariantID,
			QuantityAvailable: invReq.Quantity,
			WarehouseLocation: invReq.Location,
			QuantityReserved:  invReq.Reserved,
		}
		if err := tx.Create(&inventoryItem).Error; err != nil {
			tx.Rollback()
			return nil, fmt.Errorf("failed to create inventory: %v", err)
		}
		inventory = append(inventory, inventoryItem)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %v", err)
	}

	return &AdminProductResponse{
		Product:   &product,
		Variants:  variants,
		Images:    images,
		Inventory: inventory,
	}, nil
}

// DeleteProduct deletes a product and all related data
func (s *AdminProductService) DeleteProduct(id uuid.UUID) error {
	// Start transaction
	tx := s.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// Check if product exists
	var product models.Product
	if err := tx.First(&product, id).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("product not found: %v", err)
	}

	// Delete related data
	if err := tx.Where("product_id = ?", id).Delete(&models.ProductVariant{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete variants: %v", err)
	}

	if err := tx.Where("product_id = ?", id).Delete(&models.ProductImage{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete images: %v", err)
	}

	if err := tx.Where("product_id = ?", id).Delete(&models.Inventory{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete inventory: %v", err)
	}

	if err := tx.Where("product_id = ?", id).Delete(&models.InventoryReservation{}).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete inventory reservations: %v", err)
	}

	// Delete product
	if err := tx.Delete(&product).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to delete product: %v", err)
	}

	// Commit transaction
	if err := tx.Commit().Error; err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

// GetProductWithDetails retrieves a product with all related data
func (s *AdminProductService) GetProductWithDetails(id uuid.UUID) (*AdminProductResponse, error) {
	var product models.Product
	if err := s.db.Preload("Variants").Preload("Images").Preload("Inventory").Preload("Category").First(&product, id).Error; err != nil {
		return nil, fmt.Errorf("product not found: %v", err)
	}

	return &AdminProductResponse{
		Product:   &product,
		Variants:  product.Variants,
		Images:    product.Images,
		Inventory: product.Inventory,
	}, nil
}

// BulkImportProducts imports multiple products
func (s *AdminProductService) BulkImportProducts(req BulkImportRequest) (*BulkImportResponse, error) {
	response := &BulkImportResponse{
		TotalProcessed: len(req.Products),
		Errors:         []BulkImportError{},
	}

	for i, productReq := range req.Products {
		// Check if product exists
		var existingProduct models.Product
		err := s.db.Where("sku = ?", productReq.SKU).First(&existingProduct).Error

		if err == nil && !req.UpdateExisting {
			// Product exists and we're not updating
			response.Errors = append(response.Errors, BulkImportError{
				Index: i,
				SKU:   productReq.SKU,
				Error: "Product already exists and update_existing is false",
			})
			continue
		}

		if err == nil && req.UpdateExisting {
			// Update existing product
			_, err = s.UpdateProduct(existingProduct.ID, productReq)
			if err != nil {
				response.Errors = append(response.Errors, BulkImportError{
					Index: i,
					SKU:   productReq.SKU,
					Error: err.Error(),
				})
				continue
			}
			response.Updated++
		} else {
			// Create new product
			_, err = s.CreateProduct(productReq)
			if err != nil {
				response.Errors = append(response.Errors, BulkImportError{
					Index: i,
					SKU:   productReq.SKU,
					Error: err.Error(),
				})
				continue
			}
			response.Created++
		}
	}

	return response, nil
}

// ExportProducts exports products to CSV format
func (s *AdminProductService) ExportProducts(filters ProductFilters) ([]byte, error) {
	var products []models.Product

	query := s.db.Preload("Category").Preload("Variants").Preload("Images").Preload("Inventory")

	if filters.Status != "" {
		query = query.Where("status = ?", filters.Status)
	}

	if filters.CategoryID != uuid.Nil {
		query = query.Where("category_id = ?", filters.CategoryID)
	}

	if filters.Search != "" {
		query = query.Where("name ILIKE ? OR description ILIKE ?", "%"+filters.Search+"%", "%"+filters.Search+"%")
	}

	if err := query.Find(&products).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch products: %v", err)
	}

	// Convert to CSV format
	csvData := s.convertToCSV(products)
	return csvData, nil
}

// convertToCSV converts products to CSV format
func (s *AdminProductService) convertToCSV(products []models.Product) []byte {
	var csv strings.Builder

	// CSV header
	csv.WriteString("ID,Name,Description,Price,SKU,Status,Category,Variants,Images,Inventory\n")

	for _, product := range products {
		// Basic product info
		csv.WriteString(fmt.Sprintf("%s,%s,%s,%.2f,%s,%s,%s,",
			product.ID.String(),
			strings.ReplaceAll(product.Name, ",", "\\,"),
			strings.ReplaceAll(product.Description, ",", "\\,"),
			product.Price,
			product.SKU,
			product.Status,
			product.Category.Name,
		))

		// Variants
		var variantStrs []string
		for _, variant := range product.Variants {
			variantStrs = append(variantStrs, fmt.Sprintf("%s:%s", variant.VariantName, variant.VariantValue))
		}
		csv.WriteString(strings.Join(variantStrs, ";"))
		csv.WriteString(",")

		// Images
		var imageStrs []string
		for _, image := range product.Images {
			imageStrs = append(imageStrs, image.URL)
		}
		csv.WriteString(strings.Join(imageStrs, ";"))
		csv.WriteString(",")

		// Inventory
		var inventoryStrs []string
		for _, inv := range product.Inventory {
			inventoryStrs = append(inventoryStrs, fmt.Sprintf("%d", inv.QuantityAvailable))
		}
		csv.WriteString(strings.Join(inventoryStrs, ";"))
		csv.WriteString("\n")
	}

	return []byte(csv.String())
}

// GetProductStats returns statistics about products
func (s *AdminProductService) GetProductStats() (map[string]interface{}, error) {
	stats := make(map[string]interface{})

	// Total products
	var totalProducts int64
	if err := s.db.Model(&models.Product{}).Count(&totalProducts).Error; err != nil {
		return nil, fmt.Errorf("failed to count products: %v", err)
	}
	stats["total_products"] = totalProducts

	// Active products
	var activeProducts int64
	if err := s.db.Model(&models.Product{}).Where("status = ?", "active").Count(&activeProducts).Error; err != nil {
		return nil, fmt.Errorf("failed to count active products: %v", err)
	}
	stats["active_products"] = activeProducts

	// Products by category
	var categoryStats []struct {
		CategoryName string `json:"category_name"`
		Count        int64  `json:"count"`
	}
	if err := s.db.Table("products").
		Select("categories.name as category_name, COUNT(products.id) as count").
		Joins("LEFT JOIN categories ON products.category_id = categories.id").
		Group("categories.name").
		Scan(&categoryStats).Error; err != nil {
		return nil, fmt.Errorf("failed to get category stats: %v", err)
	}
	stats["products_by_category"] = categoryStats

	// Recent products (last 30 days)
	var recentProducts int64
	thirtyDaysAgo := time.Now().AddDate(0, 0, -30)
	if err := s.db.Model(&models.Product{}).
		Where("created_at >= ?", thirtyDaysAgo).
		Count(&recentProducts).Error; err != nil {
		return nil, fmt.Errorf("failed to count recent products: %v", err)
	}
	stats["recent_products"] = recentProducts

	return stats, nil
}
