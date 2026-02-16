# TUI Layout v3 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Rewrite sidebar, discover, postcard, columns layout, and splash to fix visual bugs, add interactivity, and elevate the brand experience.

**Architecture:** Component rewrite targeting 5 files. The app.go routing layer gets minimal changes to dispatch j/k to sidebar/discover when those columns are focused. The layout system adds header bars. All other files (client, config, views/login, views/register, views/profile, views/help) remain untouched.

**Tech Stack:** Go 1.22+, Bubble Tea, Bubbles (textarea, spinner), Lip Gloss, charmbracelet/x/ansi

---

### Task 1: Fix Post Card Line Break Bug

**Files:**
- Modify: `internal/tui/components/postcard.go:63-73`

This is the most critical bug. `ansi.Wordwrap()` destroys internal `\n` characters by treating the entire content as a single paragraph.

**Step 1: Fix the rendering logic**

Replace lines 63-73 in `postcard.go`:

```go
	// Content, word-wrapped with 2-char left padding.
	// Split by newlines first to preserve intentional line breaks,
	// then wordwrap each paragraph individually.
	contentWidth := width - 2
	if contentWidth < 10 {
		contentWidth = 10
	}
	paragraphs := strings.Split(post.Content, "\n")
	for _, para := range paragraphs {
		if para == "" {
			b.WriteString("  \n")
			continue
		}
		wrapped := ansi.Wordwrap(para, contentWidth, "")
		for _, line := range strings.Split(wrapped, "\n") {
			b.WriteString("  ")
			b.WriteString(line)
			b.WriteString("\n")
		}
	}
```

**Step 2: Run tests**

Run: `go test ./internal/tui/components/... -race -v`
Expected: PASS (existing tests should still pass)

**Step 3: Manual smoke test**

Run: `make dev-tui`
Type a multiline post (press Enter between lines in compose bar). Verify the newlines are preserved in the rendered timeline.

**Step 4: Commit**

```bash
git add internal/tui/components/postcard.go
git commit -m "fix: preserve newlines in post card rendering"
```

---

### Task 2: Add Header Bar to Column Layout

**Files:**
- Modify: `internal/tui/layout/columns.go:89-133`

Add a `RenderHeaderBar` function that renders a thick amber bar for the active column and a thin dim bar for inactive columns. Modify `RenderColumns` to prepend header bars.

**Step 1: Add header bar rendering functions**

Add after line 103 in `columns.go`:

```go
// headerBar returns a horizontal line spanning the given width.
// Active columns get a thick amber bar (━), inactive get a thin dim line (─).
func headerBar(width int, active bool) string {
	if width <= 0 {
		return ""
	}
	if active {
		return lipgloss.NewStyle().Foreground(theme.Accent).Render(strings.Repeat("━", width))
	}
	return lipgloss.NewStyle().Foreground(theme.Border).Render(strings.Repeat("─", width))
}
```

**Step 2: Update RenderColumns signature and rendering**

Change `RenderColumns` to accept `focus FocusColumn`:

```go
func RenderColumns(width, height int, focus FocusColumn, leftContent, centerContent, rightContent string) string {
	cols := ComputeColumns(width)

	// Reserve 1 line for header bar
	contentHeight := height - 1
	if contentHeight < 1 {
		contentHeight = 1
	}

	colStyle := func(w int) lipgloss.Style {
		return lipgloss.NewStyle().Width(w).Height(contentHeight)
	}

	switch cols.Mode {
	case ThreeColumn:
		leftHeader := headerBar(cols.Left, focus == FocusLeft)
		centerHeader := headerBar(cols.Center, focus == FocusCenter)
		rightHeader := headerBar(cols.Right, focus == FocusRight)

		left := leftHeader + "\n" + colStyle(cols.Left).Render(leftContent)
		center := centerHeader + "\n" + colStyle(cols.Center).Render(centerContent)
		right := rightHeader + "\n" + colStyle(cols.Right).Render(rightContent)
		div := verticalDivider(height)
		return lipgloss.JoinHorizontal(lipgloss.Top, left, div, center, div, right)

	case TwoColumn:
		leftHeader := headerBar(cols.Left, focus == FocusLeft)
		centerHeader := headerBar(cols.Center, focus == FocusCenter)

		left := leftHeader + "\n" + colStyle(cols.Left).Render(leftContent)
		center := centerHeader + "\n" + colStyle(cols.Center).Render(centerContent)
		div := verticalDivider(height)
		return lipgloss.JoinHorizontal(lipgloss.Top, left, div, center)

	default:
		centerHeader := headerBar(cols.Center, true)
		return centerHeader + "\n" + colStyle(cols.Center).Render(centerContent)
	}
}
```

**Step 3: Update the call site in app.go**

In `internal/tui/app/app.go:489`, update the `RenderColumns` call to pass `m.focus.Active()`:

```go
return layout.RenderColumns(m.width, m.height, m.focus.Active(), leftContent, centerContent, rightContent)
```

**Step 4: Run tests**

Run: `go test ./internal/tui/... -race -v`
Expected: PASS. Update any tests that call `RenderColumns` to pass the new `focus` parameter.

**Step 5: Commit**

```bash
git add internal/tui/layout/columns.go internal/tui/app/app.go
git commit -m "feat: add amber header bar for active column focus indicator"
```

---

### Task 3: Add ASCII Art Logo

**Files:**
- Modify: `internal/tui/theme/logo.go`

Add a `LogoASCII(width int) string` function that returns a multi-line figlet-style ASCII art rendering of "niotebook". The `i` dot is rendered in amber accent. Must fit within ~20 chars wide.

**Step 1: Create the ASCII art logo function**

Add to `logo.go`:

```go
// LogoASCII returns a multi-line ASCII art rendering of "niotebook".
// The dot on the letter 'i' is rendered in Accent color.
// Designed to fit within a ~20 character wide sidebar column.
func LogoASCII(width int) string {
	text := lipgloss.NewStyle().Bold(true).Foreground(Text)
	accent := lipgloss.NewStyle().Bold(true).Foreground(Accent)

	// Compact 3-line ASCII art that fits in ~20 chars
	// Line 1: n i otebook  (with dot over i in accent)
	// Using simple block-style characters
	lines := []string{
		text.Render("n") + accent.Render("•") + text.Render("otebook"),
		text.Render("━━━━━━━━━━━"),
	}

	return strings.Join(lines, "\n")
}
```

