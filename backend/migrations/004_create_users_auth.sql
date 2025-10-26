-- Migration: Add authentication fields to users table
-- Description: Adds authentication-related fields for user accounts

-- Add authentication fields to users table if they don't exist
ALTER TABLE users ADD COLUMN IF NOT EXISTS account_state VARCHAR(20) DEFAULT 'active';
ALTER TABLE users ADD COLUMN IF NOT EXISTS failed_login_attempts INT DEFAULT 0;
ALTER TABLE users ADD COLUMN IF NOT EXISTS lockout_until TIMESTAMP;
ALTER TABLE users ADD COLUMN IF NOT EXISTS last_login_at TIMESTAMP;

-- Create indexes for authentication queries
CREATE INDEX IF NOT EXISTS idx_users_account_state ON users(account_state);
CREATE INDEX IF NOT EXISTS idx_users_lockout_until ON users(lockout_until);

-- Update existing users to have active account_state
UPDATE users SET account_state = 'active' WHERE account_state IS NULL;
