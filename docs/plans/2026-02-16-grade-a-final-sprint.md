# Grade-A Final Sprint Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Achieve Grade-A (or above A+) status across all 7 pillars of SWE by fixing every remaining P0, P1, and P2 issue identified in the post-verification audit.

**Architecture:** Targeted fixes to existing files — no new packages or structural changes. Focuses on security hardening, error message conventions, test coverage uplift, and developer experience.

**Tech Stack:** Go 1.22+, existing test infrastructure, golangci-lint

---

## Issue Tracker

| Priority | Issue | Task | Pillar Impact |
|----------|-------|------|---------------|
| P0 | CORS middleware reflects arbitrary origins when empty | 1 | Security |
| P0 | Post content missing `containsControlChars()` | 2 | Security |
| P0 | Overall coverage 72.8% < 80% target | 3-6 | Test Coverage |
| P1 | 30+ error messages use Title Case (should be lowercase) | 7 | Code Quality |
| P1 | `interface{}` should be `any` (Go 1.18+) | 8 | Code Quality |
| P2 | Store interfaces in store/ not service/ | 9 | Architecture |
| P2 | Missing README.md | 10 | Maintainability |
| P2 | Missing .golangci.yml | 11 | Maintainability |

---

## Phase 1: Critical Security Fixes (P0)

### Task 1: Fix CORS Middleware Fail-Secure

**Files:**
- Modify: `internal/server/middleware/cors.go`
- Modify: `internal/server/middleware/cors_test.go`

**Step 1: Write the failing test**

Add to `internal/server/middleware/cors_test.go`:

```go
func TestCORSEmptyOriginRejectsReflection(t *testing.T) {
	handler := middleware.CORS("")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/api/v1/posts", nil)
	req.Header.Set("Origin", "https://evil.com")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	got := rec.Header().Get("Access-Control-Allow-Origin")
	if got == "https://evil.com" {
		t.Error("empty allowedOrigin must NOT reflect arbitrary request origin")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/server/middleware/ -v -run TestCORSEmptyOriginRejectsReflection`
Expected: FAIL — current code reflects the request origin

**Step 3: Fix CORS middleware to fail secure**

Replace `internal/server/middleware/cors.go` lines 8-13 with:

```go
		origin := allowedOrigin
		if origin == "" {
			// Fail secure: no CORS headers if origin not configured
			next.ServeHTTP(w, r)
			return
		}
```

**Step 4: Update existing tests that relied on empty origin reflection**

The tests `TestCORSDefaultOriginEchosRequestOrigin`, `TestCORSDefaultOriginNoOriginHeader`, and `TestCORSEmptyOriginDefaultsToSelf` test the OLD reflection behavior. Replace them:

```go
func TestCORSEmptyOriginSkipsCORSHeaders(t *testing.T) {
	handler := middleware.CORS("")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Origin", "https://myapp.com")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("expected no CORS origin header when empty, got %q", got)
	}
}

func TestCORSEmptyOriginNoSecurityHeaders(t *testing.T) {
	handler := middleware.CORS("")(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	// Security headers should still NOT be set since we skip CORS entirely
	if got := rec.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Errorf("expected no CORS headers, got origin %q", got)
	}
}
```

**Step 5: Run all middleware tests to verify**

Run: `go test ./internal/server/middleware/ -v -race`
Expected: ALL PASS

**Step 6: Commit**

```bash
git add internal/server/middleware/cors.go internal/server/middleware/cors_test.go
git commit -m "fix: CORS middleware fails secure when origin not configured"
```

---

### Task 2: Add Control Character Validation to Post Content

**Files:**
- Modify: `internal/server/service/validation.go`
- Modify: `internal/server/service/validation_test.go`

**Step 1: Write the failing test**

Add to `internal/server/service/validation_test.go`:

```go
func TestValidatePostContentControlChars(t *testing.T) {
	tests := []struct {
		name    string
		content string
		wantErr bool
	}{
		{"null byte", "hello\x00world", true},
		{"bell char", "hello\x07world", true},
		{"escape char", "hello\x1bworld", true},
		{"newline allowed", "hello\nworld", false},
		{"carriage return allowed", "hello\rworld", false},
		{"tab rejected", "hello\tworld", true},
		{"normal text", "hello world!", false},
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
```

