---
title: "ADR-0023: Health Check Endpoint"
status: accepted
created: 2026-02-15
updated: 2026-02-15
tags: [adr, api, operations]
---

# ADR-0023: Health Check Endpoint

## Status

Accepted

## Context

Operators need to know if the server is healthy. Monitoring tools and reverse proxies (Caddy) need a health endpoint.

## Decision

Add `GET /health` — unauthenticated, not rate limited. Returns 200 if server and database are reachable, 503 if database is down. Response includes server version.

See [[02-engineering/architecture/server-internals#Health Endpoint|Server Internals]] for full specification.

## Consequences

### Positive

- Caddy can use it for backend health checks
- Uptime monitoring services can ping it
- Quick way to verify deployment succeeded

### Negative

- Minimal (adds one simple handler)
