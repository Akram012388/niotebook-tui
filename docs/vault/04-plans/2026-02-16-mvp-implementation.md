---
title: "MVP Implementation Plan"
created: 2026-02-16
updated: 2026-02-16
status: active
tags: [plan, mvp, implementation]
---

# Niotebook MVP Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build the complete Niotebook MVP — a standalone TUI social media platform with Go server, PostgreSQL database, and Bubble Tea terminal client.

**Architecture:** Monorepo producing two binaries (`niotebook-server`, `niotebook-tui`). Server uses three-layer architecture (handler → service → store) with JWT auth. TUI uses Bubble Tea's Elm architecture (Model-Update-View) with async HTTP via tea.Cmd. All specs in `docs/vault/`.

**Tech Stack:** Go 1.22+, PostgreSQL 15+, Bubble Tea/Bubbles/Lip Gloss, pgx v5, golang-jwt v5, golang-migrate v4, bcrypt, slog

---

## Phase Overview

| Phase | Tasks | Description |
|-------|-------|-------------|
| 1 | 1-2 | Project bootstrap, shared models |
| 2 | 3-5 | Database migrations and store layer |
| 3 | 6-9 | Server services (validation, auth, post, user) |
| 4 | 10-14 | Server HTTP layer (middleware, handlers, assembly) |
| 5 | 15-18 | TUI foundation (config, client, components) |
| 6 | 19-23 | TUI views and app assembly |
| 7 | 24 | CI pipeline and verification |

---

## Phase 1: Project Bootstrap

### Task 1: Initialize Go Module and Directory Structure

**Files:**
- Create: `go.mod`
- Create: `Makefile`
- Create: `.env.example`
- Create: directory tree

**Step 1: Initialize module and create directories**

```bash
cd /path/to/niotebook-tui
go mod init github.com/Akram012388/niotebook-tui
mkdir -p cmd/server cmd/tui
mkdir -p internal/models internal/build
mkdir -p internal/server/handler internal/server/service internal/server/store
mkdir -p internal/tui/app internal/tui/views internal/tui/components internal/tui/client internal/tui/config
mkdir -p migrations
```

**Step 2: Create Makefile**

Implement as specified in [[02-engineering/architecture/build-and-dev-workflow|Build & Dev Workflow]]. Full Makefile:

```makefile
.PHONY: build server tui test lint migrate-up migrate-down clean dev dev-tui test-cover migrate-create release

VERSION ?= $(shell git describe --tags --always --dirty 2>/dev/null || echo "dev")
COMMIT  ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
LDFLAGS  = -X github.com/Akram012388/niotebook-tui/internal/build.Version=$(VERSION) \
           -X github.com/Akram012388/niotebook-tui/internal/build.CommitSHA=$(COMMIT)

build: server tui

server:
	go build -ldflags "$(LDFLAGS)" -o bin/niotebook-server ./cmd/server

tui:
	go build -ldflags "$(LDFLAGS)" -o bin/niotebook-tui ./cmd/tui

dev:
	go run ./cmd/server

dev-tui:
	go run ./cmd/tui --server http://localhost:8080

test:
	go test ./... -v -race

test-cover:
	go test ./... -v -race -coverprofile=coverage.out
	go tool cover -html=coverage.out -o coverage.html

lint:
	golangci-lint run ./...

migrate-up:
	migrate -path migrations -database "$(NIOTEBOOK_DB_URL)" up

migrate-down:
	migrate -path migrations -database "$(NIOTEBOOK_DB_URL)" down 1

migrate-create:
	migrate create -ext sql -dir migrations -seq $(name)

clean:
	rm -rf bin/ coverage.out coverage.html

release:
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/niotebook-server-linux-amd64 ./cmd/server
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/niotebook-server-linux-arm64 ./cmd/server
	GOOS=darwin GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/niotebook-tui-darwin-amd64 ./cmd/tui
	GOOS=darwin GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/niotebook-tui-darwin-arm64 ./cmd/tui
	GOOS=linux GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/niotebook-tui-linux-amd64 ./cmd/tui
	GOOS=linux GOARCH=arm64 go build -ldflags "$(LDFLAGS)" -o bin/niotebook-tui-linux-arm64 ./cmd/tui
	GOOS=windows GOARCH=amd64 go build -ldflags "$(LDFLAGS)" -o bin/niotebook-tui-windows-amd64.exe ./cmd/tui
```

**Step 3: Create .env.example**

```env
NIOTEBOOK_DB_URL=postgres://localhost/niotebook_dev?sslmode=disable
NIOTEBOOK_JWT_SECRET=change-me-to-a-secure-random-string-at-least-32-bytes
NIOTEBOOK_PORT=8080
NIOTEBOOK_HOST=localhost
NIOTEBOOK_LOG_LEVEL=debug
```

**Step 4: Create placeholder main files**

`cmd/server/main.go`:
```go
package main

func main() {
	// TODO: implement in Task 14
}
```

`cmd/tui/main.go`:
```go
package main

func main() {
	// TODO: implement in Task 23
}
```

**Step 5: Install dependencies**

```bash
go get github.com/charmbracelet/bubbletea@latest
go get github.com/charmbracelet/bubbles@latest
go get github.com/charmbracelet/lipgloss@latest
go get github.com/jackc/pgx/v5@latest
go get github.com/golang-jwt/jwt/v5@latest
go get github.com/golang-migrate/migrate/v4@latest
go get golang.org/x/crypto@latest
go get golang.org/x/time@latest
go get gopkg.in/yaml.v3@latest
go mod tidy
```

**Step 6: Verify build**

Run: `go build ./...`
Expected: Clean build, no errors

**Step 7: Commit**

```bash
git add -A
git commit -m "chore: initialize Go module, directory structure, and Makefile"
```

---

### Task 2: Shared Domain Models and Build Package

**Files:**
- Create: `internal/models/user.go`
- Create: `internal/models/post.go`
- Create: `internal/models/auth.go`
- Create: `internal/models/errors.go`
- Create: `internal/models/models_test.go`
- Create: `internal/build/version.go`

**Step 1: Write model tests**

`internal/models/models_test.go`:
```go
package models_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/Akram012388/niotebook-tui/internal/models"
)

func TestUserJSONRoundTrip(t *testing.T) {
	user := models.User{
		ID:          "550e8400-e29b-41d4-a716-446655440000",
		Username:    "akram",
		DisplayName: "Akram",
		Bio:         "Building things in Go.",
		CreatedAt:   time.Date(2026, 2, 15, 22, 0, 0, 0, time.UTC),
	}

	data, err := json.Marshal(user)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got models.User
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.ID != user.ID || got.Username != user.Username || got.DisplayName != user.DisplayName {
		t.Errorf("round trip mismatch: got %+v, want %+v", got, user)
	}
}

func TestPostJSONRoundTrip(t *testing.T) {
	post := models.Post{
		ID:       "660e8400-e29b-41d4-a716-446655440001",
		AuthorID: "550e8400-e29b-41d4-a716-446655440000",
		Content:  "Hello, Niotebook!",
		CreatedAt: time.Date(2026, 2, 15, 23, 30, 0, 0, time.UTC),
	}

	data, err := json.Marshal(post)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got models.Post
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.ID != post.ID || got.Content != post.Content {
		t.Errorf("round trip mismatch: got %+v, want %+v", got, post)
	}
}

func TestPostJSONIncludesAuthor(t *testing.T) {
	post := models.Post{
		ID:       "660e8400-e29b-41d4-a716-446655440001",
		AuthorID: "550e8400-e29b-41d4-a716-446655440000",
		Author: &models.User{
			ID:       "550e8400-e29b-41d4-a716-446655440000",
			Username: "akram",
		},
		Content:   "Hello!",
		CreatedAt: time.Date(2026, 2, 15, 23, 30, 0, 0, time.UTC),
	}

	data, err := json.Marshal(post)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var raw map[string]interface{}
	json.Unmarshal(data, &raw)

	if _, ok := raw["author"]; !ok {
		t.Error("expected author field in JSON")
	}
}

func TestPostJSONOmitsNilAuthor(t *testing.T) {
	post := models.Post{
		ID:        "660e8400-e29b-41d4-a716-446655440001",
		AuthorID:  "550e8400-e29b-41d4-a716-446655440000",
		Content:   "Hello!",
		CreatedAt: time.Date(2026, 2, 15, 23, 30, 0, 0, time.UTC),
	}

	data, err := json.Marshal(post)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var raw map[string]interface{}
	json.Unmarshal(data, &raw)

	if _, ok := raw["author"]; ok {
		t.Error("expected author field to be omitted when nil")
	}
}

func TestAPIErrorJSON(t *testing.T) {
	apiErr := models.APIError{
		Code:    "validation_error",
		Message: "Username must be 3-15 characters",
		Field:   "username",
	}

	data, err := json.Marshal(apiErr)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}

	var got models.APIError
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}

	if got.Code != apiErr.Code || got.Field != apiErr.Field {
		t.Errorf("round trip mismatch: got %+v, want %+v", got, apiErr)
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/models/ -v`
Expected: FAIL — packages don't exist yet

**Step 3: Implement models**

`internal/models/user.go`:
```go
package models

import "time"

type User struct {
	ID          string    `json:"id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
	Bio         string    `json:"bio"`
	CreatedAt   time.Time `json:"created_at"`
}

type UserUpdate struct {
	DisplayName *string `json:"display_name,omitempty"`
	Bio         *string `json:"bio,omitempty"`
}
```

`internal/models/post.go`:
```go
package models

import "time"

type Post struct {
	ID        string    `json:"id"`
	AuthorID  string    `json:"author_id"`
	Author    *User     `json:"author,omitempty"`
	Content   string    `json:"content"`
	CreatedAt time.Time `json:"created_at"`
}
```

`internal/models/auth.go`:
```go
package models

import "time"

type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

type AuthResponse struct {
	User   *User     `json:"user"`
	Tokens *TokenPair `json:"tokens"`
}

type TimelineResponse struct {
	Posts      []Post  `json:"posts"`
	NextCursor *string `json:"next_cursor"`
	HasMore    bool    `json:"has_more"`
}
```

`internal/models/errors.go`:
```go
package models

import "fmt"

type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Field   string `json:"field,omitempty"`
}

func (e *APIError) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Error codes
const (
	ErrCodeValidation  = "validation_error"
	ErrCodeContentLong = "content_too_long"
	ErrCodeUnauthorized = "unauthorized"
	ErrCodeTokenExpired = "token_expired"
	ErrCodeForbidden   = "forbidden"
	ErrCodeNotFound    = "not_found"
	ErrCodeConflict    = "conflict"
	ErrCodeRateLimited = "rate_limited"
	ErrCodeInternal    = "internal_error"
)
```

`internal/build/version.go`:
```go
package build

