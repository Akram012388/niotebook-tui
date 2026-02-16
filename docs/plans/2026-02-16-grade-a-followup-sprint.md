# Grade-A Follow-Up Sprint Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Complete the remaining 5 tasks (original tasks 3-7) from the Grade-A sprint to achieve Grade-A (90+/100) across all 7 SWE pillars.

**Architecture:** Targeted fixes — view test coverage expansion (login submit, register submit, timeline init/fetch, compose init), CI version alignment, git hygiene (tracked binary removal), .env.example completeness, and final comprehensive verification. No architectural changes.

**Tech Stack:** Go 1.25+, golangci-lint, go test, git

---

### Context: What's Already Done

Tasks 1-2 from the original Grade-A sprint completed successfully:
- **Task 1** (done): CORS hardened — wildcard removed, security headers added
- **Task 2** (done): Handler unit tests for helpers.go — 7 test functions

Current metrics:
- Build: clean
- Lint: 0 issues
- Tests: all passing
- Coverage: 66.3% total (views at 58.9%)
- Branch: `grade-a-sprint`

### What Remains

| Task | Original # | Description | Pillar Impact |
|------|-----------|-------------|---------------|
| 1 | 3 | Add view tests for submit/init/fetch paths | Testing & QA |
| 2 | 4 | Fix go.mod/CI Go version mismatch | DevOps, Maintainability |
| 3 | 5 | Remove tracked server binary from git | Maintainability |
| 4 | 6 | Update .env.example with missing vars | Maintainability |
| 5 | 7 | Final verification — all 7 pillars | All |

---

### Task 1: Add view tests for submit, init, and fetch paths

**Files:**
- Modify: `internal/tui/views/login_test.go`
- Modify: `internal/tui/views/register_test.go`
- Modify: `internal/tui/views/timeline_test.go`
- Modify: `internal/tui/views/compose_test.go`

**Step 1: Add login submit tests**

Append to `internal/tui/views/login_test.go`:

```go
func TestLoginSubmitEmptyEmail(t *testing.T) {
	m := views.NewLoginModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("expected nil cmd when email is empty")
	}
	view := m.View()
	if !strings.Contains(view, "email is required") {
		t.Error("expected email validation error in view")
	}
}

func TestLoginSubmitEmptyPassword(t *testing.T) {
	m := views.NewLoginModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	for _, r := range "test@example.com" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("expected nil cmd when password is empty")
	}
	view := m.View()
	if !strings.Contains(view, "password is required") {
		t.Error("expected password validation error in view")
	}
}

func TestLoginSubmitWithNilClient(t *testing.T) {
	m := views.NewLoginModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	for _, r := range "test@example.com" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	for _, r := range "password123" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected cmd from submit")
	}
	msg := cmd()
	if _, ok := msg.(app.MsgAuthError); !ok {
		t.Errorf("expected MsgAuthError with nil client, got %T", msg)
	}
}

func TestLoginAuthErrorShowsMessage(t *testing.T) {
	m := views.NewLoginModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = m.Update(app.MsgAuthError{Message: "wrong password", Field: "password"})
	view := m.View()
	if !strings.Contains(view, "wrong password") {
		t.Error("expected error message in view")
	}
}

func TestLoginKeypressClearsError(t *testing.T) {
	m := views.NewLoginModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = m.Update(app.MsgAuthError{Message: "bad password", Field: "password"})
	view := m.View()
	if !strings.Contains(view, "bad password") {
		t.Fatal("error should be visible")
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	view = m.View()
	if strings.Contains(view, "bad password") {
		t.Error("error should be cleared after keypress")
	}
}

func TestLoginTabPastPasswordSwitchesToRegister(t *testing.T) {
	m := views.NewLoginModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.FocusIndex() != 1 {
		t.Fatalf("focus = %d, want 1", m.FocusIndex())
	}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if cmd == nil {
		t.Fatal("expected cmd for switching to register")
	}
	msg := cmd()
	if _, ok := msg.(app.MsgSwitchToRegister); !ok {
		t.Errorf("expected MsgSwitchToRegister, got %T", msg)
	}
}

func TestLoginInit(t *testing.T) {
	m := views.NewLoginModel(nil)
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init should return a blink command")
	}
}
```

