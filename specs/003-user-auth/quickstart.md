# User Authentication - Quick Start Guide

**Version:** 1.0.0  
**Date:** 2024-12-19  
**Feature:** User Authentication

## Overview

This guide provides a quick start for implementing user authentication in the B2C Chat application. The system supports user registration, secure login with JWT sessions, password management, multi-device support, and security features like account lockout.

## Key Features

### Core Functionality
- **User Registration:** Email and password-based account creation (no email verification)
- **Secure Login:** JWT token-based authentication with session management
- **Multi-Device Support:** Concurrent sessions across multiple devices
- **Password Management:** Reset, update, and recovery capabilities
- **Account Security:** Brute force protection with automatic account lockout
- **Session Management:** 24-hour timeout with automatic logout

### Security Features
- **Password Hashing:** bcrypt with work factor 10
- **Brute Force Protection:** 5 failed attempts = 15-minute lockout
- **Account States:** Active, locked, suspended
- **Session Timeout:** 24 hours of inactivity
- **Password Requirements:** Minimum 8 chars with complexity
- **Generic Error Messages:** Prevents account enumeration

## API Quick Reference

### Register User
```bash
POST /api/auth/register
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!"
}

# Response (201)
{
  "success": true,
  "user_id": "550e8400-e29b-41d4-a716-446655440000",
  "message": "Account created successfully"
}
```

### Login
```bash
POST /api/auth/login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "SecurePass123!"
}

# Response (200)
{
  "success": true,
  "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "email": "user@example.com"
  },
  "expires_at": "2024-12-20T10:30:00Z"
}
```

### Use Authenticated Endpoints
```bash
# Include token in Authorization header
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...

# Example: Logout
POST /api/auth/logout
Authorization: Bearer {token}
```

### Password Reset Flow
```bash
# Step 1: Request reset
POST /api/auth/password/reset
{
  "email": "user@example.com"
}

# Step 2: Receive email with reset link
# Click link or call API with token

# Step 3: Verify token and set new password
POST /api/auth/password/reset/verify
{
  "token": "abc123xyz...",
  "new_password": "NewSecurePass123!"
}
```

## Database Setup

### Run Migrations
```bash
# The migrations are automatically run on application startup
# Manual migration:

cd backend
go run cmd/api/main.go --migrate-only
```

### Database Schema
The system uses three main tables:
1. **users:** User accounts and credentials
2. **sessions:** Active JWT sessions
3. **password_reset_tokens:** Password reset tokens

See `data-model.md` for complete schema.

## Frontend Integration

### Setup AuthContext
```typescript
// src/contexts/AuthContext.tsx
import { createContext, useContext, useState } from 'react';

interface AuthContextType {
  user: User | null;
  token: string | null;
  login: (email: string, password: string) => Promise<void>;
  logout: () => void;
  isAuthenticated: boolean;
}

export const AuthContext = createContext<AuthContextType>(...);

export function useAuth() {
  return useContext(AuthContext);
}
```

### Login Component
```typescript
import { useAuth } from '../contexts/AuthContext';

function LoginForm() {
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const { login, isLoading, error } = useAuth();

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    await login(email, password);
  };

  return (
    <form onSubmit={handleSubmit}>
      {/* Form fields */}
    </form>
  );
}
```

### Protected Routes
```typescript
import { Navigate } from 'react-router-dom';
import { useAuth } from '../contexts/AuthContext';

function ProtectedRoute({ children }) {
  const { isAuthenticated } = useAuth();
  
  if (!isAuthenticated) {
    return <Navigate to="/login" />;
  }
  
  return children;
}
```

## Backend Implementation

### Password Hashing Service
```go
// backend/internal/services/auth/password.go
package auth

import "golang.org/x/crypto/bcrypt"

func HashPassword(password string) (string, error) {
    bytes, err := bcrypt.GenerateFromPassword([]byte(password), 10)
    return string(bytes), err
}

func VerifyPassword(password, hash string) bool {
    err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
    return err == nil
}
```

