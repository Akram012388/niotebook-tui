---
title: "ADR-0015: Single Dark Theme for MVP"
status: accepted
created: 2026-02-15
updated: 2026-02-15
tags: [adr, tui, design]
---

# ADR-0015: Single Dark Theme for MVP

## Status

Accepted

## Context

The TUI needs a color scheme. Options ranged from a single hardcoded theme to a fully configurable theme system.

## Decision

Ship MVP with a **single, carefully designed dark theme**. No theme switching, no user-configurable colors.

### Design Principles

- Dark background (terminal default or ANSI black) — matches the vast majority of developer terminal setups
- High contrast text for readability
- Accent colors for interactive elements (links, selected items, cursor)
- Muted colors for metadata (timestamps, secondary info)
- Red for errors, green for confirmations (status bar)
- Styled using Lip Gloss with ANSI 256 colors for broad terminal compatibility (not true-color, which some terminals don't support)

### Color Palette (Approximate)

| Element | Color | Purpose |
|---------|-------|---------|
| Background | Terminal default | Respect user's terminal bg |
| Primary text | White/bright white | Post content |
| Secondary text | Gray/dim | Timestamps, metadata |
| Username | Cyan/bright cyan | Author handles |
| Accent | Magenta/bright magenta | Selected items, active elements |
| Error | Red | Status bar errors |
| Success | Green | Status bar confirmations |
| Border | Dark gray | Post separators, panels |

### Why Not Configurable Themes

Theme systems are a feature rabbit hole. Defining a theme schema, parsing user configs, validating colors, handling edge cases (conflicting colors, missing keys), and documenting all of it adds scope without improving the core product. For MVP, one good theme is better than a mediocre theme engine.

## Consequences

### Positive

- Zero configuration required — works out of the box
- Consistent visual experience for all users (easier to debug, document, screenshot)
- No theme-related code complexity

### Negative

- Users with light terminal backgrounds may have poor contrast
- No personalization — power users who want custom colors can't get them yet

### Neutral

- A theme system is planned for post-MVP. The Lip Gloss styles will be defined as a central `theme` struct, making it straightforward to swap palettes later.
