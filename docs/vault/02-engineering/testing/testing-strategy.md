---
title: "Testing Strategy"
created: 2026-02-15
updated: 2026-02-15
status: accepted
tags: [engineering, testing, quality]
---

# Testing Strategy

Comprehensive testing across three layers: unit tests, integration tests, and TUI rendering tests. All tests run with `go test ./... -race` (race detector enabled).

## Test Pyramid

```
         ┌───────────┐
         │  TUI Tests │  (rendering, interaction)
         │   ~15%     │
         ├───────────┤
         │Integration │  (API endpoints, DB queries)
         │   ~35%     │
         ├───────────┤
         │   Unit     │  (services, validation, utilities)
         │   ~50%     │
         └───────────┘
```

## 1. Unit Tests (~50%)

Test business logic in isolation, no external dependencies.

### What to Unit Test

| Package | What to Test | Approach |
|---------|-------------|----------|
| `internal/server/service/` | Auth logic (password hashing, JWT generation, refresh flow) | Mock store interfaces |
| `internal/server/service/` | Post validation (length, empty, trim) | Pure functions, no mocks |
| `internal/server/service/` | User validation (username rules, email format) | Pure functions, no mocks |
| `internal/models/` | JSON serialization/deserialization | Round-trip marshal/unmarshal |
| `internal/tui/client/` | HTTP client (request building, response parsing) | `httptest.Server` mock |
| `internal/tui/components/` | Relative time formatting | Pure function: `RelativeTime(time.Time) string` |

### Example: Post Validation Unit Tests

```go
func TestValidatePostContent(t *testing.T) {
    tests := []struct {
        name    string
        content string
        wantErr string
    }{
        {"valid short post", "Hello, Niotebook!", ""},
        {"exactly 140 chars", strings.Repeat("a", 140), ""},
        {"141 chars", strings.Repeat("a", 141), "content_too_long"},
        {"empty string", "", "content_empty"},
        {"whitespace only", "   \n\t  ", "content_empty"},
        {"emoji counts as 1", "Hello 🌍" + strings.Repeat("a", 133), ""},
        {"leading/trailing spaces trimmed", "  Hello  ", ""}, // stored as "Hello"
    }
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            err := service.ValidatePostContent(tt.content)
            if tt.wantErr == "" {
                assert.NoError(t, err)
            } else {
                assert.ErrorContains(t, err, tt.wantErr)
            }
        })
    }
}
```

### Example: Username Validation Unit Tests

```go
func TestValidateUsername(t *testing.T) {
    tests := []struct {
        name     string
        username string
        wantErr  bool
    }{
        {"valid simple", "akram", false},
        {"valid with underscore", "code_ninja", false},
        {"valid numbers", "user42", false},
        {"minimum length", "abc", false},
        {"maximum length", "abcdefghijklmno", false}, // 15 chars
        {"too short", "ab", true},
        {"too long", "abcdefghijklmnop", true}, // 16 chars
        {"leading underscore", "_akram", true},
        {"trailing underscore", "akram_", true},
        {"consecutive underscores", "code__ninja", true},
        {"special chars", "akram!", true},
        {"spaces", "ak ram", true},
        {"uppercase stored as lower", "Akram", false}, // application lowercases
        {"reserved", "admin", true},
        {"reserved", "api", true},
    }
    // ...
}
```

### Mocking Strategy

Define interfaces for store layer:

```go
type UserStore interface {
    CreateUser(ctx context.Context, user *models.User) error
    GetUserByEmail(ctx context.Context, email string) (*models.User, error)
    GetUserByID(ctx context.Context, id string) (*models.User, error)
    GetUserByUsername(ctx context.Context, username string) (*models.User, error)
    UpdateUser(ctx context.Context, id string, updates *models.UserUpdate) (*models.User, error)
}

type PostStore interface {
    CreatePost(ctx context.Context, post *models.Post) error
    GetPostByID(ctx context.Context, id string) (*models.Post, error)
    GetTimeline(ctx context.Context, cursor time.Time, limit int) ([]models.Post, error)
    GetUserPosts(ctx context.Context, userID string, cursor time.Time, limit int) ([]models.Post, error)
}
```

Unit tests use mock implementations (hand-rolled or generated with `mockgen`). No database dependency in unit tests.

## 2. Integration Tests (~35%)

Test API endpoints end-to-end against a real PostgreSQL instance.

### Test Database Setup

Use a separate database for tests:

```go
// internal/server/store/testutil_test.go
func setupTestDB(t *testing.T) *pgxpool.Pool {
    t.Helper()
    dbURL := os.Getenv("NIOTEBOOK_TEST_DB_URL")
    if dbURL == "" {
        dbURL = "postgres://localhost/niotebook_test?sslmode=disable"
    }

    pool, err := pgxpool.New(context.Background(), dbURL)
    require.NoError(t, err)

    // Run migrations
    runMigrations(t, dbURL)

    // Clean tables before each test
    t.Cleanup(func() {
        _, _ = pool.Exec(context.Background(),
            "TRUNCATE users, posts, refresh_tokens CASCADE")
        pool.Close()
    })

    return pool
}
```

### What to Integration Test