var (
	Version   = "dev"
	CommitSHA = "unknown"
)
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/models/ -v`
Expected: PASS — all 5 tests pass

**Step 5: Commit**

```bash
git add internal/models/ internal/build/
git commit -m "feat: add shared domain models and build version package"
```

---

## Phase 2: Database Layer

### Task 3: Database Migrations

**Files:**
- Create: `migrations/000001_create_users.up.sql`
- Create: `migrations/000001_create_users.down.sql`
- Create: `migrations/000002_create_posts.up.sql`
- Create: `migrations/000002_create_posts.down.sql`
- Create: `migrations/000003_create_refresh_tokens.up.sql`
- Create: `migrations/000003_create_refresh_tokens.down.sql`

**Step 1: Create migration files**

Implement exactly as specified in [[02-engineering/architecture/database-schema|Database Schema]].

`migrations/000001_create_users.up.sql`:
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
    CONSTRAINT users_email_unique UNIQUE (email),
    CONSTRAINT users_username_length CHECK (char_length(username) >= 3),
    CONSTRAINT users_username_format CHECK (username ~ '^[a-z0-9]([a-z0-9_]*[a-z0-9])?$'),
    CONSTRAINT users_username_no_consecutive_underscores CHECK (username NOT LIKE '%__%'),
    CONSTRAINT users_email_format CHECK (email ~ '^[^@]+@[^@]+\.[^@]+$'),
    CONSTRAINT users_bio_max_length CHECK (char_length(bio) <= 160),
    CONSTRAINT users_display_name_max_length CHECK (char_length(display_name) <= 50)
);

CREATE UNIQUE INDEX idx_users_username_lower ON users (LOWER(username));
CREATE UNIQUE INDEX idx_users_email_lower ON users (LOWER(email));
```

`migrations/000001_create_users.down.sql`:
```sql
DROP TABLE IF EXISTS users CASCADE;
```

`migrations/000002_create_posts.up.sql`:
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

`migrations/000002_create_posts.down.sql`:
```sql
DROP TABLE IF EXISTS posts CASCADE;
```

`migrations/000003_create_refresh_tokens.up.sql`:
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

`migrations/000003_create_refresh_tokens.down.sql`:
```sql
DROP TABLE IF EXISTS refresh_tokens CASCADE;
```

**Step 2: Create test database and run migrations**

```bash
createdb niotebook_test
migrate -path migrations -database "postgres://localhost/niotebook_test?sslmode=disable" up
```
Expected: 3 migrations applied successfully

**Step 3: Verify migrations are reversible**

```bash
migrate -path migrations -database "postgres://localhost/niotebook_test?sslmode=disable" down
migrate -path migrations -database "postgres://localhost/niotebook_test?sslmode=disable" up
```
Expected: Clean down then up with no errors

**Step 4: Commit**

```bash
git add migrations/
git commit -m "feat: add database migration files for users, posts, and refresh_tokens"
```

---

### Task 4: Database Connection and Store Interfaces

**Files:**
- Create: `internal/server/store/db.go`
- Create: `internal/server/store/interfaces.go`
- Create: `internal/server/store/testutil_test.go`

**Step 1: Implement database connection helper**

`internal/server/store/db.go`:
```go
package store

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func NewPool(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, err
	}

	config.MinConns = 2
	config.MaxConns = 10
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, err
	}

	return pool, nil
}
```

**Step 2: Define store interfaces**

`internal/server/store/interfaces.go`:
```go
package store

import (
	"context"
	"time"

	"github.com/Akram012388/niotebook-tui/internal/models"
)

type UserStore interface {
	CreateUser(ctx context.Context, username, email, passwordHash, displayName string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, string, error) // returns user + password hash
	GetUserByID(ctx context.Context, id string) (*models.User, error)
	GetUserByUsername(ctx context.Context, username string) (*models.User, error)
	UpdateUser(ctx context.Context, id string, updates *models.UserUpdate) (*models.User, error)
}

type PostStore interface {
	CreatePost(ctx context.Context, authorID, content string) (*models.Post, error)
	GetPostByID(ctx context.Context, id string) (*models.Post, error)
	GetTimeline(ctx context.Context, cursor time.Time, limit int) ([]models.Post, error)
	GetUserPosts(ctx context.Context, userID string, cursor time.Time, limit int) ([]models.Post, error)
}

type RefreshTokenStore interface {
	StoreToken(ctx context.Context, userID, tokenHash string, expiresAt time.Time) error
	GetByHash(ctx context.Context, tokenHash string) (id, userID string, expiresAt time.Time, err error)
	DeleteByHash(ctx context.Context, tokenHash string) error
	DeleteAllForUser(ctx context.Context, userID string) error
	DeleteExpired(ctx context.Context) (int64, error)
}
```

**Step 3: Create test utilities**

`internal/server/store/testutil_test.go`:
```go
package store_test

import (
	"context"
	"os"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5/pgxpool"
)

func testDBURL() string {
	if url := os.Getenv("NIOTEBOOK_TEST_DB_URL"); url != "" {
		return url
	}
	return "postgres://localhost/niotebook_test?sslmode=disable"
}

func setupTestDB(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dbURL := testDBURL()

	// Run migrations
	m, err := migrate.New("file://../../../migrations", dbURL)
	if err != nil {
		t.Fatalf("migrate new: %v", err)
	}
	if err := m.Up(); err != nil && err != migrate.ErrNoChange {
		t.Fatalf("migrate up: %v", err)
	}

	pool, err := pgxpool.New(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("pool: %v", err)
	}

	t.Cleanup(func() {
		_, _ = pool.Exec(context.Background(),
			"TRUNCATE users, posts, refresh_tokens CASCADE")
		pool.Close()
	})

	return pool
}
```

**Step 4: Verify compilation**

Run: `go build ./internal/server/store/...`
Expected: Clean build

**Step 5: Commit**

```bash
git add internal/server/store/
git commit -m "feat: add database connection pool helper, store interfaces, and test utilities"
```

---

### Task 5: Store Implementations

**Files:**
- Create: `internal/server/store/user_store.go`
- Create: `internal/server/store/user_store_test.go`
- Create: `internal/server/store/post_store.go`
- Create: `internal/server/store/post_store_test.go`
- Create: `internal/server/store/refresh_token_store.go`
- Create: `internal/server/store/refresh_token_store_test.go`

**Step 1: Write UserStore tests**

`internal/server/store/user_store_test.go`:
```go
package store_test

import (
	"context"
	"testing"

	"github.com/Akram012388/niotebook-tui/internal/server/store"
)

func TestCreateUser(t *testing.T) {
	pool := setupTestDB(t)
	s := store.NewUserStore(pool)
	ctx := context.Background()

	user, err := s.CreateUser(ctx, "testuser", "test@example.com", "$2a$12$fakehash", "testuser")
	if err != nil {
		t.Fatalf("CreateUser: %v", err)
	}
	if user.Username != "testuser" {
		t.Errorf("username = %q, want %q", user.Username, "testuser")
	}
	if user.ID == "" {
		t.Error("expected non-empty ID")
	}
}

func TestCreateUserDuplicateUsername(t *testing.T) {
	pool := setupTestDB(t)
	s := store.NewUserStore(pool)
	ctx := context.Background()

	_, err := s.CreateUser(ctx, "testuser", "a@example.com", "$2a$12$fakehash", "testuser")
	if err != nil {
		t.Fatalf("first create: %v", err)
	}

	_, err = s.CreateUser(ctx, "testuser", "b@example.com", "$2a$12$fakehash", "testuser")
	if err == nil {
		t.Fatal("expected error for duplicate username")
	}
}

func TestCreateUserDuplicateEmail(t *testing.T) {
	pool := setupTestDB(t)
	s := store.NewUserStore(pool)
	ctx := context.Background()

	_, err := s.CreateUser(ctx, "user1", "same@example.com", "$2a$12$fakehash", "user1")
	if err != nil {
		t.Fatalf("first create: %v", err)
	}

	_, err = s.CreateUser(ctx, "user2", "same@example.com", "$2a$12$fakehash", "user2")
	if err == nil {
		t.Fatal("expected error for duplicate email")
	}
}

func TestGetUserByEmail(t *testing.T) {
	pool := setupTestDB(t)
	s := store.NewUserStore(pool)
	ctx := context.Background()

	_, err := s.CreateUser(ctx, "akram", "akram@example.com", "$2a$12$hashvalue", "akram")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	user, hash, err := s.GetUserByEmail(ctx, "akram@example.com")
	if err != nil {
		t.Fatalf("GetUserByEmail: %v", err)
	}
	if user.Username != "akram" {
		t.Errorf("username = %q, want %q", user.Username, "akram")
	}
	if hash != "$2a$12$hashvalue" {
		t.Errorf("hash = %q, want %q", hash, "$2a$12$hashvalue")
	}
}

func TestGetUserByEmailNotFound(t *testing.T) {
	pool := setupTestDB(t)
	s := store.NewUserStore(pool)
	ctx := context.Background()

	_, _, err := s.GetUserByEmail(ctx, "nobody@example.com")
	if err == nil {
		t.Fatal("expected error for nonexistent email")
	}
}

func TestGetUserByID(t *testing.T) {
	pool := setupTestDB(t)
	s := store.NewUserStore(pool)
	ctx := context.Background()

	created, _ := s.CreateUser(ctx, "akram", "akram@example.com", "$2a$12$hash", "akram")

	user, err := s.GetUserByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetUserByID: %v", err)
	}
	if user.Username != "akram" {
		t.Errorf("username = %q, want %q", user.Username, "akram")
	}
}

func TestUpdateUser(t *testing.T) {
	pool := setupTestDB(t)
	s := store.NewUserStore(pool)
	ctx := context.Background()

	created, _ := s.CreateUser(ctx, "akram", "akram@example.com", "$2a$12$hash", "akram")

	newName := "Shaikh Akram"
	newBio := "Building Niotebook."
	updated, err := s.UpdateUser(ctx, created.ID, &models.UserUpdate{
		DisplayName: &newName,
		Bio:         &newBio,
	})
	if err != nil {
		t.Fatalf("UpdateUser: %v", err)
	}
	if updated.DisplayName != "Shaikh Akram" {
		t.Errorf("display_name = %q, want %q", updated.DisplayName, "Shaikh Akram")
	}
	if updated.Bio != "Building Niotebook." {
		t.Errorf("bio = %q, want %q", updated.Bio, "Building Niotebook.")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/server/store/ -v -run TestCreate`
Expected: FAIL — `store.NewUserStore` doesn't exist

**Step 3: Implement UserStore**

