# Research Findings: User Authentication

**Date:** 2024-12-19  
**Project:** User Authentication  
**Technology Stack:** Go, Gin, React, TypeScript, PostgreSQL, Redis  
**Phase:** Technical Research

## Password Hashing Algorithm

### Decision: bcrypt with work factor 10
**Rationale:** bcrypt is industry-standard, battle-tested, and provides configurable security vs. performance tradeoff. Work factor 10 balances security (requires ~100ms to hash) with acceptable user experience. Built into Go's crypto package, reducing external dependencies.

**Alternatives considered:**
- **Argon2:** Modern winner of Password Hashing Competition, excellent security properties. Too new - fewer production deployments, less library maturity.
- **PBKDF2:** NIST recommended but slower than bcrypt for equivalent security.
- **scrypt:** Memory-hard but more resource intensive, complex to configure.

**Implementation:** Use `golang.org/x/crypto/bcrypt` with default cost 10. Consider increasing to 12 for production if computational resources allow.

## Session Token Strategy

### Decision: JWT (JSON Web Tokens) stored in memory/client-side
**Rationale:** Stateless tokens enable horizontal scaling without session server dependencies. JWTs are self-contained (user ID, expiration) reducing database lookups for validation. Industry standard with excellent library support in Go (`github.com/golang-jwt/jwt/v5`).

**Alternatives considered:**
- **Database-backed sessions:** More secure (can revoke instantly) but requires DB lookup on every request, limits scalability.
- **Redis sessions:** Fast but adds infrastructure dependency and failover complexity.
- **Session cookies:** Simpler but less flexible for multi-device scenarios.

**Implementation:** 
- Claims: `{user_id, email, issued_at, expires_at, jti}`
- Expiration: 24 hours inactive
- Secret: 256-bit random key from environment variable
- Store in memory on client, send via Authorization header

## Session Storage & Multi-Device Support

### Decision: PostgreSQL for session metadata, in-memory token storage
**Rationale:** PostgreSQL sessions table tracks active sessions per user (for management and audit). Tokens themselves stored client-side to reduce server state. Enables multi-device tracking without infrastructure overhead.

**Database schema:** sessions table with user_id, token hash, device_info, expires_at.

## Account Lockout Strategy

### Decision: 5 failed attempts within 15-minute rolling window → 15-minute lockout
**Rationale:** Balances security (prevents brute force) with usability (short enough users can retry). Rolling window prevents distributed attacks. Specific lockout message with time remaining reduces support burden.

**Alternatives considered:**
- **Exponential backoff:** More complex implementation, potentially confusing users.
- **Permanent lock:** Too harsh for legitimate user mistakes.
- **CAPTCHA after 3 failures:** Better UX but adds dependency and complexity.

**Implementation:** Track failed attempts with timestamp. Reset on successful login or lockout expiry. Show remaining lockout time in API error response.

## Password Reset Flow

### Decision: Secure token-based reset with 24-hour validity
**Rationale:** Balance between security (limited window) and user convenience (users may check email later). Cryptographically secure random tokens. Email confirmation required to complete reset.

**Alternatives considered:**
- **One-time link via SMS:** More secure but requires SMS infrastructure and cost.
- **Security questions:** Less secure, users often forget answers.
- **Temporary password:** Less secure than allowing user to set their own.

**Implementation:**
- Generate 64-character URL-safe random token
- Store hash in database with user_id and expiration
- Send reset link via email
- One-time use token (mark as used after verification)

## Rate Limiting Strategy

### Decision: Redis-based sliding window rate limiting
**Rationale:** Redis provides fast, distributed rate limiting needed for authentication endpoints. Sliding window prevents burst attacks better than fixed window. Applied to login, registration, password reset endpoints.

**Limit thresholds:**
- Login: 5 attempts per 15 minutes per IP
- Registration: 10 per hour per IP
- Password Reset: 3 per hour per email

**Alternatives considered:**
- **Database-based:** Too slow, creates bottleneck
- **In-memory:** Doesn't work across multiple servers
- **Middleware-level only:** Need application-level enforcement too

## Password Strength Requirements

