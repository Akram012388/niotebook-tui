---
title: "API Specification"
created: 2026-02-15
updated: 2026-02-15
status: accepted
tags: [engineering, api, specification]
---

# API Specification

Base URL: `https://api.niotebook.com/api/v1`

All endpoints accept and return `Content-Type: application/json`. Authenticated endpoints require `Authorization: Bearer <access_token>` header.

## Error Response Format

All errors use a consistent format:

```json
{
  "error": {
    "code": "validation_error",
    "message": "Username must be 3-15 characters, alphanumeric and underscores only.",
    "field": "username"
  }
}
```

| Field | Type | Required | Description |
|-------|------|----------|-------------|
| `error.code` | string | Yes | Machine-readable error code |
| `error.message` | string | Yes | Human-readable message (displayed in TUI status bar) |
| `error.field` | string | No | Which input field caused the error (for form validation) |

### Error Codes

| HTTP Status | Code | When |
|-------------|------|------|
| 400 | `validation_error` | Invalid input (bad username, empty content, etc.) |
| 400 | `content_too_long` | Post exceeds 140 characters |
| 401 | `unauthorized` | Missing or invalid access token |
| 401 | `token_expired` | Access token has expired (TUI should attempt refresh) |
| 403 | `forbidden` | Authenticated but not authorized (e.g., editing another user's profile) |
| 404 | `not_found` | Resource doesn't exist |
| 409 | `conflict` | Unique constraint violation (duplicate username/email) |
| 429 | `rate_limited` | Too many requests. `Retry-After` header included. |
| 500 | `internal_error` | Server bug. Message: "Something went wrong. Please try again." |

---

## Authentication Endpoints

### POST /api/v1/auth/register

Create a new user account.

**Request:**
```json
{
  "username": "akram",
  "email": "akram@example.com",
  "password": "securepass123"
}
```

**Validation:**
- `username`: 3-15 chars, `^[a-zA-Z0-9]([a-zA-Z0-9_]*[a-zA-Z0-9])?$`, no consecutive underscores, not reserved
- `email`: valid email format, unique
- `password`: minimum 8 characters

**Success Response (201 Created):**
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "akram",
    "display_name": "akram",
    "bio": "",
    "created_at": "2026-02-15T22:00:00Z"
  },
  "tokens": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_at": "2026-02-16T22:00:00Z"
  }
}
```

**Error Responses:**
- `409 Conflict` — `{"error": {"code": "conflict", "message": "Username already taken", "field": "username"}}`
- `409 Conflict` — `{"error": {"code": "conflict", "message": "Email already registered", "field": "email"}}`
- `400 Bad Request` — `{"error": {"code": "validation_error", "message": "Password must be at least 8 characters", "field": "password"}}`

### POST /api/v1/auth/login

Authenticate an existing user.

**Request:**
```json
{
  "email": "akram@example.com",
  "password": "securepass123"
}
```

**Success Response (200 OK):**
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "akram",
    "display_name": "Akram",
    "bio": "Building things in Go.",
    "created_at": "2026-02-15T22:00:00Z"
  },
  "tokens": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_at": "2026-02-16T22:00:00Z"
  }
}
```

**Error Responses:**
- `401 Unauthorized` — `{"error": {"code": "unauthorized", "message": "Invalid email or password"}}`

### POST /api/v1/auth/refresh

Exchange a refresh token for a new token pair.

