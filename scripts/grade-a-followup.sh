#!/usr/bin/env bash
#
# Grade-A Follow-Up Sprint Runner
#
# Picks up where grade-a-sprint.sh stopped (after tasks 1-2).
# Executes the remaining 5 tasks to achieve Grade-A across all 7 SWE pillars.
#
# Usage:
#   ./scripts/grade-a-followup.sh            # Fresh start
#   ./scripts/grade-a-followup.sh --resume   # Resume from progress
#
# Requirements:
#   - Go 1.22+ installed
#   - golangci-lint installed
#   - Claude CLI installed and authenticated
#   - On the grade-a-sprint branch (or will switch to it)
#
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$PROJECT_ROOT"

BRANCH="grade-a-sprint"
PLAN="docs/plans/2026-02-16-grade-a-followup-sprint.md"
LOG_DIR="$PROJECT_ROOT/logs/grade-a-followup"
PROGRESS_FILE="$PROJECT_ROOT/scripts/grade-a-followup-progress.json"
TOTAL_TASKS=5
TASK_TIMEOUT=900  # 15 minutes per task

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
BOLD='\033[1m'
NC='\033[0m'

log()  { echo -e "${BLUE}[$(date '+%H:%M:%S')]${NC} $*"; }
ok()   { echo -e "${GREEN}[OK]${NC} $*"; }
warn() { echo -e "${YELLOW}[!!]${NC} $*"; }
fail() { echo -e "${RED}[XX]${NC} $*"; }

# ── Progress Tracking ──────────────────────────────────────────────

init_progress() {
    mkdir -p "$LOG_DIR"
    if [[ ! -f "$PROGRESS_FILE" ]]; then
        python3 -c "
import json
from datetime import datetime, timezone
d = {
    'sprint': 'grade-a-followup',
    'plan': '$PLAN',
    'branch': '$BRANCH',
    'total_tasks': $TOTAL_TASKS,
    'started': datetime.now(timezone.utc).strftime('%Y-%m-%dT%H:%M:%SZ'),
    'tasks': {}
}
for i in range(1, $TOTAL_TASKS + 1):
    d['tasks'][str(i)] = {
        'status': 'pending',
        'attempts': 0,
        'started': None,
        'finished': None,
        'error': None
    }
with open('$PROGRESS_FILE', 'w') as f:
    json.dump(d, f, indent=2)
"
        log "Progress file initialized"
    fi
}

get_task_status() {
    python3 -c "
import json
with open('$PROGRESS_FILE') as f:
    d = json.load(f)
print(d.get('tasks',{}).get('$1',{}).get('status','pending'))" 2>/dev/null || echo "pending"
}

update_task_status() {
    local task_num=$1
    local status=$2
    local error="${3:-}"
    python3 -c "
import json
from datetime import datetime, timezone
with open('$PROGRESS_FILE') as f:
    d = json.load(f)
t = d.setdefault('tasks', {}).setdefault('$task_num', {})
t['status'] = '$status'
if '$status' == 'running':
    t['started'] = datetime.now(timezone.utc).strftime('%Y-%m-%dT%H:%M:%SZ')
    t.setdefault('attempts', 0)
    t['attempts'] += 1
elif '$status' == 'completed':
    t['finished'] = datetime.now(timezone.utc).strftime('%Y-%m-%dT%H:%M:%SZ')
    t['error'] = None
elif '$status' == 'failed':
    t['finished'] = datetime.now(timezone.utc).strftime('%Y-%m-%dT%H:%M:%SZ')
    t['error'] = '''$error'''
with open('$PROGRESS_FILE', 'w') as f:
    json.dump(d, f, indent=2)
"
}

# ── Pre-Flight Checks ─────────────────────────────────────────────

