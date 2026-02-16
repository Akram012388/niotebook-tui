#!/opt/homebrew/bin/bash
set -euo pipefail

# Niotebook Verification Fixes Sprint Runner
# Fixes all critical, high-priority, and low-priority issues from the codebase audit.
# Usage:
#   ./scripts/verification-fixes-sprint.sh            # Fresh start
#   ./scripts/verification-fixes-sprint.sh --resume   # Resume from progress file

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Ensure Homebrew PostgreSQL is on PATH (keg-only on macOS)
export PATH="/opt/homebrew/opt/postgresql@15/bin:$PATH"

PLAN_FILE="docs/plans/2026-02-16-verification-fixes.md"
SPRINT_DATE=$(date +%Y-%m-%d)
LOG_DIR="$PROJECT_DIR/logs/verification-fixes-$SPRINT_DATE"
PROGRESS_FILE="$SCRIPT_DIR/verification-fixes-progress.json"
BRANCH_NAME="verification-fixes"
BASE_BRANCH="grade-a-sprint"
TOTAL_TASKS=11
TASK_TIMEOUT=900           # 15 minutes in seconds

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

log()      { echo -e "${GREEN}[vfix]${NC} $(date +%H:%M:%S) $*"; }
warn()     { echo -e "${YELLOW}[vfix]${NC} $(date +%H:%M:%S) $*"; }
err()      { echo -e "${RED}[vfix]${NC} $(date +%H:%M:%S) $*" >&2; }
task_log() { echo -e "${CYAN}[task $1]${NC} $(date +%H:%M:%S) ${*:2}"; }

# ── Dependency Graph ────────────────────────────────────────────────
# Tasks 1-10 are all independent. Task 11 depends on all.
declare -A TASK_DEPS=(
  [1]=""
  [2]=""
  [3]=""
  [4]=""
  [5]=""
  [6]=""
  [7]=""
  [8]=""
  [9]=""
  [10]=""
  [11]="1 2 3 4 5 6 7 8 9 10"
)

# Task names for display
declare -A TASK_NAMES=(
  [1]="Fix compose publish keybinding (Ctrl+E → spec keybinding)"
  [2]="Remove CORS wildcard default"
  [3]="Add missing handler tests for 80% coverage"
  [4]="Add missing service, app, and config tests"
  [5]="Fix CI Go version and raise coverage threshold"
  [6]="Add XSS/control character validation for display_name and bio"
  [7]="Use JWT library error types instead of string matching"
  [8]="Stop rate limiter background goroutine on shutdown"
  [9]="Fix MaxBytesReader to pass ResponseWriter"
  [10]="Fix double WriteHeader in health handler"
  [11]="Final verification — build, vet, test, coverage"
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
          skip_dependents "$i"
        fi
        break
      fi
    done
  done
}