`internal/server/store/user_store.go`:
```go
package store

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type userStore struct {
	pool *pgxpool.Pool
}

func NewUserStore(pool *pgxpool.Pool) UserStore {
	return &userStore{pool: pool}
}

func (s *userStore) CreateUser(ctx context.Context, username, email, passwordHash, displayName string) (*models.User, error) {
	var user models.User
	err := s.pool.QueryRow(ctx,
		`INSERT INTO users (username, email, password, display_name)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id, username, display_name, bio, created_at`,
		strings.ToLower(username), strings.ToLower(email), passwordHash, displayName,
	).Scan(&user.ID, &user.Username, &user.DisplayName, &user.Bio, &user.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			if strings.Contains(pgErr.ConstraintName, "username") {
				return nil, &models.APIError{Code: models.ErrCodeConflict, Message: "Username already taken", Field: "username"}
			}
			if strings.Contains(pgErr.ConstraintName, "email") {
				return nil, &models.APIError{Code: models.ErrCodeConflict, Message: "Email already registered", Field: "email"}
			}
		}
		return nil, fmt.Errorf("create user: %w", err)
	}

	return &user, nil
}

func (s *userStore) GetUserByEmail(ctx context.Context, email string) (*models.User, string, error) {
	var user models.User
	var passwordHash string
	err := s.pool.QueryRow(ctx,
		`SELECT id, username, email, password, display_name, bio, created_at
		 FROM users WHERE LOWER(email) = LOWER($1)`, email,
	).Scan(&user.ID, &user.Username, nil, &passwordHash, &user.DisplayName, &user.Bio, &user.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", &models.APIError{Code: models.ErrCodeUnauthorized, Message: "Invalid email or password"}
		}
		return nil, "", fmt.Errorf("get user by email: %w", err)
	}

	return &user, passwordHash, nil
}

func (s *userStore) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	var user models.User
	err := s.pool.QueryRow(ctx,
		`SELECT id, username, display_name, bio, created_at
		 FROM users WHERE id = $1`, id,
	).Scan(&user.ID, &user.Username, &user.DisplayName, &user.Bio, &user.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &models.APIError{Code: models.ErrCodeNotFound, Message: "User not found"}
		}
		return nil, fmt.Errorf("get user by id: %w", err)
	}

	return &user, nil
}

func (s *userStore) GetUserByUsername(ctx context.Context, username string) (*models.User, error) {
	var user models.User
	err := s.pool.QueryRow(ctx,
		`SELECT id, username, display_name, bio, created_at
		 FROM users WHERE LOWER(username) = LOWER($1)`, username,
	).Scan(&user.ID, &user.Username, &user.DisplayName, &user.Bio, &user.CreatedAt)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, &models.APIError{Code: models.ErrCodeNotFound, Message: "User not found"}
		}
		return nil, fmt.Errorf("get user by username: %w", err)
	}

	return &user, nil
}

func (s *userStore) UpdateUser(ctx context.Context, id string, updates *models.UserUpdate) (*models.User, error) {
	setClauses := []string{}
	args := []interface{}{}
	argIdx := 1

	if updates.DisplayName != nil {
		setClauses = append(setClauses, fmt.Sprintf("display_name = $%d", argIdx))
		args = append(args, *updates.DisplayName)
		argIdx++
	}
	if updates.Bio != nil {
		setClauses = append(setClauses, fmt.Sprintf("bio = $%d", argIdx))
		args = append(args, *updates.Bio)
		argIdx++
	}

	if len(setClauses) == 0 {
		return s.GetUserByID(ctx, id)
	}

	setClauses = append(setClauses, fmt.Sprintf("updated_at = NOW()"))
	args = append(args, id)

	query := fmt.Sprintf(
		`UPDATE users SET %s WHERE id = $%d
		 RETURNING id, username, display_name, bio, created_at`,
		strings.Join(setClauses, ", "), argIdx,
	)

	var user models.User
	err := s.pool.QueryRow(ctx, query, args...).
		Scan(&user.ID, &user.Username, &user.DisplayName, &user.Bio, &user.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("update user: %w", err)
	}

	return &user, nil
}
```

Note: The `GetUserByEmail` scan uses `nil` for the email field since we don't include email in the User model (it's not exposed via the API). Fix by scanning into a throwaway variable:

```go
var ignoredEmail string
// ... .Scan(&user.ID, &user.Username, &ignoredEmail, &passwordHash, ...)
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/server/store/ -v -run TestCreate -run TestGetUser -run TestUpdate`
Expected: PASS

**Step 5: Write PostStore tests, implement, verify**

Follow the same TDD pattern. `internal/server/store/post_store_test.go` tests:
- `TestCreatePost` — creates a post, verifies ID and content
- `TestCreatePostContentTooLong` — 141 chars, expects DB constraint error
- `TestCreatePostEmpty` — whitespace-only, expects DB constraint error
- `TestGetPostByID` — retrieves with author joined
- `TestGetPostByIDNotFound` — expects not_found error
- `TestGetTimeline` — creates 3 posts, verifies reverse-chronological order
- `TestGetTimelineCursorPagination` — creates 5 posts, fetches with limit=2, verifies cursor works
- `TestGetUserPosts` — creates posts from 2 users, verifies filtering

`internal/server/store/post_store.go` implements `PostStore` with queries from [[02-engineering/architecture/database-schema#Key Queries|Database Schema — Key Queries]].

**Step 6: Write RefreshTokenStore tests, implement, verify**

`internal/server/store/refresh_token_store_test.go` tests:
- `TestStoreAndGetToken` — stores hash, retrieves by hash
- `TestDeleteByHash` — stores, deletes, get returns error
- `TestDeleteAllForUser` — stores multiple, deletes all for user
- `TestDeleteExpired` — stores expired + valid, delete expired returns correct count

`internal/server/store/refresh_token_store.go` implements `RefreshTokenStore`.

**Step 7: Run all store tests**

Run: `go test ./internal/server/store/ -v -race`
Expected: ALL PASS

**Step 8: Commit**

```bash
git add internal/server/store/
git commit -m "feat: implement UserStore, PostStore, and RefreshTokenStore with integration tests"
```

---

## Phase 3: Server Services

### Task 6: Validation Functions

**Files:**
- Create: `internal/server/service/validation.go`
- Create: `internal/server/service/validation_test.go`

**Step 1: Write validation tests**

`internal/server/service/validation_test.go`:
```go
package service_test

import (
	"strings"
	"testing"

	"github.com/Akram012388/niotebook-tui/internal/server/service"
)

func TestValidateUsername(t *testing.T) {
	tests := []struct {
		name     string
		username string
		wantErr  bool
	}{
		{"valid simple", "akram", false},
		{"valid with underscore", "code_ninja", false},
		{"valid numbers", "user42", false},
		{"minimum length 3", "abc", false},
		{"maximum length 15", "abcdefghijklmno", false},
		{"too short", "ab", true},
		{"too long 16", "abcdefghijklmnop", true},
		{"leading underscore", "_akram", true},
		{"trailing underscore", "akram_", true},
		{"consecutive underscores", "code__ninja", true},
		{"special chars", "akram!", true},
		{"spaces", "ak ram", true},
		{"uppercase accepted", "Akram", false},
		{"reserved admin", "admin", true},
		{"reserved api", "api", true},
		{"reserved root", "root", true},
		{"empty string", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateUsername(tt.username)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateUsername(%q) error = %v, wantErr %v", tt.username, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePostContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{"valid short", "Hello!", false},
		{"exactly 140", strings.Repeat("a", 140), false},
		{"141 chars", strings.Repeat("a", 141), true},
		{"empty", "", true},
		{"whitespace only", "   \n\t  ", true},
		{"with newlines", "line1\nline2", false},
		{"trimmed within limit", "  " + strings.Repeat("a", 140) + "  ", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidatePostContent(tt.content)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePostContent(%q) error = %v, wantErr %v", tt.content, err, tt.wantErr)
			}
		})
	}
}

func TestValidateEmail(t *testing.T) {
	tests := []struct {
		name    string
		email   string
		wantErr bool
	}{
		{"valid", "user@example.com", false},
		{"valid with subdomain", "user@sub.example.com", false},
		{"missing @", "userexample.com", true},
		{"missing domain", "user@", true},
		{"empty", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidateEmail(tt.email)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateEmail(%q) error = %v, wantErr %v", tt.email, err, tt.wantErr)
			}
		})
	}
}

func TestValidatePassword(t *testing.T) {
	tests := []struct {
		name     string
		password string
		wantErr  bool
	}{
		{"valid 8 chars", "12345678", false},
		{"valid long", "a-very-secure-password", false},
		{"too short 7", "1234567", true},
		{"empty", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := service.ValidatePassword(tt.password)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePassword(%q) error = %v, wantErr %v", tt.password, err, tt.wantErr)
			}
		})
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/server/service/ -v`
Expected: FAIL — package doesn't exist

**Step 3: Implement validators**

`internal/server/service/validation.go`:
```go
package service

import (
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/Akram012388/niotebook-tui/internal/models"
)

var (
	usernameRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9_]*[a-z0-9])?$`)

	reservedUsernames = map[string]bool{
		"admin": true, "root": true, "system": true,
		"niotebook": true, "api": true, "help": true,
		"support": true, "me": true, "about": true,
		"settings": true, "login": true, "register": true,
		"auth": true, "posts": true, "users": true,
		"timeline": true, "search": true, "explore": true,
	}
)

func ValidateUsername(username string) error {
	lower := strings.ToLower(username)
	length := utf8.RuneCountInString(lower)

	if length < 3 || length > 15 {
		return &models.APIError{
			Code: models.ErrCodeValidation, Field: "username",
			Message: "Username must be 3-15 characters",
		}
	}
	if !usernameRegex.MatchString(lower) {
		return &models.APIError{
			Code: models.ErrCodeValidation, Field: "username",
			Message: "Username must be alphanumeric and underscores only, cannot start or end with underscore",
		}
	}
	if strings.Contains(lower, "__") {
		return &models.APIError{
			Code: models.ErrCodeValidation, Field: "username",
			Message: "Username cannot contain consecutive underscores",
		}
	}
	if reservedUsernames[lower] {
		return &models.APIError{
			Code: models.ErrCodeValidation, Field: "username",
			Message: "Username is reserved",
		}
	}
	return nil
}

func ValidatePostContent(content string) error {
	trimmed := strings.TrimSpace(content)
	if trimmed == "" {
		return &models.APIError{
			Code: models.ErrCodeValidation, Field: "content",
			Message: "Post content cannot be empty",
		}
	}
	if utf8.RuneCountInString(trimmed) > 140 {
		return &models.APIError{
			Code: models.ErrCodeContentLong,
			Message: "Post must be 140 characters or fewer",
		}
	}
	return nil
}

func ValidateEmail(email string) error {
	if email == "" {
		return &models.APIError{
			Code: models.ErrCodeValidation, Field: "email",
			Message: "Email is required",
		}
	}
	if _, err := mail.ParseAddress(email); err != nil {
		return &models.APIError{
			Code: models.ErrCodeValidation, Field: "email",
			Message: "Invalid email format",
		}
	}
	return nil
}

