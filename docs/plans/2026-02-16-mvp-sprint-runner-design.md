# MVP Sprint Runner — Design Document

**Date:** 2026-02-16
**Status:** Approved

## Goal

Build an automated sprint runner that executes the 24-task MVP implementation plan
(`docs/vault/04-plans/2026-02-16-mvp-implementation.md`) overnight on this machine
using non-interactive Claude Code sessions, with clean logging, progress tracking,
and resume capability.

## Decisions

| Decision | Choice |
|----------|--------|
| PostgreSQL | Homebrew install (postgresql@15) |
| Session granularity | One task per Claude session (24 sessions) |
| Failure handling | Retry once, then skip and log; skip dependent tasks |
| Parallelism | Strictly sequential (tasks 1-24 in order) |
| Execution environment | Local Mac, unattended overnight |
| Approach | Custom sprint script (no Ralph Wiggum dependency) |

## Directory Layout

```
scripts/
  sprint.sh              # Main orchestrator
  sprint-bootstrap.sh    # Dependency installation
  sprint-progress.json   # Runtime task tracking state
logs/
  sprint-YYYY-MM-DD/
    task-NN.log          # Per-task Claude session output
    task-NN-verify.log   # Build + test verification output
    summary.md           # Final completion report
```

## Bootstrap Phase (sprint-bootstrap.sh)

Idempotent script that ensures all system-level dependencies are present:

1. Verify Go >= 1.22 (already installed: 1.25.6)
2. Install PostgreSQL 15 via Homebrew, start service
3. Create `niotebook_dev` and `niotebook_test` databases
4. Install `golang-migrate` CLI via Homebrew
5. Install `golangci-lint` via Homebrew
6. Generate `.env` from `.env.example` with local connection string and JWT secret
7. Exit with clear error if any tool is missing

Go library dependencies (bubbletea, pgx, jwt, etc.) are handled by Task 1 of the
implementation plan via `go get` commands — not by bootstrap.

## Sprint Script Core Logic (sprint.sh)

### Pre-Flight Checks

Before starting, verifies: Go version, PostgreSQL running, databases exist,
CLI tools installed, Claude CLI available, disk space > 5GB.

### Main Loop

```
for task_num in 1..24:
  1. Read progress file — skip if completed/skipped
  2. Check dependency graph — skip if blocker task failed
  3. Check PostgreSQL health (tasks 3+)
  4. Invoke claude -p with focused prompt referencing the plan
  5. Run post-task verification (go build, go test)
  6. Auto-commit any uncommitted changes
  7. Update progress file
  8. On failure: retry once, then mark failed + skip dependents
```

### Claude Invocation

Each task runs as a non-interactive Claude session:

```bash
claude -p "Read docs/vault/04-plans/2026-02-16-mvp-implementation.md.
Execute Task N exactly as written. Follow every step. Commit as specified.
After all steps, run: go build ./... && go test ./... -v -race" \
  --allowedTools "Bash,Read,Write,Edit,Glob,Grep" \
  --max-turns 50
```

### Dependency Graph

Encoded from the plan's execution notes:

```
Task 1 → Task 2, 3, 10, 11, 15, 17
Task 2 → Task 4, 6, 16, 17, 18
Task 3 → Task 5
Task 4 → Task 5
Task 6 → Task 7
Task 7 → Task 8, 9
Tasks 7-11 → Task 12
Task 12 → Task 13 → Task 14
Task 15 → Task 16
Tasks 16, 18 → Task 19
Tasks 17, 19 → Task 20
Task 20 → Task 21, 22
Tasks 19-22 → Task 23
Tasks 14, 23 → Task 24
```

If a task fails, all transitive dependents are marked "skipped".

## Progress Tracking (sprint-progress.json)

```json
{
  "sprint_id": "2026-02-16",
  "branch": "mvp-sprint",
  "started_at": "ISO8601",
  "tasks": {
    "1": {"status": "completed|failed|skipped|pending", "attempts": 1, "started": "...", "finished": "..."},
    ...
  }
}
```

## Verification

After each Claude session, the script independently runs:

1. `go build ./...` — code compiles
2. `go test ./... -v -race -timeout 120s` — tests pass (skipped for Task 1)
3. Check for uncommitted changes — auto-commit if found

## Resume

```bash
./scripts/sprint.sh           # Fresh start
./scripts/sprint.sh --resume  # Resume from progress file
```

On resume: skips completed tasks, retries failed tasks (reset attempt count),
continues from next pending task.

## Safety

- 15-minute soft timeout per task (--max-turns 50)
- Disk space check before starting
- PostgreSQL health check before database tasks
- Never force-push or rewrite git history
- All commits are additive on the mvp-sprint branch

## Branch Strategy

- Create `mvp-sprint` from `main`
- All sprint work on this branch
- Conventional commit messages per the plan
- Post-sprint: user decides merge strategy

## Risk Areas

- **Task 5** (store implementations): Real PostgreSQL integration tests
- **Task 23** (root AppModel): Wires everything together, complex
- **Task 24** (CI verification): Full integration test of the entire system
