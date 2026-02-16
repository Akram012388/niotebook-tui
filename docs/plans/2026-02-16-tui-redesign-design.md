# Niotebook TUI Redesign — Design Document

**Date:** 2026-02-16
**Status:** Approved
**Branch:** `feature/tui-redesign`

## Vision

Transform the Niotebook TUI from a functional prototype into a polished, terminal.shop-caliber experience. The app should feel like a native desktop social client — full-screen, visually distinctive, with a three-column layout inspired by X (Twitter) for macOS.

**Design philosophy:** "Black & white with a shot of terracotta." High contrast, minimal chrome, warm amber accent that makes the terminal feel alive.

---

## 1. Design System — Theme Package

**New package:** `internal/tui/theme/`

### 1.1 Brand Identity

- **Brand color:** Terracotta amber `#D97757` (matches Claude's brand warmth + terminal.shop's orange energy)
- **Logo treatment:** `n·otebook` — only the dot (replacing `i`) is terracotta, all other characters are white
- **Tagline:** "a social notebook"

### 1.2 Color Palette (Adaptive)

All colors use `lipgloss.AdaptiveColor{}` for automatic light/dark terminal detection.

| Token           | Dark Terminal | Light Terminal | Usage                                    |
|-----------------|---------------|----------------|------------------------------------------|
| `Accent`        | `#D97757`     | `#C15F3C`      | Logo dot, selected items, active nav     |
| `AccentDim`     | `#B85C3A`     | `#A04E30`      | Active panel borders, secondary accent   |
| `Text`          | `#FAFAF9`     | `#141413`       | Primary body text                        |
| `TextSecondary` | `#A8A29E`     | `#57534E`       | Timestamps, metadata, counts             |
| `TextMuted`     | `#57534E`     | `#A8A29E`       | Hints, placeholders, disabled text       |
| `Border`        | `#44403C`     | `#D6D3D1`       | Panel dividers, separators               |
| `Surface`       | `#141413`     | `#FAFAF9`       | Main background                          |
| `SurfaceRaised` | `#1C1917`     | `#F5F5F4`       | Sidebar background, cards, modals        |
| `Error`         | `#EF4444`     | `#DC2626`       | Error messages                           |
| `Success`       | `#22C55E`     | `#16A34A`       | Success confirmations                    |
| `Warning`       | `#FBBF24`     | `#D97706`       | Warnings                                 |

### 1.3 Typography Tokens

| Token     | Style               | Usage                              |
|-----------|---------------------|------------------------------------|
| `Heading` | Bold + Accent       | View titles, section headers       |
| `Label`   | Bold + Text         | Form labels, nav items             |
| `Body`    | Text                | Post content, descriptions         |
| `Caption` | TextSecondary       | Timestamps, counts, metadata       |
| `Hint`    | TextMuted + Italic  | Keyboard shortcuts, placeholders   |
| `Key`     | Bold + Accent       | Keyboard key references in help    |

### 1.4 Border Tokens

| Token          | Style                           | Usage                    |
|----------------|---------------------------------|--------------------------|
| `PanelBorder`  | Rounded, Border color           | Column dividers          |
| `ActiveBorder` | Rounded, AccentDim color        | Active/focused panels    |
| `ModalBorder`  | Rounded, Accent color           | Compose, help overlays   |
| `Separator`    | Dashed line, Border color       | Between posts, sections  |

---

## 2. Splash Screen

### 2.1 Behavior

- Shows **immediately** on launch — no "connecting to..." text visible
- Server connection happens in the background via `tea.Cmd`
- Animated spinner (charm `spinner.Dot`) shows connection progress
- Auto-transitions to login (if no stored session) or timeline (if session valid)
- If connection fails: shows error on splash screen with retry option

### 2.2 Layout

Full-screen centered, no header/sidebar/status bar — just the brand.

```
                    n · o t e b o o k

                    a social notebook

                         ●○○
                      connecting...
```

- `·` in terracotta, all else in white
- Tagline in TextMuted
- Spinner in terracotta
- Status text in TextSecondary

### 2.3 States

