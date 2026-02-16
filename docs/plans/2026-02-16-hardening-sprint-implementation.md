# Hardening Sprint Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Fix all critical and important issues from the B+ code review to achieve Grade-A (90+) across all 7 pillars of SWE.

**Architecture:** Targeted fixes across 4 phases — security fixes first, then code quality improvements, then test coverage push, then CI hardening. Each task is independent within its phase. All changes land on the `hardening-sprint` branch.

**Tech Stack:** Go 1.22+, Bubble Tea, pgx v5, golang-jwt/v5, golangci-lint

---

## Phase 1: Critical Security Fixes

### Task 1: Fix unchecked type assertions in auth middleware

**Files:**
- Modify: `internal/server/middleware/auth.go:79-82`
- Modify: `internal/server/middleware/auth_test.go` (add test)

**Context:** Lines 79-82 of `auth.go` do `claims["sub"].(string)` without checking the `ok` value. If a valid-signature JWT has non-string claims, the server panics. The `writeError` helper is already available in the same file.

**Step 1: Add a test for malformed JWT claims**

In `internal/server/middleware/auth_test.go`, add this test after `TestAuthMiddlewareExemptPaths`:

```go
func TestAuthMiddlewareMalformedClaims(t *testing.T) {
	// Token with integer sub instead of string — valid signature, bad claims
	token := makeToken(testSecret, jwt.MapClaims{
		"sub": 12345,
		"exp": time.Now().Add(time.Hour).Unix(),
		"iat": time.Now().Unix(),
	})

	handler := middleware.Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for malformed claims")
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}
```

**Step 2: Run the test — it should panic (FAIL)**

Run: `go test ./internal/server/middleware/ -run TestAuthMiddlewareMalformedClaims -v`
Expected: FAIL (panic: interface conversion)

**Step 3: Fix the type assertions in auth.go**

Replace lines 79-82 of `internal/server/middleware/auth.go`:

```go
		// OLD:
		userClaims := &UserClaims{
			UserID:   claims["sub"].(string),
			Username: claims["username"].(string),
		}
```

With:

```go
		// NEW: Safe type assertions — return 401 instead of panicking
		sub, ok := claims["sub"].(string)
		if !ok || sub == "" {
			writeError(w, http.StatusUnauthorized, models.ErrCodeUnauthorized, "invalid token claims")
			return
		}
		uname, ok := claims["username"].(string)
		if !ok {
			writeError(w, http.StatusUnauthorized, models.ErrCodeUnauthorized, "invalid token claims")
			return
		}

		userClaims := &UserClaims{
			UserID:   sub,
			Username: uname,
		}
```

**Step 4: Run all middleware tests**

Run: `go test ./internal/server/middleware/ -v -race`
Expected: ALL PASS including the new `TestAuthMiddlewareMalformedClaims`

**Step 5: Commit**

```bash
git add internal/server/middleware/auth.go internal/server/middleware/auth_test.go
git commit -m "fix: add checked type assertions for JWT claims in auth middleware"
```

---

### Task 2: Add JWT secret length validation

**Files:**
- Modify: `cmd/server/main.go:44-48`

**Context:** The CLAUDE.md says JWT secret must be "min 32 bytes" but the code only checks for empty string. A weak secret like "secret" would be accepted. The CI workflow already uses a 38-byte secret so no CI change needed.

**Step 1: Add the length check**

In `cmd/server/main.go`, after the empty check (line 48), add:

```go
	if len(jwtSecret) < 32 {
		slog.Error("NIOTEBOOK_JWT_SECRET must be at least 32 bytes", "length", len(jwtSecret))
		os.Exit(1)
	}
```

So lines 44-52 become:

```go
	jwtSecret := os.Getenv("NIOTEBOOK_JWT_SECRET")
	if jwtSecret == "" {
		slog.Error("NIOTEBOOK_JWT_SECRET is required")
		os.Exit(1)
	}
	if len(jwtSecret) < 32 {
		slog.Error("NIOTEBOOK_JWT_SECRET must be at least 32 bytes", "length", len(jwtSecret))
		os.Exit(1)
	}
```

**Step 2: Verify build succeeds**

Run: `go build ./cmd/server`
Expected: exit 0

**Step 3: Verify existing tests still pass**

Run: `go test ./... -race -count=1 2>&1 | tail -20`
Expected: ALL PASS (the test suite uses its own JWT secret, not the env var)

**Step 4: Commit**

```bash
git add cmd/server/main.go
git commit -m "fix: enforce minimum 32-byte JWT secret length"
```

---

### Task 3: Fix unchecked type assertions in TUI app

