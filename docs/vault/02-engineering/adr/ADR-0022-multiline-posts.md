---
title: "ADR-0022: Multi-line Posts Allowed"
status: accepted
created: 2026-02-15
updated: 2026-02-15
tags: [adr, product, content]
---

# ADR-0022: Multi-line Posts Allowed

## Status

Accepted

## Context

The compose modal uses a `textarea` component. Should the Enter key insert a newline, or should posts be single-line only?

## Decision

**Posts can contain newlines.** The 140-character limit counts all runes including `\n`. Enter inserts a newline in the compose textarea.

### Rationale

- Code snippets and formatted output benefit from line breaks
- The target audience (developers) expects Enter to insert newlines in a textarea
- 140-char limit naturally constrains the number of lines (a post can't have more than ~5 short lines)
- `Ctrl+Enter` publishes, so Enter is free for newlines

### Rendering

- Post content renders with newlines preserved in the timeline
- Word wrapping applies per-line
- Post card height varies based on content (acceptable — 140 chars is at most 5-6 lines)

## Consequences

### Positive

- Supports code snippets, lists, and formatted thoughts
- Natural textarea behavior

### Negative

- Variable post card heights make the timeline less uniform
- Posts with many newlines could waste vertical space (mitigated by the 140-char total limit)