### Session Management
```go
// backend/internal/services/auth/session.go
package auth

func (s *Service) CreateSession(userID uuid.UUID) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "user_id": userID.String(),
        "exp": time.Now().Add(24 * time.Hour).Unix(),
    })
    
    tokenString, err := token.SignedString([]byte(jwtSecret))
    return tokenString, err
}
```

### Authentication Middleware
```go
// backend/internal/middleware/auth.go
func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := extractToken(c)
        
        claims, err := validateToken(token)
        if err != nil {
            c.JSON(401, gin.H{"error": "Unauthorized"})
            c.Abort()
            return
        }
        
        c.Set("user_id", claims.UserID)
        c.Next()
    }
}
```

## Testing

### Unit Tests
```go
// backend/tests/services/auth_test.go
func TestPasswordHashing(t *testing.T) {
    password := "TestPass123!"
    hash, err := auth.HashPassword(password)
    assert.NoError(t, err)
    
    verified := auth.VerifyPassword(password, hash)
    assert.True(t, verified)
    
    wrongPassword := "WrongPass123!"
    verified = auth.VerifyPassword(wrongPassword, hash)
    assert.False(t, verified)
}
```

### Integration Tests
```go
// backend/tests/integration/auth_test.go
func TestLoginFlow(t *testing.T) {
    // 1. Register user
    registerResp := registerUser(testEmail, testPassword)
    assert.Equal(t, 201, registerResp.StatusCode)
    
    // 2. Login
    loginResp := loginUser(testEmail, testPassword)
    assert.Equal(t, 200, loginResp.StatusCode)
    assert.NotEmpty(t, loginResp.Token)
    
    // 3. Use authenticated endpoint
    userResp := getCurrentUser(loginResp.Token)
    assert.Equal(t, 200, userResp.StatusCode)
}
```

## Security Checklist

### Before Production
- [ ] HTTPS enforced on all endpoints
- [ ] JWT secret key is strong and stored securely
- [ ] Password hashing uses bcrypt cost 10+
- [ ] Account lockout is tested and working
- [ ] Session tokens expire correctly (24 hours)
- [ ] Password reset tokens expire (24 hours)
- [ ] SQL injection prevention verified
- [ ] XSS prevention tested
- [ ] CSRF protection implemented
- [ ] Rate limiting configured
- [ ] Error messages don't reveal account existence
- [ ] Audit logging enabled

## Performance Benchmarks

### Target Metrics
- **Login API:** < 500ms (p95 < 800ms)
- **Registration:** < 1 second
- **Session Validation:** < 100ms
- **Concurrent Capacity:** 1000+ requests

### Monitoring
Track these metrics:
- Failed login rate (%)
- Account lockout frequency
- Authentication error rate
- Session creation rate
- Password reset request volume

## Troubleshooting

### User Locked Out
- **Cause:** 5 failed login attempts within 15 minutes
- **Solution:** Wait 15 minutes or check failed attempts counter
- **Message:** "Account temporarily locked. Try again in X minutes."

### Token Expired
- **Cause:** 24 hours of inactivity
- **Solution:** User must log in again
- **UX:** Auto-redirect to login page with "Session expired" message

### Password Reset Not Working
- Check email delivery (spam folder)
- Verify token hasn't expired (24 hours)
- Confirm token hasn't been used already (one-time use)
- Check database for token existence

## Next Steps

1. **Implement Backend Services**
   - Password hashing service
   - Session management
   - Authentication middleware

2. **Create API Endpoints**
   - Registration, login, logout
   - Password reset and update

3. **Frontend Components**
   - Login and registration forms
   - Password reset flow
   - Protected route wrapper

4. **Testing**
   - Unit tests for security functions
   - Integration tests for API flows
   - E2E tests for user journeys

5. **Deployment**
   - Configure environment variables
   - Set up HTTPS
   - Enable monitoring

## Additional Resources

- **API Specification:** `/contracts/api.yaml`
- **Data Model:** `/data-model.md`
- **Research Findings:** `/research.md`
- **Implementation Plan:** `/plan.md`