preflight() {
    log "Running pre-flight checks..."
    local ok_flag=true

    if command -v go >/dev/null 2>&1; then
        ok "Go $(go version | grep -oE 'go[0-9]+\.[^ ]+')"
    else
        fail "Go not found"
        ok_flag=false
    fi

    if command -v golangci-lint >/dev/null 2>&1; then
        ok "golangci-lint"
    else
        fail "golangci-lint not found"
        ok_flag=false
    fi

    if command -v claude >/dev/null 2>&1; then
        ok "Claude CLI"
    else
        fail "Claude CLI not found"
        ok_flag=false
    fi

    if [[ -f "$PROJECT_ROOT/$PLAN" ]]; then
        ok "Plan file found"
    else
        fail "Plan file not found: $PLAN"
        ok_flag=false
    fi

    # Verify build works
    if go build ./... 2>/dev/null; then
        ok "Build: clean"
    else
        fail "Build fails — fix before running sprint"
        ok_flag=false
    fi

    # Verify tests pass
    if go test ./... -race -count=1 > /dev/null 2>&1; then
        ok "Tests: all passing"
    else
        fail "Tests failing — fix before running sprint"
        ok_flag=false
    fi

    if [[ "$ok_flag" == false ]]; then
        echo ""
        fail "Pre-flight checks failed. Fix issues above first."
        exit 1
    fi

    echo ""
    ok "All pre-flight checks passed."
}

# ── Branch Setup ───────────────────────────────────────────────────

setup_branch() {
    local current_branch
    current_branch=$(git branch --show-current)

    if [[ "$current_branch" == "$BRANCH" ]]; then
        log "Already on branch $BRANCH"
        return 0
    fi

    if git show-ref --verify --quiet "refs/heads/$BRANCH"; then
        log "Switching to existing branch $BRANCH"
        git checkout "$BRANCH"
    else
        log "Creating new branch $BRANCH from current HEAD"
        git checkout -b "$BRANCH"
    fi
}

# ── Verification Gate ──────────────────────────────────────────────

verify_gate() {
    local task_num=$1
    log "Running verification gate after task $task_num..."

    local errors=0

    if go build ./... 2>/dev/null; then
        ok "Build: clean"
    else
        fail "Build failed after task $task_num"
        errors=$((errors + 1))
    fi

    if golangci-lint run ./... 2>/dev/null; then
        ok "Lint: 0 issues"
    else
        warn "Lint issues found (non-blocking)"
    fi

    if go test ./... -race -count=1 > /dev/null 2>&1; then
        ok "Tests: all passing"
    else
        fail "Tests failed after task $task_num"
        errors=$((errors + 1))
    fi

    # Auto-commit any leftover uncommitted changes
    if [[ -n "$(git status --porcelain 2>/dev/null)" ]]; then
        warn "Uncommitted changes found — auto-committing"
        git add -A
        git commit -m "chore: auto-commit remaining changes from followup task $task_num" --no-verify 2>/dev/null || true
    fi

    return "$errors"
}

# ── Task Runner ────────────────────────────────────────────────────

run_task() {
    local task_num=$1
    local task_log="$LOG_DIR/task-${task_num}.log"
    local prompt="$2"

    log "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log "TASK $task_num / $TOTAL_TASKS"
    log "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    update_task_status "$task_num" "running"

    local exit_code=0

    # Run Claude with timeout using background process + watchdog
    claude \
        -p "$prompt" \
        --permission-mode bypassPermissions \
        --allowedTools "Bash,Read,Write,Edit,Glob,Grep" \
        > "$task_log" 2>&1 &
    local claude_pid=$!

    # Watchdog: kill claude if it exceeds timeout
    (
        sleep "$TASK_TIMEOUT"
        if kill -0 "$claude_pid" 2>/dev/null; then
            kill "$claude_pid" 2>/dev/null
        fi
    ) &
    local watchdog_pid=$!

    # Wait for claude to finish
    wait "$claude_pid" 2>/dev/null || exit_code=$?

    # Clean up watchdog
    kill "$watchdog_pid" 2>/dev/null || true
    wait "$watchdog_pid" 2>/dev/null || true

    if [[ "$exit_code" -eq 143 ]]; then
        fail "Task $task_num TIMEOUT after ${TASK_TIMEOUT}s"
        return 1
    elif [[ "$exit_code" -ne 0 ]]; then
        fail "Task $task_num: Claude exited with code $exit_code"
        return 1
    fi

    ok "Task $task_num: Claude session completed"
    return 0
}