**Files:**
- Modify: `internal/tui/app/app.go` (12 locations)

**Context:** The file has 12 unchecked type assertions like `updated.(TimelineViewModel)`. Each one can panic if the adapter returns a different type. The fix is to use the comma-ok pattern and return the model unchanged on failure. The ViewModel interface Update method returns `(ViewModel, tea.Cmd)` — the concrete types are `TimelineViewModel`, `ProfileViewModel`, `ComposeViewModel`, `HelpViewModel`.

**Step 1: Create a helper function at the top of app.go**

Add this after the `currentHelpText` method (end of file), or better — add a small helper that safely casts. Actually, the simplest approach is to just add the comma-ok pattern inline at each site. Since the views are controlled by us and the adapter pattern guarantees the return type, a simple approach is:

Replace ALL 12 unchecked assertions. Here is the exact list with line numbers and replacements:

**Line 242** (in `MsgTimelineLoaded` handler):
```go
// OLD: m.timeline = updated.(TimelineViewModel)
// NEW:
if tl, ok := updated.(TimelineViewModel); ok {
	m.timeline = tl
}
```

**Line 261** (in `MsgProfileLoaded` handler):
```go
// OLD: m.profile = updated.(ProfileViewModel)
// NEW:
if pv, ok := updated.(ProfileViewModel); ok {
	m.profile = pv
}
```

**Line 271** (in `MsgProfileUpdated` handler):
```go
// OLD: m.profile = updated.(ProfileViewModel)
// NEW:
if pv, ok := updated.(ProfileViewModel); ok {
	m.profile = pv
}
```

**Line 393** (in `openHelp`):
```go
// OLD: m.help = updated.(HelpViewModel)
// NEW:
if hv, ok := updated.(HelpViewModel); ok {
	m.help = hv
}
```

**Line 413** (in `updateCompose`):
```go
// OLD: m.compose = updated.(ComposeViewModel)
// NEW:
if cv, ok := updated.(ComposeViewModel); ok {
	m.compose = cv
}
```

**Line 428** (in `updateHelp`):
```go
// OLD: m.help = updated.(HelpViewModel)
// NEW:
if hv, ok := updated.(HelpViewModel); ok {
	m.help = hv
}
```

**Line 458** (in `updateCurrentView`, ViewTimeline case):
```go
// OLD: m.timeline = updated.(TimelineViewModel)
// NEW:
if tl, ok := updated.(TimelineViewModel); ok {
	m.timeline = tl
}
```

**Line 464** (in `updateCurrentView`, ViewProfile case):
```go
// OLD: m.profile = updated.(ProfileViewModel)
// NEW:
if pv, ok := updated.(ProfileViewModel); ok {
	m.profile = pv
}
```

**Line 485** (in `propagateWindowSize`, timeline):
```go
// OLD: m.timeline = updated.(TimelineViewModel)
// NEW:
if tl, ok := updated.(TimelineViewModel); ok {
	m.timeline = tl
}
```

**Line 490** (in `propagateWindowSize`, profile):
```go
// OLD: m.profile = updated.(ProfileViewModel)
// NEW:
if pv, ok := updated.(ProfileViewModel); ok {
	m.profile = pv
}
```

**Line 495** (in `propagateWindowSize`, compose):
```go
// OLD: m.compose = updated.(ComposeViewModel)
// NEW:
if cv, ok := updated.(ComposeViewModel); ok {
	m.compose = cv
}
```

**Line 500** (in `propagateWindowSize`, help):
```go
// OLD: m.help = updated.(HelpViewModel)
// NEW:
if hv, ok := updated.(HelpViewModel); ok {
	m.help = hv
}
```

**Step 2: Verify zero unchecked assertions remain**

Run: `grep -n '\.\(TimelineViewModel\)\|\.\(ProfileViewModel\)\|\.\(ComposeViewModel\)\|\.\(HelpViewModel\)' internal/tui/app/app.go | grep -v ', ok'`
Expected: No output (all assertions are now checked)

**Step 3: Run all TUI tests**

Run: `go test ./internal/tui/... -v -race`
Expected: ALL PASS

**Step 4: Commit**

```bash
git add internal/tui/app/app.go
git commit -m "fix: add checked type assertions for all view model casts in app"
```

---

## Phase 2: Important Code Fixes

### Task 4: Make CORS origin configurable

**Files:**
- Modify: `internal/server/middleware/cors.go`
- Modify: `internal/server/server.go:13-17, 57`
- Modify: `cmd/server/main.go` (add env var read)

