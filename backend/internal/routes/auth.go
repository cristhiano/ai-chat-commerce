package routes

import (
	authhandlers "chat-ecommerce-backend/internal/handlers/auth"
	authservices "chat-ecommerce-backend/internal/services/auth"
	passwordservices "chat-ecommerce-backend/internal/services/password"
	sessionservices "chat-ecommerce-backend/internal/services/session"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// SetupAuthRoutes sets up authentication-related routes
func SetupAuthRoutes(router *gin.Engine, db *gorm.DB, jwtSecret string) {
	// Initialize services
	passwordService := passwordservices.NewService(10) // bcrypt cost 10
	passwordValidator := passwordservices.NewValidator()
	sessionService, _ := sessionservices.NewService(jwtSecret)
	sessionValidator, _ := sessionservices.NewValidator(jwtSecret)

	authService := authservices.NewService(
		db,
		passwordService,
		passwordValidator,
		sessionService,
		sessionValidator,
	)

	// Initialize handlers
	registerHandler := authhandlers.NewRegisterHandler(authService)
	loginHandler := authhandlers.NewLoginHandler(authService)
	logoutHandler := authhandlers.NewLogoutHandler(authService)

	// Auth API group
	api := router.Group("/api")
	{
		auth := api.Group("/auth")
		{
			// Registration endpoint
			auth.POST("/register", registerHandler.Register)

			// Login endpoint
			auth.POST("/login", loginHandler.Login)

			// Logout endpoint
			auth.POST("/logout", logoutHandler.Logout)
		}
	}
}