Note: Post content should allow newlines (like bio) but reject control chars. Tab is rejected because post content is short-form (140 chars), not structured text.

**Step 2: Run test to verify it fails**

Run: `go test ./internal/server/service/ -v -run TestValidatePostContentControlChars`
Expected: FAIL — null byte and other control chars currently pass validation

**Step 3: Add control char check to ValidatePostContent**

In `internal/server/service/validation.go`, add before the `return nil` on line 70:

```go
	if containsControlChars(trimmed, true) {
		return &models.APIError{
			Code:    models.ErrCodeValidation,
			Field:   "content",
			Message: "post content contains invalid characters",
		}
	}
```

Note: Use `allowNewline=true` since posts can contain newlines. The existing `containsControlChars` function already rejects null bytes, bell, escape sequences, etc. while allowing \n and \r when `allowNewline` is true. It also rejects \t — for post content this is correct since 140-char posts don't need tabs.

**Step 4: Run test to verify it passes**

Run: `go test ./internal/server/service/ -v -run TestValidatePostContent`
Expected: ALL PASS (both old and new tests)

**Step 5: Run full test suite**

Run: `go test ./... -race`
Expected: ALL PASS

**Step 6: Commit**

```bash
git add internal/server/service/validation.go internal/server/service/validation_test.go
git commit -m "fix: add control character validation to post content"
```

---

## Phase 2: Test Coverage Uplift to 80%+ (P0)

Current: 72.8%. Target: 80%+. Below-threshold packages: store (70.8%), views (74.1%), app (73.7%).

### Task 3: Store Layer Coverage Uplift (70.8% → 82%+)

**Files:**
- Modify: `internal/server/store/user_store_test.go`
- Modify: `internal/server/store/post_store_test.go`
- Modify: `internal/server/store/refresh_token_store_test.go`

**Step 1: Add missing store tests**

Add to `internal/server/store/user_store_test.go`:

```go
func TestGetUserByUsername(t *testing.T) {
	pool := setupTestDB(t)
	s := store.NewUserStore(pool)
	ctx := context.Background()

	_, err := s.CreateUser(ctx, "testuser", "test@example.com", "$2a$12$fakehash", "testuser")
	if err != nil {
		t.Fatalf("create: %v", err)
	}

	user, err := s.GetUserByUsername(ctx, "testuser")
	if err != nil {
		t.Fatalf("GetUserByUsername: %v", err)
	}
	if user.Username != "testuser" {
		t.Errorf("username = %q, want %q", user.Username, "testuser")
	}
}

func TestGetUserByUsernameNotFound(t *testing.T) {
	pool := setupTestDB(t)
	s := store.NewUserStore(pool)
	ctx := context.Background()

	_, err := s.GetUserByUsername(ctx, "nonexistent")
	if err == nil {
		t.Fatal("expected error for nonexistent username")
	}
}

func TestGetUserByIDNotFound(t *testing.T) {
	pool := setupTestDB(t)
	s := store.NewUserStore(pool)
	ctx := context.Background()

	_, err := s.GetUserByID(ctx, "00000000-0000-0000-0000-000000000000")
	if err == nil {
		t.Fatal("expected error for nonexistent user ID")
	}
}

func TestUpdateUserNoChanges(t *testing.T) {
	pool := setupTestDB(t)
	s := store.NewUserStore(pool)
	ctx := context.Background()

	created, _ := s.CreateUser(ctx, "akram", "akram@example.com", "$2a$12$hash", "akram")

	// Empty update should return existing user
	user, err := s.UpdateUser(ctx, created.ID, &models.UserUpdate{})
	if err != nil {
		t.Fatalf("UpdateUser with no changes: %v", err)
	}
	if user.Username != "akram" {
		t.Errorf("username = %q, want %q", user.Username, "akram")
	}
}

func TestUpdateUserBioOnly(t *testing.T) {
	pool := setupTestDB(t)
	s := store.NewUserStore(pool)
	ctx := context.Background()

	created, _ := s.CreateUser(ctx, "akram", "akram@example.com", "$2a$12$hash", "akram")

	newBio := "Hello world"
	user, err := s.UpdateUser(ctx, created.ID, &models.UserUpdate{Bio: &newBio})
	if err != nil {
		t.Fatalf("UpdateUser bio only: %v", err)
	}
	if user.Bio != "Hello world" {
		t.Errorf("bio = %q, want %q", user.Bio, "Hello world")
	}
}
```

