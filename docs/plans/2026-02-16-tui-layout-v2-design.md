# Niotebook TUI Layout v2 — Design Document

**Date:** 2026-02-16
**Status:** Draft
**Branch:** `feature/tui-redesign`
**Predecessor:** `docs/plans/2026-02-16-tui-redesign-design.md` (v1 — approved & implemented)

## Vision

Evolve the three-column layout from v1 into a faithful terminal adaptation of X (Twitter) for macOS. The left column becomes a full nav hub, the center column gets an always-visible inline compose bar, and the right column becomes a discover/trending space. Column navigation via Tab/Shift+Tab makes the entire interface keyboard-navigable.

**Tagline update:** "the social terminal" (replaces "a social notebook")

**Design philosophy (unchanged):** "Black & white with a shot of terracotta."

---

## 1. Splash Screen Enhancements

### 1.1 Timing

- Minimum display: **2.5 seconds** (up from 1.5s)
- Server health check runs in background; splash stays visible regardless of response speed

### 1.2 Layout — Larger Presence

Full-screen centered. Logo and tagline rendered with **letter-spacing** for a larger, more dramatic feel.

```
              n i o t e b o o k

              t h e   s o c i a l   t e r m i n a l


                     █ ░ ░
                  connecting...
```

- Logo: each character separated by a space. Bold. Letter `i` in terracotta accent.
- Tagline: each character separated by a space. TextMuted.
- Spinner: custom square block animation (see 1.3)
- Status text: TextSecondary

### 1.3 Custom Block Spinner

Replace `spinner.Dot` (circles) with a custom terminal-native block spinner:

```
Frame 0:  ░ ░ ░
Frame 1:  █ ░ ░
Frame 2:  █ █ ░
Frame 3:  █ █ █
Frame 4:  ░ ░ ░    (cycle restarts)
```

- `░` (light shade U+2591) in Border color
- `█` (full block U+2588) in Accent color
- Frame duration: ~300ms per frame
- Spacing between blocks for visual breathing room

### 1.4 States (unchanged from v1)

| State      | Spinner | Text                | Next             |
|------------|---------|---------------------|------------------|
| Connecting | Active  | "connecting..."     | → Connected      |
| Connected  | Stop    | (auto-transition)   | → Login/Timeline |
| Failed     | Stop    | "connection failed" | Show retry hint  |

---

## 2. Left Column — Nav Hub (X-style)

### 2.1 Rationale

Following X's pattern: all navigation, settings, and app chrome live in the left column. The center column is purely content. Keyboard shortcuts move from the right column to here.

### 2.2 Layout

```
 niotebook                    ← bold wordmark, 'i' in terracotta
 ─────────────────

 @akram                       ← Accent color
 akram                        ← Text color (display name)

 ● Home                       ← active: Accent + filled dot
   Profile
   Bookmarks                  ← TextMuted (placeholder)
   Settings                   ← TextMuted (placeholder)

 ┌─────────────────┐
 │      Post        │          ← Accent border, triggers compose
 └─────────────────┘

 ─────────────────
 v0.1.0-dev
 Joined Feb 2026

 Shortcuts
 j/k    scroll
 Tab    switch col
 n      compose
 ?      help
 q      quit
```

### 2.3 Components

| Element | Style | Behavior |
|---------|-------|----------|
| Wordmark | Bold + Text, `i` in Accent | Static brand |
| Username | Accent color, `@` prefix | Static |
| Display name | Text color | Static |
| Nav items | Active: Accent + `●` prefix. Inactive: TextSecondary | j/k navigable when column focused |
| Placeholder items | TextMuted | Not selectable, shown greyed out |
| Post button | AccentDim rounded border, Accent text | Enter activates → focuses inline compose |
| Version | Caption style (TextSecondary) | Static |
| Shortcuts | Key style (Bold + Accent) for keys, Caption for descriptions | Static reference |

### 2.4 Focused State

When the left column has Tab focus:
- Column border changes to AccentDim
- j/k moves highlight between nav items (Home, Profile, Bookmarks, Settings)
- Enter activates the highlighted item
- Post button is reachable via j/k and Enter

When unfocused:
- Active nav item still shown with `●` marker
- No highlight cursor visible

---

## 3. Center Column — Feed + Inline Compose

### 3.1 Rationale

Replace the modal compose overlay with an always-visible inline compose bar at the top of the feed, exactly like X's "What's happening?" input. The bottom status bar tooltip is removed (redundant — shortcuts are in the left column).