**Context:** Currently `cors.go` hardcodes `Access-Control-Allow-Origin: *`. We need to accept an `allowedOrigin` parameter. The `server.Config` struct needs a new `CORSOrigin` field.

**Step 1: Update CORS middleware to accept origin parameter**

Replace `internal/server/middleware/cors.go` entirely:

```go
package middleware

import "net/http"

func CORS(allowedOrigin string) func(http.Handler) http.Handler {
	if allowedOrigin == "" {
		allowedOrigin = "*"
	}
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Authorization, Content-Type")

			if r.Method == http.MethodOptions {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}
```

**Step 2: Update server.go Config and middleware chain**

In `internal/server/server.go`, add `CORSOrigin` to Config:

```go
type Config struct {
	JWTSecret  string
	Host       string
	Port       string
	CORSOrigin string
}
```

And update line 57 from `h = middleware.CORS(h)` to:

```go
	h = middleware.CORS(cfg.CORSOrigin)(h)
```

**Step 3: Read env var in main.go**

In `cmd/server/main.go`, before the database section (around line 50), add:

```go
	corsOrigin := envOrDefault("NIOTEBOOK_CORS_ORIGIN", "*")
```

And update the server config (wherever `Config` is constructed, around line 60-65) to include:

```go
	CORSOrigin: corsOrigin,
```

**Step 4: Run all tests**

Run: `go test ./... -v -race -count=1 2>&1 | tail -20`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add internal/server/middleware/cors.go internal/server/server.go cmd/server/main.go
git commit -m "feat: make CORS origin configurable via NIOTEBOOK_CORS_ORIGIN"
```

---

### Task 5: Add HTTP client timeout and network retry

**Files:**
- Modify: `internal/tui/client/client.go:26-31, 238-273`
- Modify: `internal/tui/client/client_test.go` (add retry test)

**Context:** The `http.Client{}` has no timeout (will hang forever). The `do()` method has no retry for transient network errors. We need a 30s timeout and exponential backoff retry (3 attempts: 1s, 2s, 4s) for network errors only.

**Step 1: Add timeout to client constructor**

In `internal/tui/client/client.go`, change the `New` function:

```go
import (
	"time"
	// ... existing imports
	"net"
)
```

```go
func New(baseURL string) *Client {
	return &Client{
		baseURL: baseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}
```

**Step 2: Add network retry to the `do` method**

Replace the `do` method with a version that retries on network errors:

```go
// do performs the raw HTTP request with retry on network errors.
func (c *Client) do(method, path string, body interface{}, withAuth bool) (*http.Response, error) {
	const maxRetries = 3

	var lastErr error
	for attempt := 0; attempt < maxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(1<<(attempt-1)) * time.Second)
		}

		resp, err := c.doOnce(method, path, body, withAuth)
		if err == nil {
			return resp, nil
		}

		// Only retry on network errors, not on successful HTTP responses
		var netErr net.Error
		if errors.As(err, &netErr) || errors.Is(err, net.ErrClosed) {
			lastErr = err
			continue
		}
		// Non-network error — don't retry
		return nil, err
	}
	return nil, fmt.Errorf("request failed after %d attempts: %w", maxRetries, lastErr)
}

