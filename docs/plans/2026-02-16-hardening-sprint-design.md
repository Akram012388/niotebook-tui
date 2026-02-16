# Niotebook Hardening Sprint — Design Document

**Date:** 2026-02-16
**Branch:** `hardening-sprint` (off `mvp-sprint`)
**Objective:** Fix all critical and important issues from the B+ code review to achieve Grade-A across all 7 pillars of SWE.

## Context

The MVP sprint completed 24/24 tasks with zero failures. Code review graded the codebase B+ (86/100). Three review agents identified 3 critical and 5 important issues preventing Grade-A status.

## Issues to Address

### Critical (Must Fix)
| # | Issue | Location | Impact |
|---|-------|----------|--------|
| C1 | Unchecked type assertions in auth middleware | `middleware/auth.go:80-81` | Server panic on malformed JWT claims |
| C2 | JWT secret length not validated | `cmd/server/main.go:44-48` | Accepts weak secrets, JWT forgery risk |
| C3 | Unchecked type assertions in TUI app | `app/app.go` (12 locations) | TUI panic on adapter failure |

### Important (Should Fix)
| # | Issue | Location | Impact |
|---|-------|----------|--------|
| I1 | CORS allows all origins | `middleware/cors.go:7` | Any website can call the API |
| I2 | No HTTP client timeout or network retry | `tui/client/client.go:29` | TUI hangs on network issues |
| I3 | Window resize commands discarded | `app/app.go:474-503` | Views can't react to resize |
| I4 | XDG_CONFIG_HOME not honored | `tui/config/config.go:40-43` | Non-standard config path |
| I5 | Recovery/CORS/Logging middleware untested | `middleware/` | 0% coverage on critical paths |
| I6 | TUI layer at 13-33% coverage | `tui/` | Below 80% target |
| I7 | No lint step in CI | `.github/workflows/ci.yml` | Style issues not caught |
| I8 | Flaky time-dependent test | `store/post_store_test.go:120` | Intermittent CI failures |

## Decisions

1. **Script pattern:** New `scripts/hardening-sprint.sh` reusing the same architecture as `sprint.sh` (bash 5, jq progress tracking, Claude non-interactive sessions, background watchdog timeout, dependency graph, retry-once-then-skip, per-task logging).

2. **Branch:** `hardening-sprint` branched from `mvp-sprint`. All fixes committed here.

3. **No new dependencies:** The plaintext token storage issue (keychain) is deferred to a future sprint — it requires `go-keyring` which has CGo dependencies on Linux. The current `0600` file permissions are acceptable for MVP.

4. **Coverage target:** 65% overall (up from 40.8%). This is achievable by adding tests for middleware (recovery, CORS, logging), TUI components (postcard, header, statusbar), TUI views (register, error cases), and TUI app (keyboard shortcuts, overlays).

5. **Task timeout:** 15 minutes per task (same as MVP sprint).

## Task Breakdown

### Phase 1: Critical Security Fixes (Tasks 1-3, no dependencies)

**Task 1: Fix unchecked type assertions in auth middleware**
- File: `internal/server/middleware/auth.go`
- Change lines 79-82: Add `ok` pattern for `claims["sub"].(string)` and `claims["username"].(string)`
- Return 401 with `ErrCodeUnauthorized` if assertions fail
- Add test for malformed JWT claims

**Task 2: Add JWT secret length validation**
- File: `cmd/server/main.go`
- After line 48: Add `len(jwtSecret) < 32` check, log error and exit
- Update CI workflow to use a 32+ byte secret (current one is 38 bytes, already OK)

**Task 3: Fix unchecked type assertions in TUI app**
- File: `internal/tui/app/app.go`
- All 12 locations: Replace `updated.(TimelineViewModel)` with comma-ok pattern
- On failure: return model unchanged with error message to status bar
- All existing tests must still pass

### Phase 2: Important Code Fixes (Tasks 4-7, depend on Phase 1)

**Task 4: Make CORS origin configurable**
- File: `internal/server/middleware/cors.go`
- Accept `allowedOrigin string` parameter in `CORS()` function
- File: `internal/server/server.go` — pass origin from config
- File: `cmd/server/main.go` — read `NIOTEBOOK_CORS_ORIGIN` env var, default `*`
- Add CORS test (OPTIONS preflight, header validation)