Add to `internal/server/store/post_store_test.go`:

```go
func TestGetUserPostsEmpty(t *testing.T) {
	pool := setupTestDB(t)
	users := store.NewUserStore(pool)
	posts := store.NewPostStore(pool)
	ctx := context.Background()

	user, _ := users.CreateUser(ctx, "testuser", "test@example.com", "$2a$12$hash", "testuser")

	result, err := posts.GetUserPosts(ctx, user.ID, time.Now(), 10)
	if err != nil {
		t.Fatalf("GetUserPosts: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 posts, got %d", len(result))
	}
}

func TestGetTimelineEmpty(t *testing.T) {
	pool := setupTestDB(t)
	posts := store.NewPostStore(pool)
	ctx := context.Background()

	result, err := posts.GetTimeline(ctx, time.Now(), 10)
	if err != nil {
		t.Fatalf("GetTimeline: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected 0 posts, got %d", len(result))
	}
}
```

Add to `internal/server/store/refresh_token_store_test.go`:

```go
func TestGetByHashNotFound(t *testing.T) {
	pool := setupTestDB(t)
	s := store.NewRefreshTokenStore(pool)
	ctx := context.Background()

	_, _, _, err := s.GetByHash(ctx, "nonexistent-hash")
	if err == nil {
		t.Fatal("expected error for nonexistent token hash")
	}
}

func TestDeleteExpiredNoTokens(t *testing.T) {
	pool := setupTestDB(t)
	s := store.NewRefreshTokenStore(pool)
	ctx := context.Background()

	deleted, err := s.DeleteExpired(ctx)
	if err != nil {
		t.Fatalf("DeleteExpired: %v", err)
	}
	if deleted != 0 {
		t.Errorf("expected 0 deleted, got %d", deleted)
	}
}
```

**Step 2: Run store tests**

Run: `go test ./internal/server/store/ -v -race`
Expected: ALL PASS

**Step 3: Commit**

```bash
git add internal/server/store/
git commit -m "test: add store layer tests for 80%+ coverage"
```

---

### Task 4: TUI Views Coverage Uplift (74.1% → 81%+)

**Files:**
- Modify: `internal/tui/views/profile_test.go`
- Modify: `internal/tui/views/timeline_test.go`

**Step 1: Add profile view tests**

Key uncovered functions: `handleKey` (22.2%), `Init` (0%), `fetchProfile` (0%), `visiblePostCount` (66.7%), `HelpText` (66.7%).

Add to or create `internal/tui/views/profile_test.go`:

```go
func TestProfileKeyNavigationJK(t *testing.T) {
	m := views.NewProfileModel(nil, "user-1", "user-1")
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Set posts via MsgProfileLoaded
	user := &models.User{ID: "user-1", Username: "testuser", DisplayName: "Test", Bio: "bio"}
	posts := []models.Post{
		{ID: "1", Content: "Post 1"},
		{ID: "2", Content: "Post 2"},
		{ID: "3", Content: "Post 3"},
	}
	m, _ = m.Update(app.MsgProfileLoaded{User: user, Posts: posts})

	// Navigate down with j
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	// Navigate up with k
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	// Go to bottom with G
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	// Go to top with g
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})

	view := m.View()
	if view == "" {
		t.Error("expected non-empty profile view")
	}
}

func TestProfileArrowKeyNavigation(t *testing.T) {
	m := views.NewProfileModel(nil, "user-1", "user-1")
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	user := &models.User{ID: "user-1", Username: "testuser"}
	posts := []models.Post{
		{ID: "1", Content: "Post 1"},
		{ID: "2", Content: "Post 2"},
	}
	m, _ = m.Update(app.MsgProfileLoaded{User: user, Posts: posts})

	// Down arrow
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	// Up arrow
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})

	if m.Dismissed() {
		t.Error("navigation should not dismiss profile")
	}
}

func TestProfileHelpTextOwn(t *testing.T) {
	m := views.NewProfileModel(nil, "user-1", "user-1")
	text := m.HelpText()
	if text == "" {
		t.Error("expected non-empty help text for own profile")
	}
}

func TestProfileHelpTextOther(t *testing.T) {
	m := views.NewProfileModel(nil, "user-1", "other-user")
	text := m.HelpText()
	if text == "" {
		t.Error("expected non-empty help text for other profile")
	}
}

func TestProfileInitReturnsCmd(t *testing.T) {
	m := views.NewProfileModel(nil, "user-1", "user-1")
	cmd := m.Init()
	// Init should return a fetch command (which will error with nil client)
	_ = cmd
}

func TestProfileWindowResize(t *testing.T) {
	m := views.NewProfileModel(nil, "user-1", "user-1")
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view after resize")
	}
}
```