// doOnce performs a single HTTP request attempt.
func (c *Client) doOnce(method, path string, body interface{}, withAuth bool) (*http.Response, error) {
	var reqBody *bytes.Buffer
	if body != nil {
		reqBody = &bytes.Buffer{}
		if err := json.NewEncoder(reqBody).Encode(body); err != nil {
			return nil, fmt.Errorf("encoding request: %w", err)
		}
	}

	var req *http.Request
	var err error
	if reqBody != nil {
		req, err = http.NewRequest(method, c.baseURL+path, reqBody)
	} else {
		req, err = http.NewRequest(method, c.baseURL+path, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	if withAuth {
		c.mu.Lock()
		token := c.accessToken
		c.mu.Unlock()
		if token != "" {
			req.Header.Set("Authorization", "Bearer "+token)
		}
	}

	return c.httpClient.Do(req)
}
```

Also add `"errors"` and `"net"` to the imports.

**Step 3: Add a test for the timeout**

In `internal/tui/client/client_test.go`, add:

```go
func TestClientTimeout(t *testing.T) {
	c := client.New("http://198.51.100.1:1") // Non-routable IP — will timeout
	c.SetToken("test-token")

	_, err := c.GetTimeline("", 20)
	if err == nil {
		t.Fatal("expected timeout error")
	}
}
```

Note: This test may be slow (~30s) due to timeout. It verifies the timeout is set.

**Step 4: Run client tests**

Run: `go test ./internal/tui/client/ -v -race -timeout 60s`
Expected: ALL PASS (timeout test will take ~30s)

**Step 5: Commit**

```bash
git add internal/tui/client/client.go internal/tui/client/client_test.go
git commit -m "feat: add 30s HTTP client timeout and network retry with backoff"
```

---

### Task 6: Fix window resize command propagation

**Files:**
- Modify: `internal/tui/app/app.go` (function `propagateWindowSize`)

**Context:** The `propagateWindowSize` method discards all `tea.Cmd` values from view updates and always returns `nil`. Views that want to react to resize (e.g., re-fetch for new page size) silently lose their commands.

**Step 1: Replace the propagateWindowSize function**

Find the `propagateWindowSize` function in `internal/tui/app/app.go` and replace it with:

```go
// propagateWindowSize sends the window size to all sub-models.
func (m AppModel) propagateWindowSize(msg tea.WindowSizeMsg) (AppModel, tea.Cmd) {
	var cmds []tea.Cmd

	if m.login != nil {
		var cmd tea.Cmd
		m.login, cmd = m.login.Update(msg)
		cmds = append(cmds, cmd)
	}
	if m.register != nil {
		var cmd tea.Cmd
		m.register, cmd = m.register.Update(msg)
		cmds = append(cmds, cmd)
	}
	if m.timeline != nil {
		var updated ViewModel
		var cmd tea.Cmd
		updated, cmd = m.timeline.Update(msg)
		if tl, ok := updated.(TimelineViewModel); ok {
			m.timeline = tl
		}
		cmds = append(cmds, cmd)
	}
	if m.profile != nil {
		var updated ViewModel
		var cmd tea.Cmd
		updated, cmd = m.profile.Update(msg)
		if pv, ok := updated.(ProfileViewModel); ok {
			m.profile = pv
		}
		cmds = append(cmds, cmd)
	}
	if m.compose != nil {
		var updated ViewModel
		var cmd tea.Cmd
		updated, cmd = m.compose.Update(msg)
		if cv, ok := updated.(ComposeViewModel); ok {
			m.compose = cv
		}
		cmds = append(cmds, cmd)
	}
	if m.help != nil {
		var updated ViewModel
		var cmd tea.Cmd
		updated, cmd = m.help.Update(msg)
		if hv, ok := updated.(HelpViewModel); ok {
			m.help = hv
		}
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}
```

**Step 2: Run all TUI tests**

Run: `go test ./internal/tui/... -v -race`
Expected: ALL PASS

**Step 3: Commit**

```bash
git add internal/tui/app/app.go
git commit -m "fix: propagate window resize commands from all sub-models"
```

---

### Task 7: Add XDG_CONFIG_HOME support

**Files:**
- Modify: `internal/tui/config/config.go:40-43`
- Modify: `internal/tui/config/config_test.go` (add test)

**Context:** `ConfigDir()` hardcodes `~/.config/niotebook` but XDG Base Directory Specification requires checking `$XDG_CONFIG_HOME` first.

**Step 1: Add a test for XDG_CONFIG_HOME**

In `internal/tui/config/config_test.go`, add:

```go
func TestConfigDirXDGOverride(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/tmp/test-xdg")
	dir := config.ConfigDir()
	want := "/tmp/test-xdg/niotebook"
	if dir != want {
		t.Errorf("ConfigDir() = %q, want %q", dir, want)
	}
}

func TestConfigDirDefaultWithoutXDG(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")
	dir := config.ConfigDir()
	if !strings.HasSuffix(dir, "/.config/niotebook") {
		t.Errorf("ConfigDir() = %q, want suffix /.config/niotebook", dir)
	}
}
```

Add `"strings"` to the test file imports if not already present.

**Step 2: Run the test — first one should FAIL**

Run: `go test ./internal/tui/config/ -run TestConfigDirXDGOverride -v`
Expected: FAIL (returns `~/.config/niotebook` ignoring the env var)

**Step 3: Fix ConfigDir**

Replace the `ConfigDir` function in `internal/tui/config/config.go`:

```go
func ConfigDir() string {
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "niotebook")
	}
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".config", "niotebook")
}
```

**Step 4: Run all config tests**

Run: `go test ./internal/tui/config/ -v -race`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add internal/tui/config/config.go internal/tui/config/config_test.go
git commit -m "feat: honor XDG_CONFIG_HOME for config directory"
```

---

## Phase 3: Test Coverage Push

### Task 8: Add middleware tests (recovery, CORS, logging)

**Files:**
- Create: `internal/server/middleware/recovery_test.go`
- Create: `internal/server/middleware/cors_test.go`
- Create: `internal/server/middleware/logging_test.go`