**Note:** The exact ASCII art design should be iterated visually. A good starting point is a simple bold rendering with accent dot. If the column width permits, use a taller figlet font. The key constraint is ~20 chars wide.

A better approach for a true figlet-style logo within 18 chars:

```go
func LogoASCII(_ int) string {
	bold := lipgloss.NewStyle().Bold(true).Foreground(Text)
	dot := lipgloss.NewStyle().Bold(true).Foreground(Accent)

	// 3-line compact art
	line1 := bold.Render("  ") + dot.Render("•")
	line2 := bold.Render("n") + bold.Render("i") + bold.Render("otebook")
	line3 := bold.Render("───────────")

	return line1 + "\n" + line2 + "\n" + line3
}
```

The implementer should experiment with figlet fonts (`figlet -f small niotebook`) to find one that fits within 18-20 chars width. Candidate fonts: `small`, `mini`, `smslant`. The critical requirement is the `i` dot character(s) render in Accent color.

**Step 2: Update LogoSplash for the animated reveal**

Add a `LogoSplashASCII` function that returns the full-screen centered ASCII art (can be the same art or a larger variant):

```go
// LogoSplashASCII returns the splash screen ASCII art variant.
// Same as LogoASCII but can be larger since it's full-screen centered.
func LogoSplashASCII() string {
	return LogoASCII(40)
}
```

**Step 3: Run tests**

Run: `go test ./internal/tui/theme/... -race -v`
Expected: PASS

**Step 4: Commit**

```bash
git add internal/tui/theme/logo.go
git commit -m "feat: add ASCII art brand logo with accent dot"
```

---

### Task 4: Rewrite Sidebar Component

**Files:**
- Modify: `internal/tui/components/sidebar.go` (full rewrite)

Restructure the sidebar: ASCII logo at top, nav items (interactive when focused), user card in middle, shortcuts pushed to bottom. Remove Post button and version string.

**Step 1: Rewrite sidebar.go**

Replace the entire file content with the new implementation:

```go
package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/Akram012388/niotebook-tui/internal/tui/theme"
)

// View identifies the active screen.
type View int

const (
	ViewSplash   View = iota
	ViewLogin
	ViewRegister
	ViewTimeline
	ViewProfile
)

// SidebarState holds interactive state for the left column when focused.
type SidebarState struct {
	NavCursor int // which nav item is highlighted (0=Home, 1=Profile, 2=Bookmarks, 3=Settings)
}

// NavItemCount is the number of navigable items.
const NavItemCount = 4

// navItem describes a sidebar navigation entry.
type navItem struct {
	label       string
	placeholder bool // greyed out, not selectable
}

var navItems = []navItem{
	{label: "Home", placeholder: false},
	{label: "Profile", placeholder: false},
	{label: "Bookmarks", placeholder: true},
	{label: "Settings", placeholder: true},
}

// shortcut is a key-description pair.
type shortcut struct {
	key  string
	desc string
}

var shortcuts = []shortcut{
	{"j/k", "scroll"},
	{"g/G", "top/bottom"},
	{"Tab", "switch col"},
	{"n", "compose"},
	{"?", "help"},
	{"q", "quit"},
}

// RenderSidebar renders the left sidebar with ASCII logo, nav, user card,
// and shortcuts. The focused parameter controls whether nav items show a
// highlight cursor. sidebarState holds the cursor position for navigation.
func RenderSidebar(user *models.User, activeView View, focused bool, sidebarState *SidebarState, width, height int) string {
	if width == 0 {
		return ""
	}

	innerWidth := width - 2
	if innerWidth < 0 {
		innerWidth = 0
	}

	var topSections []string
	var bottomSections []string

	// === TOP: Logo ===
	topSections = append(topSections, theme.LogoASCII(innerWidth))
	topSections = append(topSections, "")

	if user != nil {
		// === Nav items ===
		for i, item := range navItems {
			isActive := (item.label == "Home" && activeView == ViewTimeline) ||
				(item.label == "Profile" && activeView == ViewProfile)
			isCursor := focused && sidebarState != nil && sidebarState.NavCursor == i

			if item.placeholder {
				topSections = append(topSections, renderNavItemPlaceholder(item.label))
			} else if isCursor {
				// Focused cursor highlight
				style := lipgloss.NewStyle().Bold(true).Foreground(theme.Accent).Reverse(true)
				topSections = append(topSections, style.Render(" "+item.label+" "))
			} else if isActive {
				topSections = append(topSections, renderNavItem(item.label, true))
			} else {
				topSections = append(topSections, renderNavItem(item.label, false))
			}
		}

		topSections = append(topSections, "")

		// === User card ===
		topSections = append(topSections, theme.Separator(innerWidth))

		usernameStyle := lipgloss.NewStyle().Foreground(theme.Accent)
		topSections = append(topSections, usernameStyle.Render("@"+user.Username))

		if user.DisplayName != "" {
			displayStyle := lipgloss.NewStyle().Foreground(theme.Text)
			topSections = append(topSections, displayStyle.Render(user.DisplayName))
		}

		if !user.CreatedAt.IsZero() {
			captionStyle := lipgloss.NewStyle().Foreground(theme.TextSecondary)
			joinDate := user.CreatedAt.Format("Joined Jan 2006")
			topSections = append(topSections, captionStyle.Render(joinDate))
		}

		topSections = append(topSections, theme.Separator(innerWidth))

		// Stats line (placeholder counts for now)
		statsStyle := lipgloss.NewStyle().Foreground(theme.TextMuted)
		topSections = append(topSections, statsStyle.Render("0 niotes · 0 following"))

		// === BOTTOM: Shortcuts ===
		for _, s := range shortcuts {
			bottomSections = append(bottomSections, renderShortcut(s.key, s.desc, innerWidth))
		}
	}

	// Compute vertical spacing to push shortcuts to bottom
	topContent := strings.Join(topSections, "\n")
	bottomContent := strings.Join(bottomSections, "\n")

	topLines := strings.Count(topContent, "\n") + 1
	bottomLines := strings.Count(bottomContent, "\n") + 1
	padding := 2 // wrapper padding (top + bottom)

	gap := height - topLines - bottomLines - padding
	if gap < 1 {
		gap = 1
	}

	content := topContent + strings.Repeat("\n", gap) + bottomContent

	wrapper := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(1, 1)

	return wrapper.Render(content)
}

func renderNavItem(label string, active bool) string {
	if active {
		style := lipgloss.NewStyle().Bold(true).Foreground(theme.Accent)
		return style.Render("● " + label)
	}
	style := lipgloss.NewStyle().Foreground(theme.TextSecondary)
	return style.Render("  " + label)
}

func renderNavItemPlaceholder(label string) string {
	style := lipgloss.NewStyle().Foreground(theme.TextMuted)
	return style.Render("  " + label)
}

func renderShortcut(key, desc string, _ int) string {
	keyStyle := lipgloss.NewStyle().Foreground(theme.Accent).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(theme.TextMuted)
	// Fixed-width key column: 5 chars + 1 space + desc
	return fmt.Sprintf("%-5s %s", keyStyle.Render(key), descStyle.Render(desc))
}
```