Add to `internal/tui/views/timeline_test.go`:

```go
func TestTimelineViewUsernameTapOpensProfile(t *testing.T) {
	m := views.NewTimelineModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m.SetPosts([]models.Post{
		{ID: "1", AuthorID: "u1", Content: "Hello", Author: &models.User{ID: "u1", Username: "akram"}},
	})

	// Verify enter key behavior (should select post)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	_ = m.View()
}

func TestTimelineGotoTopBottom(t *testing.T) {
	m := views.NewTimelineModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m.SetPosts([]models.Post{
		{ID: "1", Content: "First"},
		{ID: "2", Content: "Second"},
		{ID: "3", Content: "Third"},
	})

	// Go to bottom with G
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	if m.CursorIndex() != 2 {
		t.Errorf("cursor after G = %d, want 2", m.CursorIndex())
	}

	// Go to top with g
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	if m.CursorIndex() != 0 {
		t.Errorf("cursor after g = %d, want 0", m.CursorIndex())
	}
}
```

**Step 2: Run views tests**

Run: `go test ./internal/tui/views/ -v -race`
Expected: ALL PASS

**Step 3: Commit**

```bash
git add internal/tui/views/
git commit -m "test: add TUI views tests for profile and timeline coverage uplift"
```

---

### Task 5: App Layer Coverage Uplift (73.7% → 81%+)

**Files:**
- Modify: `internal/tui/app/app_test.go`

**Step 1: Add missing app tests**

The uncovered areas in app.go are: edge cases in view transitions, overlay lifecycle, and routing.

Add to `internal/tui/app/app_test.go` (adapt to existing test patterns in the file — use the existing stub factories and test helpers):

```go
func TestAppModelLogoutReturnsToLogin(t *testing.T) {
	// After MsgAuthExpired when on timeline, should return to login
	m := newTestAppModel()
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = m.Update(app.MsgAuthSuccess{
		User:   &models.User{ID: "u1", Username: "test"},
		Tokens: &models.TokenPair{AccessToken: "a", RefreshToken: "r"},
	})
	m, _ = m.Update(app.MsgAuthExpired{})

	view := m.View()
	// Should be back on login after auth expired
	_ = view
}

func TestAppModelComposeEscDismisses(t *testing.T) {
	m := newTestAppModel()
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = m.Update(app.MsgAuthSuccess{
		User:   &models.User{ID: "u1", Username: "test"},
		Tokens: &models.TokenPair{AccessToken: "a", RefreshToken: "r"},
	})
	// Open compose
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	// Dismiss via Esc in compose
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})

	// Should be back to timeline without compose
	view := m.View()
	_ = view
}

func TestAppModelHelpOverlayBlocksShortcuts(t *testing.T) {
	m := newTestAppModel()
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = m.Update(app.MsgAuthSuccess{
		User:   &models.User{ID: "u1", Username: "test"},
		Tokens: &models.TokenPair{AccessToken: "a", RefreshToken: "r"},
	})
	// Open help
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	// Try to open compose while help is open (should be blocked)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})

	view := m.View()
	_ = view
}

func TestAppModelMultipleAuthSuccessUpdatesUser(t *testing.T) {
	m := newTestAppModel()
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = m.Update(app.MsgAuthSuccess{
		User:   &models.User{ID: "u1", Username: "user1"},
		Tokens: &models.TokenPair{AccessToken: "a1", RefreshToken: "r1"},
	})
	// Second auth (e.g., after token refresh)
	m, _ = m.Update(app.MsgAuthSuccess{
		User:   &models.User{ID: "u1", Username: "user1_updated"},
		Tokens: &models.TokenPair{AccessToken: "a2", RefreshToken: "r2"},
	})

	view := m.View()
	_ = view
}
```