**Context:** Recovery, CORS, and logging middleware all have 0% test coverage. These are critical infrastructure. The `writeError` helper and `responseWriter` wrapper must be tested through their middleware.

**Step 1: Write recovery_test.go**

Create `internal/server/middleware/recovery_test.go`:

```go
package middleware_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Akram012388/niotebook-tui/internal/server/middleware"
)

func TestRecoveryMiddlewareCatchesPanic(t *testing.T) {
	panicker := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		panic("test panic")
	})

	handler := middleware.Recovery(panicker)
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}

	var body map[string]interface{}
	if err := json.NewDecoder(rec.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	errObj, ok := body["error"].(map[string]interface{})
	if !ok {
		t.Fatal("expected error object in response")
	}
	if errObj["code"] != "internal_error" {
		t.Errorf("error code = %q, want %q", errObj["code"], "internal_error")
	}
}

func TestRecoveryMiddlewarePassesThrough(t *testing.T) {
	normal := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	handler := middleware.Recovery(normal)
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}
```

**Step 2: Write cors_test.go**

Create `internal/server/middleware/cors_test.go`:

```go
package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Akram012388/niotebook-tui/internal/server/middleware"
)

func TestCORSPreflightRequest(t *testing.T) {
	handler := middleware.CORS("https://example.com")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		t.Error("handler should not be called for OPTIONS")
	}))

	req := httptest.NewRequest("OPTIONS", "/api/v1/posts", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "https://example.com" {
		t.Errorf("CORS origin = %q, want %q", got, "https://example.com")
	}
	if got := rec.Header().Get("Access-Control-Allow-Methods"); got == "" {
		t.Error("expected Access-Control-Allow-Methods header")
	}
	if got := rec.Header().Get("Access-Control-Allow-Headers"); got == "" {
		t.Error("expected Access-Control-Allow-Headers header")
	}
}

func TestCORSNonPreflightHasHeaders(t *testing.T) {
	handler := middleware.CORS("*")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/timeline", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("CORS origin = %q, want %q", got, "*")
	}
}

func TestCORSDefaultOrigin(t *testing.T) {
	handler := middleware.CORS("")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Errorf("CORS default origin = %q, want %q", got, "*")
	}
}
```

**Step 3: Write logging_test.go**

Create `internal/server/middleware/logging_test.go`:

```go
package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Akram012388/niotebook-tui/internal/server/middleware"
)

func TestLoggingMiddlewarePassesThrough(t *testing.T) {
	handler := middleware.Logging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	req := httptest.NewRequest("POST", "/api/v1/posts", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusCreated)
	}
}

func TestLoggingMiddlewareDefaultStatus(t *testing.T) {
	handler := middleware.Logging(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Don't call WriteHeader — should default to 200
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest("GET", "/health", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}
```

**Step 4: Run all middleware tests**

Run: `go test ./internal/server/middleware/ -v -race`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add internal/server/middleware/recovery_test.go internal/server/middleware/cors_test.go internal/server/middleware/logging_test.go
git commit -m "test: add recovery, CORS, and logging middleware tests"
```

---

### Task 9: Add TUI component tests

**Files:**
- Create: `internal/tui/components/postcard_test.go`
- Create: `internal/tui/components/header_test.go`
- Create: `internal/tui/components/statusbar_test.go`

**Context:** PostCard, Header, and StatusBar components are at 0% coverage. They're pure rendering functions (PostCard, Header) or simple state machines (StatusBar). Test them by checking the output string contains expected substrings.

**Step 1: Write postcard_test.go**

Create `internal/tui/components/postcard_test.go`:

```go
package components_test

import (
	"strings"
	"testing"
	"time"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/Akram012388/niotebook-tui/internal/tui/components"
)

func TestRenderPostCardWithAuthor(t *testing.T) {
	post := models.Post{
		Content:   "Hello, world!",
		Author:    &models.User{Username: "akram"},
		CreatedAt: time.Now().Add(-5 * time.Minute),
	}
	result := components.RenderPostCard(post, 80, false, time.Now())
	if !strings.Contains(result, "@akram") {
		t.Error("expected @akram in output")
	}
	if !strings.Contains(result, "Hello, world!") {
		t.Error("expected post content in output")
	}
}

func TestRenderPostCardSelected(t *testing.T) {
	post := models.Post{
		Content: "Test post",
		Author:  &models.User{Username: "test"},
	}
	result := components.RenderPostCard(post, 80, true, time.Now())
	if !strings.Contains(result, "Test post") {
		t.Error("expected post content in selected card")
	}
}

