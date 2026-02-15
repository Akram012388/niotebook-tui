---
title: "Database Schema"
created: 2026-02-15
updated: 2026-02-15
status: accepted
tags: [engineering, database, schema, postgresql]
---

# Database Schema

PostgreSQL database schema for Niotebook MVP. All tables use UUIDs as primary keys and `TIMESTAMPTZ` for timestamps.

## Entity Relationship Diagram

```
┌─────────────────────┐         ┌─────────────────────┐
│       users         │         │       posts         │
├─────────────────────┤         ├─────────────────────┤
│ id          UUID PK │◄────────│ author_id  UUID FK  │
│ username    VARCHAR  │         │ id         UUID PK  │
│ email       VARCHAR  │         │ content    TEXT      │
│ password    VARCHAR  │         │ created_at TIMESTAMPTZ│
│ display_name VARCHAR │         └─────────────────────┘
│ bio         TEXT     │
│ created_at  TIMESTAMPTZ│      ┌─────────────────────┐
│ updated_at  TIMESTAMPTZ│      │   refresh_tokens    │
└─────────────────────┘    ◄────├─────────────────────┤
                                │ id         UUID PK  │
                                │ user_id    UUID FK  │
                                │ token_hash VARCHAR  │
                                │ expires_at TIMESTAMPTZ│
                                │ created_at TIMESTAMPTZ│
                                └─────────────────────┘
```

## Table Definitions

### users

```sql
CREATE TABLE users (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username     VARCHAR(15) NOT NULL,
    email        VARCHAR(255) NOT NULL,
    password     VARCHAR(255) NOT NULL,
    display_name VARCHAR(50) NOT NULL DEFAULT '',
    bio          TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT users_username_unique UNIQUE (username),
    CONSTRAINT users_email_unique UNIQUE (email),
    CONSTRAINT users_username_length CHECK (char_length(username) >= 3),
    CONSTRAINT users_username_format CHECK (username ~ '^[a-z0-9]([a-z0-9_]*[a-z0-9])?$'),
    CONSTRAINT users_username_no_consecutive_underscores CHECK (username NOT LIKE '%__%'),
    CONSTRAINT users_email_format CHECK (email ~ '^[^@]+@[^@]+\.[^@]+$')
);
```

**Notes:**
- `username` is stored lowercase. Application lowercases before insert.
- `password` stores bcrypt hash (60 chars). VARCHAR(255) provides headroom for algorithm changes.
- `display_name` defaults to empty string. Application sets it to username on registration if not provided.
- `bio` defaults to empty string, max 160 chars enforced at application level.
- `email` has a basic format check at DB level. Full RFC 5322 validation at application level.

**Indexes:**
```sql
-- Unique indexes created implicitly by UNIQUE constraints on username and email
-- Additional index for case-insensitive email lookup:
CREATE UNIQUE INDEX idx_users_email_lower ON users (LOWER(email));
```

### posts

```sql
CREATE TABLE posts (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content    TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT posts_content_not_empty CHECK (char_length(TRIM(content)) > 0),
    CONSTRAINT posts_content_max_length CHECK (char_length(content) <= 140)
);
```

**Notes:**
- `ON DELETE CASCADE`: if a user is deleted, their posts are deleted too. Simple for MVP, revisit for soft deletes later.
- `content` constraints: not empty (after trim), max 140 chars. Double-enforced at application level.
- No `updated_at` column — posts are immutable in MVP (no editing).

**Indexes:**
```sql
-- Timeline query: ORDER BY created_at DESC
CREATE INDEX idx_posts_created_at ON posts (created_at DESC);

-- User's posts query: WHERE author_id = $1 ORDER BY created_at DESC
CREATE INDEX idx_posts_author_created ON posts (author_id, created_at DESC);
```

The composite index `idx_posts_author_created` covers the "user's posts" query efficiently — PostgreSQL can use it to find all posts by an author in chronological order without a separate sort step.

### refresh_tokens

```sql
CREATE TABLE refresh_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    CONSTRAINT refresh_tokens_hash_unique UNIQUE (token_hash)
);
```

**Notes:**
- `token_hash`: SHA-256 hex digest of the raw refresh token (64 chars).
- `ON DELETE CASCADE`: deleting a user invalidates all their refresh tokens.
- Single-use: tokens are deleted after use. See [[02-engineering/api/jwt-implementation|JWT Implementation]].

**Indexes:**
```sql
-- Lookup by token hash during refresh:
-- (Covered by UNIQUE constraint on token_hash)

-- Cleanup query: DELETE WHERE expires_at < NOW()
CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens (expires_at);

-- Revoke all tokens for a user:
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens (user_id);
```

## Migration Strategy

### Tool

