package websocket

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/google/uuid"
)

// AuthLevel represents the authentication level
type AuthLevel int

const (
	AuthLevelAnonymous AuthLevel = iota
	AuthLevelAuthenticated
	AuthLevelAdmin
)

// AuthResult represents the result of an authentication check
type AuthResult struct {
	IsValid      bool
	UserID       *uuid.UUID
	SessionID    string
	AuthLevel    AuthLevel
	Permissions  []string
	ExpiresAt    time.Time
	Error        error
}

// WebSocketAuthManager manages WebSocket authentication and authorization
type WebSocketAuthManager struct {
	// JWT secret for token validation
	jwtSecret string
	
	// Session storage for authenticated users
	authSessions map[string]*AuthSession
	
	// Mutex for thread-safe operations
	mu sync.RWMutex
	
	// Configuration
	tokenExpiry    time.Duration
	sessionTimeout time.Duration
	cleanupInterval time.Duration
	
	// Context for cancellation
	ctx context.Context
	cancel context.CancelFunc
	
	// Cleanup control
	cleanupRunning bool
}

// AuthSession represents an authenticated session
type AuthSession struct {
	SessionID    string
	UserID       uuid.UUID
	AuthLevel    AuthLevel
	Permissions  []string
	CreatedAt    time.Time
	LastActivity time.Time
	ExpiresAt    time.Time
	IsActive     bool
	Metadata     map[string]interface{}
}

// Permission represents a permission for authorization
type Permission string

const (
	PermissionReadCart      Permission = "cart:read"
	PermissionWriteCart     Permission = "cart:write"
	PermissionReadInventory Permission = "inventory:read"
	PermissionWriteInventory Permission = "inventory:write"
	PermissionReadOrders    Permission = "orders:read"
	PermissionWriteOrders   Permission = "orders:write"
	PermissionAdminAccess   Permission = "admin:access"
	PermissionChatAccess   Permission = "chat:access"
)

// NewWebSocketAuthManager creates a new WebSocket authentication manager
func NewWebSocketAuthManager(jwtSecret string, tokenExpiry, sessionTimeout, cleanupInterval time.Duration) *WebSocketAuthManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &WebSocketAuthManager{
		jwtSecret:       jwtSecret,
		authSessions:    make(map[string]*AuthSession),
		tokenExpiry:     tokenExpiry,
		sessionTimeout:  sessionTimeout,
		cleanupInterval: cleanupInterval,
		ctx:             ctx,
		cancel:          cancel,
	}
}

// AuthenticateToken authenticates a JWT token and returns auth result
func (wam *WebSocketAuthManager) AuthenticateToken(token string) (*AuthResult, error) {
	if token == "" {
		return &AuthResult{
			IsValid:   true,
			AuthLevel: AuthLevelAnonymous,
		}, nil
	}
	
	// Parse and validate JWT token
	claims, err := wam.parseJWTToken(token)
	if err != nil {
		return &AuthResult{
			IsValid: false,
			Error:   fmt.Errorf("invalid token: %w", err),
		}, nil
	}
	
	// Extract user information from claims
	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return &AuthResult{
			IsValid: false,
			Error:   fmt.Errorf("missing user_id in token"),
		}, nil
	}
	
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return &AuthResult{
			IsValid: false,
			Error:   fmt.Errorf("invalid user_id format: %w", err),
		}, nil
	}
	
	// Determine auth level
	authLevel := AuthLevelAuthenticated
	if role, ok := claims["role"].(string); ok && role == "admin" {
		authLevel = AuthLevelAdmin
	}
	
	// Extract permissions
	permissions := wam.extractPermissions(claims)
	
	// Extract session ID
	sessionID, _ := claims["session_id"].(string)
	if sessionID == "" {
		sessionID = uuid.New().String()
	}
	
	// Calculate expiration
	expiresAt := time.Now().Add(wam.tokenExpiry)
	if exp, ok := claims["exp"].(float64); ok {
		expiresAt = time.Unix(int64(exp), 0)
	}
	
	return &AuthResult{
		IsValid:     true,
		UserID:      &userID,
		SessionID:   sessionID,
		AuthLevel:   authLevel,
		Permissions: permissions,
		ExpiresAt:   expiresAt,
	}, nil
}

