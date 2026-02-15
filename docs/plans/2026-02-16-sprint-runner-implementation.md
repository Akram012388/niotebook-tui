# Sprint Runner Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build `scripts/sprint-bootstrap.sh` and `scripts/sprint.sh` — an automated overnight sprint runner that executes the 24-task Niotebook MVP implementation plan using non-interactive Claude Code sessions.

**Architecture:** Two bash scripts. `sprint-bootstrap.sh` installs system dependencies (PostgreSQL, migrate CLI, golangci-lint) idempotently. `sprint.sh` orchestrates sequential Claude sessions — one per task — with progress tracking via JSON, per-task logging, retry-on-failure, dependency-aware skipping, and full resume capability.

**Tech Stack:** Bash, `claude` CLI (`-p` print mode, `--permission-mode bypassPermissions`), `jq` for JSON progress file, PostgreSQL 15 via Homebrew

---

## Task 1: Create sprint-bootstrap.sh

**Files:**
- Create: `scripts/sprint-bootstrap.sh`

**Step 1: Create the scripts directory**

```bash
mkdir -p scripts
```

**Step 2: Write sprint-bootstrap.sh**

Create `scripts/sprint-bootstrap.sh` with this exact content:

```bash
#!/usr/bin/env bash
set -euo pipefail

# Niotebook MVP Sprint — Bootstrap Dependencies
# Idempotent: safe to re-run. Installs system-level tools only.

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log()  { echo -e "${GREEN}[bootstrap]${NC} $*"; }
warn() { echo -e "${YELLOW}[bootstrap]${NC} $*"; }
err()  { echo -e "${RED}[bootstrap]${NC} $*" >&2; }
die()  { err "$@"; exit 1; }

# ── Go ──────────────────────────────────────────────────────────────
check_go() {
  log "Checking Go..."
  command -v go >/dev/null 2>&1 || die "Go is not installed. Install Go 1.22+ first."
  local ver
  ver=$(go version | grep -oE '[0-9]+\.[0-9]+' | head -1)
  local major minor
  major=$(echo "$ver" | cut -d. -f1)
  minor=$(echo "$ver" | cut -d. -f2)
  if (( major < 1 || (major == 1 && minor < 22) )); then
    die "Go $ver found, need >= 1.22"
  fi
  log "Go $(go version | grep -oE 'go[0-9]+\.[^ ]+') ✓"
}

# ── Homebrew ────────────────────────────────────────────────────────
check_brew() {
  command -v brew >/dev/null 2>&1 || die "Homebrew is not installed. Install from https://brew.sh"
  log "Homebrew ✓"
}

# ── PostgreSQL ──────────────────────────────────────────────────────
install_postgres() {
  log "Checking PostgreSQL..."
  if command -v psql >/dev/null 2>&1; then
    log "PostgreSQL already installed ✓"
  else
    log "Installing PostgreSQL 15 via Homebrew..."
    brew install postgresql@15
    # Add to PATH for this session
    export PATH="/opt/homebrew/opt/postgresql@15/bin:$PATH"
  fi

  # Start service if not running
  if ! pg_isready -q 2>/dev/null; then
    log "Starting PostgreSQL service..."
    brew services start postgresql@15 2>/dev/null || true
    # Wait for PostgreSQL to be ready
    local retries=10
    while ! pg_isready -q 2>/dev/null && (( retries-- > 0 )); do
      sleep 1
    done
    pg_isready -q 2>/dev/null || die "PostgreSQL failed to start"
  fi
  log "PostgreSQL running ✓"

  # Create databases if they don't exist
  local user
  user=$(whoami)
  if ! psql -lqt 2>/dev/null | cut -d\| -f1 | grep -qw niotebook_dev; then
    log "Creating niotebook_dev database..."
    createdb niotebook_dev 2>/dev/null || true
  fi
  if ! psql -lqt 2>/dev/null | cut -d\| -f1 | grep -qw niotebook_test; then
    log "Creating niotebook_test database..."
    createdb niotebook_test 2>/dev/null || true
  fi
  log "Databases niotebook_dev + niotebook_test ✓"
}

# ── golang-migrate CLI ──────────────────────────────────────────────
install_migrate() {
  log "Checking golang-migrate..."
  if command -v migrate >/dev/null 2>&1; then
    log "golang-migrate already installed ✓"
  else
    log "Installing golang-migrate via Homebrew..."
    brew install golang-migrate
  fi
  log "golang-migrate ✓"
}

# ── golangci-lint ───────────────────────────────────────────────────
install_golangci_lint() {
  log "Checking golangci-lint..."
  if command -v golangci-lint >/dev/null 2>&1; then
    log "golangci-lint already installed ✓"
  else
    log "Installing golangci-lint via Homebrew..."
    brew install golangci-lint
  fi
  log "golangci-lint ✓"
}

# ── jq (needed by sprint.sh for progress tracking) ─────────────────
install_jq() {
  log "Checking jq..."
  if command -v jq >/dev/null 2>&1; then
    log "jq already installed ✓"
  else
    log "Installing jq via Homebrew..."
    brew install jq
  fi
  log "jq ✓"
}

# ── Claude CLI ──────────────────────────────────────────────────────
check_claude() {
  log "Checking Claude CLI..."
  command -v claude >/dev/null 2>&1 || die "Claude CLI not found. Install from https://claude.ai/claude-code"
  log "Claude CLI $(claude --version 2>/dev/null || echo 'unknown') ✓"
}

# ── .env file ───────────────────────────────────────────────────────
setup_env() {
  local env_file="$PROJECT_DIR/.env"
  if [[ -f "$env_file" ]]; then
    log ".env already exists, skipping ✓"
    return
  fi

  log "Generating .env..."
  local jwt_secret
  jwt_secret=$(openssl rand -hex 32)
  local user
  user=$(whoami)

  cat > "$env_file" <<EOF
NIOTEBOOK_DB_URL=postgres://${user}@localhost/niotebook_dev?sslmode=disable
NIOTEBOOK_JWT_SECRET=${jwt_secret}
NIOTEBOOK_PORT=8080
NIOTEBOOK_HOST=localhost
NIOTEBOOK_LOG_LEVEL=debug
EOF

  log ".env created ✓"
}

# ── Main ────────────────────────────────────────────────────────────
main() {
  echo ""
  echo "=== Niotebook MVP Sprint — Bootstrap ==="
  echo ""

  check_go
  check_brew
  install_postgres
  install_migrate
  install_golangci_lint
  install_jq
  check_claude
  setup_env

  echo ""
  log "All dependencies ready. Run ./scripts/sprint.sh to start the sprint."
  echo ""
}

main "$@"
```