**Step 2: Update the call site in app.go**

The `RenderSidebar` signature changed — it now takes `*SidebarState`. Add a `sidebarState` field to `AppModel` and pass it:

In `app.go`, add to `AppModel` struct (around line 109):
```go
	// Sidebar interactive state (nav cursor when left column focused)
	sidebarState components.SidebarState
```

Update the call site at line 474-480:
```go
	leftContent := components.RenderSidebar(
		m.user,
		components.View(m.currentView),
		m.focus.Active() == layout.FocusLeft,
		&m.sidebarState,
		cols.Left,
		contentHeight,
	)
```

**Step 3: Add j/k routing for left column in app.go**

In the key handling section of `Update` (around line 262, after global shortcuts), add routing for left column focus:

```go
		// Route j/k to sidebar nav when left column is focused
		if m.focus.Active() == layout.FocusLeft && !m.isTextInputFocused() {
			switch {
			case msg.Type == tea.KeyDown || (msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'j'):
				if m.sidebarState.NavCursor < components.NavItemCount-1 {
					m.sidebarState.NavCursor++
				}
				return m, nil
			case msg.Type == tea.KeyUp || (msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'k'):
				if m.sidebarState.NavCursor > 0 {
					m.sidebarState.NavCursor--
				}
				return m, nil
			case msg.Type == tea.KeyEnter:
				switch m.sidebarState.NavCursor {
				case 0: // Home
					if m.currentView != ViewTimeline {
						m.currentView = ViewTimeline
						if m.timeline != nil {
							return m, m.timeline.FetchLatest()
						}
					}
				case 1: // Profile
					if m.user != nil {
						return m.openProfile(m.user.ID, true)
					}
				}
				return m, nil
			}
		}
```

This block should be inserted BEFORE the `return m.updateCurrentView(msg)` fall-through (line 264).

**Step 4: Run tests**

Run: `go test ./internal/tui/... -race -v`
Expected: PASS (update any tests that call `RenderSidebar` with new signature)

**Step 5: Commit**

```bash
git add internal/tui/components/sidebar.go internal/tui/app/app.go
git commit -m "feat: rewrite sidebar with ASCII logo, user card, shortcuts at bottom"
```

---

### Task 5: Rewrite Discover Panel with Two-Section Split

**Files:**
- Modify: `internal/tui/components/discover.go` (full rewrite)

Add expanded mock data, split into two scrollable sections, add interactive navigation state.

**Step 1: Rewrite discover.go**

Replace the entire file:

```go
package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/Akram012388/niotebook-tui/internal/tui/theme"
)

type trendingTag struct {
	tag   string
	count string
}

type suggestedWriter struct {
	handle string
	bio    string
}

var trendingTags = []trendingTag{
	{tag: "#niotebook", count: "12 niotes"},
	{tag: "#hello-world", count: "8 niotes"},
	{tag: "#terminal-life", count: "5 niotes"},
	{tag: "#claude-code", count: "42 niotes"},
	{tag: "#codex", count: "28 niotes"},
	{tag: "#opencode", count: "19 niotes"},
	{tag: "#skills", count: "15 niotes"},
	{tag: "#openclaw", count: "7 niotes"},
	{tag: "#mcp-servers", count: "33 niotes"},
	{tag: "#agentic-coding", count: "21 niotes"},
	{tag: "#terminal-love", count: "14 niotes"},
}

var suggestedWriters = []suggestedWriter{
	{handle: "@alice", bio: "loves terminal apps"},
	{handle: "@bob", bio: "building in public"},
	{handle: "@dev_sarah", bio: "Go enthusiast"},
	{handle: "@terminal_fan", bio: "CLI everything"},
	{handle: "@rust_rover", bio: "systems thinker"},
	{handle: "@gopher_grace", bio: "open source lover"},
	{handle: "@vim_master", bio: "keyboard warrior"},
	{handle: "@cloud_nina", bio: "infra nerd"},
}

// DiscoverSection identifies which section is active in the right column.
type DiscoverSection int

const (
	SectionTrending DiscoverSection = 0
	SectionWriters  DiscoverSection = 1
)

// DiscoverState holds interactive state for the right column.
type DiscoverState struct {
	ActiveSection  DiscoverSection
	TrendingCursor int
	WritersCursor  int
	TrendingScroll int // scroll offset for trending section
	WritersScroll  int // scroll offset for writers section
}

// TrendingCount returns the number of trending tags.
func TrendingCount() int { return len(trendingTags) }

// WritersCount returns the number of suggested writers.
func WritersCount() int { return len(suggestedWriters) }

// RenderDiscover renders the right-column with search bar, trending, and
// writers to follow. The two sections split the available height equally.
// When focused, the active section and cursor item are highlighted.
func RenderDiscover(focused bool, discoverState *DiscoverState, width, height int) string {
	if width == 0 {
		return ""
	}

	innerWidth := width - 2
	if innerWidth < 0 {
		innerWidth = 0
	}

	var sections []string

	// Search bar
	sections = append(sections, renderSearchBar(innerWidth))
	sections = append(sections, "")

	// Calculate available height for the two sections
	// Search bar takes ~3 lines (border top + content + border bottom) + 1 blank
	searchLines := 4
	availableHeight := height - searchLines - 2 // padding
	if availableHeight < 4 {
		availableHeight = 4
	}
	halfHeight := availableHeight / 2

	// Trending section
	trendingActive := focused && discoverState != nil && discoverState.ActiveSection == SectionTrending
	sections = append(sections, renderTrendingSection(innerWidth, halfHeight, focused, trendingActive, discoverState))

	// Separator between sections
	sections = append(sections, "")

	// Writers section
	writersActive := focused && discoverState != nil && discoverState.ActiveSection == SectionWriters
	sections = append(sections, renderWritersSection(innerWidth, halfHeight, focused, writersActive, discoverState))

	content := strings.Join(sections, "\n")

	wrapper := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(1, 1)

	return wrapper.Render(content)
}

func renderSearchBar(width int) string {
	placeholder := theme.Hint.Render("Search niotes...")
	bar := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.AccentDim).
		Width(width - 2).
		Padding(0, 1).
		Render(placeholder)
	return bar
}

func renderTrendingSection(width, maxHeight int, columnFocused, sectionActive bool, state *DiscoverState) string {
	var lines []string

	// Section header
	headerStyle := theme.Label
	if sectionActive {
		headerStyle = lipgloss.NewStyle().Bold(true).Foreground(theme.Accent)
	}
	lines = append(lines, headerStyle.Render("Trending"))
	lines = append(lines, theme.Separator(width))

	// Determine scroll offset and cursor
	scrollOffset := 0
	cursor := -1
	if state != nil {
		scrollOffset = state.TrendingScroll
		if sectionActive {
			cursor = state.TrendingCursor
		}
	}

	// Calculate visible items (each tag takes 2 lines: tag + count)
	headerLines := 2
	itemHeight := 2
	visibleItems := (maxHeight - headerLines) / itemHeight
	if visibleItems < 1 {
		visibleItems = 1
	}

	// Scroll indicator top
	if scrollOffset > 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(theme.TextMuted).Render("  ▲"))
	}

	tagStyle := lipgloss.NewStyle().Foreground(theme.Accent)
	selectedTagStyle := lipgloss.NewStyle().Foreground(theme.Accent).Bold(true).Reverse(true)

	end := scrollOffset + visibleItems
	if end > len(trendingTags) {
		end = len(trendingTags)
	}

	for i := scrollOffset; i < end; i++ {
		tag := trendingTags[i]
		if i == cursor {
			lines = append(lines, selectedTagStyle.Render(" "+tag.tag+" "))
		} else {
			lines = append(lines, tagStyle.Render(tag.tag))
		}
		lines = append(lines, theme.Caption.Render(tag.count))
	}

	// Scroll indicator bottom
	if end < len(trendingTags) {
		lines = append(lines, lipgloss.NewStyle().Foreground(theme.TextMuted).Render("  ▼"))
	}

	return strings.Join(lines, "\n")
}

func renderWritersSection(width, maxHeight int, columnFocused, sectionActive bool, state *DiscoverState) string {
	var lines []string

	// Section header
	headerStyle := theme.Label
	if sectionActive {
		headerStyle = lipgloss.NewStyle().Bold(true).Foreground(theme.Accent)
	}
	lines = append(lines, headerStyle.Render("Writers to follow"))
	lines = append(lines, theme.Separator(width))

	scrollOffset := 0
	cursor := -1
	if state != nil {
		scrollOffset = state.WritersScroll
		if sectionActive {
			cursor = state.WritersCursor
		}
	}

	headerLines := 2
	itemHeight := 2
	visibleItems := (maxHeight - headerLines) / itemHeight
	if visibleItems < 1 {
		visibleItems = 1
	}

	if scrollOffset > 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(theme.TextMuted).Render("  ▲"))
	}

	handleStyle := lipgloss.NewStyle().Foreground(theme.Accent)
	selectedHandleStyle := lipgloss.NewStyle().Foreground(theme.Accent).Bold(true).Reverse(true)

	end := scrollOffset + visibleItems
	if end > len(suggestedWriters) {
		end = len(suggestedWriters)
	}

	for i := scrollOffset; i < end; i++ {
		w := suggestedWriters[i]
		if i == cursor {
			lines = append(lines, selectedHandleStyle.Render(" "+w.handle+" "))
		} else {
			lines = append(lines, handleStyle.Render(w.handle))
		}
		lines = append(lines, theme.Caption.Render(w.bio))
	}

	if end < len(suggestedWriters) {
		lines = append(lines, lipgloss.NewStyle().Foreground(theme.TextMuted).Render("  ▼"))
	}

	return strings.Join(lines, "\n")
}

// scrollTrending adjusts the trending scroll offset to keep cursor visible.
func scrollTrending(state *DiscoverState, visibleItems int) {
	if state.TrendingCursor < state.TrendingScroll {
		state.TrendingScroll = state.TrendingCursor
	} else if state.TrendingCursor >= state.TrendingScroll+visibleItems {
		state.TrendingScroll = state.TrendingCursor - visibleItems + 1
	}
}

// scrollWriters adjusts the writers scroll offset to keep cursor visible.
func scrollWriters(state *DiscoverState, visibleItems int) {
	if state.WritersCursor < state.WritersScroll {
		state.WritersScroll = state.WritersCursor
	} else if state.WritersCursor >= state.WritersScroll+visibleItems {
		state.WritersScroll = state.WritersCursor - visibleItems + 1
	}
}
```

**Step 2: Update the call site in app.go**

Add `discoverState` to `AppModel` struct:
```go
	// Discover interactive state (section/cursor when right column focused)
	discoverState components.DiscoverState
```

Update the `RenderDiscover` call at line 483-487:
```go
	rightContent := components.RenderDiscover(
		m.focus.Active() == layout.FocusRight,
		&m.discoverState,
		cols.Right,
		contentHeight,
	)
```

**Step 3: Add j/k/Enter routing for right column in app.go**

Add after the left column routing block (from Task 4):

```go
		// Route j/k/Enter to discover panel when right column is focused
		if m.focus.Active() == layout.FocusRight && !m.isTextInputFocused() {
			switch {
			case msg.Type == tea.KeyDown || (msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'j'):
				if m.discoverState.ActiveSection == components.SectionTrending {
					if m.discoverState.TrendingCursor < components.TrendingCount()-1 {
						m.discoverState.TrendingCursor++
					}
				} else {
					if m.discoverState.WritersCursor < components.WritersCount()-1 {
						m.discoverState.WritersCursor++
					}
				}
				return m, nil
			case msg.Type == tea.KeyUp || (msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'k'):
				if m.discoverState.ActiveSection == components.SectionTrending {
					if m.discoverState.TrendingCursor > 0 {
						m.discoverState.TrendingCursor--
					}
				} else {
					if m.discoverState.WritersCursor > 0 {
						m.discoverState.WritersCursor--
					}
				}
				return m, nil
			case msg.Type == tea.KeyEnter:
				// Toggle between Trending and Writers sections
				if m.discoverState.ActiveSection == components.SectionTrending {
					m.discoverState.ActiveSection = components.SectionWriters
				} else {
					m.discoverState.ActiveSection = components.SectionTrending
				}
				return m, nil
			}
		}
```