### Decision: Minimum 8 characters with complexity requirements
**Rationale:** Balance between security and usability. Prevents common weak passwords while not being too restrictive. Validate: uppercase, lowercase, number, special character.

**Rationale against stricter requirements:**
- 12+ characters: Users struggle to remember, often reuse passwords
- Required special chars: Users choose predictable patterns ("P@ssword123")

**Implementation:** Frontend validation for immediate feedback, backend validation as security boundary. Client-side strength meter for visual guidance.

## Email Service Selection

### Decision: SMTP with configurable provider (SendGrid/AWS SES recommended)
**Rationale:** SMTP is universal standard, works with any provider. SendGrid and AWS SES offer reliable delivery, good infrastructure, and reasonable costs. Avoid vendor lock-in with SMTP abstraction.

**Alternatives considered:**
- **Twilio SendGrid SDK:** More features but vendor lock-in
- **Postmark:** Good but less popular than SendGrid
- **Direct SMTP relay:** Less reliable delivery, requires server configuration

## Security Best Practices

### HTTPS Only Enforcement
**Rationale:** Password transmission over HTTP is unacceptable. Enforce HTTPS in production, redirect HTTP to HTTPS. Use strong cipher suites and TLS 1.2+.

### CSRF Protection
**Rationale:** State-changing authentication operations must prevent cross-site request forgery. Use SameSite cookies for CSRF protection. For API endpoints, include CSRF tokens in forms or use double-submit cookie pattern.

### Input Validation & Sanitization
**Rationale:** All user inputs must be validated server-side as security boundary. Sanitize email addresses, validate password strength, escape HTML in error messages to prevent XSS.

### SQL Injection Prevention
**Rationale:** Use parameterized queries exclusively. GORM's query builder handles this, but must verify all raw queries use placeholders. Never concatenate user input into SQL.

## Performance Optimization Strategies

### Database Indexing
- **Email field:** Primary lookup for authentication (INDEX on email column)
- **Account state:** Frequently filtered for lockout checks (INDEX on account_state, lockout_until)
- **Session lookups:** Index on token_hash, user_id, expires_at

### Connection Pooling
- **PostgreSQL:** Use pgx connection pool with max 25 connections
- **Redis:** Use single connection pool with pipelining for batch operations
- Monitor pool size and adjust based on load

### Caching Strategy
- **Active sessions:** Cache in Redis with 1-hour TTL to reduce database load
- **User account lookups:** Cache frequently accessed user records
- **Password reset tokens:** No caching (security-sensitive, validate directly)

## Error Handling & User Experience

### Generic Error Messages
**Rationale:** Never reveal whether email exists in system. Return same error message for invalid email or invalid password: "Invalid email or password". Prevents account enumeration attacks.

### Lockout Feedback
**Rationale:** Show specific message when locked out with time remaining. Improves UX by explaining why login fails and when to retry. Better than generic error that confuses users.

### Form Validation
**Rationale:** Client-side validation for immediate feedback, server-side validation as security boundary. Show password strength meter to guide users. Clear requirements display before submission.

## Testing Strategy

### Password Hashing Tests
- Verify hashed passwords never equal plain text
- Test password verification with correct and incorrect inputs
- Benchmark hashing performance (should take ~100ms at cost 10)

### Session Management Tests
- Test token generation and parsing
- Verify token expiration enforcement
- Test multi-device session tracking
- Validate session cleanup on logout

### Security Tests
- Brute force simulation (should lock after 5 attempts)
- Session hijacking prevention (token validation)
- SQL injection attempts on all inputs
- XSS attempts in error messages

## Monitoring & Observability

### Key Metrics
- **Login success rate:** Track failed vs successful logins
- **Account lockouts:** Monitor frequency and duration
- **Session duration:** Track average and max session lifetime
- **Password reset requests:** Watch for abnormal spikes (potential attack)

### Alerting Thresholds
- Failed login rate > 10% (potential attack or UX issue)
- Account lockouts exceeding 5% of total logins (policy too strict)
- Authentication error rate > 1% (infrastructure issue)
- Password reset email delivery rate < 95% (email service issue)

## Conclusion

All critical authentication decisions have been resolved using industry best practices and security standards. The chosen approach balances security, performance, and user experience while maintaining maintainability and scalability.

