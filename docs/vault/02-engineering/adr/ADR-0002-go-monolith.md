---
title: "ADR-0002: Go Monolith Backend Architecture"
status: accepted
created: 2026-02-15
updated: 2026-02-15
tags: [adr, backend, architecture]
---

# ADR-0002: Go Monolith Backend Architecture

## Status

Accepted

## Context

The Niotebook backend needs to serve a REST API for the TUI client, handle authentication, manage user data, and serve timeline feeds. We considered three approaches:

1. **Go monolith** — single Go binary serving all API endpoints, handling auth, and connecting to the database.
2. **Go microservices** — separate services for API, real-time, and background jobs.
3. **Backend-as-a-service** — Supabase, PocketBase, or Firebase handling auth/DB/realtime.

For an MVP targeting a single VPS deployment, the primary constraints are: development velocity, operational simplicity, and the ability to iterate quickly.

## Decision

Use a **Go monolith** — a single Go binary that serves the entire HTTP API, handles JWT authentication, and communicates with PostgreSQL.

### Considered Alternatives

**Go microservices:** Premature for an MVP. Adds inter-service communication overhead, deployment complexity (multiple binaries, service discovery), and operational burden — all without clear benefit at MVP scale. Can decompose the monolith later if needed.

**Backend-as-a-service (Supabase/PocketBase):** Reduces backend development effort but introduces external dependencies, vendor lock-in, and less control over data modeling. PocketBase is the closest fit (single-binary Go, embedded SQLite) but SQLite's single-writer lock is problematic for concurrent social media writes.

## Consequences

### Positive

- Single binary to build, deploy, and operate
- Simplest possible deployment: copy binary to VPS, run it
- All code in one place: easy to refactor, debug, and understand
- Go's standard library `net/http` (Go 1.22+) provides pattern-matching routing out of the box — no framework dependency
- Can be decomposed into services later when scale demands it

### Negative

- All functionality shares a single process — a bug in one handler can affect the entire server
- Vertical scaling only (bigger VPS) until decomposed
- Must be disciplined about internal package boundaries to prevent spaghetti code

### Neutral

- This is the standard approach for Go web applications at MVP stage. Most successful Go projects (Gitea, Mattermost) started as monoliths.
