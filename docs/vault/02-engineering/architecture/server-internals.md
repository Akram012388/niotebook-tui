---
title: "Server Internals"
created: 2026-02-15
updated: 2026-02-15
status: accepted
tags: [engineering, server, architecture, internals]
---

# Server Internals

This document specifies server-side details that weren't covered elsewhere: middleware ordering, logging, graceful shutdown, background jobs, CORS, and the health endpoint.

## Middleware Chain

Middleware is applied in this exact order (outermost to innermost):

```
Request → Recovery → Logging → Rate Limiting → CORS → Auth → Handler
```

```go
// internal/server/server.go

func NewServer(cfg *Config, pool *pgxpool.Pool) *http.Server {
    mux := http.NewServeMux()

    // Register routes
    registerAuthRoutes(mux, services)
    registerPostRoutes(mux, services)
    registerUserRoutes(mux, services)
    registerTimelineRoutes(mux, services)
    registerHealthRoutes(mux, pool)

    // Apply middleware (outermost first)
    var handler http.Handler = mux
    handler = authMiddleware(handler, cfg.JWTSecret)   // innermost — sets user context
    handler = corsMiddleware(handler)                    // CORS headers
    handler = rateLimitMiddleware(handler)               // per-IP token bucket
    handler = loggingMiddleware(handler)                 // request/response logging
    handler = recoveryMiddleware(handler)                // panic recovery → 500

    return &http.Server{
        Addr:         cfg.Host + ":" + cfg.Port,
        Handler:      handler,
        ReadTimeout:  10 * time.Second,
        WriteTimeout: 15 * time.Second,
        IdleTimeout:  60 * time.Second,
    }
}
```

### Middleware Details

#### Recovery Middleware
- Catches panics in handlers, logs stack trace, returns 500
- Prevents one bad request from crashing the server

#### Logging Middleware
- Logs every request: method, path, status code, duration, client IP
- Uses **structured JSON logging** via `log/slog` (Go 1.21+ stdlib)
- No third-party logging library needed — `slog` is sufficient for MVP

```go
// Log format:
// {"time":"2026-02-15T22:00:00Z","level":"INFO","msg":"request","method":"GET","path":"/api/v1/timeline","status":200,"duration_ms":12,"ip":"1.2.3.4"}
```

- Log levels: `DEBUG` (verbose, for dev), `INFO` (requests, startup), `WARN` (rate limits hit, token refresh), `ERROR` (500s, DB failures)
- Default: `INFO`. Set via `NIOTEBOOK_LOG_LEVEL` env var.
- Output: stdout (captured by systemd journal in production)

#### Rate Limiting Middleware
- Uses `golang.org/x/time/rate` per-IP token bucket ([[02-engineering/adr/ADR-0014-rate-limiting|ADR-0014]])
- Applied **before** auth middleware — protects unauthenticated endpoints
- Exempt paths: `/health` (never rate limited)