# ── Task Prompts ───────────────────────────────────────────────────

task1_prompt() {
cat << 'TASK_PROMPT'
You are executing Task 1 of the Grade-A Follow-Up Sprint: Add view tests for submit, init, and fetch paths.

Working directory: the project root.

YOUR JOB: Add the exact tests listed below to the exact files specified. Do NOT modify any production code. Only add tests.

IMPORTANT RULES:
- Do NOT remove or modify any existing tests
- Append the new test functions to the END of each file
- Add any missing imports needed for the new tests
- After writing all tests, run: go test ./internal/tui/views/ -v -race -count=1
- If any test fails, debug and fix the test (NOT the production code)
- Then run the full suite: go test ./... -race -count=1
- Commit with: git commit -m "test: add 20+ view tests covering submit, init, fetch, and validation paths"

=== FILE 1: internal/tui/views/login_test.go ===

Add this import if not present: "github.com/Akram012388/niotebook-tui/internal/tui/app"

Append these test functions:

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

=== FILE 2: internal/tui/views/register_test.go ===

Append these test functions:

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
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyTab})
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

=== FILE 3: internal/tui/views/timeline_test.go ===

Append these test functions:

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
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeySpace})
	if m.CursorIndex() == 0 {
		t.Error("expected cursor to move after space")
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'b'}})
	if m.CursorIndex() != 0 {
		t.Errorf("cursor = %d after b, want 0", m.CursorIndex())
	}
}

=== FILE 4: internal/tui/views/compose_test.go ===

Append this test function:

func TestComposeInitReturnsBlink(t *testing.T) {
	m := views.NewComposeModel(nil)
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init should return blink command")
	}
}

=== END OF TEST CODE ===

