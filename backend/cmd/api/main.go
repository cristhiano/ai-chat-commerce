package main

import (
	"chat-ecommerce-backend/internal/handlers"
	"chat-ecommerce-backend/internal/middleware"
	"chat-ecommerce-backend/internal/services"
	"chat-ecommerce-backend/pkg/database"
	"log"
	"net/http"
	"os"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func main() {
	// Initialize database
	db, err := database.ConnectDatabase()
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Run migrations
	if err := database.MigrateDatabase(db); err != nil {
		log.Fatal("Failed to run migrations:", err)
	}

	// Seed database
	if err := database.SeedDatabase(db); err != nil {
		log.Fatal("Failed to seed database:", err)
	}

	// Set Gin mode
	if os.Getenv("ENVIRONMENT") == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// Create Gin router
	r := gin.Default()

	// Configure CORS
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{"*"}
	// config.AllowOriginFunc = func(origin string) bool {
	// 	u, err := url.Parse(origin)
	// 	if err != nil {
	// 		return false
	// 	}
	// 	switch u.Hostname() {
	// 	case "localhost", "127.0.0.1", "::1":
	// 		return true // accepts any port for these hosts
	// 	default:
	// 		return false
	// 	}
	// }
	config.AllowMethods = []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "X-Session-ID"}
	config.AllowCredentials = true
	r.Use(cors.New(config))

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "healthy",
			"message": "Chat Ecommerce API is running",
		})
	})

	// Initialize services and handlers
	productService := services.NewProductService(db)
	productHandler := handlers.NewProductHandler(productService)
	cartService := services.NewShoppingCartService(db)
	cartHandler := handlers.NewCartHandler(cartService)
	userService := services.NewUserService(db)
	userHandler := handlers.NewUserHandler(userService, os.Getenv("JWT_SECRET"))
	orderService := services.NewOrderService(db)
	orderHandler := handlers.NewOrderHandler(orderService)
	paymentService := services.NewPaymentService()
	paymentHandler := handlers.NewPaymentHandler(paymentService, orderService)
	chatService := services.NewChatService(db, productService, cartService)
	chatHandler := handlers.NewChatHandler(chatService)
	adminProductService := services.NewAdminProductService(db)
	inventoryService := services.NewInventoryService(db)
	alertService := services.NewAlertService(db)
	adminHandler := handlers.NewAdminHandler(adminProductService, productService)

	// API v1 routes
	v1 := r.Group("/api/v1")
	{
		// Public routes
		public := v1.Group("/")
		{
			// Product routes (public)
			products := public.Group("products")
			{
				products.GET("/", productHandler.GetProducts)
				products.GET("/:id", productHandler.GetProductByID)
				products.GET("/sku/:sku", productHandler.GetProductBySKU)
				products.GET("/search", productHandler.SearchProducts)
				products.GET("/featured", productHandler.GetFeaturedProducts)
				products.GET("/:id/related", productHandler.GetRelatedProducts)
			}

			// Category routes (public)
			categories := public.Group("categories")
			{
				categories.GET("/", productHandler.GetCategories)
				categories.GET("/:id", productHandler.GetCategoryByID)
				categories.GET("/slug/:slug", productHandler.GetCategoryBySlug)
			}

			// Auth routes (public)
			auth := public.Group("auth")
			{
				auth.POST("/register", userHandler.Register)
				auth.POST("/login", userHandler.Login)
				auth.POST("/refresh", userHandler.RefreshToken)
			}

			// Chat routes (public)
			chat := public.Group("chat")
			{
				chat.GET("/ws", chatHandler.HandleWebSocket)
				chat.POST("/message", chatHandler.SendMessage)
				chat.GET("/history/:session_id", chatHandler.GetChatHistory)
				chat.GET("/suggestions", chatHandler.GetProductSuggestions)
				chat.GET("/search", chatHandler.SearchProducts)
				chat.GET("/session/:session_id", chatHandler.GetChatSession)
			}

			// Payment webhook (public)
			payments := public.Group("payments")
			{
				payments.POST("/webhook", paymentHandler.HandleWebhook)
			}
		}

		// Protected routes
		protected := v1.Group("/")
		protected.Use(middleware.AuthMiddleware())
		{
			// User routes
			users := protected.Group("user")
			{
				users.GET("/profile", userHandler.GetProfile)
				users.PUT("/profile", userHandler.UpdateProfile)
				users.POST("/change-password", userHandler.ChangePassword)
				users.DELETE("/account", userHandler.DeleteAccount)
				users.POST("/verify-email", userHandler.VerifyEmail)
			}

			// Cart routes
			cart := protected.Group("cart")
			{
				cart.GET("/", cartHandler.GetCart)
				cart.POST("/add", cartHandler.AddToCart)
				cart.PUT("/update", cartHandler.UpdateCartItem)
				cart.DELETE("/remove/:product_id", cartHandler.RemoveFromCart)
				cart.DELETE("/clear", cartHandler.ClearCart)
				cart.POST("/calculate", cartHandler.CalculateTotals)
				cart.GET("/count", cartHandler.GetCartItemCount)
			}

			// Order routes
			orders := protected.Group("orders")
			{
				orders.POST("/", orderHandler.CreateOrder)
				orders.GET("/:id", orderHandler.GetOrder)
				orders.GET("/number/:number", orderHandler.GetOrderByNumber)
				orders.GET("/", orderHandler.GetUserOrders)
				orders.GET("/:id/summary", orderHandler.GetOrderSummary)
				orders.DELETE("/:id", orderHandler.CancelOrder)
			}

			// Payment routes
			payments := protected.Group("payments")
			{
				payments.POST("/create-intent", paymentHandler.CreatePaymentIntent)
				payments.POST("/confirm", paymentHandler.ConfirmPayment)
				payments.GET("/:payment_intent_id/status", paymentHandler.GetPaymentStatus)
				payments.POST("/:payment_intent_id/cancel", paymentHandler.CancelPayment)
				payments.POST("/:payment_intent_id/refund", paymentHandler.RefundPayment)
				payments.GET("/methods", paymentHandler.GetPaymentMethods)
			}
		}

		// Admin routes
		admin := v1.Group("/admin")
		admin.Use(middleware.AuthMiddleware())
		admin.Use(middleware.AdminMiddleware())
		{
			// Product management
			products := admin.Group("products")
			{
				products.POST("/", adminHandler.CreateProduct)
				products.GET("/", adminHandler.GetProducts)
				products.GET("/:id", adminHandler.GetProductWithDetails)
				products.PUT("/:id", adminHandler.UpdateProduct)
				products.DELETE("/:id", adminHandler.DeleteProduct)
				products.POST("/bulk-import", adminHandler.BulkImportProducts)
				products.GET("/export", adminHandler.ExportProducts)
				products.GET("/stats", adminHandler.GetProductStats)
			}

			// Category management
			categories := admin.Group("categories")
			{
				categories.GET("/", adminHandler.GetCategories)
				categories.POST("/", adminHandler.CreateCategory)
				categories.PUT("/:id", adminHandler.UpdateCategory)
				categories.DELETE("/:id", adminHandler.DeleteCategory)
			}

			// Inventory management
			inventory := admin.Group("inventory")
			{
				inventory.GET("/", func(c *gin.Context) {
					productIDStr := c.Query("product_id")
					var productID *uuid.UUID
					if productIDStr != "" {
						if id, err := uuid.Parse(productIDStr); err == nil {
							productID = &id
						}
					}

					levels, err := inventoryService.GetInventoryLevels(productID, nil)
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
						return
					}

					c.JSON(http.StatusOK, gin.H{"success": true, "data": levels})
				})

				inventory.POST("/update", func(c *gin.Context) {
					var req services.InventoryUpdateRequest
					if err := c.ShouldBindJSON(&req); err != nil {
						c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
						return
					}

					if err := inventoryService.UpdateInventory(req); err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
						return
					}

					c.JSON(http.StatusOK, gin.H{"success": true, "message": "Inventory updated successfully"})
				})

				inventory.GET("/report", func(c *gin.Context) {
					report, err := inventoryService.GetInventoryReport()
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
						return
					}

					c.JSON(http.StatusOK, gin.H{"success": true, "data": report})
				})
			}

			// Alert management
			alerts := admin.Group("alerts")
			{
				alerts.GET("/", func(c *gin.Context) {
					isReadStr := c.Query("is_read")
					var isRead *bool
					if isReadStr != "" {
						if isReadStr == "true" {
							read := true
							isRead = &read
						} else if isReadStr == "false" {
							read := false
							isRead = &read
						}
					}

					alerts, err := inventoryService.GetInventoryAlerts(isRead)
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
						return
					}

					c.JSON(http.StatusOK, gin.H{"success": true, "data": alerts})
				})

				alerts.POST("/mark-read", func(c *gin.Context) {
					var req struct {
						AlertIDs []uuid.UUID `json:"alert_ids" binding:"required"`
					}

					if err := c.ShouldBindJSON(&req); err != nil {
						c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
						return
					}

					if err := alertService.MarkAlertsAsRead(req.AlertIDs); err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
						return
					}

					c.JSON(http.StatusOK, gin.H{"success": true, "message": "Alerts marked as read"})
				})

				alerts.GET("/summary", func(c *gin.Context) {
					summary, err := alertService.GetAlertSummary()
					if err != nil {
						c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
						return
					}

					c.JSON(http.StatusOK, gin.H{"success": true, "data": summary})
				})
			}
		}
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server:", err)
	}
}

