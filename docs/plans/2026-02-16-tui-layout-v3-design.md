# TUI Layout v3 Design — Component Rewrite

**Date:** 2026-02-16
**Branch:** `feature/tui-redesign`
**Approach:** Component Rewrite (rewrite sidebar, discover, postcard, columns layout, splash)
**Scope:** 5 component files + mock data, keeping app.go/views/theme/client untouched

---

## 1. Layout & Column Focus System

**Files:** `layout/columns.go`, `app/app.go`

### Active Column Header Bar

Each column renders a horizontal header line at the very top:

- **Active column:** Solid amber bar using `━` (U+2501) in Accent color (`#D97757`), full column width
- **Inactive columns:** Thin dim line using `─` (U+2500) in Border color (`#44403C`), full column width
- Rendered by the layout system (not individual components) for perfect cross-column alignment

### Top Alignment

All three columns share the same content-start Y-position below header bars:

- Left column: ASCII logo starts on row 2
- Center column: compose bar starts on row 2
- Right column: search bar starts on row 2

### Tab Navigation

- `Tab` cycles columns: Left → Center → Right → Left
- `Shift+Tab` cycles reverse: Right → Center → Left → Right
- When **center** is active: j/k navigates posts, n opens compose
- When **left** is active: j/k navigates nav items (Home, Profile, Bookmarks, Settings)
- When **right** is active: j/k scrolls within active section, `Enter` toggles between Trending/Writers sub-sections

---

## 2. Left Column (Sidebar) Rewrite

**Files:** `components/sidebar.go`, `theme/logo.go`

### Layout (top to bottom)

```
━━━━━━━━━━━━━━━━━━━━  ← amber header bar (when active)

 ┐ •        ┐        ← ASCII art "niotebook" logo
 │ ┐ ┌┐┼┌┐  │        (3-5 lines, i dot in amber)
 ┘ │ └┘ └┘└┘└┘ │┌    (figlet-style, fits ~20 chars)

● Home                ← nav items (j/k when focused)
  Profile
  Bookmarks           ← greyed out (TextMuted)
  Settings            ← greyed out (TextMuted)

─────────────────────
@akram                ← Accent
Akram                 ← Text
Joined Feb 2026       ← TextSecondary
─────────────────────
12 niotes · 3 following  ← TextMuted stats
─────────────────────


                      ← flexible whitespace


j/k  scroll           ← shortcuts (pushed to bottom)
g/G  top/bottom
Tab  switch col
n    compose
?    help
q    quit
```

### Key changes from v2

- **ASCII art logo:** Multi-line figlet-style, i dot in amber
- **Post button removed:** compose via `n` key
- **Version string removed:** available via `--version` flag
- **User card added:** @handle, name, join date, stats
- **Shortcuts at bottom:** properly columnar (4-char key column + space + desc)
- **g/G shortcut added** to tooltip list
- **Nav items interactive** when left column is focused (j/k + Enter)

---

## 3. Center Column (Timeline + Compose)

**Files:** `views/compose.go`, `components/postcard.go`, `views/timeline.go`

### Compose Bar Fix

- Textarea width set to `columnWidth - border padding` for clean hard-wrap at box edge
- Long unbroken strings wrap visually to next line (no horizontal scroll)
- Character counter counts runes (actual characters), not visual lines
- Max height ~8 lines before internal textarea scroll

### Post Card Line Break Fix (Critical)

**Root cause:** `ansi.Wordwrap()` treats entire content as single paragraph, destroying `\n`.

**Fix:** Split by `\n` first, then wordwrap each line:

```go
paragraphs := strings.Split(post.Content, "\n")
for _, para := range paragraphs {
    wrapped := ansi.Wordwrap(para, contentWidth, "")
    for _, line := range strings.Split(wrapped, "\n") {
        b.WriteString("  " + line + "\n")
    }
}
```

### Mock Posts (50+ Hardcoded Variety Pack)

Curated slice of 50+ posts for dev/demo mode testing:

