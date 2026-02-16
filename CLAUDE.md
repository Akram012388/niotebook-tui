# Niotebook Development Guide

## Project Overview

Niotebook is a standalone TUI social media platform built in Go. Monorepo producing two binaries: `niotebook-server` and `niotebook-tui`.

**Documentation:** All specs live in `docs/vault/`. The vault index is at `docs/vault/00-home/index.md`.

## Tech Stack

- **Language:** Go 1.22+
- **TUI:** Bubble Tea + Bubbles + Lip Gloss (Charm ecosystem)
- **Server:** stdlib `net/http` (Go 1.22 pattern matching), three-layer architecture (handler/service/store)
- **Database:** PostgreSQL 15+ via pgx v5
- **Auth:** JWT (golang-jwt/v5), bcrypt passwords, single-use refresh tokens
- **Migrations:** golang-migrate with sequential numbered SQL files

## Project Structure

```
cmd/
  server/          # Server binary entrypoint
  tui/             # TUI binary entrypoint
internal/
  models/          # Shared domain types (User, Post, TokenPair)
  build/           # Version embedding
  server/
    handler/       # HTTP handlers (thin, delegate to services)
    service/       # Business logic, validation
    store/         # Database access (raw SQL via pgx, one store per entity)
  tui/
    app/           # Root AppModel, key bindings, view routing
    views/         # Screen-level models (timeline, compose, profile, login)
    components/    # Reusable widgets (post card, header, status bar)
    client/        # HTTP client wrapper (handles auth refresh transparently)
    config/        # Local config loading (~/.config/niotebook/)
migrations/        # Sequential SQL migration files (up/down pairs)
```

## Build Commands

```bash
make build       # Build both binaries to bin/
make server      # Build server only
make tui         # Build TUI only
make dev         # Run server in dev mode (localhost:8080)
make dev-tui     # Run TUI pointing at local server
make test        # Run all tests with race detector
make test-cover  # Tests with coverage report
make lint        # golangci-lint
make migrate-up  # Apply pending migrations
make migrate-down # Rollback last migration
```

## Code Conventions

### Go Style

- Follow standard Go conventions (gofmt, go vet, golangci-lint)
- Use `slog` (stdlib) for structured logging, not third-party loggers
- Use `flag` (stdlib) for CLI flags, not Cobra
- Error messages are lowercase, no trailing punctuation
- Store layer uses raw SQL via pgx, not an ORM
- Interfaces defined where consumed (store interfaces in service package)

### Server Patterns

- Handlers parse HTTP, call service, write response. No business logic in handlers.
- Services contain validation and business logic. Accept and return domain types.
- Stores do database access only. One store per entity. Accept context as first parameter.
- Middleware chain order: Recovery > Logging > Rate Limiting > CORS > Auth > Handler
- Request body limit: 4KB via `http.MaxBytesReader`
- All endpoints return JSON with consistent error format: `{"error": {"code": "...", "message": "...", "field": "..."}}`
- Successful responses wrap data: `{"post": {...}}`, `{"posts": [...], "next_cursor": "...", "has_more": true}`

### TUI Patterns

- Bubble Tea Elm architecture: Model holds state, Update handles messages, View renders
- All API calls return `tea.Cmd` that runs in Bubble Tea's goroutine pool
- Custom message types for async results (msgTimelineLoaded, msgPostPublished, etc.)
- Overlays (compose, help) take priority in Update routing over view-specific handling
- `isTextInputFocused()` guards global shortcuts when text inputs are active

### Testing

- Unit tests for service logic with mock store interfaces
- Integration tests against real PostgreSQL (`niotebook_test` database)
- TUI tests using `teatest` package from Charm
- Table-driven tests for validation logic
- Test database cleaned via TRUNCATE between tests
- Race detector always enabled: `go test -race`
- Coverage target: 80%+ overall

### Database

- UUIDs as primary keys (`gen_random_uuid()`)
- `TIMESTAMPTZ` for all timestamps
- Constraints enforced at both DB and application level
- Cursor-based pagination using `created_at DESC`
- Migrations are sequential numbered pairs: `000001_name.up.sql` / `000001_name.down.sql`

### Git

- Branch strategy: `main` (stable) > `dev` (integration) > `feature/*` or `fix/*`
- Conventional Commits: `feat:`, `fix:`, `docs:`, `test:`, `refactor:`, `chore:`

## Environment Variables

Required for server:
- `NIOTEBOOK_DB_URL` - PostgreSQL connection string
- `NIOTEBOOK_JWT_SECRET` - JWT signing key (min 32 bytes)

Optional: `NIOTEBOOK_PORT`, `NIOTEBOOK_HOST`, `NIOTEBOOK_LOG_LEVEL`, `NIOTEBOOK_CORS_ORIGIN`

For testing: `NIOTEBOOK_TEST_DB_URL` - PostgreSQL connection string for test database

## Go Module

```
module github.com/Akram012388/niotebook-tui
```
