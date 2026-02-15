---
title: "ADR-0006: Manual Feed Refresh for MVP"
status: accepted
created: 2026-02-15
updated: 2026-02-15
tags: [adr, tui, real-time]
---

# ADR-0006: Manual Feed Refresh for MVP

## Status

Accepted

## Context

The TUI needs a strategy for showing new posts in the timeline. Options:

1. **Manual refresh** — user presses a key (e.g., `r`) to fetch the latest posts
2. **Polling** — TUI polls the server every N seconds in a background goroutine
3. **WebSockets** — server pushes new posts to the TUI over a persistent connection

## Decision

Use **manual refresh only** for the MVP. The user presses a key to fetch new posts. No background goroutines, no persistent connections.

### Why

The MVP prioritizes shipping speed and simplicity. Manual refresh:
- Requires zero background concurrency in the TUI (no goroutine management, no channel coordination)
- Has zero server infrastructure overhead (no WebSocket upgrade handling, no connection tracking)
- Is predictable for the user (feed updates only when they ask)
- Eliminates an entire class of bugs (stale connections, reconnection logic, race conditions between user scroll position and incoming data)

### Upgrade Path

- **v1.1: Polling** — add a background goroutine that polls every 30s. Show a "N new posts" indicator at the top of the timeline; user clicks to load them (like X's "Show N posts" bar).
- **v1.2+: WebSockets** — when the platform has enough concurrent users to justify it, add server-sent push for instant updates.

## Consequences

### Positive

- Simplest possible implementation — a single HTTP GET on keypress
- No background goroutine lifecycle management in the TUI
- No WebSocket server infrastructure
- Deterministic behavior: feed state changes only on explicit user action

### Negative

- Feed feels "static" compared to a real-time social media experience
- Users must manually check for new content
- May feel primitive to users accustomed to streaming feeds

### Neutral

- Manual refresh is the standard pattern for TUI applications (e.g., `lazygit` refreshes on keypress, not automatically). It aligns with terminal UX conventions.