**Step 4: Run tests**

Run: `go test ./internal/tui/... -race -v`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/components/discover.go internal/tui/app/app.go
git commit -m "feat: rewrite discover panel with two-section split and scroll navigation"
```

---

### Task 6: Add 50+ Mock Posts for Testing

**Files:**
- Create: `internal/tui/views/mockdata.go`
- Modify: `internal/tui/views/timeline.go`

Add a curated variety pack of 50+ mock posts with diverse content, authors, and timestamps.

**Step 1: Create mock data file**

Create `internal/tui/views/mockdata.go`:

```go
package views

import (
	"time"

	"github.com/Akram012388/niotebook-tui/internal/models"
)

// mockUser creates a models.User with the given username and display name.
func mockUser(username, displayName string) *models.User {
	return &models.User{
		ID:          "mock-" + username,
		Username:    username,
		DisplayName: displayName,
		CreatedAt:   time.Date(2026, 1, 15, 0, 0, 0, 0, time.UTC),
	}
}

// GenerateMockPosts returns 50+ diverse mock posts for dev/testing.
func GenerateMockPosts() []models.Post {
	now := time.Now()

	authors := map[string]*models.User{
		"akram":        mockUser("akram", "Akram"),
		"alice":        mockUser("alice", "Alice Chen"),
		"bob":          mockUser("bob", "Bob Builder"),
		"dev_sarah":    mockUser("dev_sarah", "Sarah Dev"),
		"terminal_fan": mockUser("terminal_fan", "Terminal Fan"),
		"gopher_grace": mockUser("gopher_grace", "Grace Gopher"),
		"rust_rover":   mockUser("rust_rover", "Rover Rust"),
		"vim_master":   mockUser("vim_master", "Vim Master"),
		"cloud_nina":   mockUser("cloud_nina", "Nina Cloud"),
		"code_poet":    mockUser("code_poet", "Code Poet"),
	}

	posts := []models.Post{
		{ID: "p01", Author: authors["akram"], Content: "just launched niotebook TUI! the social terminal is alive", CreatedAt: now.Add(-2 * time.Minute)},
		{ID: "p02", Author: authors["alice"], Content: "loving the terminal vibes here\nthis is what social media should feel like\nno ads, no algorithm, just text", CreatedAt: now.Add(-5 * time.Minute)},
		{ID: "p03", Author: authors["bob"], Content: "building in public is easier when your tools are simple #terminal-life", CreatedAt: now.Add(-8 * time.Minute)},
		{ID: "p04", Author: authors["dev_sarah"], Content: "Go + Bubble Tea = perfect combo for TUIs", CreatedAt: now.Add(-12 * time.Minute)},
		{ID: "p05", Author: authors["terminal_fan"], Content: "who needs a browser when you have a terminal?", CreatedAt: now.Add(-15 * time.Minute)},
		{ID: "p06", Author: authors["gopher_grace"], Content: "the compose bar is so clean\ntype your thoughts\nhit ctrl+enter\ndone", CreatedAt: now.Add(-18 * time.Minute)},
		{ID: "p07", Author: authors["vim_master"], Content: "j/k navigation feels like home #vim #terminal-love", CreatedAt: now.Add(-22 * time.Minute)},
		{ID: "p08", Author: authors["rust_rover"], Content: "respect to the Go devs out there. Bubble Tea is a gem.", CreatedAt: now.Add(-25 * time.Minute)},
		{ID: "p09", Author: authors["cloud_nina"], Content: "deploying niotebook on my homelab this weekend", CreatedAt: now.Add(-30 * time.Minute)},
		{ID: "p10", Author: authors["code_poet"], Content: "code is poetry\nterminals are canvases\nwe paint with text", CreatedAt: now.Add(-35 * time.Minute)},
		{ID: "p11", Author: authors["akram"], Content: "working on the three-column layout. left sidebar, center feed, right discover. X-style but for the terminal.", CreatedAt: now.Add(-40 * time.Minute)},
		{ID: "p12", Author: authors["alice"], Content: "just discovered #claude-code and my productivity doubled", CreatedAt: now.Add(-45 * time.Minute)},
		{ID: "p13", Author: authors["bob"], Content: "hot take: the best UI is no UI. just give me a prompt.", CreatedAt: now.Add(-50 * time.Minute)},
		{ID: "p14", Author: authors["dev_sarah"], Content: "TDD is not optional\nwrite the test first\nwatch it fail\nmake it pass\nrefactor\nrepeat", CreatedAt: now.Add(-55 * time.Minute)},
		{ID: "p15", Author: authors["terminal_fan"], Content: "my terminal color scheme: gruvbox dark. fight me.", CreatedAt: now.Add(-1 * time.Hour)},
		{ID: "p16", Author: authors["gopher_grace"], Content: "#opencode is the future of development tools", CreatedAt: now.Add(-1*time.Hour - 5*time.Minute)},
		{ID: "p17", Author: authors["vim_master"], Content: "protip: use lipgloss adaptive colors so your TUI works in both dark and light terminals", CreatedAt: now.Add(-1*time.Hour - 10*time.Minute)},
		{ID: "p18", Author: authors["rust_rover"], Content: "async in Go: goroutines + channels = simplicity\nasync in Rust: Pin<Box<dyn Future>> = pain", CreatedAt: now.Add(-1*time.Hour - 15*time.Minute)},
		{ID: "p19", Author: authors["cloud_nina"], Content: "docker compose up -d niotebook\nthat's the whole deploy", CreatedAt: now.Add(-1*time.Hour - 20*time.Minute)},
		{ID: "p20", Author: authors["code_poet"], Content: "a", CreatedAt: now.Add(-1*time.Hour - 25*time.Minute)},
		{ID: "p21", Author: authors["akram"], Content: "keyboard shortcuts make everything better. Tab to switch columns, j/k to navigate, n to compose. #niotebook", CreatedAt: now.Add(-1*time.Hour - 30*time.Minute)},
		{ID: "p22", Author: authors["alice"], Content: "#mcp-servers are the new APIs. context-aware tool integration changes everything.", CreatedAt: now.Add(-2 * time.Hour)},
		{ID: "p23", Author: authors["bob"], Content: "spent 3 hours debugging a CSS layout. switched to terminal UI. fixed in 20 minutes.", CreatedAt: now.Add(-2*time.Hour - 10*time.Minute)},
		{ID: "p24", Author: authors["dev_sarah"], Content: "cursor-based pagination > offset pagination\nchange my mind", CreatedAt: now.Add(-2*time.Hour - 20*time.Minute)},
		{ID: "p25", Author: authors["terminal_fan"], Content: "the best feature of niotebook? no notifications. you check it when YOU want to.", CreatedAt: now.Add(-2*time.Hour - 30*time.Minute)},
		{ID: "p26", Author: authors["gopher_grace"], Content: "reading through the Bubble Tea source code and it's so clean. Elm architecture in Go just works.", CreatedAt: now.Add(-3 * time.Hour)},
		{ID: "p27", Author: authors["vim_master"], Content: ":wq", CreatedAt: now.Add(-3*time.Hour - 10*time.Minute)},
		{ID: "p28", Author: authors["rust_rover"], Content: "#agentic-coding is the next paradigm. not AI writing code for you. AI working WITH you.", CreatedAt: now.Add(-3*time.Hour - 20*time.Minute)},
		{ID: "p29", Author: authors["cloud_nina"], Content: "postgresql > everything else\nfight me\n(you won't win)", CreatedAt: now.Add(-3*time.Hour - 30*time.Minute)},
		{ID: "p30", Author: authors["code_poet"], Content: "exactly one hundred and forty characters is the perfect length for a thought on the social terminal platform niotebook check it out now!!!!", CreatedAt: now.Add(-4 * time.Hour)},
		{ID: "p31", Author: authors["akram"], Content: "the splash screen animation is coming along nicely. typewriter effect for the logo reveal.", CreatedAt: now.Add(-4*time.Hour - 15*time.Minute)},
		{ID: "p32", Author: authors["alice"], Content: "why does every app need to be in the browser?\nsome things are better in the terminal\nlike social media\napparently", CreatedAt: now.Add(-4*time.Hour - 30*time.Minute)},
		{ID: "p33", Author: authors["bob"], Content: "#codex #skills #openclaw — the open source AI tooling ecosystem is exploding", CreatedAt: now.Add(-5 * time.Hour)},
		{ID: "p34", Author: authors["dev_sarah"], Content: "JWT refresh tokens with single-use rotation. the right way to do auth.", CreatedAt: now.Add(-5*time.Hour - 20*time.Minute)},
		{ID: "p35", Author: authors["terminal_fan"], Content: "tmux + neovim + niotebook = the holy trinity", CreatedAt: now.Add(-5*time.Hour - 40*time.Minute)},
		{ID: "p36", Author: authors["gopher_grace"], Content: "interface segregation principle in Go: define interfaces where they're consumed, not where they're implemented", CreatedAt: now.Add(-6 * time.Hour)},
		{ID: "p37", Author: authors["vim_master"], Content: "hjkl is muscle memory at this point", CreatedAt: now.Add(-6*time.Hour - 30*time.Minute)},
		{ID: "p38", Author: authors["rust_rover"], Content: "monorepo life: one repo, two binaries, zero drama", CreatedAt: now.Add(-7 * time.Hour)},
		{ID: "p39", Author: authors["cloud_nina"], Content: "make build && make test && make deploy\nthat's the whole CI pipeline", CreatedAt: now.Add(-8 * time.Hour)},
		{ID: "p40", Author: authors["code_poet"], Content: "bits and bytes\nflow through wires\ntext on screens\na world entire", CreatedAt: now.Add(-9 * time.Hour)},
		{ID: "p41", Author: authors["akram"], Content: "three column layout done. responsive breakpoints at 100 and 80 cols.", CreatedAt: now.Add(-10 * time.Hour)},
		{ID: "p42", Author: authors["alice"], Content: "the amber accent color is perfect. warm. inviting. #niotebook", CreatedAt: now.Add(-12 * time.Hour)},
		{ID: "p43", Author: authors["bob"], Content: "convention: lowercase error messages, no trailing punctuation. it's the Go way.", CreatedAt: now.Add(-14 * time.Hour)},
		{ID: "p44", Author: authors["dev_sarah"], Content: "slog > logrus > zap\nstdlib wins again", CreatedAt: now.Add(-16 * time.Hour)},
		{ID: "p45", Author: authors["terminal_fan"], Content: "alacritty with JetBrains Mono. that's the setup.", CreatedAt: now.Add(-18 * time.Hour)},
		{ID: "p46", Author: authors["gopher_grace"], Content: "table-driven tests are the backbone of good Go testing", CreatedAt: now.Add(-20 * time.Hour)},
		{ID: "p47", Author: authors["vim_master"], Content: "mapped my caps lock to escape years ago. best decision ever.", CreatedAt: now.Add(-1 * 24 * time.Hour)},
		{ID: "p48", Author: authors["rust_rover"], Content: "pgx v5 > database/sql\nconnection pooling, prepared statements, batch queries, all built in", CreatedAt: now.Add(-1*24*time.Hour - 6*time.Hour)},
		{ID: "p49", Author: authors["cloud_nina"], Content: "the social terminal. I didn't know I needed this until now.", CreatedAt: now.Add(-1*24*time.Hour - 12*time.Hour)},
		{ID: "p50", Author: authors["code_poet"], Content: "first!", CreatedAt: now.Add(-2 * 24 * time.Hour)},
		{ID: "p51", Author: authors["akram"], Content: "welcome to niotebook. the social terminal.\nwrite your thoughts. share with the world.\n140 characters at a time.", CreatedAt: now.Add(-2*24*time.Hour - 6*time.Hour)},
		{ID: "p52", Author: authors["alice"], Content: "thisisaverylongwordwithnospacestotesthowthetextWrappingHandlesItInThePostCardRenderingComponentOfTheNiotebookTUI", CreatedAt: now.Add(-2*24*time.Hour - 12*time.Hour)},
		{ID: "p53", Author: authors["bob"], Content: "\n\n\n", CreatedAt: now.Add(-3 * 24 * time.Hour)},
		{ID: "p54", Author: authors["dev_sarah"], Content: "line1\nline2\nline3\nline4\nline5", CreatedAt: now.Add(-3*24*time.Hour - 12*time.Hour)},
		{ID: "p55", Author: authors["terminal_fan"], Content: "short", CreatedAt: now.Add(-4 * 24 * time.Hour)},
	}

	// Set AuthorID for all posts
	for i := range posts {
		if posts[i].Author != nil {
			posts[i].AuthorID = posts[i].Author.ID
		}
	}

	return posts
}
```

**Step 2: Wire up mock data in timeline**

In `timeline.go`, add a `SetMockData` method or modify `Init` to load mock data when no server is available. The simplest approach is to have app.go call `SetPosts(GenerateMockPosts())` when in dev mode.

Alternatively, add a check in `NewTimelineModel`:

```go
// LoadMockPosts loads hardcoded mock posts for dev/testing mode.
func (m *TimelineModel) LoadMockPosts() {
	m.posts = GenerateMockPosts()
	m.cursor = 0
	m.scrollTop = 0
	m.hasMore = false
}
```

Then in `app.go`, after timeline is created and before it fetches from server, the implementer can call `m.timeline.LoadMockPosts()` for testing. The exact wiring depends on whether a `--mock` flag or dev mode check is available — the implementer should add a simple boolean flag.

**Step 3: Run tests**

Run: `go test ./internal/tui/views/... -race -v`
Expected: PASS

**Step 4: Commit**

```bash
git add internal/tui/views/mockdata.go internal/tui/views/timeline.go
git commit -m "feat: add 55 curated mock posts for dev/testing mode"
```

---

### Task 7: Fix Compose Bar Width for Proper Hard-Wrap

**Files:**
- Modify: `internal/tui/views/compose.go:197-204`

The textarea width needs to match the compose box width minus borders/padding exactly.

**Step 1: Fix updateTextareaSize**

Replace `updateTextareaSize` in `compose.go` (lines 197-204):

```go
func (m *ComposeModel) updateTextareaSize() {
	// The expanded box style has: border (2) + padding (2) = 4 chars horizontal
	// The textarea itself needs to fit inside that, so subtract 4 from m.width
	boxWidth := m.width - 4
	if boxWidth < 20 {
		boxWidth = 20
	}
	// Textarea width should match boxWidth exactly for clean hard-wrap
	m.textarea.SetWidth(boxWidth)
}
```

**Step 2: Run tests**

Run: `go test ./internal/tui/views/... -race -v`
Expected: PASS

**Step 3: Commit**

```bash
git add internal/tui/views/compose.go
git commit -m "fix: set compose textarea width to match box for proper hard-wrap"
```

---

### Task 8: Animated Splash Screen with Typewriter Reveal

**Files:**
- Modify: `internal/tui/views/splash.go`
- Modify: `internal/tui/app/messages.go` (add tick message)

Add phased animation: typewriter logo reveal, tagline fade-in, then connection spinner.

**Step 1: Add a reveal tick message**

Add to `messages.go`:

```go
// Splash animation messages
type MsgRevealTick struct{}
```

**Step 2: Rewrite splash.go with animation phases**

Replace the entire splash model with the animated version:

```go
package views

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Akram012388/niotebook-tui/internal/tui/app"
	"github.com/Akram012388/niotebook-tui/internal/tui/theme"
)