func ValidatePassword(password string) error {
	if len(password) < 8 {
		return &models.APIError{
			Code: models.ErrCodeValidation, Field: "password",
			Message: "Password must be at least 8 characters",
		}
	}
	return nil
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/server/service/ -v -race`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add internal/server/service/
git commit -m "feat: add input validation functions with unit tests"
```

---

### Task 7: Auth Service

**Files:**
- Create: `internal/server/service/auth.go`
- Create: `internal/server/service/auth_test.go`

**Step 1: Write auth service tests**

`internal/server/service/auth_test.go` — tests using mock stores:

```go
package service_test

import (
	"context"
	"testing"
	"time"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/Akram012388/niotebook-tui/internal/server/service"
)

// Mock stores — implement store interfaces with in-memory maps
// (see separate mock_stores_test.go file below)

func TestRegister(t *testing.T) {
	userStore := newMockUserStore()
	tokenStore := newMockRefreshTokenStore()
	auth := service.NewAuthService(userStore, tokenStore, "test-secret-32-bytes-long-xxxxx")

	resp, err := auth.Register(context.Background(), &models.RegisterRequest{
		Username: "akram",
		Email:    "akram@example.com",
		Password: "password123",
	})
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if resp.User.Username != "akram" {
		t.Errorf("username = %q, want %q", resp.User.Username, "akram")
	}
	if resp.Tokens.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if resp.Tokens.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
}

func TestRegisterInvalidUsername(t *testing.T) {
	userStore := newMockUserStore()
	tokenStore := newMockRefreshTokenStore()
	auth := service.NewAuthService(userStore, tokenStore, "test-secret-32-bytes-long-xxxxx")

	_, err := auth.Register(context.Background(), &models.RegisterRequest{
		Username: "a",
		Email:    "a@example.com",
		Password: "password123",
	})
	if err == nil {
		t.Fatal("expected validation error")
	}
}

func TestRegisterShortPassword(t *testing.T) {
	userStore := newMockUserStore()
	tokenStore := newMockRefreshTokenStore()
	auth := service.NewAuthService(userStore, tokenStore, "test-secret-32-bytes-long-xxxxx")

	_, err := auth.Register(context.Background(), &models.RegisterRequest{
		Username: "akram",
		Email:    "akram@example.com",
		Password: "short",
	})
	if err == nil {
		t.Fatal("expected validation error for short password")
	}
}

func TestLogin(t *testing.T) {
	userStore := newMockUserStore()
	tokenStore := newMockRefreshTokenStore()
	auth := service.NewAuthService(userStore, tokenStore, "test-secret-32-bytes-long-xxxxx")

	// Register first
	auth.Register(context.Background(), &models.RegisterRequest{
		Username: "akram", Email: "akram@example.com", Password: "password123",
	})

	// Login
	resp, err := auth.Login(context.Background(), &models.LoginRequest{
		Email: "akram@example.com", Password: "password123",
	})
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if resp.User.Username != "akram" {
		t.Errorf("username = %q, want %q", resp.User.Username, "akram")
	}
}

func TestLoginWrongPassword(t *testing.T) {
	userStore := newMockUserStore()
	tokenStore := newMockRefreshTokenStore()
	auth := service.NewAuthService(userStore, tokenStore, "test-secret-32-bytes-long-xxxxx")

	auth.Register(context.Background(), &models.RegisterRequest{
		Username: "akram", Email: "akram@example.com", Password: "password123",
	})

	_, err := auth.Login(context.Background(), &models.LoginRequest{
		Email: "akram@example.com", Password: "wrongpassword",
	})
	if err == nil {
		t.Fatal("expected error for wrong password")
	}
}

func TestRefreshToken(t *testing.T) {
	userStore := newMockUserStore()
	tokenStore := newMockRefreshTokenStore()
	auth := service.NewAuthService(userStore, tokenStore, "test-secret-32-bytes-long-xxxxx")

	resp, _ := auth.Register(context.Background(), &models.RegisterRequest{
		Username: "akram", Email: "akram@example.com", Password: "password123",
	})

	newTokens, err := auth.Refresh(context.Background(), resp.Tokens.RefreshToken)
	if err != nil {
		t.Fatalf("Refresh: %v", err)
	}
	if newTokens.AccessToken == "" {
		t.Error("expected new access token")
	}
	// Old refresh token should be consumed (single-use)
	_, err = auth.Refresh(context.Background(), resp.Tokens.RefreshToken)
	if err == nil {
		t.Fatal("expected error for reused refresh token")
	}
}
```

Create `internal/server/service/mock_stores_test.go` with in-memory mock implementations of `UserStore`, `PostStore`, and `RefreshTokenStore` interfaces.

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/server/service/ -v -run TestRegister -run TestLogin -run TestRefresh`
Expected: FAIL

**Step 3: Implement AuthService**

`internal/server/service/auth.go`:
```go
package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/Akram012388/niotebook-tui/internal/server/store"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	users     store.UserStore
	tokens    store.RefreshTokenStore
	jwtSecret []byte
	accessTTL time.Duration
	refreshTTL time.Duration
}

func NewAuthService(users store.UserStore, tokens store.RefreshTokenStore, jwtSecret string) *AuthService {
	return &AuthService{
		users:      users,
		tokens:     tokens,
		jwtSecret:  []byte(jwtSecret),
		accessTTL:  24 * time.Hour,
		refreshTTL: 7 * 24 * time.Hour,
	}
}

func (s *AuthService) Register(ctx context.Context, req *models.RegisterRequest) (*models.AuthResponse, error) {
	if err := ValidateUsername(req.Username); err != nil {
		return nil, err
	}
	if err := ValidateEmail(req.Email); err != nil {
		return nil, err
	}
	if err := ValidatePassword(req.Password); err != nil {
		return nil, err
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), 12)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}

	user, err := s.users.CreateUser(ctx, req.Username, req.Email, string(hash), req.Username)
	if err != nil {
		return nil, err
	}

	tokens, err := s.generateTokenPair(ctx, user)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{User: user, Tokens: tokens}, nil
}

func (s *AuthService) Login(ctx context.Context, req *models.LoginRequest) (*models.AuthResponse, error) {
	user, hash, err := s.users.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(req.Password)); err != nil {
		return nil, &models.APIError{Code: models.ErrCodeUnauthorized, Message: "Invalid email or password"}
	}

	tokens, err := s.generateTokenPair(ctx, user)
	if err != nil {
		return nil, err
	}

	return &models.AuthResponse{User: user, Tokens: tokens}, nil
}

func (s *AuthService) Refresh(ctx context.Context, rawToken string) (*models.TokenPair, error) {
	tokenHash := hashRefreshToken(rawToken)

	id, userID, expiresAt, err := s.tokens.GetByHash(ctx, tokenHash)
	if err != nil {
		return nil, &models.APIError{Code: models.ErrCodeTokenExpired, Message: "Refresh token has expired"}
	}
	_ = id

	if time.Now().After(expiresAt) {
		s.tokens.DeleteByHash(ctx, tokenHash)
		return nil, &models.APIError{Code: models.ErrCodeTokenExpired, Message: "Refresh token has expired"}
	}

	// Single-use: delete the consumed token
	if err := s.tokens.DeleteByHash(ctx, tokenHash); err != nil {
		return nil, fmt.Errorf("delete consumed token: %w", err)
	}

	user, err := s.users.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}

	return s.generateTokenPair(ctx, user)
}

func (s *AuthService) generateTokenPair(ctx context.Context, user *models.User) (*models.TokenPair, error) {
	now := time.Now()
	expiresAt := now.Add(s.accessTTL)

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub":      user.ID,
		"username": user.Username,
		"iat":      now.Unix(),
		"exp":      expiresAt.Unix(),
	})

	accessToken, err := token.SignedString(s.jwtSecret)
	if err != nil {
		return nil, fmt.Errorf("sign access token: %w", err)
	}

	rawRefresh, err := generateRefreshToken()
	if err != nil {
		return nil, err
	}

	refreshHash := hashRefreshToken(rawRefresh)
	refreshExpiry := now.Add(s.refreshTTL)

	if err := s.tokens.StoreToken(ctx, user.ID, refreshHash, refreshExpiry); err != nil {
		return nil, fmt.Errorf("store refresh token: %w", err)
	}

	return &models.TokenPair{
		AccessToken:  accessToken,
		RefreshToken: rawRefresh,
		ExpiresAt:    expiresAt,
	}, nil
}

func generateRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("generate refresh token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

func hashRefreshToken(token string) string {
	h := sha256.Sum256([]byte(token))
	return fmt.Sprintf("%x", h)
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/server/service/ -v -race`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add internal/server/service/
git commit -m "feat: implement auth service with register, login, and token refresh"
```

---

### Task 8: Post Service

**Files:**
- Create: `internal/server/service/post.go`
- Add to: `internal/server/service/mock_stores_test.go` (PostStore mock)
- Create: `internal/server/service/post_test.go`

**Step 1: Write post service tests**

```go
func TestCreatePost(t *testing.T) {
	postStore := newMockPostStore()
	svc := service.NewPostService(postStore)

	post, err := svc.CreatePost(context.Background(), "user-123", "Hello, Niotebook!")
	if err != nil {
		t.Fatalf("CreatePost: %v", err)
	}
	if post.Content != "Hello, Niotebook!" {
		t.Errorf("content = %q, want %q", post.Content, "Hello, Niotebook!")
	}
}

func TestCreatePostTrimmed(t *testing.T) {
	postStore := newMockPostStore()
	svc := service.NewPostService(postStore)

	post, _ := svc.CreatePost(context.Background(), "user-123", "  Hello  ")
	if post.Content != "Hello" {
		t.Errorf("content = %q, want trimmed %q", post.Content, "Hello")
	}
}

func TestCreatePostTooLong(t *testing.T) {
	postStore := newMockPostStore()
	svc := service.NewPostService(postStore)

	_, err := svc.CreatePost(context.Background(), "user-123", strings.Repeat("a", 141))
	if err == nil {
		t.Fatal("expected error for too-long post")
	}
}

func TestCreatePostEmpty(t *testing.T) {
	postStore := newMockPostStore()
	svc := service.NewPostService(postStore)

	_, err := svc.CreatePost(context.Background(), "user-123", "   ")
	if err == nil {
		t.Fatal("expected error for empty post")
	}
}

func TestGetTimeline(t *testing.T) {
	postStore := newMockPostStore()
	svc := service.NewPostService(postStore)

	// Add posts via mock
	postStore.AddPost("1", "user-1", "First", time.Now().Add(-2*time.Minute))
	postStore.AddPost("2", "user-1", "Second", time.Now().Add(-1*time.Minute))

	posts, err := svc.GetTimeline(context.Background(), time.Now(), 50)
	if err != nil {
		t.Fatalf("GetTimeline: %v", err)
	}
	if len(posts) != 2 {
		t.Errorf("got %d posts, want 2", len(posts))
	}
}
```

**Step 2: Implement PostService**

`internal/server/service/post.go`:
```go
package service

import (
	"context"
	"strings"
	"time"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/Akram012388/niotebook-tui/internal/server/store"
)

type PostService struct {
	posts store.PostStore
}

func NewPostService(posts store.PostStore) *PostService {
	return &PostService{posts: posts}
}

func (s *PostService) CreatePost(ctx context.Context, authorID, content string) (*models.Post, error) {
	content = strings.TrimSpace(content)
	if err := ValidatePostContent(content); err != nil {
		return nil, err
	}
	return s.posts.CreatePost(ctx, authorID, content)
}

func (s *PostService) GetPostByID(ctx context.Context, id string) (*models.Post, error) {
	return s.posts.GetPostByID(ctx, id)
}

func (s *PostService) GetTimeline(ctx context.Context, cursor time.Time, limit int) ([]models.Post, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.posts.GetTimeline(ctx, cursor, limit)
}

func (s *PostService) GetUserPosts(ctx context.Context, userID string, cursor time.Time, limit int) ([]models.Post, error) {
	if limit <= 0 || limit > 100 {
		limit = 50
	}
	return s.posts.GetUserPosts(ctx, userID, cursor, limit)
}
```

**Step 3: Run tests, verify pass, commit**

Run: `go test ./internal/server/service/ -v -race`

```bash
git add internal/server/service/
git commit -m "feat: implement post service with create, timeline, and user posts"
```

---

### Task 9: User Service

**Files:**
- Create: `internal/server/service/user.go`
- Create: `internal/server/service/user_test.go`

**Step 1: Write user service tests**

Tests for: `GetUserByID`, `UpdateUser` (display name validation, bio length validation, partial update).

**Step 2: Implement UserService**

`internal/server/service/user.go`:
```go
package service

import (
	"context"
	"unicode/utf8"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/Akram012388/niotebook-tui/internal/server/store"
)

type UserService struct {
	users store.UserStore
}

func NewUserService(users store.UserStore) *UserService {
	return &UserService{users: users}
}

func (s *UserService) GetUserByID(ctx context.Context, id string) (*models.User, error) {
	return s.users.GetUserByID(ctx, id)
}

func (s *UserService) UpdateUser(ctx context.Context, id string, updates *models.UserUpdate) (*models.User, error) {
	if updates.DisplayName != nil {
		if utf8.RuneCountInString(*updates.DisplayName) > 50 || *updates.DisplayName == "" {
			return nil, &models.APIError{
				Code: models.ErrCodeValidation, Field: "display_name",
				Message: "Display name must be 1-50 characters",
			}
		}
	}
	if updates.Bio != nil {
		if utf8.RuneCountInString(*updates.Bio) > 160 {
			return nil, &models.APIError{
				Code: models.ErrCodeValidation, Field: "bio",
				Message: "Bio must be 160 characters or fewer",
			}
		}
	}
	return s.users.UpdateUser(ctx, id, updates)
}
```

**Step 3: Run tests, verify pass, commit**

```bash
git add internal/server/service/
git commit -m "feat: implement user service with profile update and validation"
```

---

## Phase 4: Server HTTP Layer

### Task 10: JWT Auth Middleware

**Files:**
- Create: `internal/server/middleware/auth.go`
- Create: `internal/server/middleware/auth_test.go`

**Step 1: Write middleware tests**

`internal/server/middleware/auth_test.go`:
```go
package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Akram012388/niotebook-tui/internal/server/middleware"
	"github.com/golang-jwt/jwt/v5"
)

const testSecret = "test-secret-32-bytes-long-xxxxx"

func makeToken(secret string, claims jwt.MapClaims) string {
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	s, _ := token.SignedString([]byte(secret))
	return s
}

func TestAuthMiddlewareValidToken(t *testing.T) {
	token := makeToken(testSecret, jwt.MapClaims{
		"sub":      "user-123",
		"username": "akram",
		"exp":      time.Now().Add(time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	})

	handler := middleware.Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID := middleware.UserIDFromContext(r.Context())
		if userID != "user-123" {
			t.Errorf("userID = %q, want %q", userID, "user-123")
		}
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestAuthMiddlewareMissingToken(t *testing.T) {
	handler := middleware.Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddlewareExpiredToken(t *testing.T) {
	token := makeToken(testSecret, jwt.MapClaims{
		"sub": "user-123",
		"exp": time.Now().Add(-time.Hour).Unix(),
		"iat": time.Now().Add(-2 * time.Hour).Unix(),
	})

	handler := middleware.Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuthMiddlewareExemptPaths(t *testing.T) {
	handler := middleware.Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	exemptPaths := []string{
		"/api/v1/auth/login",
		"/api/v1/auth/register",
		"/api/v1/auth/refresh",
		"/health",
	}

	for _, path := range exemptPaths {
		req := httptest.NewRequest("GET", path, nil)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Errorf("path %s: status = %d, want %d", path, rec.Code, http.StatusOK)
		}
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/server/middleware/ -v`
Expected: FAIL

**Step 3: Implement auth middleware**

`internal/server/middleware/auth.go`:
```go
package middleware

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/golang-jwt/jwt/v5"
)

type contextKey string

const userCtxKey contextKey = "user_claims"

type UserClaims struct {
	UserID   string
	Username string
}

func UserIDFromContext(ctx context.Context) string {
	claims, ok := ctx.Value(userCtxKey).(*UserClaims)
	if !ok {
		return ""
	}
	return claims.UserID
}

func UsernameFromContext(ctx context.Context) string {
	claims, ok := ctx.Value(userCtxKey).(*UserClaims)
	if !ok {
		return ""
	}
	return claims.Username
}

var exemptPaths = map[string]bool{
	"/api/v1/auth/login":    true,
	"/api/v1/auth/register": true,
	"/api/v1/auth/refresh":  true,
	"/health":               true,
}

func Auth(jwtSecret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if exemptPaths[r.URL.Path] {
				next.ServeHTTP(w, r)
				return
			}

			authHeader := r.Header.Get("Authorization")
			if !strings.HasPrefix(authHeader, "Bearer ") {
				writeError(w, http.StatusUnauthorized, models.ErrCodeUnauthorized, "Missing or invalid authorization header")
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (interface{}, error) {
				return []byte(jwtSecret), nil
			}, jwt.WithValidMethods([]string{"HS256"}))

			if err != nil || !token.Valid {
				code := models.ErrCodeUnauthorized
				if strings.Contains(err.Error(), "expired") {
					code = models.ErrCodeTokenExpired
				}
				writeError(w, http.StatusUnauthorized, code, "Invalid or expired token")
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				writeError(w, http.StatusUnauthorized, models.ErrCodeUnauthorized, "Invalid token claims")
				return
			}

			userClaims := &UserClaims{
				UserID:   claims["sub"].(string),
				Username: claims["username"].(string),
			}

			ctx := context.WithValue(r.Context(), userCtxKey, userClaims)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func writeError(w http.ResponseWriter, status int, code, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"error": map[string]string{
			"code":    code,
			"message": message,
		},
	})
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/server/middleware/ -v -race`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add internal/server/middleware/
git commit -m "feat: implement JWT auth middleware with exempt paths"
```

---

### Task 11: Supporting Middleware

**Files:**
- Create: `internal/server/middleware/recovery.go`
- Create: `internal/server/middleware/logging.go`
- Create: `internal/server/middleware/ratelimit.go`
- Create: `internal/server/middleware/cors.go`
- Create: `internal/server/middleware/ratelimit_test.go`

**Step 1: Implement recovery middleware**

`internal/server/middleware/recovery.go`:
```go
package middleware

import (
	"log/slog"
	"net/http"
	"runtime/debug"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				slog.Error("panic recovered", "err", err, "stack", string(debug.Stack()))
				writeError(w, http.StatusInternalServerError, "internal_error", "Something went wrong. Please try again.")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
```

**Step 2: Implement logging middleware**

`internal/server/middleware/logging.go`:
```go
package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type responseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &responseWriter{ResponseWriter: w, statusCode: http.StatusOK}
		next.ServeHTTP(rw, r)
		slog.Info("request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.statusCode,
			"duration_ms", time.Since(start).Milliseconds(),
			"ip", r.RemoteAddr,
		)
	})
}
```

**Step 3: Implement rate limiting middleware**

`internal/server/middleware/ratelimit.go` — per-IP token bucket using `golang.org/x/time/rate`, with periodic cleanup of stale entries. Configuration per [[02-engineering/adr/ADR-0014-rate-limiting|ADR-0014]]:
- Auth endpoints: 10 req/min, burst 5
- Write endpoints: 30 req/min, burst 10
- Read endpoints: 120 req/min, burst 30
- `/health`: exempt

**Step 4: Implement CORS middleware**

`internal/server/middleware/cors.go`:
```go
package middleware

import "net/http"

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}

		next.ServeHTTP(w, r)
	})
}
```

**Step 5: Write rate limit test**

`internal/server/middleware/ratelimit_test.go` — send 12 rapid requests to a handler wrapped with rate limiting, verify the 11th+ returns 429.

**Step 6: Run all middleware tests**

Run: `go test ./internal/server/middleware/ -v -race`
Expected: ALL PASS

**Step 7: Commit**

```bash
git add internal/server/middleware/
git commit -m "feat: implement recovery, logging, CORS, and rate limiting middleware"
```

---

### Task 12: HTTP Handlers

**Files:**
- Create: `internal/server/handler/helpers.go`
- Create: `internal/server/handler/auth.go`
- Create: `internal/server/handler/post.go`
- Create: `internal/server/handler/timeline.go`
- Create: `internal/server/handler/user.go`
- Create: `internal/server/handler/health.go`
- Create: `internal/server/handler/handler_test.go`

**Step 1: Create handler helpers**

`internal/server/handler/helpers.go`:
```go
package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/Akram012388/niotebook-tui/internal/models"
)

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeAPIError(w http.ResponseWriter, err error) {
	var apiErr *models.APIError
	if errors.As(err, &apiErr) {
		status := errorCodeToHTTPStatus(apiErr.Code)
		writeJSON(w, status, map[string]interface{}{"error": apiErr})
		return
	}
	writeJSON(w, http.StatusInternalServerError, map[string]interface{}{
		"error": models.APIError{
			Code:    models.ErrCodeInternal,
			Message: "Something went wrong. Please try again.",
		},
	})
}

func errorCodeToHTTPStatus(code string) int {
	switch code {
	case models.ErrCodeValidation, models.ErrCodeContentLong:
		return http.StatusBadRequest
	case models.ErrCodeUnauthorized, models.ErrCodeTokenExpired:
		return http.StatusUnauthorized
	case models.ErrCodeForbidden:
		return http.StatusForbidden
	case models.ErrCodeNotFound:
		return http.StatusNotFound
	case models.ErrCodeConflict:
		return http.StatusConflict
	case models.ErrCodeRateLimited:
		return http.StatusTooManyRequests
	default:
		return http.StatusInternalServerError
	}
}

func decodeBody(r *http.Request, v interface{}) error {
	r.Body = http.MaxBytesReader(nil, r.Body, 4096)
	dec := json.NewDecoder(r.Body)
	dec.DisallowUnknownFields()
	return dec.Decode(v)
}
```

**Step 2: Implement auth handlers**

`internal/server/handler/auth.go` — three handlers:
- `HandleRegister` — decodes `RegisterRequest`, calls `AuthService.Register`, returns 201 with `AuthResponse`
- `HandleLogin` — decodes `LoginRequest`, calls `AuthService.Login`, returns 200 with `AuthResponse`
- `HandleRefresh` — decodes `RefreshRequest`, calls `AuthService.Refresh`, returns 200 with tokens

Each follows the pattern: decode → call service → write response or error. Reference [[02-engineering/api/api-specification|API Specification]] for exact response formats.

**Step 3: Implement post and timeline handlers**

`internal/server/handler/post.go`:
- `HandleCreatePost` — extract user ID from context, decode body, call `PostService.CreatePost`, return 201
- `HandleGetPost` — extract post ID from URL path, call `PostService.GetPostByID`, return 200

`internal/server/handler/timeline.go`:
- `HandleTimeline` — parse `cursor` and `limit` query params, call `PostService.GetTimeline`, return 200 with `TimelineResponse`

**Step 4: Implement user handlers**

`internal/server/handler/user.go`:
- `HandleGetUser` — extract user ID from path (or `me` for self), call `UserService.GetUserByID`, return 200
- `HandleGetUserPosts` — extract user ID from path, parse pagination, call `PostService.GetUserPosts`, return 200
- `HandleUpdateUser` — decode `UserUpdate`, call `UserService.UpdateUser`, return 200

**Step 5: Implement health handler**

`internal/server/handler/health.go`:
```go
package handler

