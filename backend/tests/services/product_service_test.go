package services

import (
	"chat-ecommerce-backend/internal/models"
	"chat-ecommerce-backend/internal/services"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal("Failed to connect to test database:", err)
	}

	// Auto-migrate the database
	err = db.AutoMigrate(
		&models.Category{},
		&models.Product{},
		&models.ProductVariant{},
		&models.ProductImage{},
		&models.Inventory{},
	)
	if err != nil {
		t.Fatal("Failed to migrate test database:", err)
	}

	return db
}

func TestProductService_CreateProduct(t *testing.T) {
	db := setupTestDB(t)
	service := services.NewProductService(db)

	// Create test category
	category := &models.Category{
		ID:   uuid.New(),
		Name: "Test Category",
		Slug: "test-category",
	}
	db.Create(category)

	// Test product
	product := &models.Product{
		Name:        "Test Product",
		Description: "Test Description",
		Price:       99.99,
		CategoryID:  category.ID,
		SKU:         "TEST-001",
		Status:      "active",
	}

	err := service.CreateProduct(product)
	assert.NoError(t, err)

	// Verify product was created
	var createdProduct models.Product
	err = db.Where("sku = ?", "TEST-001").First(&createdProduct).Error
	assert.NoError(t, err)
	assert.Equal(t, "Test Product", createdProduct.Name)
	assert.Equal(t, 99.99, createdProduct.Price)
}

func TestProductService_GetProductByID(t *testing.T) {
	db := setupTestDB(t)
	service := services.NewProductService(db)

	// Create test category
	category := &models.Category{
		ID:   uuid.New(),
		Name: "Test Category",
		Slug: "test-category",
	}
	db.Create(category)

	// Create test product
	product := &models.Product{
		ID:          uuid.New(),
		Name:        "Test Product",
		Description: "Test Description",
		Price:       99.99,
		CategoryID:  category.ID,
		SKU:         "TEST-001",
		Status:      "active",
	}
	db.Create(product)

	// Test getting product by ID
	result, err := service.GetProductByID(product.ID)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, product.ID, result.ID)
	assert.Equal(t, "Test Product", result.Name)
}

func TestProductService_GetProducts(t *testing.T) {
	db := setupTestDB(t)
	service := services.NewProductService(db)

	// Create test category
	category := &models.Category{
		ID:   uuid.New(),
		Name: "Test Category",
		Slug: "test-category",
	}
	db.Create(category)

	// Create test products
	products := []models.Product{
		{
			ID:          uuid.New(),
			Name:        "Product 1",
			Description: "Description 1",
			Price:       99.99,
			CategoryID:  category.ID,
			SKU:         "TEST-001",
			Status:      "active",
		},
		{
			ID:          uuid.New(),
			Name:        "Product 2",
			Description: "Description 2",
			Price:       149.99,
			CategoryID:  category.ID,
			SKU:         "TEST-002",
			Status:      "active",
		},
	}
	for _, product := range products {
		db.Create(&product)
	}

	// Test getting products
	filters := services.ProductFilters{
		Status: "active",
		Page:   1,
		Limit:  10,
	}

	result, err := service.GetProducts(filters)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, int64(2), result.Total)
	assert.Len(t, result.Products, 2)
}

func TestProductService_UpdateProduct(t *testing.T) {
	db := setupTestDB(t)
	service := services.NewProductService(db)

	// Create test category
	category := &models.Category{
		ID:   uuid.New(),
		Name: "Test Category",
		Slug: "test-category",
	}
	db.Create(category)

	// Create test product
	product := &models.Product{
		ID:          uuid.New(),
		Name:        "Test Product",
		Description: "Test Description",
		Price:       99.99,
		CategoryID:  category.ID,
		SKU:         "TEST-001",
		Status:      "active",
	}
	db.Create(product)

	// Test updating product
	updates := map[string]interface{}{
		"name":        "Updated Product",
		"description": "Updated Description",
		"price":       149.99,
	}

	err := service.UpdateProduct(product.ID, updates)
	assert.NoError(t, err)

	// Verify product was updated
	var updatedProduct models.Product
	err = db.Where("id = ?", product.ID).First(&updatedProduct).Error
	assert.NoError(t, err)
	assert.Equal(t, "Updated Product", updatedProduct.Name)
	assert.Equal(t, "Updated Description", updatedProduct.Description)
	assert.Equal(t, 149.99, updatedProduct.Price)
}