| State         | Spinner | Text                | Next              |
|---------------|---------|---------------------|-------------------|
| Connecting    | Active  | "connecting..."     | → Connected       |
| Connected     | Stop    | (auto-transition)   | → Login/Timeline  |
| Failed        | Stop    | "connection failed" | Show retry hint   |

---

## 3. Three-Column Layout

**New package:** `internal/tui/layout/`

Inspired by X (Twitter) for macOS — left nav, center content, right context.

### 3.1 Column Structure

```
┌──────────────┬────────────────────────────────┬────────────────┐
│  Left (20c)  │  Center (flexible)             │  Right (18c)   │
│  Profile+Nav │  Main Content                  │  Shortcuts     │
└──────────────┴────────────────────────────────┴────────────────┘
```

### 3.2 Left Sidebar (~20 chars fixed)

**Top: Profile Card**
- Username with `@` prefix in Accent
- Display name in Text
- Separator line

**Middle: Navigation**
- `● Home` (active indicator is terracotta filled dot)
- `  Profile`
- `  Compose`
- `  Help`
- `  Logout`

Active item: Accent color + `●` marker
Inactive items: TextSecondary

**Bottom: Stats**
- Separator line
- Post count in Caption
- Join date in Caption

**Behavior:**
- When logged out: shows only the logo and Login/Register nav
- Navigation items trigger view switches via keybindings (not click — vim-style `j/k` stays)

### 3.3 Center Content (flexible width)

The main content area — renders the active view:
- **Timeline:** Scrollable post list
- **Profile (detail):** User info + their posts
- **Compose (overlay):** Modal over center content
- **Help (overlay):** Modal over center content
- **Login/Register:** Auth forms (centered within the column)

### 3.4 Right Sidebar (~18 chars fixed)

**Context-sensitive keyboard shortcuts** that change based on the active view.

**Timeline context:**
```
 Navigation
 j/k   scroll
 g/G   top/bottom
 Enter  profile

 Actions
 n     compose
 r     refresh
 ?     help
 q     quit
```

**Compose context:**
```
 Compose
 Ctrl+J  publish
 Esc     cancel
```

**Profile context:**
```
 Navigation
 j/k   scroll
 Esc   back

 Actions
 e     edit
 ?     help
 q     quit
```

### 3.5 Responsive Breakpoints

| Terminal Width | Layout                                    |
|----------------|-------------------------------------------|
| ≥100 cols      | Three columns (sidebar + content + help)  |
| 80-99 cols     | Two columns (sidebar + content)           |
| <80 cols       | Single column (content only, status bar)  |

---

## 4. Component Redesign

### 4.1 Post Card

**Normal state:**
```
  @username · 2m ago
  Post content here with word wrapping
  that respects the center column width.
  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─
```

**Selected state:**
```
▸ @username · 2m ago
  Post content here with word wrapping
  that respects the center column width.
  ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─ ─
```

- Normal username: TextSecondary
- Selected username: Accent + Bold
- Selection marker `▸`: Accent
- Time: TextMuted
- Content: Text (body)
- Separator: Border color, dashed pattern

### 4.2 Header Bar

Removed as a standalone component — replaced by the three-column layout. The left sidebar now carries the branding and navigation that the header used to provide.

For narrow terminals (single-column fallback), a minimal header is retained:
```
n·otebook                              Timeline
```

### 4.3 Status Bar

Bottom of center column (not full-width anymore):
- Help hint on left in TextMuted
- Status messages (error/success/loading) on right, color-coded
- Auto-dismiss after 5 seconds (errors/success)

### 4.4 Spinner

Replace all static "Loading..." text with charm's `spinner.Dot` model:
- Terracotta colored
- Used in: splash screen, timeline loading, post publishing

### 4.5 Compose Modal

Overlay centered on the center column:

```
┌───────────────────────────────────┐
│                                   │
│  New Post                         │
│                                   │
│  What's on your mind?             │
│  ┌─────────────────────────────┐  │
│  │                             │  │
│  │                             │  │
│  │                             │  │
│  └─────────────────────────────┘  │
│                                   │
│  42/140                           │
│                                   │
│  Ctrl+J post    Esc cancel        │
│                                   │
└───────────────────────────────────┘
```

