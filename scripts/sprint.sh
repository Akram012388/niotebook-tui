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
