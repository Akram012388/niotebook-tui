---
title: "ADR-0007: Monorepo with Shared Types"
status: accepted
created: 2026-02-15
updated: 2026-02-15
tags: [adr, project-structure, architecture]
---

# ADR-0007: Monorepo with Shared Types

## Status

Accepted

## Context

Niotebook consists of two distinct programs — a Go API server and a Go TUI client — that communicate over REST/JSON. These could live in one repository or two.

1. **Monorepo** — single repo with `cmd/server/` and `cmd/tui/`, sharing code via `internal/`
2. **Two repos** — `niotebook-server` and `niotebook-tui` as separate repositories

## Decision

Use a **monorepo** with Go's standard project layout:

```
niotebook/
├── cmd/server/       # Server entry point
├── cmd/tui/          # TUI entry point
├── internal/
│   ├── models/       # Shared domain types (User, Post, Profile)
│   ├── server/       # Server-only code (handlers, store, service)
│   └── tui/          # TUI-only code (views, components, client)
├── migrations/       # SQL migration files
├── go.mod
└── Makefile
```

### Key Design: Shared `internal/models/`

The `internal/models/` package defines domain types (`User`, `Post`, `Profile`) used by **both** the server (for database serialization and API responses) and the TUI (for API deserialization and rendering). This ensures:

- The server's JSON response format and the TUI's expected format are always in sync
- No duplicate struct definitions to maintain
- A compile-time guarantee that API contracts aren't broken

## Consequences

### Positive

- Single `go.mod`, single CI pipeline, single version history
- Shared types prevent API contract drift between server and client
- Atomic commits: a change to the API response format and the corresponding TUI parsing can land in one commit
- Simpler development workflow: one `git clone`, one IDE workspace

### Negative

- Both programs must use the same Go module version and dependency set
- CI runs tests for both programs even when only one changed (mitigated by build tags or CI path filters)
- Repository grows larger over time (acceptable for a project of this scope)

### Neutral

- Go's `internal/` convention ensures server code can't accidentally import TUI code and vice versa (only `internal/models/` is shared). The compiler enforces the boundary.
