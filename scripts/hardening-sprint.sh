#!/opt/homebrew/bin/bash
set -euo pipefail

# Niotebook Hardening Sprint Runner
# Fixes all critical and important issues from the B+ code review to achieve Grade-A.
# Usage:
#   ./scripts/hardening-sprint.sh            # Fresh start
#   ./scripts/hardening-sprint.sh --resume   # Resume from progress file

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"

# Ensure Homebrew PostgreSQL is on PATH (keg-only on macOS)
export PATH="/opt/homebrew/opt/postgresql@15/bin:$PATH"

PLAN_FILE="docs/plans/2026-02-16-hardening-sprint-implementation.md"
SPRINT_DATE=$(date +%Y-%m-%d)
LOG_DIR="$PROJECT_DIR/logs/hardening-sprint-$SPRINT_DATE"
PROGRESS_FILE="$SCRIPT_DIR/hardening-sprint-progress.json"
BRANCH_NAME="hardening-sprint"
BASE_BRANCH="mvp-sprint"
TOTAL_TASKS=12
TASK_TIMEOUT=900           # 15 minutes in seconds

RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m'

log()      { echo -e "${GREEN}[hardening]${NC} $(date +%H:%M:%S) $*"; }
warn()     { echo -e "${YELLOW}[hardening]${NC} $(date +%H:%M:%S) $*"; }
err()      { echo -e "${RED}[hardening]${NC} $(date +%H:%M:%S) $*" >&2; }
task_log() { echo -e "${CYAN}[task $1]${NC} $(date +%H:%M:%S) ${*:2}"; }

# ── Dependency Graph ────────────────────────────────────────────────
# Phase 1: Tasks 1-3 (no deps — parallel safe)
# Phase 2: Tasks 4-7 (depend on all of Phase 1)
# Phase 3: Tasks 8-11 (depend on all of Phase 2)
# Phase 4: Task 12 (depends on all of Phase 3)
declare -A TASK_DEPS=(
  [1]=""
  [2]=""
  [3]=""
  [4]="1 2 3"
  [5]="1 2 3"
  [6]="1 2 3"
  [7]="1 2 3"
  [8]="4 5 6 7"
  [9]="4 5 6 7"
  [10]="4 5 6 7"
  [11]="4 5 6 7"
  [12]="8 9 10 11"
)

# Task names for display
declare -A TASK_NAMES=(
  [1]="Fix unchecked type assertions in auth middleware"
  [2]="Add JWT secret length validation"
  [3]="Fix unchecked type assertions in TUI app"
  [4]="Make CORS origin configurable"
  [5]="Add HTTP client timeout and network retry"
  [6]="Fix window resize command propagation"
  [7]="Add XDG_CONFIG_HOME support"
  [8]="Add middleware tests (recovery, CORS, logging)"
  [9]="Add TUI component tests"
  [10]="Add TUI view tests"
  [11]="Add TUI app integration tests"
  [12]="CI pipeline hardening"
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

# ── Pre-Flight Checks ──────────────────────────────────────────────
preflight() {
  echo ""
  echo -e "${BOLD}=== Niotebook Hardening Sprint Runner ===${NC}"
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
  for tool in golangci-lint jq claude; do
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

  # Verify MVP sprint codebase compiles
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

  # 2. Tests pass (all hardening tasks have tests)
  if go test ./... -v -race -timeout 120s >> "$verify_log" 2>&1; then
    task_log "$task_num" "  Tests: PASS"
  else
    task_log "$task_num" "  Tests: FAIL"
    ((errors++))
  fi

  # 3. Auto-commit any leftover uncommitted changes
  if [[ -n "$(git status --porcelain 2>/dev/null)" ]]; then
    task_log "$task_num" "  Auto-committing remaining changes..."
    git add -A
    git commit -m "chore: auto-commit remaining changes from hardening task $task_num" --no-verify 2>/dev/null || true
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
  prompt="You are hardening the Niotebook codebase — fixing critical and important issues from a code review.

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
    # SIGTERM from watchdog = timeout
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
# Hardening Sprint Summary — $SPRINT_DATE

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
        echo "Log: \`logs/hardening-sprint-$SPRINT_DATE/task-$(printf '%02d' "$i").log\`" >> "$summary_file"
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
  echo -e "${BOLD}=== Hardening Sprint Complete ===${NC}"
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
    echo -e "  ${GREEN}${BOLD}All $TOTAL_TASKS hardening tasks completed successfully!${NC}"
    echo -e "  Next: run code review to verify Grade-A status."
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
    log "Resuming hardening sprint from progress file..."
    # Reset failed and skipped tasks to pending for retry
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
  log "Starting hardening sprint — $TOTAL_TASKS tasks on branch $BRANCH_NAME"
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
        set_task_error "$task_num" "failed after 2 attempts — check logs/hardening-sprint-$SPRINT_DATE/task-$(printf '%02d' "$task_num").log"
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
