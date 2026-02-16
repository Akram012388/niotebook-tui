#!/usr/bin/env bash
#
# Grade-A Sprint Runner
# Executes the 7-task implementation plan at:
#   docs/plans/2026-02-16-grade-a-sprint-implementation.md
#
# Usage: ./scripts/grade-a-sprint.sh
#
# This script runs non-interactively in a Claude Code session,
# executing each task sequentially with verification between tasks.
#
set -euo pipefail

PROJECT_ROOT="$(cd "$(dirname "$0")/.." && pwd)"
cd "$PROJECT_ROOT"

BRANCH="grade-a-sprint"
PLAN="docs/plans/2026-02-16-grade-a-sprint-implementation.md"
LOG_DIR="$PROJECT_ROOT/logs/grade-a-sprint"
PROGRESS_FILE="$PROJECT_ROOT/scripts/grade-a-sprint-progress.json"
TOTAL_TASKS=7

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

log()  { echo -e "${BLUE}[$(date '+%H:%M:%S')]${NC} $*"; }
ok()   { echo -e "${GREEN}[✓]${NC} $*"; }
warn() { echo -e "${YELLOW}[!]${NC} $*"; }
fail() { echo -e "${RED}[✗]${NC} $*"; }

# Initialize progress tracking
init_progress() {
    mkdir -p "$LOG_DIR"
    if [[ ! -f "$PROGRESS_FILE" ]]; then
        cat > "$PROGRESS_FILE" <<PEOF
{
  "sprint": "grade-a",
  "plan": "$PLAN",
  "branch": "$BRANCH",
  "total_tasks": $TOTAL_TASKS,
  "started": "$(date -u +%Y-%m-%dT%H:%M:%SZ)",
  "tasks": {}
}
PEOF
    fi
}

