---
title: "System Architecture Overview"
created: 2026-02-15
updated: 2026-02-15
status: draft
tags: [architecture, system-design]
---

# System Architecture Overview

## High-Level Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                        USER'S TERMINAL                       │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐  │
│  │                   niotebook-tui                         │  │
│  │                                                        │  │
│  │  ┌──────────┐  ┌───────────┐  ┌────────────────────┐  │  │
│  │  │  Views   │  │ Components│  │   HTTP Client       │  │  │
│  │  │----------│  │-----------│  │--------------------│  │  │
│  │  │ Timeline │  │ Post Card │  │ GET  /api/v1/...   │  │  │
│  │  │ Compose  │  │ Header    │  │ POST /api/v1/...   │  │  │
│  │  │ Profile  │  │ Statusbar │  │ Auth: Bearer <jwt> │  │  │
│  │  │ Login    │  │ Input     │  │                    │  │  │
│  │  └──────────┘  └───────────┘  └────────┬───────────┘  │  │
│  │                                         │              │  │
│  │        Bubble Tea (Model-Update-View)   │              │  │
│  └─────────────────────────────────────────┼──────────────┘  │
│                                            │                 │
└────────────────────────────────────────────┼─────────────────┘
                                             │ HTTPS
                                             ▼
┌─────────────────────────────────────────────────────────────┐
│                          VPS                                 │
│                                                              │
│  ┌────────────────────────────────────────────────────────┐  │
│  │                  niotebook-server                       │  │
│  │                                                        │  │
│  │  ┌──────────────┐  ┌──────────────┐  ┌─────────────┐  │  │
│  │  │   Handlers   │  │   Services   │  │    Store     │  │  │
│  │  │--------------│  │--------------│  │-------------│  │  │
│  │  │ POST /auth/* │  │ AuthService  │  │ UserStore   │  │  │
│  │  │ GET /timeline│  │ PostService  │  │ PostStore   │  │  │
│  │  │ POST /posts  │  │ UserService  │  │             │  │  │
│  │  │ GET /users/* │  │              │  │             │  │  │
│  │  └──────┬───────┘  └──────┬───────┘  └──────┬──────┘  │  │
│  │         │                 │                  │         │  │
│  │         │    Middleware (JWT Auth, Logging)   │         │  │
│  │         │                                    │         │  │
│  │         └─────── Go net/http (1.22+) ────────┘         │  │
│  └──────────────────────────────┬─────────────────────────┘  │
│                                 │                            │
│  ┌──────────────────────────────▼─────────────────────────┐  │
│  │                    PostgreSQL                           │  │
│  │  ┌──────────┐  ┌──────────┐  ┌──────────────────────┐  │  │
│  │  │  users   │  │  posts   │  │  schema_migrations   │  │  │
│  │  └──────────┘  └──────────┘  └──────────────────────┘  │  │
│  └─────────────────────────────────────────────────────────┘  │
│                                                              │
│  Caddy (reverse proxy, auto-HTTPS) ─── :443 → :8080         │
│                                                              │
└─────────────────────────────────────────────────────────────┘
```

## Component Breakdown

### TUI Client (`cmd/tui/`)

Built with [[02-engineering/architecture/tui-architecture|Bubble Tea's Elm architecture]]:

- **Model:** Application state (current view, post list, user session, form inputs)
- **Update:** Handles keyboard events, API responses, view transitions
- **View:** Renders styled terminal output using Lip Gloss

**Key packages:**
- `internal/tui/app/` — root model, key bindings, view routing
- `internal/tui/views/` — screen-level models (timeline, compose, profile, login)
- `internal/tui/components/` — reusable widgets (post card, header, status bar)
- `internal/tui/client/` — HTTP client wrapping server API calls
- `internal/tui/config/` — local config loading (`~/.config/niotebook/config.yaml`)

### Server (`cmd/server/`)

Three-layer architecture:

1. **Handlers** (`internal/server/handler/`) — HTTP request parsing, response writing. Thin layer that delegates to services.
2. **Services** (`internal/server/service/`) — business logic. Validation, authorization checks, data transformation.
3. **Store** (`internal/server/store/`) — database access. Raw SQL via `pgx`. One store per domain entity.

**Middleware chain:** Logging → CORS → Rate Limiting → JWT Auth → Handler

### Shared Models (`internal/models/`)

Domain types used by both server and TUI:

```go
type User struct {
    ID          string    `json:"id"`
    Username    string    `json:"username"`
    DisplayName string    `json:"display_name"`
    Bio         string    `json:"bio"`
    CreatedAt   time.Time `json:"created_at"`
}

type Post struct {
    ID        string    `json:"id"`
    AuthorID  string    `json:"author_id"`
    Author    *User     `json:"author,omitempty"`
    Content   string    `json:"content"`
    CreatedAt time.Time `json:"created_at"`
}
```

## API Endpoints (MVP)

| Method | Path | Auth | Description |
|--------|------|------|-------------|
| POST | `/api/v1/auth/register` | No | Create account |
| POST | `/api/v1/auth/login` | No | Get JWT tokens |
| POST | `/api/v1/auth/refresh` | Refresh token | Refresh access token |
| GET | `/api/v1/timeline` | Yes | Get global timeline (cursor pagination) |
| POST | `/api/v1/posts` | Yes | Create a post |
| GET | `/api/v1/posts/{id}` | Yes | Get a single post |
| GET | `/api/v1/users/{id}` | Yes | Get user profile |
| GET | `/api/v1/users/{id}/posts` | Yes | Get user's posts |
| PATCH | `/api/v1/users/me` | Yes | Update own profile |

## Database Schema (MVP)

```sql
CREATE TABLE users (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    username   VARCHAR(30) UNIQUE NOT NULL,
    email      VARCHAR(255) UNIQUE NOT NULL,
    password   VARCHAR(255) NOT NULL,  -- bcrypt hash
    display_name VARCHAR(50),
    bio        TEXT DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE posts (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    author_id  UUID NOT NULL REFERENCES users(id),
    content    TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_posts_created_at ON posts(created_at DESC);
CREATE INDEX idx_posts_author_id ON posts(author_id);
```

## Data Flow: Posting

```
User presses 'n' → Compose view opens
User types post, presses Ctrl+Enter
  → TUI sends POST /api/v1/posts {content: "..."}
    → Server handler parses request, extracts JWT claims
      → Service validates content (length, not empty)
        → Store inserts row into posts table
      → Service returns created post
    → Handler writes 201 Created + JSON response
  → TUI receives response, appends post to local timeline state
  → View re-renders with new post at top
```

## Data Flow: Timeline Refresh

```
User presses 'r'
  → TUI sends GET /api/v1/timeline?cursor={last_post_id}&limit=50
    → Server handler extracts cursor + limit
      → Store queries: SELECT posts + users WHERE created_at < cursor ORDER BY created_at DESC LIMIT 50
    → Handler writes 200 OK + JSON array
  → TUI replaces timeline state with fresh data
  → View re-renders
```

## Technology Stack Summary

| Component | Technology | Rationale |
|-----------|-----------|-----------|
| Language | Go 1.22+ | Concurrency, single binary, cross-compile |
| TUI Framework | Bubble Tea | Elm architecture, mature ecosystem |
| TUI Styling | Lip Gloss | CSS-like terminal styling |
| TUI Widgets | Bubbles | Viewport, textarea, list, spinner |
| HTTP Router | net/http (stdlib) | Go 1.22 pattern matching, zero dependencies |
| Database | PostgreSQL | Relational, proven for social media |
| DB Driver | pgx | High-performance native Postgres driver |
| Migrations | golang-migrate | Versioned SQL migrations |
| Auth | JWT (golang-jwt) | Stateless token auth |
| Password | bcrypt (golang.org/x/crypto) | Industry standard password hashing |
| Config | YAML (gopkg.in/yaml.v3) | User config files |
| Reverse Proxy | Caddy | Auto-HTTPS, simple config |

## Resolved Design Questions

All design questions have been resolved via ADRs:

- [x] **Post character limit:** 140 characters — [[02-engineering/adr/ADR-0010-post-character-limit|ADR-0010]]
- [x] **TUI theme:** Single dark theme, ANSI 256 colors — [[02-engineering/adr/ADR-0015-dark-theme-only|ADR-0015]]
- [x] **Error handling UX:** Inline status bar messages — [[02-engineering/adr/ADR-0013-error-handling-ux|ADR-0013]]
- [x] **Pagination:** Cursor-based, `created_at` timestamp cursor — [[02-engineering/adr/ADR-0012-cursor-pagination|ADR-0012]]
- [x] **Rate limiting:** Per-IP token bucket, `x/time/rate` — [[02-engineering/adr/ADR-0014-rate-limiting|ADR-0014]]
- [x] **Username rules:** 3-15 chars, alphanumeric + underscores, case-insensitive — [[02-engineering/adr/ADR-0011-username-rules|ADR-0011]]
- [x] **Config location:** `~/.config/niotebook/` (XDG) — [[02-engineering/adr/ADR-0016-config-xdg|ADR-0016]]