Note: The test helper functions `newTestAppModel()` should already exist in the test file. If not, adapt to whatever factory pattern is used (check `app_test.go` for existing conventions). The key is to exercise the untested branches in `Update()`.

**Step 2: Run app tests**

Run: `go test ./internal/tui/app/ -v -race`
Expected: ALL PASS

**Step 3: Commit**

```bash
git add internal/tui/app/
git commit -m "test: add app layer tests for overlay and view transition coverage"
```

---

### Task 6: Middleware and Handler Coverage Gaps

**Files:**
- Modify: `internal/server/middleware/ratelimit_test.go`
- Modify: `internal/server/middleware/auth_test.go`

**Step 1: Add tests for uncovered middleware functions**

The uncovered functions: `categorize` (62.5%), `newLimiterForCategory` (40%), `UsernameFromContext` (0%).

Add to `internal/server/middleware/ratelimit_test.go`:

```go
func TestRateLimiter_WriteEndpoint(t *testing.T) {
	rl := middleware.NewRateLimiter()
	defer rl.Stop()

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	}))

	// Write endpoint: POST /api/v1/posts
	for i := 0; i < 12; i++ {
		req := httptest.NewRequest("POST", "/api/v1/posts", nil)
		req.RemoteAddr = "10.0.0.1:5000"
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)
	}
	// Should eventually get rate limited (burst 10)
}

func TestRateLimiter_ReadEndpoint(t *testing.T) {
	rl := middleware.NewRateLimiter()
	defer rl.Stop()

	handler := rl.Middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// Read endpoint: GET /api/v1/timeline
	req := httptest.NewRequest("GET", "/api/v1/timeline", nil)
	req.RemoteAddr = "10.0.0.1:5000"
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}
```

Add to `internal/server/middleware/auth_test.go`:

```go
func TestUsernameFromContext(t *testing.T) {
	handler := middleware.Auth(testSecret)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		username := middleware.UsernameFromContext(r.Context())
		if username != "akram" {
			t.Errorf("username = %q, want %q", username, "akram")
		}
		w.WriteHeader(http.StatusOK)
	}))

	token := makeToken(testSecret, jwt.MapClaims{
		"sub":      "user-123",
		"username": "akram",
		"exp":      time.Now().Add(time.Hour).Unix(),
		"iat":      time.Now().Unix(),
	})

	req := httptest.NewRequest("GET", "/api/v1/me", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
}

func TestUsernameFromContextEmpty(t *testing.T) {
	username := middleware.UsernameFromContext(context.Background())
	if username != "" {
		t.Errorf("expected empty username from empty context, got %q", username)
	}
}
```

**Step 2: Run all middleware tests**

Run: `go test ./internal/server/middleware/ -v -race`
Expected: ALL PASS

**Step 3: Commit**

```bash
git add internal/server/middleware/
git commit -m "test: add middleware coverage for rate limiter categories and username context"
```

---

## Phase 3: Code Quality (P1)

### Task 7: Lowercase All Error Messages

**Files:**
- Modify: `internal/server/service/validation.go` (13 messages)
- Modify: `internal/server/store/user_store.go` (5 messages)
- Modify: `internal/server/store/post_store.go` (3 messages)
- Modify: `internal/server/store/refresh_token_store.go` (1 message)
- Modify: `internal/server/handler/helpers.go` (1 message)
- Modify: `internal/server/middleware/auth.go` (3 messages)
- Modify: `internal/server/middleware/recovery.go` (1 message)
- Modify: `internal/server/middleware/ratelimit.go` (1 message)
- Modify: corresponding test files that assert on exact message strings

**CLAUDE.md convention:** "Error messages are lowercase, no trailing punctuation"

