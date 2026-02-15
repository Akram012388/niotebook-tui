---
title: "TUI Layout & Navigation"
created: 2026-02-15
updated: 2026-02-15
status: accepted
tags: [design, tui, layout, navigation]
---

# TUI Layout & Navigation

## Screen Structure

The TUI uses a **header + content + status bar** layout. The header and status bar are persistent across all views. The content area swaps between views.

```
┌──────────────────────────────────────────────────────────────────────────┐
│  niotebook  @akram                                        Timeline  ↻   │  ← Header (1 line)
├──────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│                                                                          │
│                         [ View Content Area ]                            │
│                                                                          │
│                  Timeline | Profile | Login/Register                      │
│                                                                          │
│                                                                          │  ← Content (dynamic height)
├──────────────────────────────────────────────────────────────────────────┤
│  j/k: scroll  n: new post  r: refresh  ?: help                          │  ← Status bar (1 line)
└──────────────────────────────────────────────────────────────────────────┘
```

### Header (1 line)

| Position | Content | Notes |
|----------|---------|-------|
| Left | `niotebook` | App name, styled as brand |
| Center-left | `@username` | Logged-in user's handle. Hidden on login/register views. |
| Right | View name + indicator | `Timeline ↻` (↻ = refresh available), `Profile`, `Compose` |

### Content Area (terminal height - 2 lines)

Fills all remaining vertical space. Each view manages its own scrolling, layout, and input handling within this area.

### Status Bar (1 line)

Dual-purpose:
1. **Default state:** Context-sensitive key hints for the current view
2. **Feedback state:** Colored messages (error/success/loading) that replace hints temporarily

Feedback messages auto-clear after 5 seconds, reverting to key hints.

## Views

### 1. Login View

Shown on first launch or when session expires.

```
┌──────────────────────────────────────────────────────────────────────────┐
│  niotebook                                                    Login     │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│                                                                          │
│                          ┌─────────────────────┐                         │
│                          │       Login         │                         │
│                          │                     │                         │
│                          │  Email:             │                         │
│                          │  ┌─────────────────┐│                         │
│                          │  │                 ││                         │
│                          │  └─────────────────┘│                         │
│                          │                     │                         │
│                          │  Password:          │                         │
│                          │  ┌─────────────────┐│                         │
│                          │  │ ●●●●●●●●        ││                         │
│                          │  └─────────────────┘│                         │
│                          │                     │                         │
│                          │  [Enter] Login      │                         │
│                          │                     │                         │
│                          │  No account?        │                         │
│                          │  [Tab] Register     │                         │
│                          └─────────────────────┘                         │
│                                                                          │
├──────────────────────────────────────────────────────────────────────────┤
│  Tab: switch to register  Enter: submit  q: quit                         │
└──────────────────────────────────────────────────────────────────────────┘
```

### 2. Register View

Same layout as login, with additional fields.

```
                          ┌─────────────────────┐
                          │      Register       │
                          │                     │
                          │  Username:          │
                          │  ┌─────────────────┐│
                          │  │ akram           ││
                          │  └─────────────────┘│
                          │  Email:             │
                          │  ┌─────────────────┐│
                          │  │                 ││
                          │  └─────────────────┘│
                          │  Password:          │
                          │  ┌─────────────────┐│
                          │  │ ●●●●●●●●        ││
                          │  └─────────────────┘│
                          │                     │
                          │  [Enter] Register   │
                          │                     │
                          │  Have an account?   │
                          │  [Tab] Login        │
                          └─────────────────────┘
```

**Validation feedback:** Inline below each field. E.g., if username is taken: red text `"Username already taken"` below the username input.

### 3. Timeline View (Home)

The primary view after login. Scrollable list of post cards.

```
┌──────────────────────────────────────────────────────────────────────────┐
│  niotebook  @akram                                        Timeline  ↻   │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  ▸ @code_ninja · 2m                                                      │
│    Just shipped v0.3.0 of my CLI tool. Feels good.                       │
│  ────────────────────────────────────────────────────────────────────── │
│    @devgirl · 5m                                                         │
│    Anyone else using Claude Code? It's wild.                             │
│  ────────────────────────────────────────────────────────────────────── │
│    @akram · 12m                                                          │
│    Building a social media platform in the terminal. Yes, really.        │
│  ────────────────────────────────────────────────────────────────────── │
│    @rustacean · 18m                                                      │
│    TIL: Go 1.22 has pattern matching in net/http. No more gorilla/mux.  │
│  ────────────────────────────────────────────────────────────────────── │
│    @linux_guru · 25m                                                     │
│    My .vimrc is now longer than most of my projects.                     │
│                                                                          │
├──────────────────────────────────────────────────────────────────────────┤
│  j/k: scroll  n: new post  r: refresh  Enter: view profile  ?: help     │
└──────────────────────────────────────────────────────────────────────────┘
```

**Key behaviors:**
- `▸` marker indicates the currently selected post
- Selected post may have a subtle background highlight or border color change
- `j`/`k` moves selection up/down through posts
- Scrolling is smooth — viewport follows the selected post
- When reaching the bottom, loads the next page (cursor pagination)
- `r` fetches fresh data from the top

### 4. Compose Modal (Overlay)

Triggered by `n` from the timeline. Renders as a centered bordered box over the timeline (timeline is dimmed behind it).

