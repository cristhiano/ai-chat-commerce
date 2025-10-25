package database

import (
	"chat-ecommerce-backend/internal/models"
	"log"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// MigrateDatabase runs database migrations
func MigrateDatabase(db *gorm.DB) error {
	log.Println("Running database migrations...")

	err := db.AutoMigrate(
		&models.Category{},
		&models.Product{},
		&models.ProductVariant{},
		&models.ProductImage{},
		&models.Inventory{},
		&models.InventoryReservation{},
		&models.User{},
		&models.ChatSession{},
		&models.ShoppingCart{},
		&models.Order{},
		&models.OrderItem{},
	)

	if err != nil {
		return err
	}

	log.Println("Database migrations completed successfully")
	return nil
}

// SeedDatabase populates the database with initial data
func SeedDatabase(db *gorm.DB) error {
	log.Println("Seeding database...")

	// Check if products already exist
	var productCount int64
	db.Model(&models.Product{}).Count(&productCount)
	if productCount > 0 {
		log.Println("Database already seeded, skipping...")
		return nil
	}

	// Create categories
	categories := []models.Category{
		{
			ID:          uuid.New(),
			Name:        "Electronics",
			Description: "Electronic devices and gadgets",
			Slug:        "electronics",
			SortOrder:   1,
			IsActive:    true,
			CreatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Name:        "Clothing",
			Description: "Fashion and apparel",
			Slug:        "clothing",
			SortOrder:   2,
			IsActive:    true,
			CreatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Name:        "Books",
			Description: "Books and literature",
			Slug:        "books",
			SortOrder:   3,
			IsActive:    true,
			CreatedAt:   time.Now(),
		},
		{
			ID:          uuid.New(),
			Name:        "Home & Garden",
			Description: "Home improvement and garden supplies",
			Slug:        "home-garden",
			SortOrder:   4,
			IsActive:    true,
			CreatedAt:   time.Now(),
		},
	}

	for _, category := range categories {
		if err := db.Create(&category).Error; err != nil {
			return err
		}
	}

	// Create sample products
	products := []models.Product{
		{
			ID:          uuid.New(),
			Name:        "Wireless Bluetooth Headphones",
			Description: "High-quality wireless headphones with noise cancellation",
			Price:       99.99,
			CategoryID:  categories[0].ID,
			SKU:         "WBH-001",
			Status:      "active",
			// Tags:        pq.StringArray{"wireless", "bluetooth", "headphones", "audio"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          uuid.New(),
			Name:        "Cotton T-Shirt",
			Description: "Comfortable cotton t-shirt in various sizes and colors",
			Price:       19.99,
			CategoryID:  categories[1].ID,
			SKU:         "CTS-001",
			Status:      "active",
			// Tags:        pq.StringArray{"cotton", "t-shirt", "clothing", "casual"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          uuid.New(),
			Name:        "Programming Book",
			Description: "Comprehensive guide to modern programming practices",
			Price:       49.99,
			CategoryID:  categories[2].ID,
			SKU:         "PB-001",
			Status:      "active",
			// Tags:        pq.StringArray{"programming", "book", "education", "technology"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
		{
			ID:          uuid.New(),
			Name:        "Garden Tool Set",
			Description: "Complete set of essential garden tools",
			Price:       79.99,
			CategoryID:  categories[3].ID,
			SKU:         "GTS-001",
			Status:      "active",
			// Tags:        pq.StringArray{"garden", "tools", "outdoor", "gardening"},
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		},
	}

	for _, product := range products {
		if err := db.Create(&product).Error; err != nil {
			return err
		}
	}

	// Create product variants
	variants := []models.ProductVariant{
		{
			ID:            uuid.New(),
			ProductID:     products[0].ID,
			VariantName:   "Color",
			VariantValue:  "Black",
			PriceModifier: 0,
			SKUSuffix:     "BLK",
			IsDefault:     true,
			CreatedAt:     time.Now(),
		},
		{
			ID:            uuid.New(),
			ProductID:     products[0].ID,
			VariantName:   "Color",
			VariantValue:  "White",
			PriceModifier: 0,
			SKUSuffix:     "WHT",
			IsDefault:     false,
			CreatedAt:     time.Now(),
		},
		{
			ID:            uuid.New(),
			ProductID:     products[1].ID,
			VariantName:   "Size",
			VariantValue:  "Small",
			PriceModifier: 0,
			SKUSuffix:     "S",
			IsDefault:     false,
			CreatedAt:     time.Now(),
		},
		{
			ID:            uuid.New(),
			ProductID:     products[1].ID,
			VariantName:   "Size",
			VariantValue:  "Medium",
			PriceModifier: 0,
			SKUSuffix:     "M",
			IsDefault:     true,
			CreatedAt:     time.Now(),
		},
		{
			ID:            uuid.New(),
			ProductID:     products[1].ID,
			VariantName:   "Size",
			VariantValue:  "Large",
			PriceModifier: 0,
			SKUSuffix:     "L",
			IsDefault:     false,
			CreatedAt:     time.Now(),
		},
	}

	for _, variant := range variants {
		if err := db.Create(&variant).Error; err != nil {
			return err
		}
	}

	// Create inventory records
	inventory := []models.Inventory{
		{
			ID:                uuid.New(),
			ProductID:         products[0].ID,
			VariantID:         &variants[0].ID,
			WarehouseLocation: "Main Warehouse",
			QuantityAvailable: 50,
			QuantityReserved:  0,
			LowStockThreshold: 10,
			ReorderPoint:      5,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		},
		{
			ID:                uuid.New(),
			ProductID:         products[0].ID,
			VariantID:         &variants[1].ID,
			WarehouseLocation: "Main Warehouse",
			QuantityAvailable: 30,
			QuantityReserved:  0,
			LowStockThreshold: 10,
			ReorderPoint:      5,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		},
		{
			ID:                uuid.New(),
			ProductID:         products[1].ID,
			VariantID:         &variants[2].ID,
			WarehouseLocation: "Main Warehouse",
			QuantityAvailable: 100,
			QuantityReserved:  0,
			LowStockThreshold: 20,
			ReorderPoint:      10,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		},
		{
			ID:                uuid.New(),
			ProductID:         products[1].ID,
			VariantID:         &variants[3].ID,
			WarehouseLocation: "Main Warehouse",
			QuantityAvailable: 150,
			QuantityReserved:  0,
			LowStockThreshold: 20,
			ReorderPoint:      10,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		},
		{
			ID:                uuid.New(),
			ProductID:         products[1].ID,
			VariantID:         &variants[4].ID,
			WarehouseLocation: "Main Warehouse",
			QuantityAvailable: 80,
			QuantityReserved:  0,
			LowStockThreshold: 20,
			ReorderPoint:      10,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		},
		{
			ID:                uuid.New(),
			ProductID:         products[2].ID,
			WarehouseLocation: "Main Warehouse",
			QuantityAvailable: 25,
			QuantityReserved:  0,
			LowStockThreshold: 5,
			ReorderPoint:      2,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		},
		{
			ID:                uuid.New(),
			ProductID:         products[3].ID,
			WarehouseLocation: "Main Warehouse",
			QuantityAvailable: 15,
			QuantityReserved:  0,
			LowStockThreshold: 3,
			ReorderPoint:      1,
			CreatedAt:         time.Now(),
			UpdatedAt:         time.Now(),
		},
	}

	for _, inv := range inventory {
		if err := db.Create(&inv).Error; err != nil {
			return err
		}
	}

	log.Println("Database seeded successfully")
	return nil
}