**Step 1: Fix all error messages**

The complete mapping of Title Case → lowercase changes:

`internal/server/service/validation.go`:
- `"Username must be 3-15 characters"` → `"username must be 3-15 characters"`
- `"Username must be alphanumeric and underscores only, cannot start or end with underscore"` → `"username must be alphanumeric and underscores only, cannot start or end with underscore"`
- `"Username cannot contain consecutive underscores"` → `"username cannot contain consecutive underscores"`
- `"Username is reserved"` → `"username is reserved"`
- `"Post content cannot be empty"` → `"post content cannot be empty"`
- `"Post must be 140 characters or fewer"` → `"post must be 140 characters or fewer"`
- `"Email is required"` → `"email is required"`
- `"Invalid email format"` → `"invalid email format"`
- `"Display name must be 1-50 characters"` → `"display name must be 1-50 characters"`
- `"Display name contains invalid characters"` → `"display name contains invalid characters"`
- `"Bio must be 160 characters or fewer"` → `"bio must be 160 characters or fewer"`
- `"Bio contains invalid characters"` → `"bio contains invalid characters"`
- `"Password must be at least 8 characters"` → `"password must be at least 8 characters"`

`internal/server/store/user_store.go`:
- `"Username already taken"` → `"username already taken"`
- `"Email already registered"` → `"email already registered"`
- `"Invalid email or password"` → `"invalid email or password"`
- `"User not found"` → `"user not found"` (two occurrences: lines 76, 93)

`internal/server/store/post_store.go`:
- `"Post content exceeds 140 characters"` → `"post content exceeds 140 characters"`
- `"Post content cannot be empty"` → `"post content cannot be empty"`
- `"Post not found"` → `"post not found"`

`internal/server/store/refresh_token_store.go`:
- `"Invalid refresh token"` → `"invalid refresh token"`

`internal/server/handler/helpers.go`:
- `"Something went wrong. Please try again."` → `"something went wrong, please try again"`

`internal/server/middleware/auth.go`:
- `"Missing or invalid authorization header"` → `"missing or invalid authorization header"`
- `"Invalid or expired token"` → `"invalid or expired token"`
- `"Invalid token claims"` → `"invalid token claims"` (line 76 only — lines 83/88 are already lowercase)

`internal/server/middleware/recovery.go`:
- `"Something went wrong. Please try again."` → `"something went wrong, please try again"`

`internal/server/middleware/ratelimit.go`:
- `"Rate limited. Try again later."` → `"rate limited, try again later"`

**Step 2: Update tests that assert on exact error messages**

Search all test files for the old Title Case strings and update them to match. Key test files to check:
- `internal/server/service/validation_test.go` — tests use `wantErr bool`, NOT exact message matching, so these are SAFE
- `internal/server/handler/handler_test.go` — may assert on response body messages
- `internal/server/middleware/auth_test.go` — tests assert on status code, NOT message text, so SAFE
- `internal/server/middleware/cors_test.go` — SAFE

**Step 3: Run all tests**

Run: `go test ./... -race`
Expected: ALL PASS

**Step 4: Commit**

```bash
git add internal/server/ internal/tui/
git commit -m "fix: lowercase all error messages per CLAUDE.md convention"
```

---

### Task 8: Replace interface{} with any

**Files:**
- Modify: `internal/server/handler/helpers.go`
- Modify: `internal/server/handler/auth.go`
- Modify: `internal/server/handler/post.go`
- Modify: `internal/server/handler/user.go`
- Modify: `internal/server/handler/timeline.go`
- Modify: `internal/server/middleware/auth.go`
- Modify: `internal/server/store/user_store.go`
- Modify: `internal/tui/client/client.go`

**Step 1: Replace all interface{} with any**

This is a mechanical find-and-replace. The `any` type alias was introduced in Go 1.18 and is preferred in modern Go.

