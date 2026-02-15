---
title: "ADR-0008: API-First REST Architecture"
status: accepted
created: 2026-02-15
updated: 2026-02-15
tags: [adr, api, architecture]
---

# ADR-0008: API-First REST Architecture

## Status

Accepted

## Context

The TUI client needs to communicate with the server. We evaluated three approaches:

1. **API-first REST** — server exposes a JSON REST API; TUI is just another HTTP client
2. **Embedded server** — bundle the server into the TUI binary (localhost HTTP)
3. **gRPC** — use Protocol Buffers and gRPC instead of REST/JSON

## Decision

Use an **API-first REST architecture**. The server exposes a standard JSON REST API. The TUI client consumes it over HTTP(S). The TUI has no special relationship with the server — it's a regular API client.

### API Design Principles

- RESTful resource naming: `/api/v1/posts`, `/api/v1/users/{id}`, `/api/v1/timeline`
- JSON request/response bodies
- JWT Bearer token authentication via `Authorization` header
- Standard HTTP status codes (200, 201, 400, 401, 404, 500)
- Cursor-based pagination for timeline feeds
- Go 1.22+ `net/http` router with pattern matching (no framework)

### Why Not Embedded Server

Bundling the server into the TUI means each user runs their own server instance with their own database. This defeats the purpose of a social platform — users can't see each other's posts. The embedded approach works for personal tools (like PocketBase-powered apps) but not for multi-user social media.

### Why Not gRPC

gRPC offers strong typing via Protocol Buffers, efficient binary serialization, and built-in streaming. However:

- Adds a code generation toolchain (`protoc`, `protoc-gen-go`) to the build process
- Harder to debug (can't `curl` an endpoint to test)
- Streaming support is valuable for real-time features, but MVP uses manual refresh only
- The compile-time type safety benefit is already achieved via shared Go types in `internal/models/`

gRPC may be reconsidered when real-time features (WebSocket/streaming) are added post-MVP.

## Consequences

### Positive

- The TUI is decoupled from the server — other clients (web, mobile, CLI tools, bots) can use the same API later
- Easy to test: `curl`, `httpie`, or any HTTP client can exercise the API
- No code generation step in the build process
- Well-understood patterns, extensive documentation and tooling

### Negative

- JSON serialization is less efficient than Protobuf (negligible at MVP scale)
- No built-in streaming (must add WebSockets separately for real-time)
- API versioning must be managed manually (`/api/v1/`)

### Neutral

- REST is the default choice for web APIs. This is the least surprising architecture for contributors.