**Step 3: Make executable**

```bash
chmod +x scripts/sprint-bootstrap.sh
```

**Step 4: Verify it parses**

Run: `bash -n scripts/sprint-bootstrap.sh`
Expected: No output (no syntax errors)

**Step 5: Commit**

```bash
git add scripts/sprint-bootstrap.sh
git commit -m "feat: add sprint bootstrap script for dependency installation"
```

---

## Task 2: Create sprint.sh — Configuration and Utility Functions

**Files:**
- Create: `scripts/sprint.sh`

**Step 1: Write the top section of sprint.sh (configuration + utilities + dependency graph)**

Create `scripts/sprint.sh` with this exact content:

```bash
#!/usr/bin/env bash
set -euo pipefail

# Niotebook MVP Sprint Runner
# Executes the 24-task MVP implementation plan using non-interactive Claude Code sessions.
# Usage:
#   ./scripts/sprint.sh            # Fresh start
#   ./scripts/sprint.sh --resume   # Resume from progress file

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
PLAN_FILE="docs/vault/04-plans/2026-02-16-mvp-implementation.md"
SPRINT_DATE=$(date +%Y-%m-%d)
LOG_DIR="$PROJECT_DIR/logs/sprint-$SPRINT_DATE"
PROGRESS_FILE="$SCRIPT_DIR/sprint-progress.json"
BRANCH_NAME="mvp-sprint"
TOTAL_TASKS=24
MAX_BUDGET_PER_TASK=1.00  # USD safety limit per claude session
TASK_TIMEOUT=900           # 15 minutes in seconds

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

log()      { echo -e "${GREEN}[sprint]${NC} $(date +%H:%M:%S) $*"; }
warn()     { echo -e "${YELLOW}[sprint]${NC} $(date +%H:%M:%S) $*"; }
err()      { echo -e "${RED}[sprint]${NC} $(date +%H:%M:%S) $*" >&2; }
task_log() { echo -e "${CYAN}[task $1]${NC} $(date +%H:%M:%S) ${*:2}"; }

# ── Dependency Graph ────────────────────────────────────────────────
# Each task maps to a space-separated list of tasks it depends on.
# Derived from docs/vault/04-plans/2026-02-16-mvp-implementation.md "Dependency Graph"
declare -A TASK_DEPS=(
  [1]=""
  [2]="1"
  [3]="1"
  [4]="1 2"
  [5]="3 4"
  [6]="2"
  [7]="6"
  [8]="7"
  [9]="7"
  [10]="1"
  [11]="1"
  [12]="7 8 9 10 11"
  [13]="12"
  [14]="13"
  [15]="1"
  [16]="15 2"
  [17]="2"
  [18]="2"
  [19]="16 18"
  [20]="19 17"
  [21]="20"
  [22]="20"
  [23]="19 20 21 22"
  [24]="14 23"
)

# Task names for display
declare -A TASK_NAMES=(
  [1]="Initialize Go Module and Directory Structure"
  [2]="Shared Domain Models and Build Package"
  [3]="Database Migrations"
  [4]="Database Connection and Store Interfaces"
  [5]="Store Implementations"
  [6]="Validation Functions"
  [7]="Auth Service"
  [8]="Post Service"
  [9]="User Service"
  [10]="JWT Auth Middleware"
  [11]="Supporting Middleware"
  [12]="HTTP Handlers"
  [13]="Server Router and Wiring"
  [14]="Server Binary Entry Point"
  [15]="Config Loading"
  [16]="HTTP Client Wrapper"
  [17]="TUI Components — Relative Time and Post Card"
  [18]="TUI Message Types"
  [19]="Login and Register Views"
  [20]="Timeline View"
  [21]="Compose Modal"
  [22]="Profile View and Help Overlay"
  [23]="Root AppModel and TUI Binary"
  [24]="GitHub Actions CI and Final Verification"
)

# ── Progress File Helpers ───────────────────────────────────────────
init_progress() {
  if [[ -f "$PROGRESS_FILE" ]]; then
    return
  fi
  local tasks="{}"
  for i in $(seq 1 $TOTAL_TASKS); do
    tasks=$(echo "$tasks" | jq --arg i "$i" '. + {($i): {"status": "pending", "attempts": 0, "started": null, "finished": null, "error": null}}')
  done
  jq -n \
    --arg sid "$SPRINT_DATE" \
    --arg branch "$BRANCH_NAME" \
    --arg started "$(date -u +%Y-%m-%dT%H:%M:%SZ)" \
    --argjson tasks "$tasks" \
    '{sprint_id: $sid, branch: $branch, started_at: $started, tasks: $tasks}' \
    > "$PROGRESS_FILE"
  log "Progress file initialized"
}

get_task_status() {
  jq -r --arg t "$1" '.tasks[$t].status' "$PROGRESS_FILE"
}

get_task_attempts() {
  jq -r --arg t "$1" '.tasks[$t].attempts' "$PROGRESS_FILE"
}

set_task_status() {
  local task_num=$1 status=$2
  local now
  now=$(date -u +%Y-%m-%dT%H:%M:%SZ)
  local tmp
  tmp=$(mktemp)
  jq --arg t "$task_num" --arg s "$status" --arg now "$now" \
    '.tasks[$t].status = $s | .tasks[$t].finished = $now' \
    "$PROGRESS_FILE" > "$tmp" && mv "$tmp" "$PROGRESS_FILE"
}

set_task_started() {
  local task_num=$1
  local now
  now=$(date -u +%Y-%m-%dT%H:%M:%SZ)
  local tmp
  tmp=$(mktemp)
  jq --arg t "$task_num" --arg now "$now" \
    '.tasks[$t].started = $now | .tasks[$t].attempts += 1 | .tasks[$t].status = "running"' \
    "$PROGRESS_FILE" > "$tmp" && mv "$tmp" "$PROGRESS_FILE"
}

set_task_error() {
  local task_num=$1 error_msg=$2
  local tmp
  tmp=$(mktemp)
  jq --arg t "$task_num" --arg e "$error_msg" \
    '.tasks[$t].error = $e' \
    "$PROGRESS_FILE" > "$tmp" && mv "$tmp" "$PROGRESS_FILE"
}

# ── Dependency Check ────────────────────────────────────────────────
# Returns 0 if all dependencies are completed, 1 if any are failed/skipped
check_deps() {
  local task_num=$1
  local deps="${TASK_DEPS[$task_num]}"
  if [[ -z "$deps" ]]; then
    return 0
  fi
  for dep in $deps; do
    local dep_status
    dep_status=$(get_task_status "$dep")
    if [[ "$dep_status" == "failed" || "$dep_status" == "skipped" ]]; then
      return 1
    fi
    if [[ "$dep_status" != "completed" ]]; then
      return 1
    fi
  done
  return 0
}

# Mark all transitive dependents of a failed task as "skipped"
skip_dependents() {
  local failed_task=$1
  for i in $(seq 1 $TOTAL_TASKS); do
    local deps="${TASK_DEPS[$i]}"
    for dep in $deps; do
      if [[ "$dep" == "$failed_task" ]]; then
        local status
        status=$(get_task_status "$i")
        if [[ "$status" == "pending" ]]; then
          set_task_status "$i" "skipped"
          set_task_error "$i" "blocked by failed task $failed_task"
          task_log "$i" "Skipped (blocked by task $failed_task)"
          # Recursively skip dependents of this task too
          skip_dependents "$i"
        fi
        break
      fi
    done
  done
}
```