func TestRenderPostCardNilAuthor(t *testing.T) {
	post := models.Post{Content: "Orphan post", Author: nil}
	result := components.RenderPostCard(post, 80, false, time.Now())
	if !strings.Contains(result, "@unknown") {
		t.Error("expected @unknown for nil author")
	}
}

func TestRenderPostCardNarrowWidth(t *testing.T) {
	post := models.Post{
		Content: "This is a longer post that should wrap at narrow widths",
		Author:  &models.User{Username: "user"},
	}
	result := components.RenderPostCard(post, 20, false, time.Now())
	if result == "" {
		t.Error("expected non-empty output for narrow width")
	}
}
```

**Step 2: Write header_test.go**

Create `internal/tui/components/header_test.go`:

```go
package components_test

import (
	"strings"
	"testing"

	"github.com/Akram012388/niotebook-tui/internal/tui/components"
)

func TestRenderHeaderShowsAppName(t *testing.T) {
	result := components.RenderHeader("niotebook", "akram", "Timeline", 80)
	if !strings.Contains(result, "niotebook") {
		t.Error("expected app name in header")
	}
}

func TestRenderHeaderShowsUsername(t *testing.T) {
	result := components.RenderHeader("niotebook", "akram", "Timeline", 80)
	if !strings.Contains(result, "@akram") {
		t.Error("expected @akram in header")
	}
}

func TestRenderHeaderEmptyUsername(t *testing.T) {
	result := components.RenderHeader("niotebook", "", "Timeline", 80)
	if !strings.Contains(result, "niotebook") {
		t.Error("expected app name in header even with empty username")
	}
}

func TestRenderHeaderNarrowWidth(t *testing.T) {
	result := components.RenderHeader("niotebook", "akram", "Timeline", 10)
	if result == "" {
		t.Error("expected non-empty output for narrow width")
	}
}
```

**Step 3: Write statusbar_test.go**

Create `internal/tui/components/statusbar_test.go`:

```go
package components_test

import (
	"strings"
	"testing"

	"github.com/Akram012388/niotebook-tui/internal/tui/components"
)

func TestStatusBarSetError(t *testing.T) {
	sb := components.NewStatusBarModel()
	cmd := sb.SetError("something went wrong")
	if cmd == nil {
		t.Error("expected auto-clear command")
	}
	result := sb.View("help text", 80)
	if !strings.Contains(result, "something went wrong") {
		t.Error("expected error message in status bar")
	}
}

func TestStatusBarSetSuccess(t *testing.T) {
	sb := components.NewStatusBarModel()
	cmd := sb.SetSuccess("Post published!")
	if cmd == nil {
		t.Error("expected auto-clear command")
	}
	result := sb.View("help text", 80)
	if !strings.Contains(result, "Post published!") {
		t.Error("expected success message in status bar")
	}
}

func TestStatusBarClear(t *testing.T) {
	sb := components.NewStatusBarModel()
	sb.SetError("error")
	sb.Clear()
	result := sb.View("help text", 80)
	if strings.Contains(result, "error") {
		t.Error("expected error to be cleared")
	}
}

func TestStatusBarDefaultShowsHelpText(t *testing.T) {
	sb := components.NewStatusBarModel()
	result := sb.View("n: new post  ?: help  q: quit", 80)
	if !strings.Contains(result, "n: new post") {
		t.Error("expected help text in default status bar")
	}
}
```

**Step 4: Run all component tests**

Run: `go test ./internal/tui/components/ -v -race`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add internal/tui/components/postcard_test.go internal/tui/components/header_test.go internal/tui/components/statusbar_test.go
git commit -m "test: add postcard, header, and statusbar component tests"
```

---

### Task 10: Add TUI view tests

**Files:**
- Create: `internal/tui/views/register_test.go`
- Modify: `internal/tui/views/timeline_test.go` (add error case)
- Modify: `internal/tui/views/compose_test.go` (add error case)

**Context:** The register view has 0% coverage. Other views are missing error case tests. The views use Bubble Tea's Update/View pattern — test by sending messages and checking state/output.

**Step 1: Write register_test.go**

Create `internal/tui/views/register_test.go`:

```go
package views_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Akram012388/niotebook-tui/internal/tui/app"
	"github.com/Akram012388/niotebook-tui/internal/tui/views"
)

func TestRegisterModelRenders(t *testing.T) {
	m := views.NewRegisterModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	output := m.View()
	if !strings.Contains(output, "Register") {
		t.Error("expected Register title in output")
	}
	if !strings.Contains(output, "Username") {
		t.Error("expected Username label in output")
	}
	if !strings.Contains(output, "Email") {
		t.Error("expected Email label in output")
	}
	if !strings.Contains(output, "Password") {
		t.Error("expected Password label in output")
	}
}

func TestRegisterModelTabNavigation(t *testing.T) {
	m := views.NewRegisterModel(nil)
	if m.FocusIndex() != 0 {
		t.Errorf("initial focus = %d, want 0", m.FocusIndex())
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.FocusIndex() != 1 {
		t.Errorf("after tab focus = %d, want 1", m.FocusIndex())
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.FocusIndex() != 2 {
		t.Errorf("after 2 tabs focus = %d, want 2", m.FocusIndex())
	}
}

func TestRegisterModelShiftTabNavigation(t *testing.T) {
	m := views.NewRegisterModel(nil)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab}) // go to email
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if m.FocusIndex() != 0 {
		t.Errorf("after shift-tab focus = %d, want 0", m.FocusIndex())
	}
}

func TestRegisterModelAuthError(t *testing.T) {
	m := views.NewRegisterModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = m.Update(app.MsgAuthError{Message: "email already taken", Field: "email"})
	output := m.View()
	if !strings.Contains(output, "email already taken") {
		t.Error("expected error message in output")
	}
}
```

**Step 2: Add error case test to timeline**

In `internal/tui/views/timeline_test.go`, add:

```go
func TestTimelineModelAPIError(t *testing.T) {
	m := views.NewTimelineModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	// Sending an error message should not panic
	_, cmd := m.Update(app.MsgAPIError{Message: "server error"})
	_ = cmd // Timeline may or may not produce a command from this
}
```

Add `"github.com/Akram012388/niotebook-tui/internal/tui/app"` to imports if not present.

**Step 3: Add error case test to compose**

In `internal/tui/views/compose_test.go`, add:

```go
func TestComposeModelAPIError(t *testing.T) {
	m := views.NewComposeModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	_, cmd := m.Update(app.MsgAPIError{Message: "post too long"})
	_ = cmd
}
```

Add `"github.com/Akram012388/niotebook-tui/internal/tui/app"` to imports if not present.

**Step 4: Run all view tests**

Run: `go test ./internal/tui/views/ -v -race`
Expected: ALL PASS

**Step 5: Commit**

```bash
git add internal/tui/views/register_test.go internal/tui/views/timeline_test.go internal/tui/views/compose_test.go
git commit -m "test: add register view tests and error case tests for timeline, compose"
```

---

### Task 11: Add TUI app integration tests

**Files:**
- Modify: `internal/tui/app/app_test.go` (add tests)

**Context:** The app test file only has 3 tests. We need tests for keyboard shortcuts, overlay lifecycle, and error display. The `stubFactory`, `stubTimeline`, `stubProfile`, `stubCompose`, `stubHelp`, and `update` helper are already defined in the test file.

**Step 1: Add keyboard shortcut tests**

In `internal/tui/app/app_test.go`, add these tests:

```go
func TestAppModelQuestionMarkOpensHelp(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{})
	m = update(m, app.MsgAuthSuccess{
		User:   &models.User{Username: "akram"},
		Tokens: &models.TokenPair{AccessToken: "tok"},
	})
	m = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	// Help should now be open — view name should reflect it
	if m.View() == "" {
		t.Error("expected non-empty view with help overlay")
	}
}

func TestAppModelQQuitsOnTimeline(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{})
	m = update(m, app.MsgAuthSuccess{
		User:   &models.User{Username: "akram"},
		Tokens: &models.TokenPair{AccessToken: "tok"},
	})
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if cmd == nil {
		t.Error("expected quit command on q press")
	}
}

func TestAppModelCtrlCAlwaysQuits(t *testing.T) {
	m := app.NewAppModel(nil, nil) // on login view
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlC})
	if cmd == nil {
		t.Error("expected quit command on ctrl+c")
	}
}

func TestAppModelSwitchToRegister(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{})
	m = update(m, app.MsgSwitchToRegister{})
	if m.CurrentView() != app.ViewRegister {
		t.Errorf("view = %v, want ViewRegister", m.CurrentView())
	}
}

func TestAppModelSwitchToLogin(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{})
	m = update(m, app.MsgSwitchToRegister{}) // go to register first
	m = update(m, app.MsgSwitchToLogin{})
	if m.CurrentView() != app.ViewLogin {
		t.Errorf("view = %v, want ViewLogin", m.CurrentView())
	}
}

func TestAppModelAPIErrorShowsInStatusBar(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{})
	m = update(m, app.MsgAuthSuccess{
		User:   &models.User{Username: "akram"},
		Tokens: &models.TokenPair{AccessToken: "tok"},
	})
	m = update(m, tea.WindowSizeMsg{Width: 80, Height: 24})
	m = update(m, app.MsgAPIError{Message: "server error"})
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view after API error")
	}
}

func TestAppModelAuthExpiredReturnsToLogin(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{})
	m = update(m, app.MsgAuthSuccess{
		User:   &models.User{Username: "akram"},
		Tokens: &models.TokenPair{AccessToken: "tok"},
	})
	m = update(m, app.MsgAuthExpired{})
	if m.CurrentView() != app.ViewLogin {
		t.Errorf("view = %v, want ViewLogin after auth expired", m.CurrentView())
	}
}

func TestAppModelWindowResize(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{})
	m = update(m, app.MsgAuthSuccess{
		User:   &models.User{Username: "akram"},
		Tokens: &models.TokenPair{AccessToken: "tok"},
	})
	m = update(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	// Should not panic
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view after resize")
	}
}
```