```
┌──────────────────────────────────────────────────────────────────────────┐
│  niotebook  @akram                                        Timeline  ↻   │
├──────────────────────────────────────────────────────────────────────────┤
│  ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ │
│  ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ │
│  ░░░░░░░░  ┌─── New Post ──────────────────────────────┐  ░░░░░░░░░░░░ │
│  ░░░░░░░░  │                                           │  ░░░░░░░░░░░░ │
│  ░░░░░░░░  │  What's on your mind?                     │  ░░░░░░░░░░░░ │
│  ░░░░░░░░  │  ┌───────────────────────────────────────┐│  ░░░░░░░░░░░░ │
│  ░░░░░░░░  │  │ Building a social media platform in   ││  ░░░░░░░░░░░░ │
│  ░░░░░░░░  │  │ the terminal. Yes, really.█           ││  ░░░░░░░░░░░░ │
│  ░░░░░░░░  │  └───────────────────────────────────────┘│  ░░░░░░░░░░░░ │
│  ░░░░░░░░  │                                  98/140   │  ░░░░░░░░░░░░ │
│  ░░░░░░░░  │  Ctrl+Enter: post    Esc: cancel          │  ░░░░░░░░░░░░ │
│  ░░░░░░░░  └───────────────────────────────────────────┘  ░░░░░░░░░░░░ │
│  ░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░ │
├──────────────────────────────────────────────────────────────────────────┤
│  Ctrl+Enter: publish  Esc: cancel                                        │
└──────────────────────────────────────────────────────────────────────────┘
```

**Key behaviors:**
- Character counter updates live: `98/140`
- Counter turns red when within 10 chars of limit (`131/140`)
- Counter turns red and shows negative when over limit (`145/140` — Ctrl+Enter is disabled)
- Textarea supports multi-line editing, word wrap
- `Esc` cancels without confirmation (140 chars is short, low cost to re-type)
- On successful post: modal closes, timeline refreshes, status bar shows green `"Post published."`

### 5. Profile View

Shown when pressing `Enter` on a selected post (views that post's author), or when pressing `p` (views own profile).

```
┌──────────────────────────────────────────────────────────────────────────┐
│  niotebook  @akram                                        Profile       │
├──────────────────────────────────────────────────────────────────────────┤
│                                                                          │
│  @code_ninja                                                             │
│  Code Ninja                                                              │
│  ──────────────────────────────────────────────────────────────────────  │
│  Building tools for developers. Go enthusiast. Open source contributor.  │
│  Joined Feb 2026                                                         │
│  ──────────────────────────────────────────────────────────────────────  │
│                                                                          │
│  Posts                                                                   │
│  ──────────────────────────────────────────────────────────────────────  │
│  · 2m                                                                    │
│    Just shipped v0.3.0 of my CLI tool. Feels good.                       │
│  ──────────────────────────────────────────────────────────────────────  │
│  · 1h                                                                    │
│    Hot take: Makefiles are still the best build tool for Go projects.    │
│  ──────────────────────────────────────────────────────────────────────  │
│  · 3h                                                                    │
│    Reading the Go 1.22 release notes. Pattern matching in net/http!      │
│                                                                          │
├──────────────────────────────────────────────────────────────────────────┤
│  j/k: scroll  e: edit bio (own profile)  Esc: back to timeline           │
└──────────────────────────────────────────────────────────────────────────┘
```

**Own profile** shows an `[e] Edit` option. Opens an inline modal (like compose) for editing display name and bio.

## View Transitions

```
                    ┌──────────────┐
        ┌──────────│  Login View  │──────────┐
        │ Tab      └──────┬───────┘          │ Enter (success)
        ▼                 │ Tab               ▼
┌──────────────┐          │          ┌───────────────┐
│ Register View│──────────┘          │ Timeline View │◄──── Esc (from profile)
└──────────────┘                     │               │◄──── Esc/Post (from compose)
        │ Enter (success)            │               │
        └───────────────────────────►│               ├───── n ────► Compose Modal
                                     │               │                    │
                                     │               │◄──── Esc/Ctrl+Enter┘
                                     │               │
                                     │               ├───── Enter ──► Profile View
                                     │               │                    │
                                     │               │◄──── Esc ──────────┘
                                     │               │
                                     │               ├───── p ────► Own Profile
                                     │               │                    │
                                     │               │◄──── Esc ──────────┘
                                     └───────────────┘

Auth expired (401 + refresh fail) → Login View (from any view)
```

## Terminal Resize Handling

- All views re-render on `WindowSizeMsg` from Bubble Tea
- Header and status bar maintain fixed 1-line height
- Content area recalculates available height: `termHeight - 2`
- Post cards re-wrap content to new terminal width
- Compose modal re-centers and resizes (min width: 40 cols, max: 80 cols)
- If terminal is too small (< 40 cols or < 10 rows), show a message: `"Terminal too small. Resize to at least 40x10."`

## Scroll Behavior

- **Viewport scrolling** via Bubbles `viewport` component
- `j`/`k` moves the cursor (selected post), viewport follows
- `G` jumps to bottom (newest loaded), `g` jumps to top (oldest loaded)
- When cursor reaches the last loaded post and moves down: triggers pagination fetch (loads next page)
- During fetch: spinner appears at the bottom of the list
- `Home`/`End` keys also work for jumping to top/bottom