**Step 2: Make executable**

```bash
chmod +x scripts/sprint.sh
```

**Step 3: Verify it parses**

Run: `bash -n scripts/sprint.sh`
Expected: No output (no syntax errors)

**Step 4: Commit**

```bash
git add scripts/sprint.sh
git commit -m "feat: add sprint.sh with config, dependency graph, and progress tracking"
```

---

## Task 3: Add Pre-Flight Checks and Verification Functions to sprint.sh

**Files:**
- Modify: `scripts/sprint.sh` (append after the existing content)

**Step 1: Append the pre-flight, verification, and Claude invocation functions**

Append this content to the end of `scripts/sprint.sh`:

```bash

# ── Pre-Flight Checks ──────────────────────────────────────────────
preflight() {
  echo ""
  echo -e "${BOLD}=== Niotebook MVP Sprint Runner ===${NC}"
  echo ""
  log "Running pre-flight checks..."
  local ok=true

  # Go
  if command -v go >/dev/null 2>&1; then
    log "  [✓] Go $(go version | grep -oE 'go[0-9]+\.[^ ]+')"
  else
    err "  [✗] Go not found"
    ok=false
  fi

  # PostgreSQL
  if pg_isready -q 2>/dev/null; then
    log "  [✓] PostgreSQL running"
  else
    err "  [✗] PostgreSQL not running — run ./scripts/sprint-bootstrap.sh first"
    ok=false
  fi

  # Databases
  local user
  user=$(whoami)
  if psql -lqt 2>/dev/null | cut -d\| -f1 | grep -qw niotebook_dev; then
    log "  [✓] niotebook_dev database"
  else
    err "  [✗] niotebook_dev database missing"
    ok=false
  fi
  if psql -lqt 2>/dev/null | cut -d\| -f1 | grep -qw niotebook_test; then
    log "  [✓] niotebook_test database"
  else
    err "  [✗] niotebook_test database missing"
    ok=false
  fi

  # Tools
  for tool in migrate golangci-lint jq claude; do
    if command -v "$tool" >/dev/null 2>&1; then
      log "  [✓] $tool"
    else
      err "  [✗] $tool not found"
      ok=false
    fi
  done

  # Disk space (warn if < 5GB)
  local free_gb
  free_gb=$(df -g "$PROJECT_DIR" | tail -1 | awk '{print $4}')
  if (( free_gb >= 5 )); then
    log "  [✓] Disk space: ${free_gb}GB free"
  else
    warn "  [⚠] Low disk space: ${free_gb}GB free (< 5GB recommended)"
  fi

  # Plan file
  if [[ -f "$PROJECT_DIR/$PLAN_FILE" ]]; then
    log "  [✓] Implementation plan found"
  else
    err "  [✗] Implementation plan not found at $PLAN_FILE"
    ok=false
  fi

  if [[ "$ok" == false ]]; then
    echo ""
    die "Pre-flight checks failed. Fix issues above or run ./scripts/sprint-bootstrap.sh"
  fi

  echo ""
  log "All pre-flight checks passed."
}

# ── Git Branch Setup ────────────────────────────────────────────────
setup_branch() {
  cd "$PROJECT_DIR"
  local current_branch
  current_branch=$(git branch --show-current)

  if [[ "$current_branch" == "$BRANCH_NAME" ]]; then
    log "Already on branch $BRANCH_NAME"
    return
  fi

  if git show-ref --verify --quiet "refs/heads/$BRANCH_NAME" 2>/dev/null; then
    log "Switching to existing branch $BRANCH_NAME"
    git checkout "$BRANCH_NAME"
  else
    log "Creating branch $BRANCH_NAME from main"
    git checkout -b "$BRANCH_NAME"
  fi
}

# ── Post-Task Verification ─────────────────────────────────────────
verify_task() {
  local task_num=$1
  local verify_log="$LOG_DIR/task-$(printf '%02d' "$task_num")-verify.log"
  local errors=0

  task_log "$task_num" "Running verification..."

  # 1. Code compiles
  if go build ./... >> "$verify_log" 2>&1; then
    task_log "$task_num" "  Build: PASS"
  else
    task_log "$task_num" "  Build: FAIL"
    ((errors++))
  fi

  # 2. Tests pass (skip for task 1 which has no tests)
  if [[ "$task_num" -gt 1 ]]; then
    if go test ./... -v -race -timeout 120s >> "$verify_log" 2>&1; then
      task_log "$task_num" "  Tests: PASS"
    else
      task_log "$task_num" "  Tests: FAIL"
      ((errors++))
    fi
  else
    task_log "$task_num" "  Tests: SKIPPED (task 1)"
  fi

  # 3. Auto-commit any leftover uncommitted changes
  if [[ -n "$(git status --porcelain 2>/dev/null)" ]]; then
    task_log "$task_num" "  Auto-committing remaining changes..."
    git add -A
    git commit -m "chore: auto-commit remaining changes from task $task_num" --no-verify 2>/dev/null || true
  fi

  return "$errors"
}

# ── Run Claude Session for a Task ──────────────────────────────────
run_task() {
  local task_num=$1
  local task_log_file="$LOG_DIR/task-$(printf '%02d' "$task_num").log"
  local task_name="${TASK_NAMES[$task_num]}"

  task_log "$task_num" "Starting: ${BOLD}$task_name${NC}"
  set_task_started "$task_num"

  # Build the prompt
  local prompt
  prompt="You are implementing the Niotebook MVP.

Your working directory is $(pwd).

Read the implementation plan at $PLAN_FILE.

Execute ONLY Task $task_num: \"$task_name\" — follow every step exactly as written in the plan:
- Create/modify the exact files specified
- Run the exact commands specified
- Verify the exact expected outputs
- Commit with the exact commit message specified

IMPORTANT:
- Do NOT work on any other task besides Task $task_num.
- Do NOT skip any steps.
- Do NOT modify the plan's instructions.
- If a test fails, debug and fix it before moving on.
- After completing all steps, run: go build ./... && go test ./... -v -race
- If the final build or tests fail, fix the issues.

The .env file is at the project root with database connection details.
Source it if needed: source .env"

  # Check if PostgreSQL is needed (tasks 3+)
  if [[ "$task_num" -ge 3 ]]; then
    if ! pg_isready -q 2>/dev/null; then
      task_log "$task_num" "PostgreSQL not responding, attempting restart..."
      brew services restart postgresql@15 2>/dev/null || true
      sleep 3
      if ! pg_isready -q 2>/dev/null; then
        err "PostgreSQL is down and could not be restarted"
        return 1
      fi
    fi
  fi

  # Run Claude in non-interactive mode with timeout
  local exit_code=0
  CLAUDECODE= timeout "$TASK_TIMEOUT" claude \
    -p "$prompt" \
    --permission-mode "bypassPermissions" \
    --allowedTools "Bash,Read,Write,Edit,Glob,Grep" \
    --max-budget-usd "$MAX_BUDGET_PER_TASK" \
    --no-session-persistence \
    > "$task_log_file" 2>&1 || exit_code=$?

  if [[ "$exit_code" -eq 124 ]]; then
    task_log "$task_num" "TIMEOUT after ${TASK_TIMEOUT}s"
    return 1
  elif [[ "$exit_code" -ne 0 ]]; then
    task_log "$task_num" "Claude exited with code $exit_code"
    return 1
  fi

  task_log "$task_num" "Claude session completed"
  return 0
}
```