After adding ALL tests to ALL files:
1. Run: go test ./internal/tui/views/ -v -race -count=1
2. If any test fails, debug and fix ONLY the test, not the production code
3. Run: go test ./... -race -count=1 (full suite must pass)
4. Run: git add internal/tui/views/*_test.go
5. Run: git commit -m "test: add 20+ view tests covering submit, init, fetch, and validation paths"
TASK_PROMPT
}

task2_prompt() {
cat << 'TASK_PROMPT'
You are executing Task 2 of the Grade-A Follow-Up Sprint: Fix CI Go version mismatch.

Working directory: the project root.

The go.mod says `go 1.25.0` but .github/workflows/ci.yml uses `go-version: '1.22'`.
Update BOTH occurrences of go-version in .github/workflows/ci.yml from '1.22' to '1.25'.

Steps:
1. Read .github/workflows/ci.yml
2. Change BOTH go-version values from '1.22' to '1.25'
3. Verify: go build ./...
4. Commit: git add .github/workflows/ci.yml && git commit -m "chore: update CI Go version to match go.mod (1.22 → 1.25)"

Do NOT modify any other files. Do NOT work on any other task.
TASK_PROMPT
}

task3_prompt() {
cat << 'TASK_PROMPT'
You are executing Task 3 of the Grade-A Follow-Up Sprint: Remove tracked server binary from git.

Working directory: the project root.

There is a 15MB server binary at the project root that is tracked by git. Remove it from tracking and add it to .gitignore.

Steps:
1. Read .gitignore
2. Add the line `server` under the `# Binaries` section (after `bin/` and `*.exe` lines)
3. Run: git rm --cached server
4. Verify with: git ls-files server (should output nothing)
5. Commit:
   git add .gitignore
   git commit -m "chore: remove tracked server binary, add to gitignore"

Do NOT modify any other files. Do NOT work on any other task.
TASK_PROMPT
}

task4_prompt() {
cat << 'TASK_PROMPT'
You are executing Task 4 of the Grade-A Follow-Up Sprint: Update .env.example with missing variables.

Working directory: the project root.

Steps:
1. Read .env.example
2. Append these lines at the end:

NIOTEBOOK_CORS_ORIGIN=http://localhost:3000
# For testing:
# NIOTEBOOK_TEST_DB_URL=postgres://localhost/niotebook_test?sslmode=disable

3. Commit:
   git add .env.example
   git commit -m "docs: add CORS_ORIGIN and TEST_DB_URL to .env.example"

Do NOT modify any other files. Do NOT work on any other task.
TASK_PROMPT
}

task5_prompt() {
cat << 'TASK_PROMPT'
You are executing Task 5 of the Grade-A Follow-Up Sprint: Final verification of all 7 SWE pillars.

Working directory: the project root.

Run each of these verification steps and report the results:

STEP 1 — LINT:
  golangci-lint run ./...
  Expected: 0 issues

STEP 2 — BUILD:
  go build ./...
  Expected: clean exit

STEP 3 — TESTS:
  go test ./... -race -count=1 -coverprofile=coverage.out
  Expected: ALL PASS

STEP 4 — COVERAGE:
  go tool cover -func=coverage.out | grep total
  Expected: 68%+ (target was 65% minimum)

STEP 5 — SECURITY CHECKS:
  Run these grep commands to verify security requirements:
  - grep -n 'ok :=' internal/server/middleware/auth.go (verify comma-ok pattern)
  - grep -n 'len(jwtSecret)' cmd/server/main.go (verify 32-byte minimum)
  - grep -c 'if.*ok :=' internal/tui/app/app.go (verify checked type assertions — should be 12)
  - grep -n 'X-Content-Type-Options' internal/server/middleware/cors.go (verify nosniff)
  - grep -n 'X-Frame-Options' internal/server/middleware/cors.go (verify DENY)

STEP 6 — MAINTAINABILITY CHECKS:
  - git ls-files server (should be empty — binary removed from tracking)
  - grep NIOTEBOOK_CORS_ORIGIN .env.example (should show the variable)
  - grep "go-version" .github/workflows/ci.yml (should show 1.25)
  - grep XDG_CONFIG_HOME internal/tui/config/config.go (should show XDG support)

STEP 7 — PRINT SCORECARD:
  Print a final Grade-A scorecard summarizing all 7 pillars with pass/fail status.

If any check fails that you can fix (e.g., lint issue, test failure), fix it and re-verify.
If there's a fundamental issue, report it clearly.

After verification, commit any fixes if needed:
  git add -A && git commit -m "chore: final verification fixes for Grade-A sprint"
TASK_PROMPT
}

# ── Task Descriptions (for display) ───────────────────────────────

declare -a TASK_NAMES=(
    ""
    "Add 20+ view tests (submit, init, fetch, validation)"
    "Fix CI Go version mismatch (1.22 → 1.25)"
    "Remove tracked server binary from git"
    "Update .env.example with missing variables"
    "Final verification — all 7 SWE pillars"
)

# ── Main ───────────────────────────────────────────────────────────

main() {
    local resume=false
    if [[ "${1:-}" == "--resume" ]]; then
        resume=true
    fi

    echo ""
    log "╔══════════════════════════════════════════════════╗"
    log "║     GRADE-A FOLLOW-UP SPRINT RUNNER              ║"
    log "║     5 Tasks · Tests + CI + Git + Verification    ║"
    log "╚══════════════════════════════════════════════════╝"
    log ""
    log "Plan: $PLAN"
    log "Branch: $BRANCH"
    log ""

    # Pre-flight
    preflight

    # Branch setup
    setup_branch

    # Initialize or load progress
    if [[ "$resume" == true && -f "$PROGRESS_FILE" ]]; then
        log "Resuming from progress file..."
    else
        # Remove old progress for fresh start
        rm -f "$PROGRESS_FILE"
        init_progress
    fi

    echo ""
    log "Starting follow-up sprint — $TOTAL_TASKS tasks"
    echo ""

    local completed=0
    local failed=0

    for task_num in $(seq 1 $TOTAL_TASKS); do
        local status
        status=$(get_task_status "$task_num")

        # Skip already completed tasks
        if [[ "$status" == "completed" ]]; then
            ok "Task $task_num: Already completed — skipping"
            completed=$((completed + 1))
            continue
        fi

        echo ""
        log "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
        log "${BOLD}Task $task_num / $TOTAL_TASKS: ${TASK_NAMES[$task_num]}${NC}"
        log "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

        # Get the prompt for this task
        local prompt
        prompt=$(eval "task${task_num}_prompt")

        # Attempt 1
        if run_task "$task_num" "$prompt"; then
            # Verify (skip for final verification task — it IS the verification)
            if [[ $task_num -lt $TOTAL_TASKS ]]; then
                if verify_gate "$task_num"; then
                    update_task_status "$task_num" "completed"
                    ok "Task $task_num: PASSED"
                    completed=$((completed + 1))
                else
                    warn "Task $task_num: Verification failed — retrying..."
                    # Retry
                    if run_task "$task_num" "$prompt" && verify_gate "$task_num"; then
                        update_task_status "$task_num" "completed"
                        ok "Task $task_num: PASSED on retry"
                        completed=$((completed + 1))
                    else
                        update_task_status "$task_num" "failed" "verification failed after retry"
                        fail "Task $task_num: FAILED after retry"
                        failed=$((failed + 1))
                    fi
                fi
            else
                # Final verification task — just mark as completed
                update_task_status "$task_num" "completed"
                ok "Task $task_num: PASSED"
                completed=$((completed + 1))
            fi
        else
            warn "Task $task_num: Claude failed — retrying..."
            # Retry
            if run_task "$task_num" "$prompt"; then
                if [[ $task_num -lt $TOTAL_TASKS ]]; then
                    verify_gate "$task_num" || true
                fi
                update_task_status "$task_num" "completed"
                ok "Task $task_num: PASSED on retry"
                completed=$((completed + 1))
            else
                update_task_status "$task_num" "failed" "claude failed after retry"
                fail "Task $task_num: FAILED after retry"
                failed=$((failed + 1))
            fi
        fi
    done

    # ── Summary ─────────────────────────────────────────────────────
    echo ""
    log "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log "${BOLD}FOLLOW-UP SPRINT SUMMARY${NC}"
    log "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    for i in $(seq 1 $TOTAL_TASKS); do
        local status
        status=$(get_task_status "$i")
        if [[ "$status" == "completed" ]]; then
            ok "Task $i: ${TASK_NAMES[$i]}"
        elif [[ "$status" == "failed" ]]; then
            fail "Task $i: ${TASK_NAMES[$i]}"
        else
            warn "Task $i: ${TASK_NAMES[$i]} ($status)"
        fi
    done

    echo ""
    log "Completed: $completed / $TOTAL_TASKS"
    log "Failed:    $failed"
    log "Logs:      $LOG_DIR/"
    log "Progress:  $PROGRESS_FILE"
    echo ""

    if [[ $completed -eq $TOTAL_TASKS ]]; then
        echo ""
        ok "${BOLD}ALL TASKS COMPLETE — Grade-A follow-up sprint finished!${NC}"
        echo ""
        log "Next steps:"
        log "  1. Review changes: git log --oneline grade-a-sprint"
        log "  2. Check coverage: go test ./... -coverprofile=c.out && go tool cover -func=c.out | grep total"
        log "  3. Merge to dev:   git checkout dev && git merge grade-a-sprint"
        echo ""
        return 0
    else
        fail "Sprint incomplete. $completed/$TOTAL_TASKS tasks done. Check logs."
        return 1
    fi
}

main "$@"