#### CORS Middleware
- For MVP (TUI-only client), CORS is minimal:
  - `Access-Control-Allow-Origin: *` (the TUI doesn't send Origin headers, but `*` allows curl/httpie testing)
  - `Access-Control-Allow-Methods: GET, POST, PATCH, OPTIONS`
  - `Access-Control-Allow-Headers: Authorization, Content-Type`
- If a web frontend is added post-MVP, tighten to specific origins

#### Auth Middleware
- Extracts `Authorization: Bearer <token>` header
- Validates JWT signature and expiry
- Sets user claims in request context: `ctx = context.WithValue(ctx, userCtxKey, claims)`
- **Exempt paths** (no auth required): `/api/v1/auth/register`, `/api/v1/auth/login`, `/api/v1/auth/refresh`, `/health`
- On invalid/expired token: returns 401 with appropriate error code (`unauthorized` or `token_expired`)

## Brute Force Protection

The rate limiting middleware handles brute force protection. There is **no separate account lockout mechanism**. The JWT document's mention of "5 failed attempts → 15 minute block" is superseded by the rate limit configuration:

| Endpoint | Rate | Burst | Effect |
|----------|------|-------|--------|
| `/api/v1/auth/login` | 10 req/min | 5 | After 5 rapid attempts, subsequent requests are throttled. After 10 in a minute, all are rejected with 429. |
| `/api/v1/auth/register` | 10 req/min | 5 | Same protection against mass registration. |

This is simpler and more robust than tracking failed login attempts per account (which requires DB writes on every failed login and is vulnerable to account-locking DoS attacks).

## Graceful Shutdown

```go
// cmd/server/main.go

func main() {
    // ... setup ...

    srv := server.NewServer(cfg, pool)

    // Start server in goroutine
    go func() {
        slog.Info("server starting", "addr", srv.Addr)
        if err := srv.ListenAndServe(); err != http.ErrServerClosed {
            slog.Error("server error", "err", err)
            os.Exit(1)
        }
    }()

    // Start background cleanup job
    cleanupCtx, cleanupCancel := context.WithCancel(context.Background())
    go runTokenCleanup(cleanupCtx, pool)

    // Wait for interrupt signal
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
    <-quit

    slog.Info("shutting down server...")

    // Cancel background jobs
    cleanupCancel()

    // Drain in-flight requests (30s timeout)
    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
    defer shutdownCancel()

    if err := srv.Shutdown(shutdownCtx); err != nil {
        slog.Error("forced shutdown", "err", err)
    }

    // Close database pool
    pool.Close()

    slog.Info("server stopped")
}
```

**Behavior on SIGINT/SIGTERM:**
1. Stop accepting new connections
2. Cancel background goroutines (token cleanup)
3. Wait up to 30 seconds for in-flight requests to complete
4. Close database connection pool
5. Exit cleanly

## Background Jobs

### Token Cleanup

```go
func runTokenCleanup(ctx context.Context, pool *pgxpool.Pool) {
    ticker := time.NewTicker(1 * time.Hour)
    defer ticker.Stop()

    for {
        select {
        case <-ctx.Done():
            return
        case <-ticker.C:
            n, err := deleteExpiredRefreshTokens(ctx, pool)
            if err != nil {
                slog.Error("token cleanup failed", "err", err)
                // Don't crash — retry next hour
            } else if n > 0 {
                slog.Info("cleaned up expired tokens", "count", n)
            }
        }
    }
}
```

- Runs every hour, starting 1 hour after server boot
- Deletes tokens where `expires_at < NOW()`
- Logs errors but does not crash — retries on next tick
- Cancelled on graceful shutdown via context

## Health Endpoint

```
GET /health
```

No authentication required. Not rate limited.

**Response (200 OK):**
```json
{
  "status": "ok",
  "version": "0.1.0"
}
```

**Response (503 Service Unavailable)** — if DB is unreachable:
```json
{
  "status": "error",
  "message": "database connection failed"
}
```

Implementation:
```go
func healthHandler(pool *pgxpool.Pool) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
        defer cancel()

        if err := pool.Ping(ctx); err != nil {
            w.WriteHeader(http.StatusServiceUnavailable)
            json.NewEncoder(w).Encode(map[string]string{
                "status":  "error",
                "message": "database connection failed",
            })
            return
        }

        json.NewEncoder(w).Encode(map[string]string{
            "status":  "ok",
            "version": version, // set at build time
        })
    }
}
```

Caddy can use this for health checks:
```
api.niotebook.com {
    reverse_proxy localhost:8080 {
        health_uri /health
        health_interval 30s
    }
}
```

## .env Loading

The server uses **`os.Getenv()`** directly — no third-party `.env` library.

In development, developers source the `.env` file before running:
```bash
# Option A: source in shell
source .env && go run ./cmd/server

# Option B: use direnv (auto-loads .env when entering directory)
# .envrc: dotenv

# Option C: Makefile loads it
# The Makefile `dev` target assumes .env is sourced
```

This avoids adding a `godotenv` dependency. In production, environment variables are set in the systemd `EnvironmentFile`.

## Request Body Size Limit

All endpoints enforce a maximum request body of **4KB** (posts are max 140 chars, registration fields are short). Applied in the handler layer:

```go
r.Body = http.MaxBytesReader(w, r.Body, 4096)
```

If exceeded, returns 400 with `{"error": {"code": "validation_error", "message": "Request body too large"}}`.

## Database Connection Pool

pgx pool configuration:

```go
poolConfig, _ := pgxpool.ParseConfig(cfg.DatabaseURL)
poolConfig.MinConns = 2
poolConfig.MaxConns = 10
poolConfig.MaxConnLifetime = 1 * time.Hour
poolConfig.MaxConnIdleTime = 30 * time.Minute
poolConfig.HealthCheckPeriod = 1 * time.Minute
```

- **MinConns 2:** Keep 2 warm connections ready
- **MaxConns 10:** Sufficient for single-server MVP (PostgreSQL default max is 100)
- **Health check every minute:** Detect stale connections proactively

## Binary Version Embedding

Version is embedded at build time via Go linker flags:

```go
// internal/build/version.go
var (
    Version   = "dev"      // overridden at build time
    CommitSHA = "unknown"  // overridden at build time
)
```

```makefile
VERSION ?= $(shell git describe --tags --always --dirty)
COMMIT  ?= $(shell git rev-parse --short HEAD)
LDFLAGS  = -X github.com/Akram012388/niotebook-tui/internal/build.Version=$(VERSION) \
           -X github.com/Akram012388/niotebook-tui/internal/build.CommitSHA=$(COMMIT)

server:
	go build -ldflags "$(LDFLAGS)" -o bin/niotebook-server ./cmd/server

tui:
	go build -ldflags "$(LDFLAGS)" -o bin/niotebook-tui ./cmd/tui
```

The `--version` flag on either binary prints: `niotebook-tui v0.1.0 (abc1234)`.

## Go Module Path

The Go module is:

```
module github.com/Akram012388/niotebook-tui
```

This matches the GitHub repository URL. All internal imports use this path prefix:
- `github.com/Akram012388/niotebook-tui/internal/models`
- `github.com/Akram012388/niotebook-tui/internal/server/handler`
- `github.com/Akram012388/niotebook-tui/cmd/server`