const MinSplashDuration = 2500 * time.Millisecond

// Animation timing
const (
	revealTickInterval = 40 * time.Millisecond // per-character reveal speed
	taglinePause       = 200 * time.Millisecond
	connectedPause     = 300 * time.Millisecond
)

// splashPhase tracks the current animation phase.
type splashPhase int

const (
	phaseReveal    splashPhase = iota // typewriter logo reveal
	phaseTagline                      // tagline appearing
	phaseConnecting                   // spinner + status
)

func BlockSpinnerFrames() []string {
	border := lipgloss.NewStyle().Foreground(theme.Border)
	accent := lipgloss.NewStyle().Foreground(theme.Accent)
	light := border.Render("░")
	full := accent.Render("█")
	return []string{
		light + " " + light + " " + light,
		full + " " + light + " " + light,
		full + " " + full + " " + light,
		full + " " + full + " " + full,
	}
}

func newBlockSpinner() spinner.Model {
	s := spinner.New()
	s.Spinner = spinner.Spinner{
		Frames: BlockSpinnerFrames(),
		FPS:    300 * time.Millisecond,
	}
	return s
}

type SplashModel struct {
	serverURL    string
	spinner      spinner.Model
	done         bool
	failed       bool
	err          string
	width        int
	height       int
	phase        splashPhase
	revealIndex  int    // how many chars of the logo to show
	logoText     string // the full plain-text logo for counting
	showTagline  bool
}