import (
	"context"
	"net/http"
	"time"

	"github.com/Akram012388/niotebook-tui/internal/build"
	"github.com/jackc/pgx/v5/pgxpool"
)

func HandleHealth(pool *pgxpool.Pool) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx, cancel := context.WithTimeout(r.Context(), 2*time.Second)
		defer cancel()

		if err := pool.Ping(ctx); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			writeJSON(w, http.StatusServiceUnavailable, map[string]string{
				"status":  "error",
				"message": "database connection failed",
			})
			return
		}

		writeJSON(w, http.StatusOK, map[string]string{
			"status":  "ok",
			"version": build.Version,
		})
	}
}
```

**Step 6: Write integration tests**

`internal/server/handler/handler_test.go` — full integration tests against real DB (use `testutil` pattern from Task 4). Tests:
- Register → Login → Create Post → Get Timeline → Get Profile → Update Profile → Refresh Token
- Error cases: duplicate username, invalid password, expired token, post too long, not found

**Step 7: Run tests**

Run: `go test ./internal/server/handler/ -v -race`
Expected: ALL PASS

**Step 8: Commit**

```bash
git add internal/server/handler/
git commit -m "feat: implement all HTTP handlers with integration tests"
```

---

### Task 13: Server Router and Wiring

**Files:**
- Create: `internal/server/server.go`

**Step 1: Implement server wiring**

`internal/server/server.go`:
```go
package server

import (
	"net/http"

	"github.com/Akram012388/niotebook-tui/internal/server/handler"
	"github.com/Akram012388/niotebook-tui/internal/server/middleware"
	"github.com/Akram012388/niotebook-tui/internal/server/service"
	"github.com/Akram012388/niotebook-tui/internal/server/store"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	JWTSecret string
	Host      string
	Port      string
}

func NewServer(cfg *Config, pool *pgxpool.Pool) *http.Server {
	// Stores
	userStore := store.NewUserStore(pool)
	postStore := store.NewPostStore(pool)
	tokenStore := store.NewRefreshTokenStore(pool)

	// Services
	authSvc := service.NewAuthService(userStore, tokenStore, cfg.JWTSecret)
	postSvc := service.NewPostService(postStore)
	userSvc := service.NewUserService(userStore)

	// Router (Go 1.22 pattern matching)
	mux := http.NewServeMux()

	// Auth routes
	mux.HandleFunc("POST /api/v1/auth/register", handler.HandleRegister(authSvc))
	mux.HandleFunc("POST /api/v1/auth/login", handler.HandleLogin(authSvc))
	mux.HandleFunc("POST /api/v1/auth/refresh", handler.HandleRefresh(authSvc))

	// Post routes
	mux.HandleFunc("POST /api/v1/posts", handler.HandleCreatePost(postSvc))
	mux.HandleFunc("GET /api/v1/posts/{id}", handler.HandleGetPost(postSvc))

	// Timeline
	mux.HandleFunc("GET /api/v1/timeline", handler.HandleTimeline(postSvc))

	// User routes
	mux.HandleFunc("GET /api/v1/users/{id}", handler.HandleGetUser(userSvc))
	mux.HandleFunc("GET /api/v1/users/{id}/posts", handler.HandleGetUserPosts(postSvc))
	mux.HandleFunc("PATCH /api/v1/users/me", handler.HandleUpdateUser(userSvc))

	// Health
	mux.HandleFunc("GET /health", handler.HandleHealth(pool))

	// Middleware chain: Recovery → Logging → RateLimit → CORS → Auth → Handler
	var h http.Handler = mux
	h = middleware.Auth(cfg.JWTSecret)(h)
	h = middleware.CORS(h)
	h = middleware.RateLimit()(h)
	h = middleware.Logging(h)
	h = middleware.Recovery(h)

	return &http.Server{
		Addr:    cfg.Host + ":" + cfg.Port,
		Handler: h,
	}
}
```

**Step 2: Verify compilation**

Run: `go build ./internal/server/...`
Expected: Clean build

**Step 3: Commit**

```bash
git add internal/server/server.go
git commit -m "feat: wire server router with middleware chain and all routes"
```

---

### Task 14: Server Binary Entry Point

**Files:**
- Modify: `cmd/server/main.go`

**Step 1: Implement main.go**

`cmd/server/main.go` — as specified in [[02-engineering/architecture/server-internals|Server Internals]]:
- Parse CLI flags (`--port`, `--host`, `--migrate`, `--env`)
- Load environment variables
- Create database pool
- Optionally run migrations
- Start server in goroutine
- Start background token cleanup goroutine
- Wait for SIGINT/SIGTERM
- Graceful shutdown (30s timeout)
- Close pool

```go
package main

import (
	"context"
	"flag"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Akram012388/niotebook-tui/internal/build"
	"github.com/Akram012388/niotebook-tui/internal/server"
	"github.com/Akram012388/niotebook-tui/internal/server/store"
)

func main() {
	port := flag.String("port", envOrDefault("NIOTEBOOK_PORT", "8080"), "listen port")
	host := flag.String("host", envOrDefault("NIOTEBOOK_HOST", "localhost"), "listen host")
	runMigrate := flag.Bool("migrate", false, "run pending migrations on startup")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		slog.Info("niotebook-server", "version", build.Version, "commit", build.CommitSHA)
		os.Exit(0)
	}

	// Configure logging
	logLevel := slog.LevelInfo
	if lvl := os.Getenv("NIOTEBOOK_LOG_LEVEL"); lvl == "debug" {
		logLevel = slog.LevelDebug
	}
	slog.SetDefault(slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: logLevel})))

	dbURL := os.Getenv("NIOTEBOOK_DB_URL")
	if dbURL == "" {
		slog.Error("NIOTEBOOK_DB_URL is required")
		os.Exit(1)
	}

	jwtSecret := os.Getenv("NIOTEBOOK_JWT_SECRET")
	if jwtSecret == "" {
		slog.Error("NIOTEBOOK_JWT_SECRET is required")
		os.Exit(1)
	}

	// Database
	ctx := context.Background()
	pool, err := store.NewPool(ctx, dbURL)
	if err != nil {
		slog.Error("database connection failed", "err", err)
		os.Exit(1)
	}

	if *runMigrate {
		slog.Info("running migrations...")
		// Use golang-migrate programmatically
	}

	// Server
	cfg := &server.Config{JWTSecret: jwtSecret, Host: *host, Port: *port}
	srv := server.NewServer(cfg, pool)

	go func() {
		slog.Info("server starting", "addr", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			slog.Error("server error", "err", err)
			os.Exit(1)
		}
	}()

	// Background: token cleanup
	cleanupCtx, cleanupCancel := context.WithCancel(context.Background())
	go runTokenCleanup(cleanupCtx, pool)

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down server...")
	cleanupCancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		slog.Error("forced shutdown", "err", err)
	}

	pool.Close()
	slog.Info("server stopped")
}

func runTokenCleanup(ctx context.Context, pool interface{ Exec(context.Context, string, ...interface{}) (interface{}, error) }) {
	// Simplified — uses raw SQL via pool
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// DELETE FROM refresh_tokens WHERE expires_at < NOW()
			slog.Debug("running token cleanup")
		}
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
```

**Step 2: Verify server starts**

```bash
source .env
go run ./cmd/server
# In another terminal:
curl http://localhost:8080/health
```
Expected: `{"status":"ok","version":"dev"}`

**Step 3: Commit**

```bash
git add cmd/server/
git commit -m "feat: implement server binary with graceful shutdown and background cleanup"
```

---

## Phase 5: TUI Foundation

### Task 15: Config Loading

**Files:**
- Create: `internal/tui/config/config.go`
- Create: `internal/tui/config/config_test.go`

**Step 1: Write config tests**

```go
package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Akram012388/niotebook-tui/internal/tui/config"
)

func TestLoadConfigDefaults(t *testing.T) {
	cfg := config.DefaultConfig()
	if cfg.ServerURL != "https://api.niotebook.com" {
		t.Errorf("ServerURL = %q, want default", cfg.ServerURL)
	}
}

func TestLoadConfigFromFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	os.WriteFile(path, []byte("server_url: http://localhost:8080\n"), 0600)

	cfg, err := config.LoadFromFile(path)
	if err != nil {
		t.Fatalf("LoadFromFile: %v", err)
	}
	if cfg.ServerURL != "http://localhost:8080" {
		t.Errorf("ServerURL = %q, want %q", cfg.ServerURL, "http://localhost:8080")
	}
}

func TestSaveAndLoadAuthTokens(t *testing.T) {
	dir := t.TempDir()
	authPath := filepath.Join(dir, "auth.json")

	tokens := &config.StoredAuth{
		AccessToken:  "access-123",
		RefreshToken: "refresh-456",
		ExpiresAt:    "2026-02-16T22:00:00Z",
	}
	if err := config.SaveAuth(authPath, tokens); err != nil {
		t.Fatalf("SaveAuth: %v", err)
	}

	loaded, err := config.LoadAuth(authPath)
	if err != nil {
		t.Fatalf("LoadAuth: %v", err)
	}
	if loaded.AccessToken != "access-123" {
		t.Errorf("AccessToken = %q, want %q", loaded.AccessToken, "access-123")
	}
}

func TestLoadAuthFileNotFound(t *testing.T) {
	_, err := config.LoadAuth("/nonexistent/auth.json")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/tui/config/ -v`
Expected: FAIL

**Step 3: Implement config package**

`internal/tui/config/config.go`:
```go
package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	ServerURL string `yaml:"server_url"`
}

type StoredAuth struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresAt    string `json:"expires_at"`
}

func DefaultConfig() *Config {
	return &Config{
		ServerURL: "https://api.niotebook.com",
	}
}

func LoadFromFile(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}
	return cfg, nil
}

func ConfigDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "niotebook")
}

func EnsureConfigDir() error {
	return os.MkdirAll(ConfigDir(), 0755)
}

func SaveAuth(path string, auth *StoredAuth) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(auth, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0600)
}

