---
title: "ADR-0016: XDG Config Directory"
status: accepted
created: 2026-02-15
updated: 2026-02-15
tags: [adr, tui, config]
---

# ADR-0016: XDG Config Directory

## Status

Accepted

## Context

The TUI client needs to persist local state: server URL, JWT tokens, and user preferences. This needs a home on the filesystem.

## Decision

Store all TUI configuration in **`~/.config/niotebook/`**, following the XDG Base Directory Specification.

### File Layout

```
~/.config/niotebook/
├── config.yaml      # User-editable settings (server URL, preferences)
└── auth.json        # JWT tokens (access + refresh). Not user-edited.
```

### config.yaml

```yaml
server_url: "https://api.niotebook.com"
```

For MVP, this is minimal. Future additions: keybindings, theme overrides, default view.

### auth.json

```json
{
  "access_token": "eyJ...",
  "refresh_token": "eyJ...",
  "expires_at": "2026-02-16T10:00:00Z"
}
```

Separated from `config.yaml` because:
- Tokens are machine-managed (written by the TUI after login), not user-edited
- Users may want to version-control or share `config.yaml` (e.g., in dotfiles repos) without exposing tokens
- File permissions can be set more restrictively on `auth.json` (0600)

### XDG Compliance

- If `$XDG_CONFIG_HOME` is set, use `$XDG_CONFIG_HOME/niotebook/` instead of `~/.config/niotebook/`
- The TUI creates the directory on first run if it doesn't exist

## Consequences

### Positive

- Follows platform conventions — Linux users expect `~/.config/`
- Clean separation between user config and machine-managed state
- Token file can be excluded from dotfile repos easily
- Works on macOS (where `~/.config/` is increasingly standard for CLI tools)

### Negative

- Windows doesn't follow XDG (would need `%APPDATA%\niotebook\` — handled if Windows support is added)

### Neutral

- macOS has `~/Library/Application Support/` as the "correct" location, but `~/.config/` is the de facto standard for CLI tools on macOS. Most Go CLI tools (gh, lazygit) use `~/.config/`.
