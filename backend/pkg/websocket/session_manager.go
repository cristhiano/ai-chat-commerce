package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/google/uuid"
)

// SessionManager manages WebSocket sessions and their lifecycle
type SessionManager struct {
	// Active sessions
	sessions map[string]*SessionInfo
	
	// Mutex for thread-safe operations
	mu sync.RWMutex
	
	// Configuration
	sessionTimeout    time.Duration
	cleanupInterval   time.Duration
	maxSessions       int
	
	// Context for cancellation
	ctx context.Context
	cancel context.CancelFunc
	
	// Cleanup control
	cleanupRunning bool
}

// SessionInfo represents information about a WebSocket session
type SessionInfo struct {
	ID                string                 `json:"id"`
	UserID            *uuid.UUID             `json:"user_id"`
	CreatedAt         time.Time              `json:"created_at"`
	LastActivity      time.Time              `json:"last_activity"`
	ExpiresAt         time.Time              `json:"expires_at"`
	IsActive          bool                   `json:"is_active"`
	ConnectionCount   int                    `json:"connection_count"`
	Metadata          map[string]interface{} `json:"metadata"`
	ClientIDs         []string               `json:"client_ids"`
}

// SessionContext represents the context of a WebSocket session
type SessionContext struct {
	SessionID     string                 `json:"session_id"`
	UserID        *uuid.UUID             `json:"user_id"`
	CartState     map[string]interface{} `json:"cart_state"`
	Preferences   map[string]interface{} `json:"preferences"`
	LastActivity  time.Time              `json:"last_activity"`
	Metadata      map[string]interface{} `json:"metadata"`
}

// NewSessionManager creates a new session manager
func NewSessionManager(sessionTimeout, cleanupInterval time.Duration, maxSessions int) *SessionManager {
	ctx, cancel := context.WithCancel(context.Background())
	
	return &SessionManager{
		sessions:        make(map[string]*SessionInfo),
		sessionTimeout:  sessionTimeout,
		cleanupInterval: cleanupInterval,
		maxSessions:     maxSessions,
		ctx:             ctx,
		cancel:          cancel,
	}
}

// CreateSession creates a new WebSocket session
func (sm *SessionManager) CreateSession(sessionID string, userID *uuid.UUID) (*SessionInfo, error) {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	// Check session limit
	if len(sm.sessions) >= sm.maxSessions {
		return nil, fmt.Errorf("maximum number of sessions reached")
	}
	
	// Check if session already exists
	if existingSession, exists := sm.sessions[sessionID]; exists {
		if existingSession.IsActive {
			return existingSession, nil
		}
		// Reactivate existing session
		existingSession.IsActive = true
		existingSession.LastActivity = time.Now()
		existingSession.ExpiresAt = time.Now().Add(sm.sessionTimeout)
		return existingSession, nil
	}
	
	// Create new session
	now := time.Now()
	session := &SessionInfo{
		ID:              sessionID,
		UserID:          userID,
		CreatedAt:       now,
		LastActivity:    now,
		ExpiresAt:       now.Add(sm.sessionTimeout),
		IsActive:        true,
		ConnectionCount: 0,
		Metadata:        make(map[string]interface{}),
		ClientIDs:       make([]string, 0),
	}
	
	sm.sessions[sessionID] = session
	
	// Start cleanup if not already running
	if !sm.cleanupRunning {
		go sm.startCleanup()
		sm.cleanupRunning = true
	}
	
	log.Printf("Session created: %s (User: %v)", sessionID, userID)
	
	return session, nil
}

// GetSession retrieves session information
func (sm *SessionManager) GetSession(sessionID string) (*SessionInfo, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	session, exists := sm.sessions[sessionID]
	if !exists || !session.IsActive {
		return nil, false
	}
	
	// Return a copy to prevent external modifications
	sessionCopy := *session
	return &sessionCopy, true
}

// UpdateSessionActivity updates the last activity time for a session
func (sm *SessionManager) UpdateSessionActivity(sessionID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	
	if !session.IsActive {
		return fmt.Errorf("session is not active: %s", sessionID)
	}
	
	session.LastActivity = time.Now()
	session.ExpiresAt = time.Now().Add(sm.sessionTimeout)
	
	return nil
}

// AddConnectionToSession adds a connection to a session
func (sm *SessionManager) AddConnectionToSession(sessionID, clientID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	
	if !session.IsActive {
		return fmt.Errorf("session is not active: %s", sessionID)
	}
	
	// Check if client is already in session
	for _, existingClientID := range session.ClientIDs {
		if existingClientID == clientID {
			return nil // Already exists
		}
	}
	
	session.ClientIDs = append(session.ClientIDs, clientID)
	session.ConnectionCount++
	session.LastActivity = time.Now()
	
	log.Printf("Connection %s added to session %s (Total connections: %d)", 
		clientID, sessionID, session.ConnectionCount)
	
	return nil
}

// RemoveConnectionFromSession removes a connection from a session
func (sm *SessionManager) RemoveConnectionFromSession(sessionID, clientID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	
	// Remove client from session
	for i, existingClientID := range session.ClientIDs {
		if existingClientID == clientID {
			session.ClientIDs = append(session.ClientIDs[:i], session.ClientIDs[i+1:]...)
			session.ConnectionCount--
			break
		}
	}
	
	// If no connections left, mark session as inactive
	if session.ConnectionCount <= 0 {
		session.IsActive = false
		log.Printf("Session %s marked as inactive (no connections)", sessionID)
	}
	
	log.Printf("Connection %s removed from session %s (Remaining connections: %d)", 
		clientID, sessionID, session.ConnectionCount)
	
	return nil
}