**Step 2: Verify it parses**

Run: `bash -n scripts/sprint.sh`
Expected: No output (no syntax errors)

**Step 3: Commit**

```bash
git add scripts/sprint.sh
git commit -m "feat: add pre-flight checks, verification, and claude invocation to sprint.sh"
```

---

## Task 4: Add Main Loop and Summary Generation to sprint.sh

**Files:**
- Modify: `scripts/sprint.sh` (append after existing content)

**Step 1: Append the main loop and summary functions**

Append this content to the end of `scripts/sprint.sh`:

```bash

# ── Summary Generation ──────────────────────────────────────────────
generate_summary() {
  local summary_file="$LOG_DIR/summary.md"
  local completed=0 failed=0 skipped=0 pending=0

  for i in $(seq 1 $TOTAL_TASKS); do
    local status
    status=$(get_task_status "$i")
    case "$status" in
      completed) ((completed++)) ;;
      failed)    ((failed++)) ;;
      skipped)   ((skipped++)) ;;
      *)         ((pending++)) ;;
    esac
  done

  local finished_at
  finished_at=$(date -u +%Y-%m-%dT%H:%M:%SZ)
  local started_at
  started_at=$(jq -r '.started_at' "$PROGRESS_FILE")

  cat > "$summary_file" <<SUMMARY
# Sprint Summary — $SPRINT_DATE

| Metric | Value |
|--------|-------|
| Branch | $BRANCH_NAME |
| Started | $started_at |
| Finished | $finished_at |
| Completed | $completed / $TOTAL_TASKS |
| Failed | $failed |
| Skipped | $skipped |
| Pending | $pending |

## Task Results

| Task | Name | Status | Attempts |
|------|------|--------|----------|
SUMMARY

  for i in $(seq 1 $TOTAL_TASKS); do
    local status attempts name
    status=$(get_task_status "$i")
    attempts=$(get_task_attempts "$i")
    name="${TASK_NAMES[$i]}"
    local icon
    case "$status" in
      completed) icon="✅" ;;
      failed)    icon="❌" ;;
      skipped)   icon="⏭️" ;;
      *)         icon="⬜" ;;
    esac
    echo "| $i | $name | $icon $status | $attempts |" >> "$summary_file"
  done

  # Add failed task details
  if (( failed > 0 )); then
    echo "" >> "$summary_file"
    echo "## Failed Tasks" >> "$summary_file"
    echo "" >> "$summary_file"
    for i in $(seq 1 $TOTAL_TASKS); do
      local status
      status=$(get_task_status "$i")
      if [[ "$status" == "failed" ]]; then
        local error
        error=$(jq -r --arg t "$i" '.tasks[$t].error // "unknown"' "$PROGRESS_FILE")
        echo "### Task $i: ${TASK_NAMES[$i]}" >> "$summary_file"
        echo "" >> "$summary_file"
        echo "Error: $error" >> "$summary_file"
        echo "" >> "$summary_file"
        echo "Log: \`logs/sprint-$SPRINT_DATE/task-$(printf '%02d' "$i").log\`" >> "$summary_file"
        echo "" >> "$summary_file"
      fi
    done
  fi

  log "Summary written to $summary_file"
}

# ── Print Final Report ──────────────────────────────────────────────
print_report() {
  local completed=0 failed=0 skipped=0

  for i in $(seq 1 $TOTAL_TASKS); do
    local status
    status=$(get_task_status "$i")
    case "$status" in
      completed) ((completed++)) ;;
      failed)    ((failed++)) ;;
      skipped)   ((skipped++)) ;;
    esac
  done

  echo ""
  echo -e "${BOLD}=== Sprint Complete ===${NC}"
  echo ""
  echo -e "  Completed: ${GREEN}$completed${NC} / $TOTAL_TASKS"
  echo -e "  Failed:    ${RED}$failed${NC}"
  echo -e "  Skipped:   ${YELLOW}$skipped${NC}"
  echo ""
  echo -e "  Progress:  $PROGRESS_FILE"
  echo -e "  Logs:      $LOG_DIR/"
  echo -e "  Summary:   $LOG_DIR/summary.md"
  echo -e "  Branch:    $BRANCH_NAME"
  echo ""

  if (( failed == 0 && skipped == 0 )); then
    echo -e "  ${GREEN}${BOLD}All $TOTAL_TASKS tasks completed successfully!${NC}"
    echo -e "  Next: review the branch and merge when ready."
  else
    echo -e "  ${YELLOW}Review failed/skipped tasks in the summary and logs.${NC}"
  fi
  echo ""
}

# ── Main ────────────────────────────────────────────────────────────
main() {
  local resume=false
  if [[ "${1:-}" == "--resume" ]]; then
    resume=true
  fi

  cd "$PROJECT_DIR"

  # Create log directory
  mkdir -p "$LOG_DIR"

  # Pre-flight
  preflight

  # Branch setup
  setup_branch

  # Initialize or load progress
  if [[ "$resume" == true && -f "$PROGRESS_FILE" ]]; then
    log "Resuming sprint from progress file..."
    # Reset failed tasks to pending for retry
    for i in $(seq 1 $TOTAL_TASKS); do
      local status
      status=$(get_task_status "$i")
      if [[ "$status" == "failed" ]]; then
        local tmp
        tmp=$(mktemp)
        jq --arg t "$i" \
          '.tasks[$t].status = "pending" | .tasks[$t].attempts = 0 | .tasks[$t].error = null' \
          "$PROGRESS_FILE" > "$tmp" && mv "$tmp" "$PROGRESS_FILE"
        log "Reset task $i from failed to pending"
      fi
    done
  else
    init_progress
  fi

  echo ""
  log "Starting sprint — $TOTAL_TASKS tasks on branch $BRANCH_NAME"
  echo ""

  # Main loop
  for task_num in $(seq 1 $TOTAL_TASKS); do
    local status
    status=$(get_task_status "$task_num")
    local task_name="${TASK_NAMES[$task_num]}"

    # Skip completed
    if [[ "$status" == "completed" ]]; then
      task_log "$task_num" "Already completed, skipping"
      continue
    fi

    # Skip already-skipped
    if [[ "$status" == "skipped" ]]; then
      task_log "$task_num" "Skipped (blocked by dependency)"
      continue
    fi

    # Check dependencies
    if ! check_deps "$task_num"; then
      set_task_status "$task_num" "skipped"
      set_task_error "$task_num" "dependency not met"
      task_log "$task_num" "Skipped — dependency not met"
      continue
    fi

    # Attempt 1
    echo ""
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    task_log "$task_num" "${BOLD}Task $task_num / $TOTAL_TASKS: $task_name${NC}"
    echo -e "${BLUE}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"

    if run_task "$task_num" && verify_task "$task_num"; then
      set_task_status "$task_num" "completed"
      task_log "$task_num" "${GREEN}PASSED${NC}"
    else
      # Attempt 2 (retry)
      task_log "$task_num" "${YELLOW}FAILED — retrying...${NC}"

      if run_task "$task_num" && verify_task "$task_num"; then
        set_task_status "$task_num" "completed"
        task_log "$task_num" "${GREEN}PASSED on retry${NC}"
      else
        # Mark as failed, skip dependents
        set_task_status "$task_num" "failed"
        set_task_error "$task_num" "failed after 2 attempts — check logs/sprint-$SPRINT_DATE/task-$(printf '%02d' "$task_num").log"
        task_log "$task_num" "${RED}FAILED after 2 attempts${NC}"
        skip_dependents "$task_num"
      fi
    fi
  done

  # Generate summary
  generate_summary
  print_report
}

main "$@"
```