// CreateAuthSession creates an authenticated session
func (wam *WebSocketAuthManager) CreateAuthSession(userID uuid.UUID, sessionID string, authLevel AuthLevel, permissions []string) (*AuthSession, error) {
	wam.mu.Lock()
	defer wam.mu.Unlock()
	
	now := time.Now()
	session := &AuthSession{
		SessionID:    sessionID,
		UserID:       userID,
		AuthLevel:    authLevel,
		Permissions:  permissions,
		CreatedAt:    now,
		LastActivity: now,
		ExpiresAt:    now.Add(wam.sessionTimeout),
		IsActive:     true,
		Metadata:     make(map[string]interface{}),
	}
	
	wam.authSessions[sessionID] = session
	
	// Start cleanup if not already running
	if !wam.cleanupRunning {
		go wam.startCleanup()
		wam.cleanupRunning = true
	}
	
	log.Printf("Auth session created: %s (User: %s, Level: %d)", 
		sessionID, userID, authLevel)
	
	return session, nil
}

// GetAuthSession retrieves an authenticated session
func (wam *WebSocketAuthManager) GetAuthSession(sessionID string) (*AuthSession, bool) {
	wam.mu.RLock()
	defer wam.mu.RUnlock()
	
	session, exists := wam.authSessions[sessionID]
	if !exists || !session.IsActive {
		return nil, false
	}
	
	// Return a copy to prevent external modifications
	sessionCopy := *session
	return &sessionCopy, true
}

// UpdateSessionActivity updates the last activity time for a session
func (wam *WebSocketAuthManager) UpdateSessionActivity(sessionID string) error {
	wam.mu.Lock()
	defer wam.mu.Unlock()
	
	session, exists := wam.authSessions[sessionID]
	if !exists {
		return fmt.Errorf("auth session not found: %s", sessionID)
	}
	
	if !session.IsActive {
		return fmt.Errorf("auth session is not active: %s", sessionID)
	}
	
	session.LastActivity = time.Now()
	session.ExpiresAt = time.Now().Add(wam.sessionTimeout)
	
	return nil
}

// CheckPermission checks if a session has a specific permission
func (wam *WebSocketAuthManager) CheckPermission(sessionID string, permission Permission) bool {
	wam.mu.RLock()
	defer wam.mu.RUnlock()
	
	session, exists := wam.authSessions[sessionID]
	if !exists || !session.IsActive {
		return false
	}
	
	for _, perm := range session.Permissions {
		if Permission(perm) == permission {
			return true
		}
	}
	
	return false
}

// CheckAuthLevel checks if a session meets the minimum auth level
func (wam *WebSocketAuthManager) CheckAuthLevel(sessionID string, requiredLevel AuthLevel) bool {
	wam.mu.RLock()
	defer wam.mu.RUnlock()
	
	session, exists := wam.authSessions[sessionID]
	if !exists || !session.IsActive {
		return requiredLevel == AuthLevelAnonymous
	}
	
	return session.AuthLevel >= requiredLevel
}

// RevokeSession revokes an authenticated session
func (wam *WebSocketAuthManager) RevokeSession(sessionID string) error {
	wam.mu.Lock()
	defer wam.mu.Unlock()
	
	session, exists := wam.authSessions[sessionID]
	if !exists {
		return fmt.Errorf("auth session not found: %s", sessionID)
	}
	
	session.IsActive = false
	
	log.Printf("Auth session revoked: %s (User: %s)", sessionID, session.UserID)
	
	return nil
}

// RevokeUserSessions revokes all sessions for a specific user
func (wam *WebSocketAuthManager) RevokeUserSessions(userID uuid.UUID) error {
	wam.mu.Lock()
	defer wam.mu.Unlock()
	
	var revokedSessions []string
	
	for sessionID, session := range wam.authSessions {
		if session.UserID == userID && session.IsActive {
			session.IsActive = false
			revokedSessions = append(revokedSessions, sessionID)
		}
	}
	
	log.Printf("Revoked %d sessions for user %s", len(revokedSessions), userID)
	
	return nil
}

// GetUserSessions returns all active sessions for a user
func (wam *WebSocketAuthManager) GetUserSessions(userID uuid.UUID) []*AuthSession {
	wam.mu.RLock()
	defer wam.mu.RUnlock()
	
	var sessions []*AuthSession
	for _, session := range wam.authSessions {
		if session.UserID == userID && session.IsActive {
			sessions = append(sessions, session)
		}
	}
	return sessions
}

