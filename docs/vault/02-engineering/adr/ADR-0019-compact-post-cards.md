---
title: "ADR-0019: Compact Post Card Design"
status: accepted
created: 2026-02-15
updated: 2026-02-15
tags: [adr, tui, design, component]
---

# ADR-0019: Compact Post Card Design

## Status

Accepted

## Context

Post cards are the most-rendered component. The design must balance information density with readability.

## Decision

Use **compact post cards** with a single header line and minimal metadata:

```
▸ @username · 2m
  Post content here, word-wrapped.
──────────────────────────────────
```

- **Header:** `@username · relative_time` — one line
- **Content:** Post text, word-wrapped to terminal width minus 2 chars padding
- **Separator:** Horizontal rule between posts
- **Selection:** `▸` marker + accent color on selected post

No display name, no footer, no engagement counts (MVP has no engagement features).

## Consequences

### Positive

- Maximum posts visible per screen
- Every post occupies 3-4 lines (header + 1-2 content lines + separator)
- In a 24-row terminal with 2 rows for header/status bar, ~5-6 posts are visible at once

### Negative

- No display name shown (only @username) — less personal
- No engagement counts even when likes/reposts are added (will need a redesign for v1.1)

### Neutral

- Display name can be added to the header line in v1.1 without layout changes: `Display Name @username · 2m`