// ExtendSession extends the session expiration time
func (sm *SessionManager) ExtendSession(sessionID string, duration time.Duration) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	
	if !session.IsActive {
		return fmt.Errorf("session is not active: %s", sessionID)
	}
	
	session.ExpiresAt = time.Now().Add(duration)
	session.LastActivity = time.Now()
	
	log.Printf("Session %s extended by %v", sessionID, duration)
	
	return nil
}

// CloseSession closes a session and marks it as inactive
func (sm *SessionManager) CloseSession(sessionID string) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	
	session.IsActive = false
	session.ConnectionCount = 0
	session.ClientIDs = make([]string, 0)
	
	log.Printf("Session %s closed", sessionID)
	
	return nil
}

// GetSessionsByUser returns all sessions for a specific user
func (sm *SessionManager) GetSessionsByUser(userID uuid.UUID) []*SessionInfo {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	var sessions []*SessionInfo
	for _, session := range sm.sessions {
		if session.UserID != nil && *session.UserID == userID && session.IsActive {
			sessions = append(sessions, session)
		}
	}
	return sessions
}

// GetActiveSessionCount returns the number of active sessions
func (sm *SessionManager) GetActiveSessionCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	count := 0
	for _, session := range sm.sessions {
		if session.IsActive {
			count++
		}
	}
	return count
}

// GetTotalSessionCount returns the total number of sessions (including inactive)
func (sm *SessionManager) GetTotalSessionCount() int {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	return len(sm.sessions)
}

// startCleanup starts the cleanup routine
func (sm *SessionManager) startCleanup() {
	ticker := time.NewTicker(sm.cleanupInterval)
	defer ticker.Stop()
	
	for {
		select {
		case <-sm.ctx.Done():
			return
		case <-ticker.C:
			sm.cleanupExpiredSessions()
		}
	}
}

// cleanupExpiredSessions removes expired sessions
func (sm *SessionManager) cleanupExpiredSessions() {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	now := time.Now()
	var expiredSessions []string
	
	for sessionID, session := range sm.sessions {
		if session.ExpiresAt.Before(now) {
			expiredSessions = append(expiredSessions, sessionID)
		}
	}
	
	for _, sessionID := range expiredSessions {
		session := sm.sessions[sessionID]
		session.IsActive = false
		session.ConnectionCount = 0
		session.ClientIDs = make([]string, 0)
		
		log.Printf("Cleaned up expired session: %s (User: %v, Last activity: %v)", 
			sessionID, session.UserID, session.LastActivity)
	}
	
	if len(expiredSessions) > 0 {
		log.Printf("Cleaned up %d expired sessions", len(expiredSessions))
	}
}

// GetSessionStats returns session statistics
func (sm *SessionManager) GetSessionStats() map[string]interface{} {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	activeCount := 0
	userCount := make(map[string]int)
	totalConnections := 0
	
	for _, session := range sm.sessions {
		if session.IsActive {
			activeCount++
			totalConnections += session.ConnectionCount
			
			if session.UserID != nil {
				userCount[session.UserID.String()]++
			}
		}
	}
	
	return map[string]interface{}{
		"total_sessions":      len(sm.sessions),
		"active_sessions":     activeCount,
		"unique_users":        len(userCount),
		"total_connections":   totalConnections,
		"max_sessions":        sm.maxSessions,
		"session_timeout_s":   sm.sessionTimeout.Seconds(),
		"cleanup_interval_s":  sm.cleanupInterval.Seconds(),
	}
}

// SetSessionMetadata sets metadata for a session
func (sm *SessionManager) SetSessionMetadata(sessionID string, key string, value interface{}) error {
	sm.mu.Lock()
	defer sm.mu.Unlock()
	
	session, exists := sm.sessions[sessionID]
	if !exists {
		return fmt.Errorf("session not found: %s", sessionID)
	}
	
	session.Metadata[key] = value
	return nil
}

// GetSessionMetadata retrieves metadata for a session
func (sm *SessionManager) GetSessionMetadata(sessionID string, key string) (interface{}, bool) {
	sm.mu.RLock()
	defer sm.mu.RUnlock()
	
	session, exists := sm.sessions[sessionID]
	if !exists {
		return nil, false
	}
	
	value, exists := session.Metadata[key]
	return value, exists
}

// CreateSessionContext creates a session context from session info
func (sm *SessionManager) CreateSessionContext(sessionID string) (*SessionContext, error) {
	session, exists := sm.GetSession(sessionID)
	if !exists {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}
	
	context := &SessionContext{
		SessionID:    session.ID,
		UserID:       session.UserID,
		CartState:    make(map[string]interface{}),
		Preferences:  make(map[string]interface{}),
		LastActivity: session.LastActivity,
		Metadata:     make(map[string]interface{}),
	}
	
	// Copy metadata
	for key, value := range session.Metadata {
		context.Metadata[key] = value
	}
	
	return context, nil
}

// SerializeSessionContext serializes session context to JSON
func (sm *SessionManager) SerializeSessionContext(context *SessionContext) ([]byte, error) {
	return json.Marshal(context)
}

// DeserializeSessionContext deserializes session context from JSON
func (sm *SessionManager) DeserializeSessionContext(data []byte) (*SessionContext, error) {
	var context SessionContext
	err := json.Unmarshal(data, &context)
	if err != nil {
		return nil, err
	}
	return &context, nil
}

// Stop stops the session manager
func (sm *SessionManager) Stop() {
	sm.cancel()
}