// GetActiveSessionCount returns the number of active authenticated sessions
func (wam *WebSocketAuthManager) GetActiveSessionCount() int {
	wam.mu.RLock()
	defer wam.mu.RUnlock()
	
	count := 0
	for _, session := range wam.authSessions {
		if session.IsActive {
			count++
		}
	}
	return count
}

// startCleanup starts the cleanup routine
func (wam *WebSocketAuthManager) startCleanup() {
	ticker := time.NewTicker(wam.cleanupInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-wam.ctx.Done():
			return
		case <-ticker.C:
			wam.cleanupExpiredSessions()
		}
	}
}

// cleanupExpiredSessions removes expired sessions
func (wam *WebSocketAuthManager) cleanupExpiredSessions() {
	wam.mu.Lock()
	defer wam.mu.Unlock()
	
	now := time.Now()
	var expiredSessions []string
	
	for sessionID, session := range wam.authSessions {
		if session.ExpiresAt.Before(now) {
			expiredSessions = append(expiredSessions, sessionID)
		}
	}
	
	for _, sessionID := range expiredSessions {
		session := wam.authSessions[sessionID]
		session.IsActive = false
		
		log.Printf("Cleaned up expired auth session: %s (User: %s)", 
			sessionID, session.UserID)
	}
	
	if len(expiredSessions) > 0 {
		log.Printf("Cleaned up %d expired auth sessions", len(expiredSessions))
	}
}

// parseJWTToken parses and validates a JWT token
func (wam *WebSocketAuthManager) parseJWTToken(token string) (map[string]interface{}, error) {
	// This is a simplified JWT parser - in production, use a proper JWT library
	// like github.com/golang-jwt/jwt/v4
	
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, fmt.Errorf("invalid token format")
	}
	
	// For now, return a mock claims map
	// In production, you would decode and verify the JWT signature
	return map[string]interface{}{
		"user_id":    "123e4567-e89b-12d3-a456-426614174000", // Mock user ID
		"role":       "user",
		"session_id": "session-123",
		"exp":        float64(time.Now().Add(time.Hour).Unix()),
		"permissions": []string{"cart:read", "cart:write", "chat:access"},
	}, nil
}

// extractPermissions extracts permissions from JWT claims
func (wam *WebSocketAuthManager) extractPermissions(claims map[string]interface{}) []string {
	permissions := make([]string, 0)
	
	if perms, ok := claims["permissions"].([]interface{}); ok {
		for _, perm := range perms {
			if permStr, ok := perm.(string); ok {
				permissions = append(permissions, permStr)
			}
		}
	}
	
	// Add default permissions based on role
	if role, ok := claims["role"].(string); ok {
		switch role {
		case "admin":
			permissions = append(permissions, string(PermissionAdminAccess))
			permissions = append(permissions, string(PermissionWriteInventory))
		case "user":
			permissions = append(permissions, string(PermissionReadCart))
			permissions = append(permissions, string(PermissionWriteCart))
			permissions = append(permissions, string(PermissionChatAccess))
		}
	}
	
	return permissions
}

// GetAuthStats returns authentication statistics
func (wam *WebSocketAuthManager) GetAuthStats() map[string]interface{} {
	wam.mu.RLock()
	defer wam.mu.RUnlock()
	
	activeCount := 0
	userCount := make(map[string]int)
	levelCount := make(map[AuthLevel]int)
	
	for _, session := range wam.authSessions {
		if session.IsActive {
			activeCount++
			userCount[session.UserID.String()]++
			levelCount[session.AuthLevel]++
		}
	}
	
	return map[string]interface{}{
		"total_sessions":      len(wam.authSessions),
		"active_sessions":     activeCount,
		"unique_users":        len(userCount),
		"auth_level_distribution": levelCount,
		"token_expiry_s":      wam.tokenExpiry.Seconds(),
		"session_timeout_s":   wam.sessionTimeout.Seconds(),
		"cleanup_interval_s":  wam.cleanupInterval.Seconds(),
	}
}

// Stop stops the authentication manager
func (wam *WebSocketAuthManager) Stop() {
	wam.cancel()
}
