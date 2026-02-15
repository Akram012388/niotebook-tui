---
title: "ADR-0021: Comprehensive Testing Strategy"
status: accepted
created: 2026-02-15
updated: 2026-02-15
tags: [adr, testing, quality]
---

# ADR-0021: Comprehensive Testing Strategy

## Status

Accepted

## Context

The testing approach determines the confidence level for shipping and the development velocity. Options ranged from minimal (auth + validation only) to comprehensive (unit + integration + TUI rendering).

## Decision

Use **comprehensive testing** across three layers:

1. **Unit tests (~50%):** Business logic, validation, utilities. Mock store interfaces.
2. **Integration tests (~35%):** API endpoints against a real PostgreSQL test database.
3. **TUI rendering tests (~15%):** Bubble Tea model tests using `teatest` package.

### Coverage Targets

- `internal/server/service/`: 90%+
- `internal/server/handler/`: 80%+
- `internal/tui/components/`: 80%+
- Overall project: 80%+

### Test Infrastructure

- CI runs all tests on every push/PR via GitHub Actions
- PostgreSQL service container in CI for integration tests
- Race detector enabled (`-race` flag) on all test runs

## Consequences

### Positive

- High confidence in core logic correctness
- Integration tests catch issues that unit tests miss (serialization, DB constraints, middleware chain)
- TUI tests prevent rendering regressions
- CI prevents broken code from reaching main branch

### Negative

- More test code to write and maintain
- Integration tests require PostgreSQL (slower than pure unit tests)
- TUI tests can be brittle (sensitive to styling changes)

### Neutral

- Test infrastructure (helpers, fixtures, setup) is reusable as the project grows. The investment compounds.