Add missing import to login_test.go: `"github.com/Akram012388/niotebook-tui/internal/tui/app"`

**Step 2: Add register submit tests**

Append to `internal/tui/views/register_test.go`:

```go
func TestRegisterSubmitEmptyUsername(t *testing.T) {
	m := views.NewRegisterModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("expected nil cmd when username is empty")
	}
	view := m.View()
	if !strings.Contains(view, "username is required") {
		t.Error("expected username validation error")
	}
}

func TestRegisterSubmitShortUsername(t *testing.T) {
	m := views.NewRegisterModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	for _, r := range "ab" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("expected nil cmd for short username")
	}
	view := m.View()
	if !strings.Contains(view, "at least 3 characters") {
		t.Error("expected short username error")
	}
}

func TestRegisterSubmitInvalidEmail(t *testing.T) {
	m := views.NewRegisterModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	for _, r := range "akram" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	for _, r := range "notanemail" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("expected nil cmd for invalid email")
	}
	view := m.View()
	if !strings.Contains(view, "invalid email") {
		t.Error("expected invalid email error")
	}
}

func TestRegisterSubmitShortPassword(t *testing.T) {
	m := views.NewRegisterModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	for _, r := range "akram" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	for _, r := range "test@example.com" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	for _, r := range "short" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("expected nil cmd for short password")
	}
	view := m.View()
	if !strings.Contains(view, "at least 8 characters") {
		t.Error("expected short password error")
	}
}

func TestRegisterSubmitWithNilClient(t *testing.T) {
	m := views.NewRegisterModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	for _, r := range "akram" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	for _, r := range "test@example.com" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	for _, r := range "password123" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected cmd from submit")
	}
	msg := cmd()
	if _, ok := msg.(app.MsgAuthError); !ok {
		t.Errorf("expected MsgAuthError with nil client, got %T", msg)
	}
}

func TestRegisterTabPastPasswordSwitchesToLogin(t *testing.T) {
	m := views.NewRegisterModel(nil)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab}) // username → email
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab}) // email → password
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyTab}) // password → switch to login
	if cmd == nil {
		t.Fatal("expected cmd for switching to login")
	}
	msg := cmd()
	if _, ok := msg.(app.MsgSwitchToLogin); !ok {
		t.Errorf("expected MsgSwitchToLogin, got %T", msg)
	}
}

func TestRegisterInit(t *testing.T) {
	m := views.NewRegisterModel(nil)
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init should return a blink command")
	}
}

func TestRegisterKeypressClearsError(t *testing.T) {
	m := views.NewRegisterModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = m.Update(app.MsgAuthError{Message: "email taken", Field: "email"})
	view := m.View()
	if !strings.Contains(view, "email taken") {
		t.Fatal("error should be visible")
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	view = m.View()
	if strings.Contains(view, "email taken") {
		t.Error("error should be cleared after keypress")
	}
}
```

**Step 3: Add timeline init/fetch tests**

Append to `internal/tui/views/timeline_test.go`:

```go
func TestTimelineInitReturnsCmd(t *testing.T) {
	m := views.NewTimelineModel(nil)
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init should return a fetch command")
	}
	msg := cmd()
	if _, ok := msg.(app.MsgAPIError); !ok {
		t.Errorf("expected MsgAPIError with nil client, got %T", msg)
	}
}

func TestTimelineFetchLatestReturnsCmd(t *testing.T) {
	m := views.NewTimelineModel(nil)
	cmd := m.FetchLatest()
	if cmd == nil {
		t.Error("FetchLatest should return a command")
	}
	msg := cmd()
	if _, ok := msg.(app.MsgAPIError); !ok {
		t.Errorf("expected MsgAPIError with nil client, got %T", msg)
	}
}

func TestTimelineSpaceAndBPagination(t *testing.T) {
	posts := make([]models.Post, 20)
	for i := range posts {
		posts[i] = models.Post{
			ID:      fmt.Sprintf("%d", i),
			Author:  &models.User{Username: "user"},
			Content: fmt.Sprintf("Post %d", i),
		}
	}
	m := views.NewTimelineModel(nil)
	m.SetPosts(posts)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Space pages down
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	if m.CursorIndex() == 0 {
		t.Error("expected cursor to move after space")
	}

	// b pages back up
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	if m.CursorIndex() != 0 {
		t.Errorf("cursor = %d after b, want 0", m.CursorIndex())
	}
}
```