### 3.2 Compose Bar — Collapsed State (always visible)

```
 ┌─────────────────────────────────────────────────────────┐
 │  What's on your mind?                          0/140    │
 └─────────────────────────────────────────────────────────┘
```

- Single-line height
- AccentDim rounded border
- Placeholder text in TextMuted
- Character counter in TextSecondary (right-aligned)
- Pressing `n` (from any column) or `Enter` (when bar is highlighted) → expands

### 3.3 Compose Bar — Expanded State

```
 ┌─────────────────────────────────────────────────────────┐
 │  What's on your mind?                                   │
 │                                                         │
 │  This is my new post and the textarea grows             │
 │  as I type more content across multiple lines...        │
 │                                                         │
 │  89/140              Ctrl+Enter: post    Esc: cancel    │
 └─────────────────────────────────────────────────────────┘
```

- Border becomes Accent (brighter when active)
- Textarea **grows in height** as content is typed (minimum 3 lines, maximum ~8 lines)
- Character counter: TextSecondary normally, Error when >130
- Hints (Ctrl+Enter / Esc) in Hint style (TextMuted + italic)
- All hints inside the compose box — no external tooltip needed
- `Esc` collapses back to single-line bar
- `Ctrl+Enter` publishes and collapses

### 3.4 Feed

Below the compose bar, the niotes feed renders exactly as v1:

```
 ▸ @akram · now
   testing testing testing testing testing
 ─────────────────────────────────────────────────────────
   @akram · 2m
   checking 123
 ─────────────────────────────────────────────────────────
```

- Post cards unchanged from v1 (marker, username, timestamp, content, separator)
- j/k scrolls through posts when center column is focused and compose is collapsed
- Enter on a post opens that user's profile

### 3.5 Status Bar — Removed

The bottom tooltip bar (`j/k: navigate  n: compose  r: refresh  ?: help  q: quit`) is **removed**. Shortcuts now live in the left column sidebar. Status messages (errors, success confirmations) can briefly flash at the bottom of the compose bar or be shown inline.

---

## 4. Right Column — Discover + Trending

### 4.1 Rationale

The right column transforms from a static shortcuts panel into a dynamic discover/trending space. All content is placeholder for MVP — backend support for search, trending, and recommendations comes later.

### 4.2 Layout

```
 ┌─────────────────────┐
 │  Search niotes...   │      ← placeholder search bar
 └─────────────────────┘

 Trending                      ← section header (Label style)
 ─────────────────────
 #niotebook                    ← Accent color
 12 niotes                     ← Caption style

 #hello-world
 8 niotes

 #terminal-life
 5 niotes

 Writers to follow             ← section header (Label style)
 ─────────────────────
 @alice                        ← Accent color
 loves terminal apps           ← Caption style

 @bob
 building in public
```

### 4.3 Components

| Element | Style | Behavior |
|---------|-------|----------|
| Search bar | AccentDim rounded border, TextMuted placeholder | Non-functional placeholder for MVP |
| Section headers | Label style (Bold + Text) | Static |
| Trending tags | Accent color for `#tag`, Caption for count | j/k navigable when column focused (future) |
| Writer names | Accent color for `@name`, Caption for bio | j/k navigable when column focused (future) |
| Separators | Border color dashed line | Visual divider |

### 4.4 Focused State

When the right column has Tab focus:
- Column border changes to AccentDim
- j/k moves highlight between trending items and suggested writers
- Enter on a trending tag → filters feed (future)
- Enter on a writer → opens their profile (future)

For MVP: focus is visual only (accent border), items are not yet interactive.

---

## 5. Column Navigation

### 5.1 Focus Model

Three focusable columns. Center is the default. Tab cycles forward, Shift+Tab cycles backward.

```
           Tab →                Tab →
  ┌──────────┐       ┌──────────────┐       ┌──────────┐
  │  Left    │  →    │   Center     │  →    │  Right   │
  │  Nav Hub │       │   Feed       │       │  Discover│
  └──────────┘       └──────────────┘       └──────────┘
           ← Shift+Tab           ← Shift+Tab
```

### 5.2 Visual Indicator

| State | Border Style |
|-------|-------------|
| Unfocused column | Border color (subtle) |
| Focused column | AccentDim color (visible but not loud) |

### 5.3 Key Behavior by Column

