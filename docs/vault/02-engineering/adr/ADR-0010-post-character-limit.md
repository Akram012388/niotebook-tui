---
title: "ADR-0010: 140-Character Post Limit"
status: accepted
created: 2026-02-15
updated: 2026-02-15
tags: [adr, product, content]
---

# ADR-0010: 140-Character Post Limit

## Status

Accepted

## Context

Every social platform must decide how much text a single post can contain. This affects the culture of the platform, the UX of the timeline, and the rendering complexity in the TUI.

Options considered: 140, 280, 500, 1000, or no limit.

## Decision

Posts are limited to **140 characters**, matching the original Twitter constraint.

### Rationale

- **Concision breeds quality.** A tight limit forces users to distill their thoughts. This aligns with the "notebook for nerds" ethos — sharp, high-signal posts over rambling.
- **Terminal-friendly.** Most terminals are 80-120 columns wide. A 140-character post fits in 1-2 lines of rendered text, making the timeline dense and scannable.
- **Cultural signal.** The 140-char limit is an intentional design choice that differentiates Niotebook from platforms trending toward longer content (Bluesky 300, Mastodon 500, X 280). It attracts users who value brevity.
- **Simpler rendering.** No need for post truncation/expansion in the timeline view — every post fits without collapsing.

## Consequences

### Positive

- Clean, scannable timeline — more posts visible per screen
- Forces high-quality, concise writing
- Simplifies TUI rendering (no expand/collapse logic)
- Strong cultural identity for the platform

### Negative

- Some users will find it too restrictive for technical content (code snippets, explanations)
- No room for context or nuance in a single post (threads can address this in v1.1)
- May discourage long-form engagement

### Neutral

- Character count is enforced server-side (validation) and displayed client-side (compose view counter). Unicode characters count as their Go `len([]rune(content))` equivalent.