**Task 5: Add HTTP client timeout + network retry**
- File: `internal/tui/client/client.go`
- Set `http.Client{Timeout: 30 * time.Second}`
- Add `doWithRetry()` wrapper: 3 attempts with exponential backoff (1s, 2s, 4s) for network errors only (not HTTP errors)
- Add test with httptest server that drops connections

**Task 6: Fix window resize command propagation**
- File: `internal/tui/app/app.go` function `propagateWindowSize`
- Collect all commands from view updates into `[]tea.Cmd` slice
- Return `tea.Batch(cmds...)` instead of `nil`
- Also fix the unchecked type assertions in this function (if not already fixed by Task 3)

**Task 7: Add XDG_CONFIG_HOME support**
- File: `internal/tui/config/config.go`
- `ConfigDir()`: Check `$XDG_CONFIG_HOME` env var first, fall back to `~/.config/niotebook`
- Add test verifying env var override

### Phase 3: Test Coverage Push (Tasks 8-11, depend on Phase 2)

**Task 8: Add middleware tests**
- Files: `internal/server/middleware/recovery_test.go`, `cors_test.go`, `logging_test.go`
- Recovery: handler that panics, verify 500 response + no crash
- CORS: OPTIONS preflight returns correct headers, non-OPTIONS gets CORS headers, configurable origin
- Logging: verify status code capture via responseWriter wrapper

**Task 9: Add TUI component tests**
- Files: `internal/tui/components/postcard_test.go`, `header_test.go`, `statusbar_test.go`
- PostCard: selected/unselected rendering, narrow width, long content, nil author
- Header: various widths, empty username
- StatusBar: SetError, SetSuccess, auto-clear timer

**Task 10: Add TUI view tests**
- Files: `internal/tui/views/register_test.go` (new), enhanced `login_test.go`, `timeline_test.go`, `compose_test.go`, `profile_test.go`
- Register: form rendering, field navigation, submit
- All views: error case tests (API failure messages)
- Timeline: pagination, refresh key

**Task 11: Add TUI app integration tests**
- File: `internal/tui/app/app_test.go` (enhanced)
- Keyboard shortcuts: 'r' refresh, 'p' profile, 'q' quit, '?' help
- Overlay lifecycle: open compose, cancel, open help, close
- Error propagation: API error displays in status bar

### Phase 4: CI Hardening (Task 12, depends on Phase 3)

**Task 12: CI pipeline hardening**
- File: `.github/workflows/ci.yml`
- Add `golangci-lint run ./...` step after build
- Add coverage report step with 65% threshold
- Fix flaky test in `store/post_store_test.go` (increase `time.Sleep` to 50ms or use explicit timestamps)
- Verify all tests pass with race detector

## Dependency Graph

```
Phase 1 (no deps):  1  2  3
                      \|/
Phase 2 (depends on 1-3):  4  5  6  7
                              \|/
Phase 3 (depends on 4-7):  8  9  10  11
                              \|/
Phase 4 (depends on 8-11):  12
```

Simplified: Tasks within each phase are independent of each other. Each phase depends on all tasks in the prior phase completing.

## Verification

After all 12 tasks complete:
1. `go build ./...` — must succeed
2. `go test ./... -v -race` — all pass, zero failures
3. `go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out | tail -1` — >= 65% overall
4. `golangci-lint run ./...` — zero errors
5. No unchecked type assertions: `grep -rn '\.\(.*ViewModel\)' internal/tui/app/app.go` returns zero matches without comma-ok
6. No unchecked JWT claims: `grep -n 'claims\[.*\]\.\(string\)' internal/server/middleware/auth.go` returns zero matches without comma-ok

## Success Criteria

Grade-A (90+) across all 7 pillars:
- Architecture & Design: already A- (91), maintain
- Code Quality: B+ (87) → A- (fix type assertions, CORS, window resize)
- Testing & QA: C+ (76) → B+/A- (middleware, components, views, app tests)
- Security: B (83) → A- (JWT secret validation, checked assertions, configurable CORS)
- Performance: A- (90) → A (client timeout + retry)
- Maintainability: B+ (86) → A- (XDG compliance, CI lint)
- DevOps & CI/CD: A (93) → A+ (lint step, coverage threshold, flaky test fix)