| Key | Left (Nav) | Center (Feed) | Right (Discover) |
|-----|-----------|---------------|-----------------|
| `j/k` | Move nav highlight | Scroll posts | Scroll sections |
| `Enter` | Activate nav item / Post button | Open post author's profile | Activate item (future) |
| `n` | Focus center + open compose | Open compose (if collapsed) | Focus center + open compose |
| `?` | Open help overlay | Open help overlay | Open help overlay |
| `q` | Quit | Quit | Quit |
| `Esc` | Return focus to center | Close compose (if open) | Return focus to center |

### 5.4 Compose Focus Override

When compose is expanded:
- All keyboard input routes to the compose textarea
- `Tab`/`Shift+Tab` are disabled (they're valid text input characters in some contexts)
- Only `Esc` (cancel) and `Ctrl+Enter` (publish) escape the compose
- After publish/cancel, focus returns to center column feed

---

## 6. Brand Wordmark Fix

### 6.1 Current Problem

The wordmark uses a different rendering style (middle dot `·`, character substitution) that makes it look like a different font from the rest of the UI.

### 6.2 Fix

- Render `niotebook` as a single word in the **same typeface as all body text**
- Apply **Bold** weight
- Color the letter `i` in **Accent** (terracotta)
- All other letters in **Text** color (white in dark mode)
- No middle dot, no character substitution, no special spacing (except in splash screen where letter-spacing is applied to ALL characters for dramatic effect)

### 6.3 Splash vs. Sidebar

| Context | Rendering |
|---------|-----------|
| Splash screen | `n i o t e b o o k` — spaced out, bold, `i` in Accent |
| Left sidebar | `niotebook` — normal spacing, bold, `i` in Accent |

---

## 7. Files to Create/Modify

### New/Modified Files

| File | Change |
|------|--------|
| `internal/tui/theme/logo.go` | Fix wordmark: bold, `i` accent, remove `·`. Add `LogoSplash()` with letter-spacing. |
| `internal/tui/theme/theme.go` | Update tagline to "the social terminal" |
| `internal/tui/views/splash.go` | 2.5s minimum, custom block spinner, spaced logo/tagline |
| `internal/tui/components/sidebar.go` | Full nav hub: nav items, Post button, shortcuts, version, placeholders |
| `internal/tui/components/shortcuts.go` | **Remove** or repurpose — shortcuts move to left sidebar |
| `internal/tui/components/discover.go` | **New** — Right column: search placeholder, trending, writers to follow |
| `internal/tui/views/compose.go` | Rewrite as inline compose bar (not modal). Auto-expanding textarea. |
| `internal/tui/components/compose_bar.go` | **New** — Inline compose bar component (collapsed/expanded states) |
| `internal/tui/app/app.go` | Column focus state, Tab/Shift+Tab routing, remove status bar, inline compose integration |
| `internal/tui/app/messages.go` | Add column focus messages if needed |
| `internal/tui/components/statusbar.go` | **Remove** — bottom tooltip bar eliminated |
| `internal/tui/layout/columns.go` | Add focus state tracking to Columns struct |

### Removed Components
- `components/statusbar.go` — bottom tooltip bar (redundant)
- `components/shortcuts.go` — right panel shortcuts (moved to left sidebar)

---

## 8. Non-Goals (for this iteration)

- No functional search (placeholder only)
- No real trending data (hardcoded placeholder)
- No real user recommendations (hardcoded placeholder)
- No mouse support
- No right-column item interaction (visual focus only)
- No bookmarks or settings screens (placeholder nav items only)

---

## 9. Summary of Changes from v1

| Area | v1 (current) | v2 (this design) |
|------|-------------|-----------------|
| Splash | 1.5s, dot spinner, `n·otebook` | 2.5s, block spinner, spaced `n i o t e b o o k` |
| Tagline | "a social notebook" | "the social terminal" |
| Left column | Profile + 2 nav items + join date | Full nav hub (X-style): nav, Post button, shortcuts, version |
| Center column | Feed + bottom status bar | Inline compose bar + feed. No status bar. |
| Right column | Keyboard shortcuts | Discover: search, trending, writers to follow |
| Compose | Modal overlay centered on screen | Inline bar at top of feed, auto-expanding |
| Column nav | None (center always focused) | Tab/Shift+Tab with accent border indicators |
| Wordmark | `n·otebook` (middle dot) | `niotebook` (bold, `i` in accent) |

---

## References

- X (Twitter) for macOS — primary layout reference (see attached screenshot)
- [terminal.shop](https://www.terminal.shop) — terminal aesthetic inspiration
- v1 design doc: `docs/plans/2026-02-16-tui-redesign-design.md`