| Endpoint | Test Cases |
|----------|-----------|
| `POST /auth/register` | Successful registration, duplicate username, duplicate email, invalid username, short password, reserved username |
| `POST /auth/login` | Successful login, wrong password, nonexistent email, returns valid JWT |
| `POST /auth/refresh` | Successful refresh (rotates tokens), expired refresh token, reused token (revokes all), invalid token |
| `GET /timeline` | Returns posts newest-first, cursor pagination works, empty timeline, respects limit, includes author data |
| `POST /posts` | Successful post, empty content, too-long content, unauthorized (no token), returns created post with author |
| `GET /posts/{id}` | Existing post, nonexistent post, includes author |
| `GET /users/{id}` | Existing user, nonexistent user, `me` shorthand |
| `GET /users/{id}/posts` | Returns user's posts, cursor pagination, empty list for user with no posts |
| `PATCH /users/me` | Update display name, update bio, both fields, empty bio (clear), too-long display name |

### Example: Timeline Integration Test

```go
func TestTimeline(t *testing.T) {
    pool := setupTestDB(t)
    srv := httptest.NewServer(setupRouter(pool))
    defer srv.Close()

    // Create user and get token
    token := registerAndLogin(t, srv, "testuser", "test@example.com", "password123")

    // Create 3 posts
    createPost(t, srv, token, "First post")
    time.Sleep(10 * time.Millisecond) // ensure different timestamps
    createPost(t, srv, token, "Second post")
    time.Sleep(10 * time.Millisecond)
    createPost(t, srv, token, "Third post")

    // Fetch timeline
    resp := getTimeline(t, srv, token, "", 50)

    assert.Len(t, resp.Posts, 3)
    assert.Equal(t, "Third post", resp.Posts[0].Content)  // newest first
    assert.Equal(t, "First post", resp.Posts[2].Content)   // oldest last
    assert.Equal(t, "testuser", resp.Posts[0].Author.Username)

    // Test cursor pagination
    resp2 := getTimeline(t, srv, token, resp.NextCursor, 2)
    // ... verify correct page
}
```

### Rate Limiting Tests

```go
func TestRateLimiting(t *testing.T) {
    // Send 11 requests to /auth/login rapidly (limit is 10/min)
    // Assert 11th returns 429 with Retry-After header
}
```

### Middleware Tests

Test JWT middleware independently:
- Valid token → passes through, sets user context
- Expired token → 401 with `token_expired` code
- Missing token → 401 with `unauthorized` code
- Malformed token → 401 with `unauthorized` code

## 3. TUI Rendering Tests (~15%)

Test Bubble Tea models using the `teatest` package from Charm.

### What to TUI Test

| Component | Test Cases |
|-----------|-----------|
| Timeline View | Renders posts correctly, scroll moves selection, refresh triggers fetch, empty state |
| Post Card | Renders username, content, relative time correctly, selected state shows marker |
| Compose Modal | Opens on 'n', char counter updates, rejects over-140, Ctrl+Enter sends, Esc cancels |
| Login View | Tab switches to register, Enter submits, shows error on failure |
| Status Bar | Shows error in red, success in green, auto-clears, shows key hints |

### Example: Compose Modal Test

```go
func TestComposeModal(t *testing.T) {
    m := NewComposeModel()

    // Open modal
    m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
    assert.True(t, m.showCompose)

    // Type content
    for _, r := range "Hello Niotebook!" {
        m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
    }

    // Verify character count
    view := m.View()
    assert.Contains(t, view, "16/140")

    // Cancel with Esc
    m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEscape})
    assert.False(t, m.showCompose)
}
```

### Using teatest

```go
func TestTimelineView(t *testing.T) {
    // Create model with mock data
    posts := []models.Post{
        {Author: &models.User{Username: "akram"}, Content: "Hello!", CreatedAt: time.Now()},
    }
    m := NewTimelineModel(posts)

    // Render and assert
    tm := teatest.NewModel(t, m)
    tm.Send(tea.WindowSizeMsg{Width: 80, Height: 24})

    out := tm.FinalOutput(t)
    assert.Contains(t, out, "@akram")
    assert.Contains(t, out, "Hello!")
}
```

## Test Infrastructure

### CI Pipeline (GitHub Actions)

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
      - run: make migrate-up
        env:
          NIOTEBOOK_DB_URL: postgres://postgres:postgres@localhost/niotebook_test?sslmode=disable
      - run: make test
        env:
          NIOTEBOOK_TEST_DB_URL: postgres://postgres:postgres@localhost/niotebook_test?sslmode=disable
          NIOTEBOOK_JWT_SECRET: test-secret-for-ci-only
      - run: make lint
```

### Test Helpers

A shared `testutil` package provides:

```go
// Create a registered user and return auth token
func RegisterAndLogin(t *testing.T, srv *httptest.Server, username, email, password string) string

// Create a post and return the created post
func CreatePost(t *testing.T, srv *httptest.Server, token, content string) models.Post

// Fetch timeline with optional cursor
func GetTimeline(t *testing.T, srv *httptest.Server, token, cursor string, limit int) TimelineResponse

// Make an authenticated HTTP request
func AuthRequest(t *testing.T, method, url, token string, body interface{}) *http.Response
```

## Coverage Targets

| Package | Target | Rationale |
|---------|--------|-----------|
| `internal/server/service/` | 90%+ | Core business logic, must be thoroughly tested |
| `internal/server/handler/` | 80%+ | Covered by integration tests |
| `internal/server/store/` | 70%+ | Covered by integration tests against real DB |
| `internal/tui/views/` | 60%+ | TUI rendering, harder to test exhaustively |
| `internal/tui/components/` | 80%+ | Reusable components with clear inputs/outputs |
| `internal/tui/client/` | 80%+ | HTTP client, testable with httptest |
| `internal/models/` | 90%+ | Simple types, JSON round-trips |

Overall project target: **80%+ coverage**.
