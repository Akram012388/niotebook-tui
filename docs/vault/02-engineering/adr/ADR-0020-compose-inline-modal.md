---
title: "ADR-0020: Inline Compose Modal"
status: accepted
created: 2026-02-15
updated: 2026-02-15
tags: [adr, tui, compose, ux]
---

# ADR-0020: Inline Compose Modal

## Status

Accepted

## Context

When a user presses `n` to compose a new post, the compose interface can be: a full-screen view, an overlay modal, or a bottom panel.

## Decision

Use an **inline modal (overlay)** — a bordered box that appears centered over the dimmed timeline.

### Rationale

- **Context preserved:** Users can see the timeline behind the modal. Useful for referencing recent posts while composing.
- **Focus without disruption:** The modal captures input without navigating away from the timeline.
- **Natural mental model:** Modals are universally understood as temporary, focused interactions.

### Modal Behavior

- Appears on `n` keypress from timeline
- Centered horizontally and vertically in the content area
- Width: min(80, terminal_width - 4)
- Timeline behind the modal is rendered with dim styling
- `Esc` closes without confirmation (140 chars is trivially re-typed)
- `Ctrl+Enter` publishes and closes
- Live character counter: `42/140`, turns red at 131+

## Consequences

### Positive

- User stays "in" the timeline, reducing context switching
- Simple overlay rendering with Lip Gloss positioning
- No view transition animation needed

### Negative

- Modal implementation requires layering (render timeline, then overlay compose box)
- Cannot scroll the timeline while composing (acceptable — the modal captures all input)