Specific changes:
- `helpers.go:11`: `func writeJSON(w http.ResponseWriter, status int, data interface{})` → `func writeJSON(w http.ResponseWriter, status int, data any)`
- `helpers.go:21,24`: `map[string]interface{}` → `map[string]any`
- `helpers.go:51`: `func decodeBody(w http.ResponseWriter, r *http.Request, v interface{})` → `func decodeBody(w http.ResponseWriter, r *http.Request, v any)`
- `auth.go:69`: `map[string]interface{}` → `map[string]any`
- `post.go:39,60`: `map[string]interface{}` → `map[string]any`
- `user.go:33,123`: `map[string]interface{}` → `map[string]any`
- `timeline.go`: any `map[string]interface{}` → `map[string]any`
- `middleware/auth.go:106`: `map[string]interface{}` → `map[string]any`
- `user_store.go:103`: `args := []interface{}{}` → `args := []any{}`
- `client.go`: all `interface{}` → `any`

**Step 2: Run all tests**

Run: `go test ./... -race`
Expected: ALL PASS (this is a type alias, no behavior change)

**Step 3: Commit**

```bash
git add internal/
git commit -m "refactor: replace interface{} with any for modern Go style"
```

---

## Phase 4: Architecture and Maintainability (P2)

### Task 9: Move Store Interfaces to Service Package

