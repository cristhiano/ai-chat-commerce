# Data Model: User Authentication

**Date:** 2024-12-19  
**Project:** User Authentication  
**Technology Stack:** Go, PostgreSQL, GORM  
**Version:** 1.0.0

## Overview

The authentication system uses three primary entities: Users, Sessions, and Password Reset Tokens. All tables include standard audit fields (created_at, updated_at) and use UUID primary keys.

## Entity Definitions

### User

**Purpose:** Store user account information including authentication credentials and account state.

```go
type User struct {
    ID                   uuid.UUID  `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    Email                string     `gorm:"size:255;uniqueIndex;not null"`
    PasswordHash         string     `gorm:"type:text;not null"`
    AccountState         string     `gorm:"size:20;default:'active';index"`
    FailedLoginAttempts int        `gorm:"default:0;not null"`
    LockoutUntil         *time.Time `gorm:"index"`
    LastLoginAt          *time.Time
    CreatedAt            time.Time
    UpdatedAt            time.Time
}
```

**Field Descriptions:**
- `ID`: Unique identifier for user account
- `Email`: User's email address, must be unique, used for login
- `PasswordHash`: bcrypt-hashed password (never store plain text)
- `AccountState`: One of: `active`, `locked`, `suspended`
- `FailedLoginAttempts`: Count of consecutive failed login attempts
- `LockoutUntil`: Timestamp when lockout expires (null if not locked)
- `LastLoginAt`: Timestamp of most recent successful login
- `CreatedAt`: Account creation timestamp
- `UpdatedAt`: Last modification timestamp

**Validation Rules:**
- Email must be valid format and unique
- PasswordHash must be bcrypt hash (starts with `$2a$`, `$2b$`, etc.)
- AccountState must be one of: `active`, `locked`, `suspended`
- FailedLoginAttempts must be non-negative
- LockoutUntil must be in future when set

**Indexes:**
```sql
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_account_state ON users(account_state);
CREATE INDEX idx_users_lockout_until ON users(lockout_until);
```

### Session

**Purpose:** Track active user sessions across multiple devices with token metadata.

```go
type Session struct {
    ID        uuid.UUID           `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    UserID    uuid.UUID           `gorm:"type:uuid;not null;index"`
    Token     string              `gorm:"size:512;uniqueIndex;not null"`
    DeviceInfo string             `gorm:"type:text"` // JSON: browser, OS, IP
    LastAccessAt time.Time        `gorm:"index"`
    ExpiresAt    time.Time        `gorm:"index;not null"`
    CreatedAt    time.Time
}
```

**Field Descriptions:**
- `ID`: Unique identifier for session record
- `UserID`: Foreign key to users table
- `Token`: JWT token string (512 chars for base64-encoded JWT)
- `DeviceInfo`: JSON object with browser, OS, IP address for user visibility
- `LastAccessAt`: Timestamp of most recent API request using this session
- `ExpiresAt`: Timestamp when session expires (24 hours from last activity)
- `CreatedAt`: Session creation timestamp

**Validation Rules:**
- UserID must reference existing user
- Token must be unique
- ExpiresAt must be in future when created
- DeviceInfo must be valid JSON

**Indexes:**
```sql
CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_token ON sessions(token);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX idx_sessions_last_access ON sessions(last_access_at);
```

**Lifecycle:**
1. **Created:** On successful login
2. **Updated:** On each API request (last_access_at)
3. **Expired:** After 24 hours of inactivity (background job cleanup)
4. **Deleted:** Explicit logout or expiration

### PasswordResetToken

**Purpose:** One-time use tokens for password reset requests with expiration.

```go
type PasswordResetToken struct {
    ID        uuid.UUID `gorm:"type:uuid;primary_key;default:gen_random_uuid()"`
    UserID    uuid.UUID `gorm:"type:uuid;not null;index"`
    Token     string    `gorm:"size:128;uniqueIndex;not null"`
    ExpiresAt time.Time `gorm:"index;not null"`
    Used      bool      `gorm:"default:false"`
    CreatedAt time.Time
}
```

**Field Descriptions:**
- `ID`: Unique identifier for reset token
- `UserID`: Foreign key to users table
- `Token`: Cryptographically secure random token (URL-safe, 64 chars)
- `ExpiresAt`: Timestamp when token expires (24 hours from creation)
- `Used`: Boolean flag indicating if token has been consumed
- `CreatedAt`: Token creation timestamp

**Validation Rules:**
- UserID must reference existing user
- Token must be unique and cryptographically random
- ExpiresAt must be exactly 24 hours after CreatedAt
- Used must be false when created

**Indexes:**
```sql
CREATE INDEX idx_password_reset_tokens_token ON password_reset_tokens(token);
CREATE INDEX idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at);
CREATE INDEX idx_password_reset_tokens_user_id ON password_reset_tokens(user_id);
```

**Lifecycle:**
1. **Created:** On password reset request
2. **Used:** Set to true when password is successfully changed
3. **Expired:** After 24 hours (background job cleanup)
4. **Deleted:** After successful use or expiration

## Relationships

### User → Sessions (One-to-Many)
- One user can have multiple active sessions
- Sessions are independently managed (logout from one device doesn't affect others)
- Cascade delete: Deleting user deletes all their sessions

### User → PasswordResetTokens (One-to-Many)
- One user can have multiple reset tokens (if they request multiple resets)
- Only the most recent unused token is valid
- Cascade delete: Deleting user deletes all their reset tokens

## State Transitions

### User Account State Machine

```
Registration → active
              ↓
         active ──failed logins──> locked ──timeout──> active
              ↓                    ↓
         suspended ←────────── admin action
```

**State Definitions:**
- **active:** Normal account access, can log in
- **locked:** Temporarily locked after 5 failed login attempts, auto-unlocks after 15 minutes
- **suspended:** Permanently disabled by admin (requires manual intervention to restore)

**Transition Rules:**
- Registration always creates `active` account (no email verification)
- 5 failed logins within 15 minutes transitions to `locked`
- Lockout period expires automatically (transitions back to `active`)
- Admin action transitions to `suspended` (security breach, TOS violation, etc.)
- Successful login resets `failed_login_attempts` counter
- Logout does not change state

## Data Integrity Rules

### User Table
- Email uniqueness enforced by unique index
- PasswordHash cannot be empty or null
- AccountState defaults to 'active' on insert
- FailedLoginAttempts resets to 0 on successful login
- LockoutUntil cleared when lockout expires or user logs in successfully

### Session Table
- Token must be unique across all sessions
- ExpiresAt must be within 24 hours of CreatedAt
- LastAccessAt updated by middleware on each API request
- Sessions older than 24 hours automatically cleaned up by background job

### PasswordResetToken Table
- Token must be unique and cryptographically random
- Token can only be used once (Used flag prevents reuse)
- Expired tokens (>24 hours) cleaned up by background job
- Multiple tokens per user allowed (each reset request creates new token)

## Performance Considerations

### Database Indexes Summary
- **Primary keys:** Auto-indexed by PostgreSQL
- **Foreign keys:** Indexed for join performance (user_id in sessions, password_reset_tokens)
- **Lookup fields:** Email, token, expires_at indexed for fast queries
- **Range queries:** ExpiresAt, LockoutUntil indexed for time-based queries

### Query Optimization
- Use prepared statements for all queries (GORM handles this)
- Batch session cleanup queries for expired sessions
- Limit session history (keep last 100 sessions per user for audit trail)
- Use connection pooling (configured in database connection)

### Data Retention
- **Sessions:** Keep for 7 days after expiration for audit (auto-delete older)
- **PasswordResetTokens:** Delete immediately after use or 48 hours after expiration
- **Users:** Retain indefinitely (soft delete for GDPR compliance if needed)

## Migration Strategy

### Initial Setup
```sql
-- Create users table with authentication fields
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash TEXT NOT NULL,
    account_state VARCHAR(20) DEFAULT 'active',
    failed_login_attempts INT DEFAULT 0,
    lockout_until TIMESTAMP,
    last_login_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Create indexes
CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_account_state ON users(account_state);
CREATE INDEX idx_users_lockout_until ON users(lockout_until);
```

### Add Session Management
```sql
-- Create sessions table for JWT session tracking
CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(512) UNIQUE NOT NULL,
    device_info TEXT,
    last_access_at TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_sessions_user_id ON sessions(user_id);
CREATE INDEX idx_sessions_token ON sessions(token);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);
CREATE INDEX idx_sessions_last_access ON sessions(last_access_at);
```

### Add Password Reset Support
```sql
-- Create password reset tokens table
CREATE TABLE password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(128) UNIQUE NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW()
);

CREATE INDEX idx_password_reset_tokens_token ON password_reset_tokens(token);
CREATE INDEX idx_password_reset_tokens_expires_at ON password_reset_tokens(expires_at);
CREATE INDEX idx_password_reset_tokens_user_id ON password_reset_tokens(user_id);
```

## Backup and Recovery

### Backup Strategy
- Full database backups include authentication data (users, sessions, tokens)
- Sessions and tokens can be regenerated (no critical data loss)
- User passwords cannot be recovered if database backup lost (by design - secure)

### Disaster Recovery
- Authentication system can be restored from database backup
- Users must reset passwords after disaster recovery (cannot recover password hashes)
- New sessions will be created (old sessions invalid even if restored)