# ── Pre-Flight Checks ──────────────────────────────────────────────
preflight() {
  echo ""
  echo -e "${BOLD}=== Niotebook Verification Fixes Sprint Runner ===${NC}"
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
  if psql -lqt 2>/dev/null | cut -d\| -f1 | grep -qw niotebook_test; then
    log "  [✓] niotebook_test database"
  else
    err "  [✗] niotebook_test database missing"
    ok=false
  fi

  # Tools
  for tool in jq claude; do
    if command -v "$tool" >/dev/null 2>&1; then
      log "  [✓] $tool"
    else
      err "  [✗] $tool not found"
      ok=false
    fi
  done

  # Plan file
  if [[ -f "$PROJECT_DIR/$PLAN_FILE" ]]; then
    log "  [✓] Implementation plan found"
  else
    err "  [✗] Implementation plan not found at $PLAN_FILE"
    ok=false
  fi

  # Verify codebase compiles
  if go build ./... 2>/dev/null; then
    log "  [✓] Codebase compiles"
  else
    err "  [✗] Codebase does not compile — fix build errors first"
    ok=false
  fi

  if [[ "$ok" == false ]]; then
    echo ""
    err "Pre-flight checks failed. Fix issues above before running."
    exit 1
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
    log "Creating branch $BRANCH_NAME from $BASE_BRANCH"
    git checkout -b "$BRANCH_NAME" "$BASE_BRANCH"
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

  # 2. Tests pass
  if go test ./... -v -race -timeout 300s >> "$verify_log" 2>&1; then
    task_log "$task_num" "  Tests: PASS"
  else
    task_log "$task_num" "  Tests: FAIL"
    ((errors++))
  fi

  # 3. Auto-commit any leftover uncommitted changes
  if [[ -n "$(git status --porcelain 2>/dev/null)" ]]; then
    task_log "$task_num" "  Auto-committing remaining changes..."
    git add -A
    git commit -m "chore: auto-commit remaining changes from verification task $task_num" --no-verify 2>/dev/null || true
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
  prompt="You are fixing issues identified in a rigorous codebase verification audit of the Niotebook project.

Your working directory is $(pwd).

Read the implementation plan at $PLAN_FILE.

Execute ONLY Task $task_num: \"$task_name\" — follow every step exactly as written in the plan:
- Read the files mentioned in the task first
- Create/modify the exact files specified
- Run the exact commands specified
- Verify the exact expected outputs
- Commit with the exact commit message specified

IMPORTANT:
- Do NOT work on any other task besides Task $task_num.
- Do NOT skip any steps.
- Do NOT modify the plan's instructions unless you discover the plan has a technical error (e.g., wrong key constant name). In that case, use the closest correct alternative and document what you did.
- If a test fails, debug and fix it before moving on.
- After completing all steps, run: go build ./... && go test ./... -v -race -timeout 300s
- If the final build or tests fail, fix the issues.

Reference files in docs/vault/ for specs:
- docs/vault/03-design/keybindings.md — keybinding specifications
- docs/vault/02-engineering/adr/ADR-0020 — compose behavior
- docs/vault/04-plans/2026-02-16-mvp-implementation.md — original MVP plan
- CLAUDE.md — project conventions

The .env file is at the project root with database connection details.
Source it if needed: source .env"

  # Check PostgreSQL is alive for tests
  if ! pg_isready -q 2>/dev/null; then
    task_log "$task_num" "PostgreSQL not responding, attempting restart..."
    brew services restart postgresql@15 2>/dev/null || true
    sleep 3
    if ! pg_isready -q 2>/dev/null; then
      err "PostgreSQL is down and could not be restarted"
      return 1
    fi
  fi

  # Run Claude in non-interactive mode with timeout
  local exit_code=0
  CLAUDECODE= claude \
    -p "$prompt" \
    --permission-mode "bypassPermissions" \
    --allowedTools "Bash,Read,Write,Edit,Glob,Grep" \
    --no-session-persistence \
    > "$task_log_file" 2>&1 &
  local claude_pid=$!

  # Watchdog: kill claude if it exceeds TASK_TIMEOUT
  (
    sleep "$TASK_TIMEOUT"
    if kill -0 "$claude_pid" 2>/dev/null; then
      kill "$claude_pid" 2>/dev/null
    fi
  ) &
  local watchdog_pid=$!

  # Wait for claude to finish
  wait "$claude_pid" 2>/dev/null || exit_code=$?

  # Clean up watchdog if claude finished before timeout
  kill "$watchdog_pid" 2>/dev/null || true
  wait "$watchdog_pid" 2>/dev/null || true

  if [[ "$exit_code" -eq 143 ]]; then
    task_log "$task_num" "TIMEOUT after ${TASK_TIMEOUT}s"
    return 1
  elif [[ "$exit_code" -ne 0 ]]; then
    task_log "$task_num" "Claude exited with code $exit_code"
    return 1
  fi

  task_log "$task_num" "Claude session completed"
  return 0
}

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
# Verification Fixes Sprint Summary — $SPRINT_DATE

| Metric | Value |
|--------|-------|
| Branch | $BRANCH_NAME |
| Base | $BASE_BRANCH |
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
      completed) icon="pass" ;;
      failed)    icon="FAIL" ;;
      skipped)   icon="skip" ;;
      *)         icon="----" ;;
    esac
    echo "| $i | $name | $icon $status | $attempts |" >> "$summary_file"
  done

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
        echo "Log: \`logs/verification-fixes-$SPRINT_DATE/task-$(printf '%02d' "$i").log\`" >> "$summary_file"
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
  echo -e "${BOLD}=== Verification Fixes Sprint Complete ===${NC}"
  echo ""
  echo -e "  Completed: ${GREEN}$completed${NC} / $TOTAL_TASKS"
  echo -e "  Failed:    ${RED}$failed${NC}"
  echo -e "  Skipped:   ${YELLOW}$skipped${NC}"
  echo ""
  echo -e "  Progress:  $PROGRESS_FILE"
  echo -e "  Logs:      $LOG_DIR/"
  echo -e "  Summary:   $LOG_DIR/summary.md"
  echo -e "  Branch:    $BRANCH_NAME (from $BASE_BRANCH)"
  echo ""

  if (( failed == 0 && skipped == 0 )); then
    echo -e "  ${GREEN}${BOLD}All $TOTAL_TASKS verification tasks completed!${NC}"
    echo -e "  Next: run a fresh verification audit to confirm Grade-A status."
  else
    echo -e "  ${YELLOW}Review failed/skipped tasks in the summary and logs.${NC}"
    echo -e "  Re-run with --resume to retry failed tasks."
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
    log "Resuming verification fixes sprint from progress file..."
    for i in $(seq 1 $TOTAL_TASKS); do
      local status
      status=$(get_task_status "$i")
      if [[ "$status" == "failed" || "$status" == "skipped" ]]; then
        local tmp
        tmp=$(mktemp)
        jq --arg t "$i" \
          '.tasks[$t].status = "pending" | .tasks[$t].attempts = 0 | .tasks[$t].error = null' \
          "$PROGRESS_FILE" > "$tmp" && mv "$tmp" "$PROGRESS_FILE"
        log "Reset task $i from $status to pending"
      fi
    done
  else
    init_progress
  fi

  echo ""
  log "Starting verification fixes sprint — $TOTAL_TASKS tasks on branch $BRANCH_NAME"
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
        set_task_error "$task_num" "failed after 2 attempts — check logs/verification-fixes-$SPRINT_DATE/task-$(printf '%02d' "$task_num").log"
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
