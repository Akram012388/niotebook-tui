---
title: "ADR-0018: Header + Content + Status Bar Layout"
status: accepted
created: 2026-02-15
updated: 2026-02-15
tags: [adr, tui, layout]
---

# ADR-0018: Header + Content + Status Bar Layout

## Status

Accepted

## Context

The TUI needs a screen structure to organize information across all views. Three options were considered: full-screen single view, header/content/statusbar, and sidebar/main.

## Decision

Use **header (1 line) + content area (dynamic) + status bar (1 line)**. The header and status bar persist across all views. The content area swaps between views.

### Rationale

- **Persistent context:** Header shows app name, current user, and view name. Users always know where they are.
- **Maximum content width:** No sidebar consuming horizontal space. With 140-char posts, every column matters.
- **Established pattern:** The standard layout for polished TUI apps (OpenCode, lazygit, k9s).
- **Sidebar is premature:** Only 3 views in MVP. Key presses switch instantly; a navigation sidebar adds no value.

## Consequences

### Positive

- Clean, focused UI with maximum content real estate
- Status bar doubles as feedback channel (errors, confirmations) and key hint display
- Simple implementation — three fixed components composed vertically

### Negative

- No persistent navigation indicator beyond the header's view name
- Status bar is limited to one line for feedback messages

### Neutral

- A sidebar or tab bar can be introduced post-MVP when more views exist.