[golang-migrate](https://github.com/golang-migrate/migrate) with PostgreSQL driver.

### File Convention

```
migrations/
├── 000001_create_users.up.sql
├── 000001_create_users.down.sql
├── 000002_create_posts.up.sql
├── 000002_create_posts.down.sql
├── 000003_create_refresh_tokens.up.sql
└── 000003_create_refresh_tokens.down.sql
```

Each migration has an `up` (apply) and `down` (rollback) file. Numbered sequentially.

### Running Migrations

```bash
# Apply all pending migrations
migrate -path migrations -database "postgres://user:pass@localhost/niotebook?sslmode=disable" up

# Rollback last migration
migrate -path migrations -database "postgres://..." down 1
```

The server binary also embeds migrations and can run them on startup with a `--migrate` flag:
```bash
niotebook-server --migrate
```

### Migration 000001: create_users

**Up:**
```sql
CREATE TABLE users (
    id           UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username     VARCHAR(15) NOT NULL,
    email        VARCHAR(255) NOT NULL,
    password     VARCHAR(255) NOT NULL,
    display_name VARCHAR(50) NOT NULL DEFAULT '',
    bio          TEXT NOT NULL DEFAULT '',
    created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT users_username_unique UNIQUE (username),
    CONSTRAINT users_email_unique UNIQUE (email),
    CONSTRAINT users_username_length CHECK (char_length(username) >= 3),
    CONSTRAINT users_username_format CHECK (username ~ '^[a-z0-9]([a-z0-9_]*[a-z0-9])?$'),
    CONSTRAINT users_username_no_consecutive_underscores CHECK (username NOT LIKE '%__%'),
    CONSTRAINT users_email_format CHECK (email ~ '^[^@]+@[^@]+\.[^@]+$')
);

CREATE UNIQUE INDEX idx_users_email_lower ON users (LOWER(email));
```

**Down:**
```sql
DROP TABLE IF EXISTS users CASCADE;
```

### Migration 000002: create_posts

**Up:**
```sql
CREATE TABLE posts (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id  UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content    TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT posts_content_not_empty CHECK (char_length(TRIM(content)) > 0),
    CONSTRAINT posts_content_max_length CHECK (char_length(content) <= 140)
);

CREATE INDEX idx_posts_created_at ON posts (created_at DESC);
CREATE INDEX idx_posts_author_created ON posts (author_id, created_at DESC);
```

**Down:**
```sql
DROP TABLE IF EXISTS posts CASCADE;
```

### Migration 000003: create_refresh_tokens

**Up:**
```sql
CREATE TABLE refresh_tokens (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token_hash VARCHAR(64) NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    CONSTRAINT refresh_tokens_hash_unique UNIQUE (token_hash)
);

CREATE INDEX idx_refresh_tokens_expires_at ON refresh_tokens (expires_at);
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens (user_id);
```

**Down:**
```sql
DROP TABLE IF EXISTS refresh_tokens CASCADE;
```

## Key Queries

### Timeline (Global Feed)

```sql
SELECT p.id, p.author_id, p.content, p.created_at,
       u.id AS user_id, u.username, u.display_name, u.bio, u.created_at AS user_created_at
FROM posts p
JOIN users u ON p.author_id = u.id
WHERE p.created_at < $1  -- cursor (or NOW() if no cursor)
ORDER BY p.created_at DESC
LIMIT $2;  -- limit (default 50, max 100)
```

Uses `idx_posts_created_at` for ordering and cursor filtering.

### User's Posts

```sql
SELECT p.id, p.author_id, p.content, p.created_at
FROM posts p
WHERE p.author_id = $1
  AND p.created_at < $2  -- cursor
ORDER BY p.created_at DESC
LIMIT $3;
```

Uses `idx_posts_author_created` composite index.

### Login (Find User by Email)

```sql
SELECT id, username, email, password, display_name, bio, created_at
FROM users
WHERE LOWER(email) = LOWER($1);
```

Uses `idx_users_email_lower`.

### Refresh Token Lookup

```sql
SELECT id, user_id, expires_at
FROM refresh_tokens
WHERE token_hash = $1;
```

Uses the unique index on `token_hash`.

### Cleanup Expired Refresh Tokens

```sql
DELETE FROM refresh_tokens
WHERE expires_at < NOW();
```

Run periodically by a background goroutine (every hour).

## Data Integrity

### Constraints Summary

| Table | Constraint | Type | Purpose |
|-------|-----------|------|---------|
| users | username unique | UNIQUE | No duplicate usernames |
| users | email unique | UNIQUE | No duplicate emails |
| users | username length >= 3 | CHECK | Minimum username length |
| users | username format | CHECK (regex) | Alphanumeric + underscores only |
| users | no consecutive underscores | CHECK (LIKE) | Clean usernames |
| users | email format | CHECK (regex) | Basic email validation |
| posts | content not empty | CHECK | No blank posts |
| posts | content <= 140 chars | CHECK | Character limit |
| posts | author_id FK | FOREIGN KEY | Referential integrity |
| refresh_tokens | token_hash unique | UNIQUE | No duplicate tokens |
| refresh_tokens | user_id FK | FOREIGN KEY | Referential integrity |

### Cascade Behavior

- Deleting a user cascades to: posts, refresh_tokens
- No soft deletes in MVP. Hard deletes only.

## Reserved Usernames

Enforced at application level (not DB constraint) to keep the list maintainable:

```go
var reservedUsernames = map[string]bool{
    "admin": true, "root": true, "system": true,
    "niotebook": true, "api": true, "help": true,
    "support": true, "me": true, "about": true,
    "settings": true, "login": true, "register": true,
    "auth": true, "posts": true, "users": true,
    "timeline": true, "search": true, "explore": true,
}
```
