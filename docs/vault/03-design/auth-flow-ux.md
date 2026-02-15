---
title: "Auth Flow UX"
created: 2026-02-15
updated: 2026-02-15
status: accepted
tags: [design, tui, auth, ux]
---

# Auth Flow UX

## Startup Flow

```
TUI launches
  │
  ├─ Config exists (~/.config/niotebook/auth.json)?
  │   │
  │   ├─ Yes: access token present?
  │   │   │
  │   │   ├─ Yes: access token expired?
  │   │   │   │
  │   │   │   ├─ No: → Timeline View (instant)
  │   │   │   │
  │   │   │   └─ Yes: refresh token valid?
  │   │   │       │
  │   │   │       ├─ Yes: refresh tokens → Timeline View
  │   │   │       │
  │   │   │       └─ No: → Login View ("Session expired")
  │   │   │
  │   │   └─ No: → Login View
  │   │
  │   └─ No: → Login View
  │
  └─ No: → Login View (first launch)
```

## First Launch Experience

1. TUI opens to **Login View** with the Register tab hint visible
2. New users press `Tab` to switch to **Register View**
3. User fills in username, email, password
4. On success: config directory and files are created, JWT stored, transition to Timeline

### Registration Validation

Validation happens in two stages:

**Client-side (instant, as user types):**
- Username: show character count `3/15`, turn red if invalid characters typed
- Email: basic format check (contains `@` and `.`)
- Password: minimum 8 characters, show character count

**Server-side (on submit):**
- Username uniqueness: `"Username already taken"`
- Email uniqueness: `"Email already registered"`
- Password strength: server may reject weak passwords

Server errors appear as red text below the relevant field, inline in the form.

## Login Flow

1. User enters email + password
2. Presses `Enter`
3. Status bar: `"Logging in..."`
4. On success:
   - JWT pair written to `~/.config/niotebook/auth.json`
   - Transition to Timeline View
   - Status bar: green `"Welcome back, @akram"`
5. On failure:
   - Status bar: red `"Invalid email or password."`
   - Form fields remain filled (except password, which is cleared)
   - Cursor returns to password field

## Token Refresh Flow

Happens transparently — the user never sees it unless it fails.

1. TUI makes an API request
2. Server returns `401 Unauthorized`
3. TUI checks: is there a refresh token?
   - **Yes:** POST `/api/v1/auth/refresh` with refresh token
     - **Success:** New token pair stored, original request retried automatically
     - **Failure:** Transition to Login View with `"Session expired. Please log in again."`
   - **No:** Transition to Login View

This retry logic lives in the HTTP client wrapper (`internal/tui/client/`). All API calls go through it, so refresh is transparent to view code.

## Logout

MVP does not have an explicit logout action (no button/keybinding). To "log out," the user can:
- Delete `~/.config/niotebook/auth.json`
- Or quit and re-launch (expired token triggers login)

Explicit logout (`Ctrl+l` or similar) is a post-MVP addition.

## Password Input

- Characters are masked as `●` (bullet) in the terminal
- No password visibility toggle in MVP
- Passwords are sent over HTTPS, never logged, never stored in config

## Config File Creation

On first successful registration or login:

```
mkdir -p ~/.config/niotebook/
```

Write `config.yaml`:
```yaml
server_url: "https://api.niotebook.com"
```

Write `auth.json` (permissions: 0600):
```json
{
  "access_token": "eyJ...",
  "refresh_token": "eyJ...",
  "expires_at": "2026-02-16T22:00:00Z"
}
```

## Server URL Configuration

The TUI needs to know where the server is. For MVP:

1. **Default:** `https://api.niotebook.com` (production server)
2. **Override:** `--server` CLI flag: `niotebook --server http://localhost:8080`
3. **Persistent override:** Set `server_url` in `~/.config/niotebook/config.yaml`

Priority: CLI flag > config file > default.

This allows developers to run local server instances during development.