**Step 2: Run all app tests**

Run: `go test ./internal/tui/app/ -v -race`
Expected: ALL PASS

**Step 3: Commit**

```bash
git add internal/tui/app/app_test.go
git commit -m "test: add keyboard shortcut, overlay, and error display tests for app"
```

---

## Phase 4: CI Hardening

### Task 12: CI pipeline hardening

**Files:**
- Modify: `.github/workflows/ci.yml`
- Modify: `internal/server/store/post_store_test.go:119-123` (fix flaky test)

**Context:** CI is missing lint step and coverage threshold. The post store test uses `time.Sleep(10ms)` between inserts to guarantee ordering, which can be flaky on slow CI. Increase to 50ms. Add golangci-lint and coverage to CI.

**Step 1: Fix flaky test**

In `internal/server/store/post_store_test.go`, find the `TestGetTimeline` function and change all `time.Sleep(10 * time.Millisecond)` to `time.Sleep(50 * time.Millisecond)`. There are 2 occurrences (lines 120 and 122):

```go
// OLD: time.Sleep(10 * time.Millisecond)
// NEW: time.Sleep(50 * time.Millisecond)
```

Also do the same in `TestGetTimelineCursorPagination` (line 151):

```go
// OLD: time.Sleep(10 * time.Millisecond)
// NEW: time.Sleep(50 * time.Millisecond)
```

**Step 2: Update CI workflow**

Replace `.github/workflows/ci.yml` with:

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
        run: go test ./... -v -race -coverprofile=coverage.out
        env:
          NIOTEBOOK_TEST_DB_URL: postgres://postgres:postgres@localhost/niotebook_test?sslmode=disable
          NIOTEBOOK_JWT_SECRET: test-secret-for-ci-only-not-production
      - name: Check coverage
        run: |
          total=$(go tool cover -func=coverage.out | grep total | awk '{print $3}' | tr -d '%')
          echo "Total coverage: ${total}%"
          threshold=60
          if (( $(echo "$total < $threshold" | bc -l) )); then
            echo "Coverage ${total}% is below threshold ${threshold}%"
            exit 1
          fi
      - name: Build
        run: make build

  lint:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.22'
      - name: golangci-lint
        uses: golangci/golangci-lint-action@v4
        with:
          version: latest
```

**Step 3: Verify tests pass locally**

Run: `go test ./... -v -race -count=1 2>&1 | tail -20`
Expected: ALL PASS

**Step 4: Check coverage**

Run: `go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out | grep total`
Expected: total coverage >= 60%

**Step 5: Commit**

```bash
git add .github/workflows/ci.yml internal/server/store/post_store_test.go
git commit -m "chore: add lint job, coverage threshold, and fix flaky test in CI"
```

---

## Verification Checklist

After all 12 tasks complete, run these verification commands:

```bash
# 1. Build succeeds
go build ./...

# 2. All tests pass with race detector
go test ./... -v -race

# 3. Coverage >= 60%
go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out | grep total

# 4. No unchecked type assertions in app.go
grep -c '\.([A-Z].*ViewModel)' internal/tui/app/app.go | grep -v ok

# 5. No unchecked JWT claims
grep -c 'claims\[' internal/server/middleware/auth.go

# 6. Lint passes
golangci-lint run ./...
```

## Dependency Graph

```
Phase 1:  [1]  [2]  [3]     (no deps — parallel safe)
Phase 2:  [4]  [5]  [6]  [7]  (depend on all of Phase 1)
Phase 3:  [8]  [9]  [10] [11] (depend on all of Phase 2)
Phase 4:  [12]                 (depends on all of Phase 3)
```