- Short one-liners, multiline, posts with #hashtags, @mentions
- Multiple fake authors: @akram, @alice, @bob, @dev_sarah, @terminal_fan, @gopher_grace, etc.
- Varied timestamps: minutes, hours, days ago
- Edge cases: exactly 140 chars, single word, many newlines, long unbroken strings

---

## 4. Right Column (Discover Panel)

**File:** `components/discover.go`

### Two-Section Split

Right column viewport (below header bar and search) divided into 2 equal halves:

- **Top half:** Trending section
- **Bottom half:** Writers to follow section
- Thin dim separator line between sections

### Expanded Trending Data (11 topics)

```
#niotebook       12 niotes
#hello-world      8 niotes
#terminal-life    5 niotes
#claude-code     42 niotes
#codex           28 niotes
#opencode        19 niotes
#skills          15 niotes
#openclaw         7 niotes
#mcp-servers     33 niotes
#agentic-coding  21 niotes
#terminal-love   14 niotes
```

### Expanded Writers to Follow (8 writers)

```
@alice           loves terminal apps
@bob             building in public
@dev_sarah       Go enthusiast
@terminal_fan    CLI everything
@rust_rover      systems thinker
@gopher_grace    open source lover
@vim_master      keyboard warrior
@cloud_nina      infra nerd
```

### Navigation (when right column is focused)

- j/k scrolls within active section (cursor highlight in accent)
- `Enter` toggles between Trending and Writers sections
- Active section header: accent bold; inactive: TextSecondary
- Scroll indicators: `▲`/`▼` at section edges when items overflow viewport

### Search Bar

- Same style: `Search niotes...` placeholder, AccentDim border
- Aligned to same row as compose bar in center column
- Non-functional (placeholder)

---

## 5. Splash Screen — Animated ASCII Reveal

**File:** `views/splash.go`

### Animation Phases

**Phase 1 — ASCII Logo Reveal** (0–1.5s)

- Full-screen centered ASCII art banner (same art as sidebar but larger/centered)
- Characters appear one-by-one with typewriter effect (~30-50ms per character tick)
- The `i` dot appears in amber accent as revealed
- All other UI hidden during reveal

**Phase 2 — Tagline Fade-in** (1.5–2.0s)

- `the social terminal` appears below logo in letter-spaced style (TextMuted)
- Appears all at once after 200ms pause from logo completion

**Phase 3 — Connection Status** (2.0–3.0s)

- Block spinner: `░ ░ ░` → `█ ░ ░` → `█ █ ░` → `█ █ █`
- Status: "connecting..." in TextMuted
- On success: "connected" in Success color, 300ms pause, then transition
- On failure: "connection failed" in Error + "press r to retry" hint

### Implementation

- Bubble Tea `tick` command for typewriter effect (reuse existing tick mechanism)
- `revealIndex` counter increments each tick
- View renders characters up to `revealIndex`
- Phase transitions triggered by index thresholds
- Minimum display: ~2.5s (animation naturally fills this)

---

## Files Changed Summary

| File | Change |
|------|--------|
| `layout/columns.go` | Add header bar rendering, pass focus state to renderer |
| `components/sidebar.go` | Full rewrite: ASCII logo, nav with j/k, user card, shortcuts at bottom |
| `components/discover.go` | Full rewrite: two-section split, scroll viewports, expanded data |
| `components/postcard.go` | Fix line break bug (split by \n before wordwrap) |
| `views/compose.go` | Fix textarea width for proper hard-wrap |
| `views/timeline.go` | Add 50+ mock posts for dev mode |
| `views/splash.go` | Add typewriter reveal animation phases |
| `theme/logo.go` | Add ASCII art banner function |
| `app/app.go` | Route j/k to sidebar/discover when those columns are focused |

---

## Non-Goals (Explicitly Out of Scope)

- Functional search (backend not ready)
- Real trending/writer data from API
- Bookmarks/Settings screens
- Profile editing
- Dark/light mode toggle
- Responsive breakpoint changes (keep current 100/80/80 thresholds)