- Modal border: Accent color
- Title: Heading style
- Counter: Caption (normal), Error (when >130)
- Hint: Hint style

### 4.6 Auth Forms (Login/Register)

Centered within the center column area. Same modal style as compose but larger:
- Rounded border in AccentDim
- Form title in Heading
- Labels in Label
- Error messages in Error
- Hints in Hint

### 4.7 Help Overlay

Modal over center column:

```
┌───────────────────────────────────┐
│                                   │
│  Keyboard Shortcuts               │
│                                   │
│  j/k          Scroll up/down      │
│  n            New post            │
│  r            Refresh             │
│  Enter        View profile        │
│  p            Own profile         │
│  g/G          Top/bottom          │
│  ?            Close help          │
│  q            Quit                │
│                                   │
│  Press ? or Esc to close          │
│                                   │
└───────────────────────────────────┘
```

- Keys in Key style (Bold + Accent)
- Descriptions in Body
- Fixed-width key column for alignment

---

## 5. View Flow

```
Launch → Splash Screen (connect in background)
                ↓ connected
        ┌── Has session? ──┐
        │ no               │ yes
        ↓                  ↓
    Login/Register    Timeline (3-col)
        ↓ success          │
    Timeline (3-col) ──────┤
        │                  │
    Nav: Profile, Compose, Help, Logout
```

### 5.1 Navigation via Sidebar

The left sidebar nav items map to the same keybindings as today:
- Home/Timeline: default view
- Profile: `p` key
- Compose: `n` or `c` key (opens overlay)
- Help: `?` key (opens overlay)
- Logout: new feature (clears session, returns to login)

---

## 6. Files to Create/Modify

### New Files
- `internal/tui/theme/theme.go` — Color palette, typography, border tokens
- `internal/tui/theme/logo.go` — Brand logo rendering (n·otebook with terracotta dot)
- `internal/tui/layout/columns.go` — Three-column layout manager
- `internal/tui/layout/responsive.go` — Breakpoint detection and column toggling
- `internal/tui/views/splash.go` — Splash screen view model
- `internal/tui/components/sidebar.go` — Left sidebar (profile + nav)
- `internal/tui/components/shortcuts.go` — Right sidebar (context shortcuts)
- `internal/tui/components/spinner.go` — Themed spinner wrapper

### Modified Files
- `internal/tui/app/app.go` — Integrate layout manager, splash screen, sidebar state
- `internal/tui/app/messages.go` — Add splash/connection messages
- `internal/tui/views/timeline.go` — Use theme tokens, adapt to column width
- `internal/tui/views/compose.go` — Use theme tokens, themed modal
- `internal/tui/views/profile.go` — Use theme tokens, adapt layout
- `internal/tui/views/login.go` — Use theme tokens, center in content area
- `internal/tui/views/register.go` — Use theme tokens, center in content area
- `internal/tui/views/help.go` — Use theme tokens, themed modal
- `internal/tui/views/factory.go` — Add splash screen factory method
- `internal/tui/components/header.go` — Minimal header for narrow terminals only
- `internal/tui/components/postcard.go` — Use theme tokens
- `internal/tui/components/statusbar.go` — Use theme tokens, column-aware width
- `cmd/tui/main.go` — Start with splash screen instead of login

---

## 7. Non-Goals (for this sprint)

- No mouse/click support (vim-style keyboard only)
- No notification system or real-time updates
- No image/media rendering
- No right sidebar content beyond shortcuts (future: trending, suggestions)
- No custom themes or user-configurable colors
- No SSH delivery (like terminal.shop's `ssh` access)

---

## References

- [terminal.shop](https://www.terminal.shop) — Design inspiration (black/white + orange accent)
- [Charm blog: terminal.shop](https://charm.land/blog/terminaldotshop/) — How it was built
- [Claude brand colors](https://www.brandcolorcode.com/claude) — Terracotta `#D97757`
- [Lip Gloss](https://github.com/charmbracelet/lipgloss) — Styling library
- [Bubble Tea](https://github.com/charmbracelet/bubbletea) — TUI framework
