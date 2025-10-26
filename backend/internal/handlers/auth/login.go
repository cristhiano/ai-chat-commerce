package auth

import (
	"net/http"
	"strings"

	"chat-ecommerce-backend/internal/services/auth"

	"github.com/gin-gonic/gin"
)

// LoginHandler handles user login requests
type LoginHandler struct {
	authService *auth.Service
}

// NewLoginHandler creates a new login handler
func NewLoginHandler(authService *auth.Service) *LoginHandler {
	return &LoginHandler{
		authService: authService,
	}
}

// Login handles POST /api/auth/login
func (h *LoginHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	// Bind and validate request
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": gin.H{
				"code":    "VALIDATION_ERROR",
				"message": "Invalid request data",
				"details": err.Error(),
			},
		})
		return
	}

	// Login user
	response, err := h.authService.Login(&auth.LoginRequest{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		// Check if account is locked
		if strings.Contains(err.Error(), "account is locked") {
			// Extract remaining minutes (basic parsing)
			// In production, this would be more robust
			c.JSON(http.StatusLocked, gin.H{
				"error": gin.H{
					"code":    "ACCOUNT_LOCKED",
					"message": "Account temporarily locked due to multiple failed login attempts",
				},
			})
			return
		}

		// Generic error for invalid credentials (security: don't reveal if email exists)
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": gin.H{
				"code":    "INVALID_CREDENTIALS",
				"message": "Invalid email or password",
			},
		})
		return
	}

	// Success
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"token":   response.Token,
		"user": gin.H{
			"id":    response.UserID,
			"email": response.Email,
		},
		"expires_at": response.ExpiresAt,
	})
}