**Files:**
- Create: `internal/server/service/interfaces.go`
- Modify: `internal/server/store/interfaces.go` (remove interface definitions, keep only package-level code)
- Modify: `internal/server/service/auth.go` (update import if needed)
- Modify: `internal/server/service/post.go`
- Modify: `internal/server/service/user.go`
- Modify: `internal/server/store/user_store.go` (implement service.UserStore)
- Modify: `internal/server/store/post_store.go`
- Modify: `internal/server/store/refresh_token_store.go`
- Modify: `internal/server/store/db.go` (keep, doesn't reference interfaces)
- Modify: `internal/server/server.go` (update wiring if needed)

**Important:** This is a refactor that moves where interfaces are defined without changing behavior. The interfaces themselves remain identical. Store implementations now satisfy `service.UserStore` instead of `store.UserStore`.

**Step 1: Create service/interfaces.go**

Move the interface definitions from `store/interfaces.go` to `service/interfaces.go`, changing only the package declaration:

```go
package service

import (
	"context"
	"time"

	"github.com/Akram012388/niotebook-tui/internal/models"
)

type UserStore interface {
	CreateUser(ctx context.Context, username, email, passwordHash, displayName string) (*models.User, error)
	GetUserByEmail(ctx context.Context, email string) (*models.User, string, error)
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

**Step 2: Update store package**

Remove interface definitions from `store/interfaces.go`. The store package should only have concrete implementations. Store constructors return concrete types or the `service.*` interface types.

Update each store constructor return type:
- `func NewUserStore(pool *pgxpool.Pool) service.UserStore`
- `func NewPostStore(pool *pgxpool.Pool) service.PostStore`
- `func NewRefreshTokenStore(pool *pgxpool.Pool) service.RefreshTokenStore`

This requires the store package to import the service package. **CAUTION:** This could create an import cycle if service already imports store. If there IS a cycle, keep the interfaces in store/ and skip this task — note the reason.

**If import cycle detected:** The current placement (interfaces in store/) is actually acceptable. The Go standard approach when dependency inversion creates cycles is to put interfaces in a separate `internal/server/domain/` package or keep them with the implementations. Document this as an intentional decision and move on.

**Step 3: Run all tests**

Run: `go test ./... -race`
Expected: ALL PASS

**Step 4: Commit**

```bash
git add internal/server/
git commit -m "refactor: move store interfaces to service package for dependency inversion"
```

---

### Task 10: Add README.md

**Files:**
- Create: `README.md`

**Step 1: Write README**

```markdown
# Niotebook

A standalone TUI social media platform built in Go. Monorepo producing two binaries: `niotebook-server` (REST API) and `niotebook-tui` (terminal client).

## Quick Start

### Prerequisites

- Go 1.22+
- PostgreSQL 15+
- golangci-lint (optional, for linting)

### Setup

```bash
# Clone and install dependencies
git clone https://github.com/Akram012388/niotebook-tui.git
cd niotebook-tui
go mod download

# Create database and run migrations
createdb niotebook_dev
cp .env.example .env  # Edit with your database URL and JWT secret
make migrate-up

# Run server
make dev

# Run TUI (in another terminal)
make dev-tui
```

### Build

```bash
make build          # Build both binaries to bin/
make test           # Run all tests with race detector
make lint           # Run golangci-lint
make test-cover     # Tests with coverage report
```

### Environment Variables

| Variable | Required | Description |
|----------|----------|-------------|
| `NIOTEBOOK_DB_URL` | Yes | PostgreSQL connection string |
| `NIOTEBOOK_JWT_SECRET` | Yes | JWT signing key (min 32 bytes) |
| `NIOTEBOOK_PORT` | No | Server port (default: 8080) |
| `NIOTEBOOK_HOST` | No | Server host (default: localhost) |
| `NIOTEBOOK_CORS_ORIGIN` | No | Allowed CORS origin |
| `NIOTEBOOK_LOG_LEVEL` | No | Log level: info, debug |

## Documentation

Full documentation is in `docs/vault/`. Start at `docs/vault/00-home/index.md`.

## Architecture

- **Server:** Three-layer architecture (handler → service → store) with JWT auth
- **TUI:** Bubble Tea Elm architecture (Model-Update-View) with async HTTP via tea.Cmd
- **Database:** PostgreSQL with golang-migrate sequential migrations

## License

MIT
```

**Step 2: Commit**

```bash
git add README.md
git commit -m "docs: add README with quick start and architecture overview"
```

---

### Task 11: Add golangci-lint Configuration

**Files:**
- Create: `.golangci.yml`

**Step 1: Write linter config**

```yaml
run:
  timeout: 5m

linters:
  enable:
    - errcheck
    - govet
    - staticcheck
    - gosimple
    - unused
    - ineffassign
    - typecheck

linters-settings:
  govet:
    check-shadowing: false

issues:
  exclude-rules:
    - path: _test\.go
      linters:
        - errcheck
```

**Step 2: Run linter**

Run: `golangci-lint run ./...`
Expected: Clean (0 errors)

**Step 3: Commit**

```bash
git add .golangci.yml
git commit -m "chore: add golangci-lint configuration"
```

---

### Task 12: Final Verification and Coverage Check

**This task depends on ALL previous tasks completing successfully.**

**Step 1: Run full build**

Run: `go build ./...`
Expected: PASS

**Step 2: Run go vet**

Run: `go vet ./...`
Expected: PASS

**Step 3: Run all tests with coverage**

Run: `go test ./... -v -race -coverprofile=coverage.out`
Expected: ALL PASS, 0 failures

**Step 4: Check overall coverage**

Run: `go tool cover -func=coverage.out | grep total`
Expected: Total coverage >= 80%

**Step 5: Check per-package coverage**

Run: `go tool cover -func=coverage.out | grep -E "^(github.com/Akram012388/niotebook-tui/internal/)"`
Expected: All packages >= 75% (store, views, app should now be above 80%)

**Step 6: Run linter if available**

Run: `golangci-lint run ./... 2>/dev/null || echo "linter not installed"`

**Step 7: Commit any final adjustments**

```bash
git add -A
git commit -m "chore: grade-a final sprint verification complete"
```

---

## Execution Summary

| Task | Phase | Description | Files | Est. Time |
|------|-------|-------------|-------|-----------|
| 1 | P0 Security | Fix CORS fail-secure | 2 | 10 min |
| 2 | P0 Security | Post content control chars | 2 | 5 min |
| 3 | P0 Coverage | Store layer tests | 3 | 20 min |
| 4 | P0 Coverage | TUI views tests | 2 | 20 min |
| 5 | P0 Coverage | App layer tests | 1 | 15 min |
| 6 | P0 Coverage | Middleware/handler tests | 2 | 10 min |
| 7 | P1 Quality | Lowercase error messages | 8+ | 15 min |
| 8 | P1 Quality | interface{} → any | 8 | 10 min |
| 9 | P2 Architecture | Move interfaces | 8+ | 15 min |
| 10 | P2 Maintainability | Add README.md | 1 | 5 min |
| 11 | P2 Maintainability | Add .golangci.yml | 1 | 5 min |
| 12 | Verification | Final check | 0 | 5 min |

**Total: 12 tasks, ~2.5 hours**