**Step 4: Add compose init test**

Append to `internal/tui/views/compose_test.go`:

```go
func TestComposeInitReturnsBlink(t *testing.T) {
	m := views.NewComposeModel(nil)
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init should return blink command")
	}
}
```

**Step 5: Run tests to verify**

Run: `go test ./internal/tui/views/ -v -race -count=1`
Expected: ALL PASS

**Step 6: Run full suite with coverage**

Run: `go test ./... -race -count=1 -coverprofile=coverage.out && go tool cover -func=coverage.out | grep total`
Expected: Coverage increases from 66.3% to 68%+

**Step 7: Commit**

```bash
git add internal/tui/views/*_test.go
git commit -m "test: add 20+ view tests covering submit, init, fetch, and validation paths"
```

---

### Task 2: Fix CI Go version mismatch

**Files:**
- Modify: `.github/workflows/ci.yml`

**Step 1: Update Go version in CI**

Change both occurrences of `go-version: '1.22'` to `go-version: '1.25'` in `.github/workflows/ci.yml`.

**Step 2: Verify build**

Run: `go build ./...`
Expected: Clean exit

**Step 3: Commit**

```bash
git add .github/workflows/ci.yml
git commit -m "chore: update CI Go version to match go.mod (1.22 → 1.25)"
```

---

### Task 3: Remove tracked server binary from git

**Files:**
- Modify: `.gitignore`

**Step 1: Add server binary to .gitignore**

Add `server` on a new line under the `# Binaries` section of `.gitignore`.

**Step 2: Remove from git tracking**

```bash
git rm --cached server
```

**Step 3: Verify**

```bash
git status
```
Expected: `server` shows as deleted from index, `.gitignore` shows as modified.

**Step 4: Commit**

```bash
git add .gitignore
git commit -m "chore: remove tracked server binary, add to gitignore"
```

---

### Task 4: Update .env.example with missing variables

**Files:**
- Modify: `.env.example`

**Step 1: Add missing variables**

Append these lines to `.env.example`:

```
NIOTEBOOK_CORS_ORIGIN=http://localhost:3000
# For testing:
# NIOTEBOOK_TEST_DB_URL=postgres://localhost/niotebook_test?sslmode=disable
```

**Step 2: Commit**

```bash
git add .env.example
git commit -m "docs: add CORS_ORIGIN and TEST_DB_URL to .env.example"
```

---

### Task 5: Final verification — all 7 Grade-A pillars

**Step 1: Run lint**

Run: `golangci-lint run ./...`
Expected: 0 issues

**Step 2: Run build**

Run: `go build ./...`
Expected: Clean exit

**Step 3: Run full test suite with coverage**

Run: `go test ./... -race -count=1 -coverprofile=coverage.out`
Expected: ALL PASS, 0 failures

**Step 4: Check coverage**

Run: `go tool cover -func=coverage.out | grep total`
Expected: 68%+ total

**Step 5: Verify security checklist**

- `grep -rn 'claims\[' internal/server/middleware/auth.go` — all use comma-ok
- `grep -n 'len(jwtSecret)' cmd/server/main.go` — min 32 bytes check exists
- `grep -rn 'updated\.(' internal/tui/app/app.go` — all use `if x, ok := ...`
- `grep -n 'X-Content-Type-Options' internal/server/middleware/cors.go` — nosniff header set
- `grep -n 'X-Frame-Options' internal/server/middleware/cors.go` — DENY header set

**Step 6: Verify maintainability checklist**

- `git ls-files server` — empty (binary no longer tracked)
- `grep NIOTEBOOK_CORS_ORIGIN .env.example` — present
- `grep go-version .github/workflows/ci.yml` — shows 1.25
- `grep XDG_CONFIG_HOME internal/tui/config/config.go` — XDG honored

**Step 7: Report Grade-A scorecard**

Print final pillar scores based on evidence collected.