get_last_completed_task() {
    if [[ ! -f "$PROGRESS_FILE" ]]; then
        echo 0
        return
    fi
    # Find the highest completed task number
    local max=0
    for i in $(seq 1 $TOTAL_TASKS); do
        local status
        status=$(python3 -c "
import json
with open('$PROGRESS_FILE') as f:
    d = json.load(f)
print(d.get('tasks',{}).get('$i',{}).get('status',''))" 2>/dev/null || echo "")
        if [[ "$status" == "completed" ]]; then
            max=$i
        fi
    done
    echo "$max"
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

# Setup branch
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

# Run a single task via Claude Code
run_task() {
    local task_num=$1
    local task_log="$LOG_DIR/task-${task_num}.log"
    local task_descriptions=(
        ""
        "Task 1: Harden CORS — require explicit origin, add security headers (X-Content-Type-Options, X-Frame-Options). Remove wildcard default."
        "Task 2: Add handler unit tests for helpers.go — writeJSON, writeAPIError, errorCodeToHTTPStatus, decodeBody."
        "Task 3: Add 25+ targeted view tests — login validation/submit/error, register validation/submit, profile loading/navigation/update, timeline pagination/init, compose init/error."
        "Task 4: Fix CI Go version mismatch — update .github/workflows/ci.yml go-version from 1.22 to 1.25 to match go.mod."
        "Task 5: Remove tracked server binary from git — git rm --cached server, add to .gitignore."
        "Task 6: Update .env.example — add NIOTEBOOK_CORS_ORIGIN and NIOTEBOOK_TEST_DB_URL."
        "Task 7: Final verification — run lint, build, test, coverage. Confirm Grade-A metrics."
    )

    log "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log "TASK $task_num / $TOTAL_TASKS"
    log "${task_descriptions[$task_num]}"
    log "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    update_task_status "$task_num" "running"

    local prompt
    prompt="You are executing Task $task_num of the Grade-A Sprint plan.

PLAN FILE: $PLAN
Read the plan file first, then execute ONLY Task $task_num exactly as specified.

${task_descriptions[$task_num]}

RULES:
- Follow the plan EXACTLY as written — do not deviate
- Run all verification commands specified in the task
- Commit with the exact commit message specified in the plan
- If tests fail, fix them before committing
- Do NOT proceed to other tasks — only complete Task $task_num
- After committing, run 'go test ./... -race -count=1' to verify nothing is broken
- Report back with: files changed, tests passing, coverage if applicable"

    if claude --print --dangerously-skip-permissions \
        -p "$prompt" \
        > "$task_log" 2>&1; then
        ok "Task $task_num completed"
        update_task_status "$task_num" "completed"
        return 0
    else
        fail "Task $task_num failed (exit code $?)"
        update_task_status "$task_num" "failed" "claude exited with non-zero"
        return 1
    fi
}

# Verification gate between tasks
verify_between_tasks() {
    log "Running verification gate..."

    if ! go build ./... 2>/dev/null; then
        fail "Build failed after task"
        return 1
    fi
    ok "Build: clean"

    if ! golangci-lint run ./... 2>/dev/null; then
        warn "Lint issues found (non-blocking)"
    else
        ok "Lint: 0 issues"
    fi

    if ! go test ./... -race -count=1 > /dev/null 2>&1; then
        fail "Tests failed after task"
        return 1
    fi
    ok "Tests: all passing"

    return 0
}

# Main execution
main() {
    log "╔══════════════════════════════════════╗"
    log "║     GRADE-A SPRINT RUNNER            ║"
    log "║     7 Tasks · Security + Tests +     ║"
    log "║     Maintainability                  ║"
    log "╚══════════════════════════════════════╝"
    log ""
    log "Plan: $PLAN"
    log "Branch: $BRANCH"
    log ""

    init_progress
    setup_branch

    local last_completed
    last_completed=$(get_last_completed_task)

    if [[ $last_completed -gt 0 ]]; then
        warn "Resuming from task $((last_completed + 1)) (tasks 1-$last_completed completed)"
    fi

    local failed=0
    for task_num in $(seq $((last_completed + 1)) $TOTAL_TASKS); do
        if ! run_task "$task_num"; then
            fail "Task $task_num failed. Check log: $LOG_DIR/task-${task_num}.log"
            failed=1
            # Retry once
            warn "Retrying task $task_num..."
            if ! run_task "$task_num"; then
                fail "Task $task_num failed on retry. Stopping."
                break
            fi
        fi

        # Skip verification on final task (it IS verification)
        if [[ $task_num -lt $TOTAL_TASKS ]]; then
            if ! verify_between_tasks; then
                fail "Verification gate failed after task $task_num"
                break
            fi
        fi

        log ""
    done

    log ""
    log "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"
    log "SPRINT SUMMARY"
    log "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

    local completed=0
    for i in $(seq 1 $TOTAL_TASKS); do
        local status
        status=$(python3 -c "
import json
with open('$PROGRESS_FILE') as f:
    d = json.load(f)
print(d.get('tasks',{}).get('$i',{}).get('status','pending'))" 2>/dev/null || echo "pending")

        if [[ "$status" == "completed" ]]; then
            ok "Task $i: completed"
            completed=$((completed + 1))
        elif [[ "$status" == "failed" ]]; then
            fail "Task $i: failed"
        else
            warn "Task $i: $status"
        fi
    done

    log ""
    log "Completed: $completed / $TOTAL_TASKS"
    log "Logs: $LOG_DIR/"
    log "Progress: $PROGRESS_FILE"

    if [[ $completed -eq $TOTAL_TASKS ]]; then
        ok "ALL TASKS COMPLETE — Grade-A sprint finished!"
        log ""
        log "Next steps:"
        log "  1. Review changes: git log --oneline $BRANCH"
        log "  2. Run final verification: go test ./... -race -coverprofile=coverage.out"
        log "  3. Check coverage: go tool cover -func=coverage.out | grep total"
        log "  4. Create PR: gh pr create --base hardening-sprint --head $BRANCH"
        return 0
    else
        fail "Sprint incomplete. $completed/$TOTAL_TASKS tasks done."
        return 1
    fi
}

main "$@"