// Placeholder handlers - these will be implemented in the next phase
func getProducts(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get products - to be implemented"})
}

func getProduct(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get product - to be implemented"})
}

func searchProducts(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Search products - to be implemented"})
}

func getCategories(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get categories - to be implemented"})
}

func getCategory(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get category - to be implemented"})
}

func registerUser(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Register user - to be implemented"})
}

func loginUser(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Login user - to be implemented"})
}

func refreshToken(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Refresh token - to be implemented"})
}

func getUserProfile(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get user profile - to be implemented"})
}

func updateUserProfile(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Update user profile - to be implemented"})
}

func getUserOrders(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get user orders - to be implemented"})
}

func getCart(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get cart - to be implemented"})
}

func addToCart(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Add to cart - to be implemented"})
}

func updateCartItem(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Update cart item - to be implemented"})
}

func removeFromCart(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Remove from cart - to be implemented"})
}

func clearCart(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Clear cart - to be implemented"})
}

func createOrder(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Create order - to be implemented"})
}

func getOrder(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get order - to be implemented"})
}

func createProduct(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Create product - to be implemented"})
}

func updateProduct(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Update product - to be implemented"})
}

func deleteProduct(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Delete product - to be implemented"})
}

func getInventory(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get inventory - to be implemented"})
}

func updateInventory(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Update inventory - to be implemented"})
}

func adjustInventory(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Adjust inventory - to be implemented"})
}

func getAllOrders(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Get all orders - to be implemented"})
}

func updateOrderStatus(c *gin.Context) {
	c.JSON(200, gin.H{"message": "Update order status - to be implemented"})
}

func handleWebSocket(c *gin.Context) {
	c.JSON(200, gin.H{"message": "WebSocket handler - to be implemented"})
}
