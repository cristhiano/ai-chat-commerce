package auth

import (
	"net/http"

	"chat-ecommerce-backend/internal/services/auth"

	"github.com/gin-gonic/gin"
)

// RegisterHandler handles user registration requests
type RegisterHandler struct {
	authService *auth.Service
}

// NewRegisterHandler creates a new registration handler
func NewRegisterHandler(authService *auth.Service) *RegisterHandler {
	return &RegisterHandler{
		authService: authService,
	}
}

// Register handles POST /api/auth/register
func (h *RegisterHandler) Register(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
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

	// Register user
	response, err := h.authService.Register(&auth.RegisterRequest{
		Email:    req.Email,
		Password: req.Password,
	})

	if err != nil {
		// Check if email already exists
		if err.Error() == "email already exists" {
			c.JSON(http.StatusConflict, gin.H{
				"error": gin.H{
					"code":    "EMAIL_EXISTS",
					"message": "Email already registered",
				},
			})
			return
		}

		// Check if password validation failed
		if err.Error()[:26] == "password validation failed" {
			c.JSON(http.StatusBadRequest, gin.H{
				"error": gin.H{
					"code":    "PASSWORD_VALIDATION",
					"message": "Password does not meet requirements",
					"details": err.Error(),
				},
			})
			return
		}

		// Generic error
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": gin.H{
				"code":    "INTERNAL_ERROR",
				"message": "Failed to create account",
			},
		})
		return
	}

	// Success
	c.JSON(http.StatusCreated, gin.H{
		"success": true,
		"user_id": response.UserID,
		"message": response.Message,
	})
}
