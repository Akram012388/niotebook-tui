---
title: "Key Bindings"
created: 2026-02-15
updated: 2026-02-15
status: accepted
tags: [design, tui, keybindings]
---

# Key Bindings

All key bindings follow Vim conventions where applicable. Every action is keyboard-accessible. Mouse support is not included in MVP.

## Global Keys (Available in All Views)

| Key | Action |
|-----|--------|
| `q` | Quit application (with confirmation if composing) |
| `?` | Toggle help overlay (shows all keybindings for current view) |
| `Ctrl+c` | Force quit (no confirmation) |

## Login / Register View

| Key | Action |
|-----|--------|
| `Tab` | Switch between Login and Register forms |
| `Enter` | Submit form (login or register) |
| `Tab` (within form) | Move to next input field |
| `Shift+Tab` | Move to previous input field |
| `q` | Quit application |

## Timeline View

| Key | Action |
|-----|--------|
| `j` / `↓` | Move selection down (next post) |
| `k` / `↑` | Move selection up (previous post) |
| `g` / `Home` | Jump to top of loaded posts |
| `G` / `End` | Jump to bottom of loaded posts |
| `n` | Open compose modal (new post) |
| `r` | Refresh timeline (fetch latest posts) |
| `Enter` | View selected post's author profile |
| `p` | View own profile |
| `Space` / `Page Down` | Scroll down one page |
| `b` / `Page Up` | Scroll up one page |

### Why These Specific Bindings

- `j`/`k` — Vim movement, universal in TUI apps (lazygit, k9s, tig)
- `n` — "new" — common in email clients and TUI tools
- `r` — "refresh" — intuitive mnemonic
- `g`/`G` — Vim top/bottom, familiar to target audience
- `Enter` — universal "select/open"
- `p` — "profile" — avoids conflict with other bindings
- `Space`/`b` — `less` pager conventions (page down/up)

## Compose Modal

| Key | Action |
|-----|--------|
| `Ctrl+Enter` | Publish post |
| `Esc` | Cancel and close modal |
| Regular typing | Edit post content |
| `Ctrl+a` | Move cursor to start of line |
| `Ctrl+e` | Move cursor to end of line |
| `Ctrl+k` | Delete to end of line |
| `Ctrl+u` | Delete to start of line |
| `Ctrl+w` | Delete previous word |

Note: `Ctrl+Enter` is used instead of `Enter` because the textarea needs `Enter` for newlines (even though 140 chars rarely needs them, the option should exist).

## Profile View

| Key | Action |
|-----|--------|
| `j` / `↓` | Scroll down through user's posts |
| `k` / `↑` | Scroll up through user's posts |
| `g` / `Home` | Jump to top |
| `G` / `End` | Jump to bottom |
| `Esc` | Return to timeline |
| `e` | Edit bio/display name (own profile only) |

## Edit Profile Modal

| Key | Action |
|-----|--------|
| `Tab` | Move between display name and bio fields |
| `Shift+Tab` | Move to previous field |
| `Ctrl+Enter` | Save changes |
| `Esc` | Cancel and close modal |

## Help Overlay

| Key | Action |
|-----|--------|
| `?` | Close help overlay |
| `Esc` | Close help overlay |
| `q` | Close help overlay |

## Design Notes

### Conflict Resolution

- `q` quits globally EXCEPT when a text input is focused (compose modal, edit profile, login fields). In those contexts, `q` types the character `q`.
- `Esc` always closes/cancels the current modal or overlay, never quits.
- Arrow keys are always available as alternatives to `j`/`k` for users who don't know Vim.

### Future Bindings (Reserved, Not Implemented in MVP)

These keys are intentionally unused in MVP to avoid conflicts with planned features:

| Key | Reserved For |
|-----|-------------|
| `l` | Like post (v1.1) |
| `t` | Repost (v1.1) |
| `/` | Search (v1.1) |
| `f` | Follow/unfollow user on profile (v1.1) |
| `d` | DMs view (v1.2+) |
| `1`-`9` | Quick-switch between views (post-MVP) |