**Request:**
```json
{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Success Response (200 OK):**
```json
{
  "tokens": {
    "access_token": "eyJhbGciOiJIUzI1NiIs...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
    "expires_at": "2026-02-16T22:00:00Z"
  }
}
```

**Error Responses:**
- `401 Unauthorized` — `{"error": {"code": "token_expired", "message": "Refresh token has expired"}}`

Note: Refresh also rotates the refresh token (single-use refresh tokens). The old refresh token is invalidated.

---

## Post Endpoints

### POST /api/v1/posts

Create a new post. Requires authentication.

**Request:**
```json
{
  "content": "Building a social media platform in the terminal. Yes, really."
}
```

**Validation:**
- `content`: 1-140 characters (rune count), not empty, not whitespace-only
- Newlines are allowed (multi-line posts are supported, see [[02-engineering/adr/ADR-0022-multiline-posts|ADR-0022]])
- Content is **trimmed** of leading/trailing whitespace before validation and storage
- The 140-char limit is checked **after** trimming (e.g., `"  hello  "` becomes `"hello"` = 5 chars)
- The TUI compose modal shows character count of **raw** (untrimmed) input, since trimming is server-side

**Success Response (201 Created):**
```json
{
  "post": {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "author_id": "550e8400-e29b-41d4-a716-446655440000",
    "author": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "username": "akram",
      "display_name": "Akram",
      "bio": "Building things in Go.",
      "created_at": "2026-02-15T22:00:00Z"
    },
    "content": "Building a social media platform in the terminal. Yes, really.",
    "created_at": "2026-02-15T23:30:00Z"
  }
}
```

**Error Responses:**
- `400 Bad Request` — `{"error": {"code": "content_too_long", "message": "Post must be 140 characters or fewer"}}`
- `400 Bad Request` — `{"error": {"code": "validation_error", "message": "Post content cannot be empty"}}`

### GET /api/v1/posts/{id}

Get a single post by ID. Requires authentication.

**Success Response (200 OK):**
```json
{
  "post": {
    "id": "660e8400-e29b-41d4-a716-446655440001",
    "author_id": "550e8400-e29b-41d4-a716-446655440000",
    "author": {
      "id": "550e8400-e29b-41d4-a716-446655440000",
      "username": "akram",
      "display_name": "Akram",
      "bio": "Building things in Go.",
      "created_at": "2026-02-15T22:00:00Z"
    },
    "content": "Building a social media platform in the terminal. Yes, really.",
    "created_at": "2026-02-15T23:30:00Z"
  }
}
```

**Error Responses:**
- `404 Not Found` — `{"error": {"code": "not_found", "message": "Post not found"}}`

---

## Timeline Endpoints

### GET /api/v1/timeline

Get the global timeline (all posts, reverse chronological). Requires authentication.

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `cursor` | string (RFC3339 timestamp) | none (latest) | Return posts older than this timestamp |
| `limit` | integer | 50 | Number of posts to return (max 100) |

**Example Request:**
```
GET /api/v1/timeline?limit=50
GET /api/v1/timeline?cursor=2026-02-15T22:00:00Z&limit=50
```

**Success Response (200 OK):**
```json
{
  "posts": [
    {
      "id": "660e8400-e29b-41d4-a716-446655440001",
      "author_id": "550e8400-e29b-41d4-a716-446655440000",
      "author": {
        "id": "550e8400-e29b-41d4-a716-446655440000",
        "username": "akram",
        "display_name": "Akram",
        "bio": "",
        "created_at": "2026-02-15T22:00:00Z"
      },
      "content": "Hello, Niotebook!",
      "created_at": "2026-02-15T23:30:00Z"
    }
  ],
  "next_cursor": "2026-02-15T20:00:00Z",
  "has_more": true
}
```

**Notes:**
- `posts` is ordered newest-first (descending `created_at`)
- `next_cursor` is the `created_at` of the last post in the array. Pass it as `cursor` to get the next page.
- `has_more` is `false` when there are no more posts to load.
- If the timeline is empty, `posts` is `[]`, `next_cursor` is `null`, `has_more` is `false`.

---

## User Endpoints

### GET /api/v1/users/{id}

Get a user's public profile. Requires authentication. The `{id}` can be a UUID or the special value `me` (returns the authenticated user).

**Success Response (200 OK):**
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "akram",
    "display_name": "Akram",
    "bio": "Building things in Go.",
    "created_at": "2026-02-15T22:00:00Z"
  }
}
```

### GET /api/v1/users/{id}/posts

Get a user's posts (reverse chronological). Same pagination as timeline.

**Query Parameters:**
| Parameter | Type | Default | Description |
|-----------|------|---------|-------------|
| `cursor` | string (RFC3339) | none | Posts older than this timestamp |
| `limit` | integer | 50 | Number of posts (max 100) |

**Success Response (200 OK):**
```json
{
  "posts": [...],
  "next_cursor": "2026-02-15T20:00:00Z",
  "has_more": true
}
```

Each post in this endpoint includes `author_id` but omits the nested `author` object (it's redundant — the caller already has the user profile from `GET /users/{id}`). Post structure:

```json
{
  "id": "660e8400-e29b-41d4-a716-446655440001",
  "author_id": "550e8400-e29b-41d4-a716-446655440000",
  "content": "Hello Niotebook!",
  "created_at": "2026-02-15T23:30:00Z"
}
```

### PATCH /api/v1/users/me

Update the authenticated user's profile. Only provided fields are updated.

**Request:**
```json
{
  "display_name": "Shaikh Akram",
  "bio": "Building Niotebook. Go enthusiast."
}
```

**Validation:**
- `display_name`: 1-50 characters, optional
- `bio`: 0-160 characters, optional

**Success Response (200 OK):**
```json
{
  "user": {
    "id": "550e8400-e29b-41d4-a716-446655440000",
    "username": "akram",
    "display_name": "Shaikh Akram",
    "bio": "Building Niotebook. Go enthusiast.",
    "created_at": "2026-02-15T22:00:00Z"
  }
}
```

**Error Responses:**
- `400 Bad Request` — `{"error": {"code": "validation_error", "message": "Display name must be 50 characters or fewer", "field": "display_name"}}`

---

## Health Endpoint

### GET /health

Server health check. No authentication required. Not rate limited.

**Success Response (200 OK):**
```json
{
  "status": "ok",
  "version": "0.1.0"
}
```

**Failure Response (503 Service Unavailable):**
```json
{
  "status": "error",
  "message": "database connection failed"
}
```

---

## Response Envelope Convention

All successful responses wrap data in a named key:
- Single resource: `{"post": {...}}`, `{"user": {...}}`
- Collections: `{"posts": [...], "next_cursor": "...", "has_more": true}`
- Token responses: `{"tokens": {...}}` (sometimes alongside `{"user": {...}}`)

This prevents top-level arrays (a security concern in older browsers, and makes the response self-documenting).
