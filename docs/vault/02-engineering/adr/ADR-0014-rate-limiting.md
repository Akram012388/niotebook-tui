---
title: "ADR-0014: Per-IP Token Bucket Rate Limiting"
status: accepted
created: 2026-02-15
updated: 2026-02-15
tags: [adr, security, api]
---

# ADR-0014: Per-IP Token Bucket Rate Limiting

## Status

Accepted

## Context

The API must be protected from abuse (spam, brute-force login, scraping). Rate limiting is the first line of defense.

## Decision

Implement **per-IP token bucket rate limiting** as HTTP middleware on the server.

### Configuration

| Endpoint Category | Limit | Burst |
|-------------------|-------|-------|
| Auth (`/auth/*`) | 10 req/min | 5 |
| Write (`POST /posts`) | 30 req/min | 10 |
| Read (timeline, profiles) | 120 req/min | 30 |

### Implementation

Use Go's `golang.org/x/time/rate` package (standard extended library). Each unique IP gets a `rate.Limiter` instance stored in a `sync.Map`. Stale entries are cleaned up periodically (every 10 minutes, remove IPs not seen in 15 minutes).

### Response on Limit

Return `429 Too Many Requests` with a `Retry-After` header. The TUI displays: `"Rate limited. Try again in Xs."` in the status bar.

### Why Per-IP, Not Per-User

Per-user rate limiting requires the auth middleware to run first (to extract the user ID from the JWT). Per-IP rate limiting runs before auth, protecting unauthenticated endpoints (login, register) from brute force. Per-user limits can be layered on top in a future iteration.

## Consequences

### Positive

- Protects all endpoints, including unauthenticated ones
- Uses stdlib-adjacent package (`x/time/rate`) — no external dependency
- Simple in-memory implementation — no Redis needed for MVP
- Token bucket algorithm allows short bursts while enforcing sustained rate

### Negative

- In-memory state is lost on server restart (acceptable — limiters rebuild quickly)
- Per-IP doesn't work behind shared IPs (VPNs, corporate proxies) — legitimate users may share limits
- Cannot distinguish between abusive and legitimate high-volume users on the same IP

### Neutral

- Per-user rate limiting is a natural addition for v1.1, layered as a second middleware after auth.