func TestProductService_DeleteProduct(t *testing.T) {
	db := setupTestDB(t)
	service := services.NewProductService(db)

	// Create test category
	category := &models.Category{
		ID:   uuid.New(),
		Name: "Test Category",
		Slug: "test-category",
	}
	db.Create(category)

	// Create test product
	product := &models.Product{
		ID:          uuid.New(),
		Name:        "Test Product",
		Description: "Test Description",
		Price:       99.99,
		CategoryID:  category.ID,
		SKU:         "TEST-001",
		Status:      "active",
	}
	db.Create(product)

	// Test deleting product
	err := service.DeleteProduct(product.ID)
	assert.NoError(t, err)

	// Verify product was deleted
	var deletedProduct models.Product
	err = db.Where("id = ?", product.ID).First(&deletedProduct).Error
	assert.Error(t, err)
	assert.Equal(t, gorm.ErrRecordNotFound, err)
}

func TestProductService_SearchProducts(t *testing.T) {
	db := setupTestDB(t)
	service := services.NewProductService(db)

	// Create test category
	category := &models.Category{
		ID:   uuid.New(),
		Name: "Test Category",
		Slug: "test-category",
	}
	db.Create(category)

	// Create test products
	products := []models.Product{
		{
			ID:          uuid.New(),
			Name:        "Test Product",
			Description: "Test Description",
			Price:       99.99,
			CategoryID:  category.ID,
			SKU:         "TEST-001",
			Status:      "active",
		},
		{
			ID:          uuid.New(),
			Name:        "Another Product",
			Description: "Another Description",
			Price:       149.99,
			CategoryID:  category.ID,
			SKU:         "TEST-002",
			Status:      "active",
		},
	}
	for _, product := range products {
		db.Create(&product)
	}

	// Test searching products
	result, err := service.SearchProducts("Test", 10)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, "Test Product", result[0].Name)
}

func TestProductService_GetFeaturedProducts(t *testing.T) {
	db := setupTestDB(t)
	service := services.NewProductService(db)

	// Create test category
	category := &models.Category{
		ID:   uuid.New(),
		Name: "Test Category",
		Slug: "test-category",
	}
	db.Create(category)

	// Create test products
	products := []models.Product{
		{
			ID:          uuid.New(),
			Name:        "Product 1",
			Description: "Description 1",
			Price:       99.99,
			CategoryID:  category.ID,
			SKU:         "TEST-001",
			Status:      "active",
		},
		{
			ID:          uuid.New(),
			Name:        "Product 2",
			Description: "Description 2",
			Price:       149.99,
			CategoryID:  category.ID,
			SKU:         "TEST-002",
			Status:      "active",
		},
	}
	for _, product := range products {
		db.Create(&product)
	}

	// Test getting featured products
	result, err := service.GetFeaturedProducts(5)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
}

func TestProductService_GetRelatedProducts(t *testing.T) {
	db := setupTestDB(t)
	service := services.NewProductService(db)

	// Create test category
	category := &models.Category{
		ID:   uuid.New(),
		Name: "Test Category",
		Slug: "test-category",
	}
	db.Create(category)

	// Create test products
	products := []models.Product{
		{
			ID:          uuid.New(),
			Name:        "Product 1",
			Description: "Description 1",
			Price:       99.99,
			CategoryID:  category.ID,
			SKU:         "TEST-001",
			Status:      "active",
		},
		{
			ID:          uuid.New(),
			Name:        "Product 2",
			Description: "Description 2",
			Price:       149.99,
			CategoryID:  category.ID,
			SKU:         "TEST-002",
			Status:      "active",
		},
	}
	for _, product := range products {
		db.Create(&product)
	}

	// Test getting related products
	result, err := service.GetRelatedProducts(products[0].ID, 5)
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 1)
	assert.Equal(t, "Product 2", result[0].Name)
}