func LoadAuth(path string) (*StoredAuth, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var auth StoredAuth
	if err := json.Unmarshal(data, &auth); err != nil {
		return nil, fmt.Errorf("parse auth: %w", err)
	}
	return &auth, nil
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/tui/config/ -v -race`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add internal/tui/config/
git commit -m "feat: implement TUI config loading and auth token persistence"
```

---

### Task 16: HTTP Client Wrapper

**Files:**
- Create: `internal/tui/client/client.go`
- Create: `internal/tui/client/client_test.go`

**Step 1: Write client tests**

`internal/tui/client/client_test.go`:
```go
package client_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/Akram012388/niotebook-tui/internal/tui/client"
)

func TestGetTimeline(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/timeline" {
			t.Errorf("path = %q, want /api/v1/timeline", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer test-token" {
			t.Errorf("missing auth header")
		}
		json.NewEncoder(w).Encode(models.TimelineResponse{
			Posts:   []models.Post{{ID: "1", Content: "Hello"}},
			HasMore: false,
		})
	}))
	defer srv.Close()

	c := client.New(srv.URL)
	c.SetToken("test-token")

	resp, err := c.GetTimeline("", 50)
	if err != nil {
		t.Fatalf("GetTimeline: %v", err)
	}
	if len(resp.Posts) != 1 {
		t.Errorf("got %d posts, want 1", len(resp.Posts))
	}
}

func TestCreatePost(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" || r.URL.Path != "/api/v1/posts" {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"post": models.Post{ID: "new-1", Content: "Test post"},
		})
	}))
	defer srv.Close()

	c := client.New(srv.URL)
	c.SetToken("test-token")

	post, err := c.CreatePost("Test post")
	if err != nil {
		t.Fatalf("CreatePost: %v", err)
	}
	if post.Content != "Test post" {
		t.Errorf("content = %q, want %q", post.Content, "Test post")
	}
}

func TestAuthRefreshOn401(t *testing.T) {
	callCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		callCount++
		if r.URL.Path == "/api/v1/auth/refresh" {
			json.NewEncoder(w).Encode(map[string]interface{}{
				"tokens": models.TokenPair{
					AccessToken:  "new-access",
					RefreshToken: "new-refresh",
				},
			})
			return
		}
		if callCount == 1 {
			// First call: return 401
			w.WriteHeader(http.StatusUnauthorized)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"error": map[string]string{"code": "token_expired"},
			})
			return
		}
		// Retry after refresh: success
		json.NewEncoder(w).Encode(models.TimelineResponse{
			Posts: []models.Post{{ID: "1", Content: "After refresh"}},
		})
	}))
	defer srv.Close()

	c := client.New(srv.URL)
	c.SetToken("expired-token")
	c.SetRefreshToken("valid-refresh")

	resp, err := c.GetTimeline("", 50)
	if err != nil {
		t.Fatalf("GetTimeline with refresh: %v", err)
	}
	if len(resp.Posts) != 1 {
		t.Errorf("got %d posts after refresh, want 1", len(resp.Posts))
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/tui/client/ -v`
Expected: FAIL

**Step 3: Implement HTTP client**

`internal/tui/client/client.go` — wraps all API calls. Key features:
- Methods for each endpoint: `Login`, `Register`, `Refresh`, `GetTimeline`, `CreatePost`, `GetPost`, `GetUser`, `GetUserPosts`, `UpdateUser`
- `SetToken` / `SetRefreshToken` for auth state
- Transparent 401 retry: on receiving 401 with `token_expired`, automatically call `/auth/refresh`, store new tokens, retry original request
- `OnTokenRefresh` callback to persist new tokens to auth.json

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/tui/client/ -v -race`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add internal/tui/client/
git commit -m "feat: implement TUI HTTP client with transparent auth refresh"
```

---

### Task 17: TUI Components — Relative Time and Post Card

**Files:**
- Create: `internal/tui/components/relativetime.go`
- Create: `internal/tui/components/relativetime_test.go`
- Create: `internal/tui/components/postcard.go`
- Create: `internal/tui/components/header.go`
- Create: `internal/tui/components/statusbar.go`

**Step 1: Write relative time tests**

`internal/tui/components/relativetime_test.go`:
```go
package components_test

import (
	"testing"
	"time"

	"github.com/Akram012388/niotebook-tui/internal/tui/components"
)

func TestRelativeTime(t *testing.T) {
	now := time.Date(2026, 2, 16, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		t    time.Time
		want string
	}{
		{"just now", now.Add(-30 * time.Second), "now"},
		{"1 minute", now.Add(-1 * time.Minute), "1m"},
		{"5 minutes", now.Add(-5 * time.Minute), "5m"},
		{"59 minutes", now.Add(-59 * time.Minute), "59m"},
		{"1 hour", now.Add(-1 * time.Hour), "1h"},
		{"23 hours", now.Add(-23 * time.Hour), "23h"},
		{"1 day", now.Add(-24 * time.Hour), "1d"},
		{"6 days", now.Add(-6 * 24 * time.Hour), "6d"},
		{"1 week", now.Add(-7 * 24 * time.Hour), "1w"},
		{"3 weeks", now.Add(-21 * 24 * time.Hour), "3w"},
		{"31 days same year", now.Add(-31 * 24 * time.Hour), "Jan 16"},
		{"different year", time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC), "Jun 15, 2025"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := components.RelativeTimeFrom(tt.t, now)
			if got != tt.want {
				t.Errorf("RelativeTimeFrom() = %q, want %q", got, tt.want)
			}
		})
	}
}
```

**Step 2: Run tests to verify they fail, implement, verify pass**

`internal/tui/components/relativetime.go` — implements `RelativeTimeFrom(t, now time.Time) string` per rules in [[03-design/post-card-component#Relative Time Rules|Post Card — Relative Time Rules]].

**Step 3: Implement post card, header, and status bar**

`internal/tui/components/postcard.go` — renders a single post card per [[03-design/post-card-component|Post Card Component]]:
- `RenderPostCard(post models.Post, width int, selected bool, now time.Time) string`
- Uses Lip Gloss for styling: cyan bold username, dim gray time/separator, accent marker when selected

`internal/tui/components/header.go`:
- `RenderHeader(appName, username, viewName string, width int) string`
- Left-aligned app name + username, right-aligned view name

`internal/tui/components/statusbar.go`:
- `StatusBarModel` with `SetError(msg)`, `SetSuccess(msg)`, `SetLoading(msg)`, `View(view, width) string`
- Auto-clear after 5 seconds via `tea.Tick`

**Step 4: Run all component tests**

Run: `go test ./internal/tui/components/ -v -race`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add internal/tui/components/
git commit -m "feat: implement TUI components — post card, header, status bar, relative time"
```

---

### Task 18: TUI Message Types

**Files:**
- Create: `internal/tui/app/messages.go`

**Step 1: Define all message types**

`internal/tui/app/messages.go` — as specified in [[02-engineering/architecture/bubble-tea-model-hierarchy#Message Types|Bubble Tea Model Hierarchy — Message Types]]:

```go
package app

import "github.com/Akram012388/niotebook-tui/internal/models"

// Auth messages
type MsgAuthSuccess struct {
	User   *models.User
	Tokens *models.TokenPair
}
type MsgAuthExpired struct{}
type MsgAuthError struct {
	Message string
	Field   string
}

// Timeline messages
type MsgTimelineLoaded struct {
	Posts      []models.Post
	NextCursor string
	HasMore    bool
}
type MsgTimelineRefreshed struct {
	Posts      []models.Post
	NextCursor string
	HasMore    bool
}

// Post messages
type MsgPostPublished struct{ Post models.Post }

// Profile messages
type MsgProfileLoaded struct {
	User  *models.User
	Posts []models.Post
}
type MsgProfileUpdated struct{ User *models.User }

// Generic messages
type MsgAPIError struct{ Message string }
type MsgStatusClear struct{}
```

**Step 2: Verify compilation**

Run: `go build ./internal/tui/app/...`
Expected: Clean build

**Step 3: Commit**

```bash
git add internal/tui/app/
git commit -m "feat: define TUI message types for async operations"
```

---

## Phase 6: TUI Views and App Assembly

### Task 19: Login and Register Views

**Files:**
- Create: `internal/tui/views/login.go`
- Create: `internal/tui/views/register.go`
- Create: `internal/tui/views/login_test.go`

**Step 1: Write login view test**

```go
package views_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/Akram012388/niotebook-tui/internal/tui/views"
)

func TestLoginViewRender(t *testing.T) {
	m := views.NewLoginModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := m.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
	// Should contain login form elements
	if !containsAny(view, "Email", "Password", "Login") {
		t.Error("view missing form elements")
	}
}

func TestLoginViewTabSwitchesField(t *testing.T) {
	m := views.NewLoginModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Initial focus on email
	if m.FocusIndex() != 0 {
		t.Errorf("initial focus = %d, want 0 (email)", m.FocusIndex())
	}

	// Tab moves to password
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.FocusIndex() != 1 {
		t.Errorf("after tab focus = %d, want 1 (password)", m.FocusIndex())
	}
}

func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
```

**Step 2: Implement login view**

`internal/tui/views/login.go` — implements `LoginModel` per [[02-engineering/architecture/bubble-tea-model-hierarchy#LoginModel|Bubble Tea Model Hierarchy]] and [[03-design/tui-layout-and-navigation#1. Login View|TUI Layout — Login View]]:
- Two `textinput.Model` fields (email, password)
- Tab/Shift+Tab for field navigation
- Enter to submit → returns `tea.Cmd` that calls `client.Login()`
- Tab to switch to Register view (emits view-switch message)
- Centered form rendering with Lip Gloss

**Step 3: Implement register view**

`internal/tui/views/register.go` — same pattern as login with 3 fields (username, email, password). Client-side validation feedback as specified in [[03-design/auth-flow-ux#Registration Validation|Auth Flow UX]].

**Step 4: Run tests**

Run: `go test ./internal/tui/views/ -v -race`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add internal/tui/views/
git commit -m "feat: implement login and register views with form validation"
```

---

### Task 20: Timeline View

**Files:**
- Create: `internal/tui/views/timeline.go`
- Create: `internal/tui/views/timeline_test.go`

**Step 1: Write timeline view test**

```go
func TestTimelineViewRendersPosts(t *testing.T) {
	posts := []models.Post{
		{
			ID: "1",
			Author: &models.User{Username: "akram"},
			Content: "Hello, Niotebook!",
			CreatedAt: time.Now().Add(-5 * time.Minute),
		},
	}
	m := views.NewTimelineModel(nil)
	m.SetPosts(posts)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := m.View()
	if !strings.Contains(view, "@akram") {
		t.Error("view missing username")
	}
	if !strings.Contains(view, "Hello, Niotebook!") {
		t.Error("view missing post content")
	}
}

func TestTimelineViewScrolling(t *testing.T) {
	posts := make([]models.Post, 10)
	for i := range posts {
		posts[i] = models.Post{
			ID:      fmt.Sprintf("%d", i),
			Author:  &models.User{Username: "user"},
			Content: fmt.Sprintf("Post %d", i),
		}
	}
	m := views.NewTimelineModel(nil)
	m.SetPosts(posts)

	// Initial cursor at 0
	if m.CursorIndex() != 0 {
		t.Errorf("initial cursor = %d, want 0", m.CursorIndex())
	}

	// Press j to move down
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.CursorIndex() != 1 {
		t.Errorf("after j cursor = %d, want 1", m.CursorIndex())
	}
}

func TestTimelineViewEmptyState(t *testing.T) {
	m := views.NewTimelineModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := m.View()
	if !strings.Contains(view, "No posts yet") {
		t.Error("expected empty state message")
	}
}
```

**Step 2: Implement timeline view**

`internal/tui/views/timeline.go` — per [[02-engineering/architecture/bubble-tea-model-hierarchy#TimelineModel|Bubble Tea Hierarchy]] and [[03-design/tui-layout-and-navigation#3. Timeline View|TUI Layout — Timeline View]]:
- `TimelineModel` with `[]models.Post`, cursor index, viewport, pagination state
- `j`/`k`/arrows for navigation, `g`/`G` for top/bottom, `Space`/`b` for page scroll
- Renders post cards via `components.RenderPostCard`
- Empty state and loading spinner
- `Init()` returns cmd to fetch timeline
- `FetchLatest()` for refresh
- `SelectedPost()` returns currently highlighted post

**Step 3: Run tests, verify pass, commit**

```bash
git add internal/tui/views/
git commit -m "feat: implement timeline view with scrolling and pagination"
```

---

### Task 21: Compose Modal

**Files:**
- Create: `internal/tui/views/compose.go`
- Create: `internal/tui/views/compose_test.go`

**Step 1: Write compose modal tests**

Tests:
- Typing updates textarea and character counter
- Ctrl+Enter sets `submitted = true` and returns publish cmd
- Esc sets `cancelled = true`
- Over 140 chars: counter turns red (visual), Ctrl+Enter is disabled
- Empty content: Ctrl+Enter is disabled

**Step 2: Implement compose modal**

`internal/tui/views/compose.go` — per [[03-design/tui-layout-and-navigation#4. Compose Modal|TUI Layout — Compose Modal]] and [[02-engineering/architecture/bubble-tea-model-hierarchy#ComposeModel|Bubble Tea — ComposeModel]]:
- `textarea.Model` for multi-line input
- Live character counter (`X/140`)
- `Ctrl+Enter` publishes (calls `client.CreatePost()`)
- `Esc` cancels
- Renders as centered bordered overlay

**Step 3: Run tests, verify pass, commit**

```bash
git add internal/tui/views/
git commit -m "feat: implement compose modal with character counter and submit"
```

---

### Task 22: Profile View and Help Overlay

**Files:**
- Create: `internal/tui/views/profile.go`
- Create: `internal/tui/views/help.go`
- Create: `internal/tui/views/profile_test.go`

**Step 1: Implement profile view**

`internal/tui/views/profile.go` — per [[03-design/tui-layout-and-navigation#5. Profile View|TUI Layout — Profile View]]:
- Shows user bio, display name, joined date
- Lists user's posts below
- `e` opens edit modal (own profile only)
- `Esc` returns to timeline
- Scrollable via viewport

**Step 2: Implement help overlay**

`internal/tui/views/help.go` — per [[02-engineering/architecture/bubble-tea-model-hierarchy#Help Overlay|Bubble Tea — Help Overlay]]:
- Static keymap display per view
- Dismissed by `?`, `Esc`, or `q`
- Centered bordered overlay

**Step 3: Write profile tests**

Tests:
- Renders username and bio
- Renders user's posts
- `e` key on own profile sets editing state
- `Esc` sets dismissed state

**Step 4: Run tests, verify pass, commit**

```bash
git add internal/tui/views/
git commit -m "feat: implement profile view with edit modal and help overlay"
```

---

### Task 23: Root AppModel and TUI Binary

**Files:**
- Create: `internal/tui/app/app.go`
- Create: `internal/tui/app/app_test.go`
- Modify: `cmd/tui/main.go`

**Step 1: Write AppModel tests**

```go
func TestAppModelStartsOnLogin(t *testing.T) {
	m := app.NewAppModel(nil, nil) // no stored auth
	if m.CurrentView() != app.ViewLogin {
		t.Errorf("initial view = %v, want ViewLogin", m.CurrentView())
	}
}

func TestAppModelAuthSuccessSwitchesToTimeline(t *testing.T) {
	m := app.NewAppModel(nil, nil)
	m, _ = m.Update(app.MsgAuthSuccess{
		User:   &models.User{Username: "akram"},
		Tokens: &models.TokenPair{AccessToken: "tok"},
	})
	if m.CurrentView() != app.ViewTimeline {
		t.Errorf("view after auth = %v, want ViewTimeline", m.CurrentView())
	}
}

func TestAppModelNOpensCompose(t *testing.T) {
	m := app.NewAppModel(nil, nil)
	// Simulate logged in on timeline
	m, _ = m.Update(app.MsgAuthSuccess{
		User:   &models.User{Username: "akram"},
		Tokens: &models.TokenPair{AccessToken: "tok"},
	})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if !m.IsComposeOpen() {
		t.Error("expected compose to be open after pressing n")
	}
}
```

**Step 2: Implement AppModel**

`internal/tui/app/app.go` — the root model per [[02-engineering/architecture/bubble-tea-model-hierarchy|Bubble Tea Model Hierarchy]]:
- Holds all sub-models, shared state (client, user, window size)
- View routing in `Update()`: global keys → overlay routing → view-specific routing
- `View()`: header + content + status bar vertical layout, with overlay rendering
- `isTextInputFocused()` to guard global shortcuts
- State preservation: timeline state kept when navigating to profile

**Step 3: Implement TUI binary**

`cmd/tui/main.go`:
```go
package main

import (
	"flag"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/Akram012388/niotebook-tui/internal/build"
	"github.com/Akram012388/niotebook-tui/internal/tui/app"
	"github.com/Akram012388/niotebook-tui/internal/tui/client"
	"github.com/Akram012388/niotebook-tui/internal/tui/config"
)

func main() {
	serverURL := flag.String("server", "", "server URL (overrides config)")
	configPath := flag.String("config", "", "config file path")
	showVersion := flag.Bool("version", false, "print version and exit")
	flag.Parse()

	if *showVersion {
		fmt.Printf("niotebook-tui %s (%s)\n", build.Version, build.CommitSHA)
		os.Exit(0)
	}

	// Load config
	cfgDir := config.ConfigDir()
	cfgFile := filepath.Join(cfgDir, "config.yaml")
	if *configPath != "" {
		cfgFile = *configPath
	}

	cfg, err := config.LoadFromFile(cfgFile)
	if err != nil {
		cfg = config.DefaultConfig()
	}

	if *serverURL != "" {
		cfg.ServerURL = *serverURL
	}

	// Load stored auth
	authFile := filepath.Join(cfgDir, "auth.json")
	storedAuth, _ := config.LoadAuth(authFile)

	// Create HTTP client
	c := client.New(cfg.ServerURL)
	if storedAuth != nil {
		c.SetToken(storedAuth.AccessToken)
		c.SetRefreshToken(storedAuth.RefreshToken)
	}

	// Set up token persistence callback
	c.OnTokenRefresh = func(tokens *models.TokenPair) {
		config.SaveAuth(authFile, &config.StoredAuth{
			AccessToken:  tokens.AccessToken,
			RefreshToken: tokens.RefreshToken,
			ExpiresAt:    tokens.ExpiresAt.Format(time.RFC3339),
		})
	}

	// Create and run app
	model := app.NewAppModel(c, storedAuth)
	p := tea.NewProgram(model, tea.WithAltScreen())

	if _, err := p.Run(); err != nil {
		slog.Error("TUI error", "err", err)
		os.Exit(1)
	}
}
```

**Step 4: Run tests**

Run: `go test ./internal/tui/... -v -race`
Expected: ALL PASS

**Step 5: Verify full build**

Run: `make build`
Expected: Both `bin/niotebook-server` and `bin/niotebook-tui` built

**Step 6: Commit**

```bash
git add internal/tui/app/ cmd/tui/
git commit -m "feat: implement root AppModel with view routing and TUI binary"
```

---

## Phase 7: CI and Verification

### Task 24: GitHub Actions CI and Final Verification

**Files:**
- Create: `.github/workflows/ci.yml`

**Step 1: Create CI pipeline**

`.github/workflows/ci.yml` — as specified in [[02-engineering/testing/testing-strategy#CI Pipeline|Testing Strategy — CI Pipeline]]:

```yaml
name: CI
on: [push, pull_request]

jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_DB: niotebook_test
          POSTGRES_USER: postgres
          POSTGRES_PASSWORD: postgres
        ports:
          - 5432:5432
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - run: go mod download
      - name: Run migrations
        run: |
          go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
          migrate -path migrations -database "postgres://postgres:postgres@localhost/niotebook_test?sslmode=disable" up
      - name: Run tests
        run: make test
        env:
          NIOTEBOOK_TEST_DB_URL: postgres://postgres:postgres@localhost/niotebook_test?sslmode=disable
          NIOTEBOOK_JWT_SECRET: test-secret-for-ci-only-not-production
      - name: Build
        run: make build
```

**Step 2: Run full local verification**

```bash
# Ensure test DB is ready
createdb niotebook_test 2>/dev/null || true
migrate -path migrations -database "postgres://localhost/niotebook_test?sslmode=disable" up

# Run all tests
make test

# Build both binaries
make build

# Verify server starts
source .env
./bin/niotebook-server &
sleep 2
curl -s http://localhost:8080/health | grep '"ok"'
kill %1
```

Expected: All tests pass, both binaries build, health endpoint responds.

**Step 3: Commit**

```bash
git add .github/
git commit -m "chore: add GitHub Actions CI pipeline"
```

---

## Execution Notes

### Dependency Graph

```
Task 1 (bootstrap)
├── Task 2 (models)
│   ├── Task 4 (store interfaces) ── Task 5 (store impls)
│   ├── Task 6 (validators) ── Task 7 (auth svc) ── Task 8 (post svc) ── Task 9 (user svc)
│   └── Task 18 (messages)
├── Task 3 (migrations) ── Task 5 (store impls)
├── Task 10 (auth middleware)
├── Task 11 (other middleware)
├── Task 12 (handlers) ── Task 13 (server wiring) ── Task 14 (server binary)
├── Task 15 (config) ── Task 16 (client)
├── Task 17 (components)
├── Task 19 (auth views) ── Task 20 (timeline) ── Task 21 (compose) ── Task 22 (profile+help)
│   └── Task 23 (app model + TUI binary)
└── Task 24 (CI)
```

### Parallelization Opportunities

These task groups can be worked on independently:
- **Server stores** (Tasks 4-5) and **Server services** (Tasks 6-9) after models are done
- **TUI config+client** (Tasks 15-16) and **TUI components** (Task 17) after models are done
- **Server middleware** (Tasks 10-11) independent of services

### Reference Documents

During implementation, consult these vault docs frequently:
- [[02-engineering/api/api-specification|API Specification]] — exact request/response formats
- [[02-engineering/architecture/database-schema|Database Schema]] — SQL, constraints, key queries
- [[02-engineering/architecture/bubble-tea-model-hierarchy|Bubble Tea Model Hierarchy]] — TUI model code
- [[02-engineering/architecture/server-internals|Server Internals]] — middleware, shutdown, background jobs
- [[03-design/tui-layout-and-navigation|TUI Layout]] — screen wireframes
- [[03-design/keybindings|Key Bindings]] — complete keymap
- [[03-design/post-card-component|Post Card]] — rendering rules
- [[03-design/auth-flow-ux|Auth Flow UX]] — startup and token refresh flow