func NewSplashModel(serverURL string) SplashModel {
	// Get the plain text of the splash logo for character counting
	logoPlain := "n i o t e b o o k" // letter-spaced logo text
	return SplashModel{
		serverURL: serverURL,
		spinner:   newBlockSpinner(),
		phase:     phaseReveal,
		logoText:  logoPlain,
	}
}

func (m SplashModel) Done() bool         { return m.done }
func (m SplashModel) Failed() bool        { return m.failed }
func (m SplashModel) ErrorMessage() string { return m.err }
func (m SplashModel) HelpText() string    { return "" }

func (m SplashModel) Init() tea.Cmd {
	return tea.Batch(
		m.revealTick(),
		m.checkHealth(),
	)
}

func (m SplashModel) revealTick() tea.Cmd {
	return tea.Tick(revealTickInterval, func(_ time.Time) tea.Msg {
		return app.MsgRevealTick{}
	})
}

func (m SplashModel) taglinePauseCmd() tea.Cmd {
	return tea.Tick(taglinePause, func(_ time.Time) tea.Msg {
		return msgTaglineShow{}
	})
}

// Internal messages for splash animation
type msgTaglineShow struct{}

func (m SplashModel) Update(msg tea.Msg) (SplashModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case app.MsgServerConnected:
		m.done = true
		return m, nil

	case app.MsgServerFailed:
		m.failed = true
		m.err = msg.Err
		return m, nil

	case app.MsgRevealTick:
		if m.phase == phaseReveal {
			m.revealIndex++
			if m.revealIndex >= len([]rune(m.logoText)) {
				// Logo fully revealed, pause then show tagline
				m.phase = phaseTagline
				return m, m.taglinePauseCmd()
			}
			return m, m.revealTick()
		}
		return m, nil

	case msgTaglineShow:
		m.showTagline = true
		m.phase = phaseConnecting
		return m, m.spinner.Tick

	case tea.KeyMsg:
		if m.failed {
			if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'r' {
				m.failed = false
				m.err = ""
				return m, tea.Batch(m.spinner.Tick, m.checkHealth())
			}
		}
		if msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'q' {
			return m, tea.Quit
		}
		if msg.Type == tea.KeyCtrlC {
			return m, tea.Quit
		}
		return m, nil

	case spinner.TickMsg:
		if m.phase == phaseConnecting && !m.done && !m.failed {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	return m, nil
}

func (m SplashModel) View() string {
	var b strings.Builder

	// Render logo with typewriter reveal
	b.WriteString(m.renderRevealLogo())
	b.WriteString("\n")

	// Tagline (only after reveal completes)
	if m.showTagline {
		b.WriteString(theme.TaglineSplash())
		b.WriteString("\n\n")
	} else {
		b.WriteString("\n\n")
	}

	// Connection status (only in connecting phase)
	if m.phase == phaseConnecting {
		if m.failed {
			errStyle := lipgloss.NewStyle().Foreground(theme.Error)
			b.WriteString(errStyle.Render(fmt.Sprintf("connection failed: %s", m.err)))
			b.WriteString("\n\n")
			b.WriteString(theme.Hint.Render("press r to retry · q to quit"))
		} else if !m.done {
			b.WriteString(m.spinner.View())
			b.WriteString("\n")
			b.WriteString(theme.Caption.Render("connecting..."))
		}
	}

	content := b.String()
	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m SplashModel) renderRevealLogo() string {
	text := lipgloss.NewStyle().Bold(true).Foreground(theme.Text)
	accent := lipgloss.NewStyle().Bold(true).Foreground(theme.Accent)

	letters := []struct {
		char  string
		style lipgloss.Style
	}{
		{"n", text}, {"i", accent}, {"o", text}, {"t", text},
		{"e", text}, {"b", text}, {"o", text}, {"o", text}, {"k", text},
	}

	// Build the spaced logo: "n i o t e b o o k"
	// Each letter takes 2 chars in the plain text (char + space), except the last
	var revealed strings.Builder
	charIndex := 0
	for i, l := range letters {
		if charIndex >= m.revealIndex {
			break
		}
		revealed.WriteString(l.style.Render(l.char))
		charIndex++
		// Add space between letters (except after last)
		if i < len(letters)-1 && charIndex < m.revealIndex {
			revealed.WriteString(" ")
			charIndex++
		}
	}

	return revealed.String()
}

func (m SplashModel) checkHealth() tea.Cmd {
	url := m.serverURL
	return func() tea.Msg {
		start := time.Now()
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(url + "/health")
		if elapsed := time.Since(start); elapsed < MinSplashDuration {
			time.Sleep(MinSplashDuration - elapsed)
		}
		if err != nil {
			return app.MsgServerFailed{Err: err.Error()}
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			return app.MsgServerConnected{}
		}
		return app.MsgServerFailed{
			Err: fmt.Sprintf("server returned status %d", resp.StatusCode),
		}
	}
}
```

**Step 3: Update app.go to handle MsgRevealTick**

In `app.go`, the `MsgRevealTick` message needs to reach the splash model. It's already handled: the default case at line 416-428 routes unknown messages to the splash model when `currentView == ViewSplash`. No changes needed here.

**Step 4: Run tests**

Run: `go test ./internal/tui/views/... -race -v`
Expected: PASS (update splash tests for new model fields)

**Step 5: Commit**

```bash
git add internal/tui/views/splash.go internal/tui/app/messages.go
git commit -m "feat: add typewriter reveal animation to splash screen"
```

---

### Task 9: Integration Testing & Polish

**Files:**
- All modified files from tasks 1-8

This task validates everything works together.

**Step 1: Build and run**

```bash
make build
```

Expected: Clean build, no errors.

**Step 2: Run all tests**

```bash
make test
```

Expected: All tests pass with race detector.

**Step 3: Manual smoke test checklist**

Run: `make dev-tui`

Verify:
- [ ] Splash screen: typewriter reveal animation, tagline appears after logo
- [ ] Header bars: amber thick bar on active column, dim bar on inactive
- [ ] Tab/Shift+Tab: cycles column focus, header bar follows
- [ ] Left column: ASCII logo visible, nav items present, user card shows, shortcuts at bottom
- [ ] Left column focused: j/k moves nav cursor, Enter navigates
- [ ] Center column: compose bar aligns with search bar on right
- [ ] Compose: multiline text wraps properly at box edge
- [ ] Posts: newlines preserved in rendered post cards
- [ ] Right column: trending and writers sections split equally
- [ ] Right column focused: j/k scrolls items, Enter toggles sections
- [ ] Scroll indicators (▲/▼) appear when items overflow

**Step 4: Fix any issues found**

Address any visual alignment, rendering, or interaction bugs discovered during smoke testing.

**Step 5: Final commit**

```bash
git add -A
git commit -m "polish: fix integration issues from TUI v3 component rewrite"
```

---

### Summary: Task Order & Dependencies

```
Task 1: Fix post card line break bug       (independent, critical)
Task 2: Add header bars to columns         (independent)
Task 3: Add ASCII art logo                 (independent)
Task 4: Rewrite sidebar                    (depends on Task 3 for LogoASCII)
Task 5: Rewrite discover panel             (independent)
Task 6: Add 50+ mock posts                 (depends on Task 1 for correct rendering)
Task 7: Fix compose bar width              (independent)
Task 8: Animated splash screen             (depends on Task 3 for logo)
Task 9: Integration testing                (depends on all above)
```

Parallel-safe groups:
- **Group A (independent):** Tasks 1, 2, 3, 5, 7 can all be done in parallel
- **Group B (depends on Group A):** Tasks 4, 6, 8
- **Group C (final):** Task 9
