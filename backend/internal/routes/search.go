package routes

import (
	searchhandlers "chat-ecommerce-backend/internal/handlers/search"
	searchservices "chat-ecommerce-backend/internal/services/search"

	"github.com/gin-gonic/gin"
)

// SetupSearchRoutes sets up search-related routes
func SetupSearchRoutes(router *gin.Engine, searchService *searchservices.Service) {
	searchHandler := searchhandlers.NewHandler(searchService)

	// Search API group
	api := router.Group("/api")
	{
		search := api.Group("/search")
		{
			// Main search endpoint
			search.POST("", searchHandler.SearchProducts)

			// Autocomplete suggestions
			search.GET("/suggestions", searchHandler.GetSuggestions)

			// Filter options
			search.GET("/filters", searchHandler.GetFilterOptions)

			// Analytics logging
			search.POST("/analytics", searchHandler.LogAnalytics)
		}
	}
}
