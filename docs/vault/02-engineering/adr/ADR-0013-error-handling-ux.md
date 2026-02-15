---
title: "ADR-0013: Inline Status Bar Error Handling"
status: accepted
created: 2026-02-15
updated: 2026-02-15
tags: [adr, tui, ux]
---

# ADR-0013: Inline Status Bar Error Handling

## Status

Accepted

## Context

The TUI must communicate errors (network failures, server 500s, validation errors, auth expiry) to the user without breaking workflow. Options:

1. **Inline status bar message** — colored text in the bottom bar
2. **Modal popup** — centered overlay that requires dismissal
3. **Toast notification** — auto-dismissing message

## Decision

Use the **bottom status bar** as the primary feedback channel for all transient messages: errors, confirmations, and loading state.

### Behavior

- **Errors** render in red: `"Failed to publish post. Press r to retry."`
- **Confirmations** render in green: `"Post published."`
- **Loading** renders in dim/gray: `"Refreshing timeline..."`
- Messages persist until the next action replaces them, or auto-clear after 5 seconds
- The status bar is always visible at the bottom of every view

### Auth Expiry

Special case: when a 401 response indicates the JWT has expired and refresh fails, the TUI transitions to the login view with a status message: `"Session expired. Please log in again."`

### Why Not Modals

Modals interrupt flow. In a keyboard-driven TUI, a modal requires the user to press a key to dismiss before doing anything else. For non-critical errors (network timeout, failed refresh), this is unnecessarily disruptive.

### Why Not Toasts

Auto-dismissing messages can be missed if the user isn't looking at the terminal. The status bar persists until replaced, ensuring the message is seen on the next glance.

## Consequences

### Positive

- Non-intrusive — errors don't block the UI or require explicit dismissal
- Consistent feedback location — users always know where to look
- Simple implementation — single status bar component shared across all views

### Negative

- Easy to overlook if the user isn't looking at the bottom of the screen
- Limited space for detailed error messages (single line)
- Cannot show multiple simultaneous errors

### Neutral

- Modals may be introduced later for critical confirmations (e.g., "Delete this post?" in v1.1 when delete is added). The status bar handles informational feedback; modals handle destructive action confirmations.
