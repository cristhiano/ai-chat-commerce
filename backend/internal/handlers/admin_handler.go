package handlers

import (
	"chat-ecommerce-backend/internal/services"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// AdminHandler handles admin-specific HTTP requests
type AdminHandler struct {
	adminProductService *services.AdminProductService
	productService      *services.ProductService
}

// NewAdminHandler creates a new AdminHandler
func NewAdminHandler(adminProductService *services.AdminProductService, productService *services.ProductService) *AdminHandler {
	return &AdminHandler{
		adminProductService: adminProductService,
		productService:      productService,
	}
}

// CreateProduct handles POST /api/v1/admin/products
func (h *AdminHandler) CreateProduct(c *gin.Context) {
	var req services.AdminProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.adminProductService.CreateProduct(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"data":    response,
	})
}

// UpdateProduct handles PUT /api/v1/admin/products/:id
func (h *AdminHandler) UpdateProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	var req services.AdminProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.adminProductService.UpdateProduct(id, req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// DeleteProduct handles DELETE /api/v1/admin/products/:id
func (h *AdminHandler) DeleteProduct(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	err = h.adminProductService.DeleteProduct(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": "Product deleted successfully",
	})
}

// GetProductWithDetails handles GET /api/v1/admin/products/:id
func (h *AdminHandler) GetProductWithDetails(c *gin.Context) {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid product ID"})
		return
	}

	response, err := h.adminProductService.GetProductWithDetails(id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// BulkImportProducts handles POST /api/v1/admin/products/bulk-import
func (h *AdminHandler) BulkImportProducts(c *gin.Context) {
	var req services.BulkImportRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	response, err := h.adminProductService.BulkImportProducts(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// ExportProducts handles GET /api/v1/admin/products/export
func (h *AdminHandler) ExportProducts(c *gin.Context) {
	// Parse query parameters
	filters := services.ProductFilters{
		Status: c.Query("status"),
		Search: c.Query("search"),
	}

	if categoryIDStr := c.Query("category_id"); categoryIDStr != "" {
		if categoryID, err := uuid.Parse(categoryIDStr); err == nil {
			filters.CategoryID = categoryID
		}
	}

	csvData, err := h.adminProductService.ExportProducts(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Header("Content-Type", "text/csv")
	c.Header("Content-Disposition", "attachment; filename=products.csv")
	c.Data(http.StatusOK, "text/csv", csvData)
}

// GetProductStats handles GET /api/v1/admin/products/stats
func (h *AdminHandler) GetProductStats(c *gin.Context) {
	stats, err := h.adminProductService.GetProductStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    stats,
	})
}

// GetProducts handles GET /api/v1/admin/products
func (h *AdminHandler) GetProducts(c *gin.Context) {
	// Parse query parameters
	filters := services.ProductFilters{
		Status:    c.Query("status"),
		Search:    c.Query("search"),
		SortBy:    c.Query("sort_by"),
		SortOrder: c.Query("sort_order"),
	}

	if pageStr := c.Query("page"); pageStr != "" {
		if page, err := strconv.Atoi(pageStr); err == nil {
			filters.Page = page
		}
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		if limit, err := strconv.Atoi(limitStr); err == nil {
			filters.Limit = limit
		}
	}

	if categoryIDStr := c.Query("category_id"); categoryIDStr != "" {
		if categoryID, err := uuid.Parse(categoryIDStr); err == nil {
			filters.CategoryID = categoryID
		}
	}

	if minPriceStr := c.Query("min_price"); minPriceStr != "" {
		if minPrice, err := strconv.ParseFloat(minPriceStr, 64); err == nil {
			filters.MinPrice = minPrice
		}
	}

	if maxPriceStr := c.Query("max_price"); maxPriceStr != "" {
		if maxPrice, err := strconv.ParseFloat(maxPriceStr, 64); err == nil {
			filters.MaxPrice = maxPrice
		}
	}

	response, err := h.productService.GetProducts(filters)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    response,
	})
}

// GetCategories handles GET /api/v1/admin/categories
func (h *AdminHandler) GetCategories(c *gin.Context) {
	categories, err := h.productService.GetCategories()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"data":    categories,
	})
}

// CreateCategory handles POST /api/v1/admin/categories
func (h *AdminHandler) CreateCategory(c *gin.Context) {
	var req struct {
		Name        string     `json:"name" binding:"required"`
		Description string     `json:"description"`
		ParentID    *uuid.UUID `json:"parent_id"`
		Slug        string     `json:"slug" binding:"required"`
		SortOrder   int        `json:"sort_order"`
		IsActive    bool       `json:"is_active"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// This would need to be implemented in ProductService
	// For now, return a placeholder response
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Category creation not yet implemented",
	})
}

// UpdateCategory handles PUT /api/v1/admin/categories/:id
func (h *AdminHandler) UpdateCategory(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Category update not yet implemented",
	})
}

// DeleteCategory handles DELETE /api/v1/admin/categories/:id
func (h *AdminHandler) DeleteCategory(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error": "Category deletion not yet implemented",
	})
}
