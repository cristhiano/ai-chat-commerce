-- Migration: Create additional authentication indexes
-- Description: Optimize authentication queries with composite and covering indexes

-- Composite index for active users lookup
CREATE INDEX IF NOT EXISTS idx_users_email_state ON users(email, account_state) WHERE account_state = 'active';

-- Composite index for lockout queries
CREATE INDEX IF NOT EXISTS idx_users_lockout_active ON users(lockout_until) WHERE lockout_until IS NOT NULL;

-- Composite index for session cleanup
CREATE INDEX IF NOT EXISTS idx_sessions_expires_cleanup ON sessions(expires_at, user_id) WHERE expires_at < NOW();

-- Partial index for active sessions
CREATE INDEX IF NOT EXISTS idx_sessions_active ON sessions(user_id, expires_at) WHERE expires_at > NOW();