**Step 2: Verify it parses**

Run: `bash -n scripts/sprint.sh`
Expected: No output (no syntax errors)

**Step 3: Commit**

```bash
git add scripts/sprint.sh
git commit -m "feat: add main loop, summary generation, and resume to sprint.sh"
```

---

## Task 5: Add logs/ to .gitignore and Final Verification

**Files:**
- Modify: `.gitignore`

**Step 1: Add logs directory and progress file to .gitignore**

Append to `.gitignore`:

```
# Sprint runner
logs/
scripts/sprint-progress.json
```

**Step 2: Verify both scripts parse cleanly**

Run: `bash -n scripts/sprint-bootstrap.sh && bash -n scripts/sprint.sh && echo "Both scripts OK"`
Expected: `Both scripts OK`

**Step 3: Verify sprint.sh --help-like behavior**

Run: `bash -c 'source scripts/sprint.sh; echo "${TASK_NAMES[1]}"; echo "${TASK_DEPS[24]}"'`

This won't work because `main` runs on source — instead just verify the script doesn't crash with a quick parse check:

Run: `bash -n scripts/sprint.sh && echo "OK"`
Expected: `OK`

**Step 4: Commit**

```bash
git add .gitignore scripts/sprint-bootstrap.sh scripts/sprint.sh
git commit -m "chore: finalize sprint runner scripts, add logs to gitignore"
```

