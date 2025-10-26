package auth

import (
	"net/http"

	"chat-ecommerce-backend/internal/services/auth"

	"github.com/gin-gonic/gin"
)

// LogoutHandler handles user logout requests
type LogoutHandler struct {
	authService *auth.Service
}

// NewLogoutHandler creates a new logout handler
func NewLogoutHandler(authService *auth.Service) *LogoutHandler {
	return &LogoutHandler{
		authService: authService,
	}
}

// Logout handles POST /api/auth/logout
func (h *LogoutHandler) Logout(c *gin.Context) {
	// Get token from Authorization header
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "NO_TOKEN",
				"message": "Authorization token required",
			},
		})
		return
	}

	// Extract token from "Bearer <token>"
	token := ""
	if len(authHeader) > 7 && authHeader[:7] == "Bearer " {
		token = authHeader[7:]
	} else {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "INVALID_TOKEN",
				"message": "Invalid authorization format",
			},
		})
		return
	}

	// Logout user
	response, err := h.authService.Logout(token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "LOGOUT_ERROR",
				"message": "Failed to logout",
			},
		})
		return
	}

	// Success
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"message": response.Message,
	})
}
