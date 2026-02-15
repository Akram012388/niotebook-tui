---
title: "JWT Implementation Details"
created: 2026-02-15
updated: 2026-02-15
status: accepted
tags: [engineering, auth, jwt, security]
---

# JWT Implementation Details

## Token Types

Niotebook uses a two-token system: **access tokens** for API authentication and **refresh tokens** for session renewal.

| Property | Access Token | Refresh Token |
|----------|-------------|---------------|
| Purpose | Authenticate API requests | Obtain new access token |
| Lifetime | 24 hours | 7 days |
| Storage (server) | Stateless (not stored) | Stored in DB (for revocation) |
| Storage (client) | `~/.config/niotebook/auth.json` | `~/.config/niotebook/auth.json` |
| Rotation | New one on each refresh | Rotated on each use (single-use) |

## Access Token Claims

```json
{
  "sub": "550e8400-e29b-41d4-a716-446655440000",
  "username": "akram",
  "iat": 1739660400,
  "exp": 1739746800
}
```

| Claim | Type | Description |
|-------|------|-------------|
| `sub` | string (UUID) | User ID. Used to identify the authenticated user in handlers. |
| `username` | string | Username. Included for convenience (avoids DB lookup for display). |
| `iat` | integer (Unix timestamp) | Issued at time. |
| `exp` | integer (Unix timestamp) | Expiration time. 24h after `iat`. |

### Signing

- **Algorithm:** HS256 (HMAC-SHA256)
- **Secret:** 256-bit random key, stored as environment variable `NIOTEBOOK_JWT_SECRET`
- **Library:** `github.com/golang-jwt/jwt/v5`

HS256 is chosen over RS256 because there's a single server (no need to distribute a public key for verification). The secret never leaves the server.

## Refresh Token

The refresh token is an opaque random string (not a JWT), stored in the database.

### Database Table

```sql
CREATE TABLE refresh_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL UNIQUE,  -- SHA-256 hash of token
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);
CREATE INDEX idx_refresh_tokens_token_hash ON refresh_tokens(token_hash);
```

### Why Opaque Tokens (Not JWT)

Refresh tokens are stored server-side anyway (for revocation). Making them JWTs adds no benefit — the server must look them up in the DB regardless. An opaque string is simpler and avoids exposing claims unnecessarily.

### Token Generation

```
token = base64url(random_bytes(32))  // 256-bit random, URL-safe
hash  = sha256(token)                // stored in DB
```

The raw token is sent to the client. The server only stores the hash. This means if the DB is compromised, the attacker can't use the hashed tokens directly.

## Refresh Flow

```
1. TUI sends request with access token
2. Server returns 401 (token expired)
3. TUI sends POST /api/v1/auth/refresh { refresh_token: "raw_token" }
4. Server:
   a. Hash the received token: sha256(raw_token)
   b. Look up hash in refresh_tokens table
   c. Verify: token exists AND not expired AND user exists
   d. Delete the used refresh token (single-use)
   e. Generate new access token + new refresh token
   f. Store new refresh token hash in DB
   g. Return both tokens to client
5. TUI stores new tokens, retries original request
```

### Single-Use Refresh Tokens

Each refresh token can only be used once. After use, it's deleted and a new one is issued. This limits the damage of a stolen refresh token — the legitimate user's next refresh will fail (because the token was already consumed), signaling compromise.

If a refresh token is reused (already deleted from DB), the server deletes ALL refresh tokens for that user (nuclear option) — forcing re-login on all sessions. This is a standard defense against token theft.

## Token Lifecycle Diagram

```
Registration/Login
    │
    ├── Access Token (24h) ──── API requests ──── Expires
    │                                                │
    │                                                ▼
    │                                          401 Unauthorized
    │                                                │
    └── Refresh Token (7d) ──── POST /auth/refresh ──┘
            │                         │
            │                         ├── Success: new token pair
            │                         │
            │                         └── Failure: → Login View
            │
            └── Expires after 7d ──── → Login View
```

## Security Considerations

### Token Storage on Client

- `auth.json` file permissions: `0600` (read/write only for the user)
- Tokens are stored in plaintext (standard for CLI tools — `gh`, `docker`, `kubectl` all do this)
- No keychain integration for MVP (post-MVP enhancement)

### Server-Side Security

- JWT secret MUST be at least 256 bits of entropy
- JWT secret is loaded from environment variable, never hardcoded or committed
- Refresh token hashes are stored, not raw tokens
- Expired refresh tokens are cleaned up by a periodic goroutine (every hour, delete WHERE expires_at < NOW())

### Brute Force Protection

- Login endpoint: rate-limited to 10 requests/minute per IP with burst of 5 ([[02-engineering/adr/ADR-0014-rate-limiting|ADR-0014]])
- After 5 rapid requests, the token bucket throttles. After 10 in a minute, all are rejected with 429.
- No separate account lockout mechanism — rate limiting per IP handles brute force. Account lockout is avoided because it enables DoS attacks (attacker locks out a legitimate user by spamming login attempts).
- See [[02-engineering/architecture/server-internals#Brute Force Protection|Server Internals]] for implementation details.

## Environment Variables

| Variable | Required | Default | Description |
|----------|----------|---------|-------------|
| `NIOTEBOOK_JWT_SECRET` | Yes | — | HS256 signing key (min 32 bytes) |
| `NIOTEBOOK_ACCESS_TOKEN_TTL` | No | `24h` | Access token lifetime |
| `NIOTEBOOK_REFRESH_TOKEN_TTL` | No | `168h` | Refresh token lifetime (7 days) |