---

## Execution Notes

### Running the Sprint

```bash
# Step 1: Bootstrap dependencies (run once, interactively)
./scripts/sprint-bootstrap.sh

# Step 2: Verify bootstrap worked
pg_isready && echo "PostgreSQL OK"
migrate -version && echo "migrate OK"
golangci-lint version && echo "lint OK"

# Step 3: Start the overnight sprint
nohup ./scripts/sprint.sh > logs/sprint-$(date +%Y-%m-%d)/runner.log 2>&1 &
echo $! > scripts/sprint.pid
echo "Sprint started as PID $(cat scripts/sprint.pid)"

# Or, if you prefer to watch:
./scripts/sprint.sh | tee logs/sprint-$(date +%Y-%m-%d)/runner.log
```

### Checking Progress (from another terminal)

```bash
# Quick status
jq '.tasks | to_entries | map({task: .key, status: .value.status}) | sort_by(.task | tonumber)' scripts/sprint-progress.json

# Count by status
jq '.tasks | to_entries | group_by(.value.status) | map({status: .[0].value.status, count: length})' scripts/sprint-progress.json

# View latest task log
ls -t logs/sprint-*/task-*.log | head -1 | xargs tail -20
```

### Resuming After Interruption

```bash
./scripts/sprint.sh --resume
```
