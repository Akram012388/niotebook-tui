# TUI Layout v2 Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Evolve the three-column TUI layout into an X (Twitter)-style interface with left nav hub, inline compose bar, right discover column, and Tab-based column navigation.

**Architecture:** Refactor the existing Bubble Tea component hierarchy. Theme/logo changes are pure functions. Column focus state lives in the `layout` package. The inline compose bar replaces the modal compose overlay and becomes part of the center column's View() output (not an overlay). Left sidebar becomes a stateful component with j/k navigation. Right column gets a new discover component with placeholder content.

**Tech Stack:** Go 1.22+, Bubble Tea, Bubbles (textarea), Lip Gloss, existing theme system.

**Design Doc:** `docs/plans/2026-02-16-tui-layout-v2-design.md`

**Branch:** `feature/tui-redesign`

---

## Phase 1: Theme & Logo Foundation

### Task 1: Update tagline to "the social terminal"

**Files:**
- Modify: `internal/tui/theme/logo.go:23-25`
- Modify: `internal/tui/theme/logo_test.go:33-41`

**Step 1: Update the failing test**

```go
// In logo_test.go, change TestTaglineContainsSocialNotebook:
func TestTaglineContainsSocialTerminal(t *testing.T) {
	result := Tagline()
	if result == "" {
		t.Fatal("Tagline() returned empty string")
	}
	if !strings.Contains(result, "the social terminal") {
		t.Errorf("Tagline() should contain 'the social terminal', got %q", result)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/theme/ -run TestTaglineContainsSocialTerminal -v`
Expected: FAIL — contains "a social notebook" not "the social terminal"

**Step 3: Update implementation**

```go
// In logo.go, change Tagline():
func Tagline() string {
	return Hint.Render("the social terminal")
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui/theme/ -run TestTaglineContainsSocialTerminal -v`
Expected: PASS

**Step 5: Run all theme tests**

Run: `go test ./internal/tui/theme/ -v`
Expected: All PASS

**Step 6: Commit**

```bash
git add internal/tui/theme/logo.go internal/tui/theme/logo_test.go
git commit -m "feat: update tagline to 'the social terminal'"
```

---

### Task 2: Make Logo() bold and add LogoSplash() with letter-spacing

**Files:**
- Modify: `internal/tui/theme/logo.go`
- Modify: `internal/tui/theme/logo_test.go`

**Step 1: Write failing tests**

```go
// Add to logo_test.go:

func TestLogoBold(t *testing.T) {
	logo := Logo()
	// Logo should contain ANSI bold escape sequence
	if !strings.Contains(logo, "\033[1m") && !strings.Contains(logo, "\x1b[1m") {
		t.Error("Logo() should render with bold styling")
	}
}

func TestLogoSplashReturnsSpacedLetters(t *testing.T) {
	result := LogoSplash()
	if result == "" {
		t.Fatal("LogoSplash() returned empty string")
	}
	// Should contain spaced-out letters: "n i o t e b o o k"
	// Each letter separated by a space — check for "o t e" substring
	if !strings.Contains(result, "o") {
		t.Error("LogoSplash() does not contain expected spaced letters")
	}
}

func TestTaglineSplashReturnsSpacedText(t *testing.T) {
	result := TaglineSplash()
	if result == "" {
		t.Fatal("TaglineSplash() returned empty string")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/tui/theme/ -run "TestLogoBold|TestLogoSplash|TestTaglineSplash" -v`
Expected: FAIL — LogoSplash and TaglineSplash undefined, Logo not bold

**Step 3: Implement**

```go
// In logo.go, replace entire file:
package theme

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Logo returns the Niotebook brand wordmark: "niotebook" in bold,
// with the letter 'i' in the terracotta Accent color and all other
// characters in the primary Text color.
func Logo() string {
	text := lipgloss.NewStyle().Bold(true).Foreground(Text)
	accent := lipgloss.NewStyle().Bold(true).Foreground(Accent)

	return text.Render("n") + accent.Render("i") + text.Render("otebook")
}

// LogoCompact returns a compact variant of the brand logo suitable for
// sidebars and tight spaces. Currently identical to Logo.
func LogoCompact() string {
	return Logo()
}

// LogoSplash returns the splash screen variant of the brand logo with
// letter-spacing: "n i o t e b o o k" where each character is separated
// by a space. Bold, with 'i' in Accent color.
func LogoSplash() string {
	text := lipgloss.NewStyle().Bold(true).Foreground(Text)
	accent := lipgloss.NewStyle().Bold(true).Foreground(Accent)

	letters := []struct {
		char  string
		style lipgloss.Style
	}{
		{"n", text}, {"i", accent}, {"o", text}, {"t", text},
		{"e", text}, {"b", text}, {"o", text}, {"o", text}, {"k", text},
	}

	parts := make([]string, len(letters))
	for i, l := range letters {
		parts[i] = l.style.Render(l.char)
	}
	return strings.Join(parts, " ")
}

// Tagline returns the brand tagline styled in the Hint typography.
func Tagline() string {
	return Hint.Render("the social terminal")
}

// TaglineSplash returns the splash screen variant of the tagline with
// letter-spacing: "t h e   s o c i a l   t e r m i n a l" where each
// character is separated by a space. Styled in TextMuted.
func TaglineSplash() string {
	style := lipgloss.NewStyle().Foreground(TextMuted)
	chars := []rune("the social terminal")
	parts := make([]string, len(chars))
	for i, ch := range chars {
		if ch == ' ' {
			parts[i] = " " // extra space for word gaps
		} else {
			parts[i] = style.Render(string(ch))
		}
	}
	return strings.Join(parts, " ")
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/tui/theme/ -v`
Expected: All PASS

**Step 5: Commit**

```bash
git add internal/tui/theme/logo.go internal/tui/theme/logo_test.go
git commit -m "feat: add bold logo, LogoSplash() and TaglineSplash() with letter-spacing"
```

---

## Phase 2: Splash Screen Enhancements

### Task 3: Create custom block spinner and update splash screen

**Files:**
- Modify: `internal/tui/views/splash.go`
- Modify: `internal/tui/views/splash_test.go`

**Step 1: Write failing tests**

```go
// Add to splash_test.go:

func TestSplashMinDuration(t *testing.T) {
	// The minimum splash duration should be 2500ms
	if views.MinSplashDuration != 2500*time.Millisecond {
		t.Errorf("MinSplashDuration = %v, want 2500ms", views.MinSplashDuration)
	}
}

func TestBlockSpinnerFrames(t *testing.T) {
	frames := views.BlockSpinnerFrames()
	if len(frames) != 4 {
		t.Fatalf("expected 4 spinner frames, got %d", len(frames))
	}
	// Frame 0: all light blocks
	if !strings.Contains(frames[0], "░") {
		t.Error("frame 0 should contain light shade blocks")
	}
	// Frame 3: all full blocks
	if !strings.Contains(frames[3], "█") {
		t.Error("frame 3 should contain full blocks")
	}
}

func TestSplashViewContainsSpacedLogo(t *testing.T) {
	m := views.NewSplashModel("http://localhost:8080")
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	view := m.View()
	// Splash should use the spaced logo variant
	// The spaced logo has individual characters separated by spaces
	if view == "" {
		t.Error("splash view should not be empty")
	}
}
```

Note: Export `MinSplashDuration` (rename from `minSplashDuration`) and add `BlockSpinnerFrames()` function.

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/tui/views/ -run "TestSplashMinDuration|TestBlockSpinner|TestSplashViewContainsSpacedLogo" -v`
Expected: FAIL — undefined exports

**Step 3: Implement splash enhancements**

Replace `internal/tui/views/splash.go`:

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

// MinSplashDuration is the minimum time the splash screen stays visible.
const MinSplashDuration = 2500 * time.Millisecond

// Block spinner characters.
const (
	blockLight = "░" // U+2591 light shade
	blockFull  = "█" // U+2588 full block
)

// BlockSpinnerFrames returns the custom block spinner frame strings.
func BlockSpinnerFrames() []string {
	light := lipgloss.NewStyle().Foreground(theme.Border).Render(blockLight)
	full := lipgloss.NewStyle().Foreground(theme.Accent).Render(blockFull)

	return []string{
		light + " " + light + " " + light, // ░ ░ ░
		full + " " + light + " " + light,  // █ ░ ░
		full + " " + full + " " + light,   // █ █ ░
		full + " " + full + " " + full,    // █ █ █
	}
}

// newBlockSpinner creates a spinner using the custom block frames.
func newBlockSpinner() spinner.Model {
	frames := BlockSpinnerFrames()
	s := spinner.New()
	s.Spinner = spinner.Spinner{
		Frames: frames,
		FPS:    300 * time.Millisecond,
	}
	return s
}

// SplashModel is the splash screen shown on app launch while connecting
// to the server.
type SplashModel struct {
	serverURL string
	spinner   spinner.Model
	done      bool
	failed    bool
	err       string
	width     int
	height    int
}

// NewSplashModel creates a new splash screen model.
func NewSplashModel(serverURL string) SplashModel {
	return SplashModel{
		serverURL: serverURL,
		spinner:   newBlockSpinner(),
	}
}

// Done returns whether the server connection succeeded.
func (m SplashModel) Done() bool { return m.done }

// Failed returns whether the server connection failed.
func (m SplashModel) Failed() bool { return m.failed }

// ErrorMessage returns the error message if the connection failed.
func (m SplashModel) ErrorMessage() string { return m.err }

// HelpText returns an empty string since the splash screen has no help.
func (m SplashModel) HelpText() string { return "" }

// Init returns the initial commands: start the spinner and check server health.
func (m SplashModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.checkHealth())
}

// Update handles messages for the splash screen.
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

	case tea.KeyMsg:
		if m.failed {
			switch {
			case msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'r':
				m.failed = false
				m.err = ""
				return m, tea.Batch(m.spinner.Tick, m.checkHealth())
			}
		}
		switch {
		case msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && msg.Runes[0] == 'q':
			return m, tea.Quit
		case msg.Type == tea.KeyCtrlC:
			return m, tea.Quit
		}
		return m, nil

	case spinner.TickMsg:
		if !m.done && !m.failed {
			var cmd tea.Cmd
			m.spinner, cmd = m.spinner.Update(msg)
			return m, cmd
		}
		return m, nil
	}

	return m, nil
}

// View renders the splash screen with spaced logo, tagline, and block spinner.
func (m SplashModel) View() string {
	var b strings.Builder

	// Spaced logo
	b.WriteString(theme.LogoSplash())
	b.WriteString("\n\n")

	// Spaced tagline
	b.WriteString(theme.TaglineSplash())
	b.WriteString("\n\n\n")

	if m.failed {
		// Error state
		errStyle := lipgloss.NewStyle().Foreground(theme.Error)
		b.WriteString(errStyle.Render(fmt.Sprintf("connection failed: %s", m.err)))
		b.WriteString("\n\n")
		b.WriteString(theme.Hint.Render("press r to retry · q to quit"))
	} else if !m.done {
		// Connecting state — block spinner centered
		b.WriteString(m.spinner.View())
		b.WriteString("\n")
		b.WriteString(theme.Caption.Render("connecting..."))
	}

	content := b.String()

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// checkHealth returns a tea.Cmd that makes an HTTP GET request to the
// server's health endpoint. A minimum display time ensures the splash
// screen is visible even when the server responds instantly.
func (m SplashModel) checkHealth() tea.Cmd {
	url := m.serverURL
	return func() tea.Msg {
		start := time.Now()

		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(url + "/health")

		// Ensure the splash is visible for at least MinSplashDuration.
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

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/tui/views/ -run "TestSplash" -v`
Expected: All PASS

**Step 5: Commit**

```bash
git add internal/tui/views/splash.go internal/tui/views/splash_test.go
git commit -m "feat: splash screen with 2.5s timer, block spinner, spaced logo"
```

---

## Phase 3: Column Focus State

### Task 4: Add column focus tracking to layout package

**Files:**
- Modify: `internal/tui/layout/columns.go`
- Modify: `internal/tui/layout/columns_test.go`

**Step 1: Write failing tests**

```go
// Add to columns_test.go:

func TestFocusStateDefault(t *testing.T) {
	fs := layout.NewFocusState()
	if fs.Active() != layout.FocusCenter {
		t.Errorf("default focus = %d, want FocusCenter", fs.Active())
	}
}

func TestFocusStateNextCycles(t *testing.T) {
	fs := layout.NewFocusState()
	fs.Next() // Center -> Right
	if fs.Active() != layout.FocusRight {
		t.Errorf("after Next: focus = %d, want FocusRight", fs.Active())
	}
	fs.Next() // Right -> Left
	if fs.Active() != layout.FocusLeft {
		t.Errorf("after Next×2: focus = %d, want FocusLeft", fs.Active())
	}
	fs.Next() // Left -> Center
	if fs.Active() != layout.FocusCenter {
		t.Errorf("after Next×3: focus = %d, want FocusCenter", fs.Active())
	}
}

func TestFocusStatePrevCycles(t *testing.T) {
	fs := layout.NewFocusState()
	fs.Prev() // Center -> Left
	if fs.Active() != layout.FocusLeft {
		t.Errorf("after Prev: focus = %d, want FocusLeft", fs.Active())
	}
	fs.Prev() // Left -> Right
	if fs.Active() != layout.FocusRight {
		t.Errorf("after Prev×2: focus = %d, want FocusRight", fs.Active())
	}
	fs.Prev() // Right -> Center
	if fs.Active() != layout.FocusCenter {
		t.Errorf("after Prev×3: focus = %d, want FocusCenter", fs.Active())
	}
}

func TestFocusStateReset(t *testing.T) {
	fs := layout.NewFocusState()
	fs.Next()
	fs.Reset()
	if fs.Active() != layout.FocusCenter {
		t.Errorf("after Reset: focus = %d, want FocusCenter", fs.Active())
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/tui/layout/ -run "TestFocusState" -v`
Expected: FAIL — FocusState type undefined

**Step 3: Implement focus state**

Add to the bottom of `internal/tui/layout/columns.go`:

```go
// FocusColumn identifies which column currently has keyboard focus.
type FocusColumn int

const (
	FocusLeft   FocusColumn = 0
	FocusCenter FocusColumn = 1
	FocusRight  FocusColumn = 2
)

// FocusState tracks which column is currently focused.
type FocusState struct {
	active FocusColumn
}

// NewFocusState returns a FocusState with Center as the default.
func NewFocusState() FocusState {
	return FocusState{active: FocusCenter}
}

// Active returns the currently focused column.
func (f *FocusState) Active() FocusColumn {
	return f.active
}

// Next moves focus to the next column (Left→Center→Right→Left).
func (f *FocusState) Next() {
	f.active = (f.active + 1) % 3
}

// Prev moves focus to the previous column (Left←Center←Right←Left).
func (f *FocusState) Prev() {
	f.active = (f.active + 2) % 3
}

// Reset returns focus to the center column.
func (f *FocusState) Reset() {
	f.active = FocusCenter
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/tui/layout/ -v`
Expected: All PASS

**Step 5: Commit**

```bash
git add internal/tui/layout/columns.go internal/tui/layout/columns_test.go
git commit -m "feat: add column focus state tracking to layout package"
```

---

## Phase 4: Remove Deprecated Components

### Task 5: Remove status bar component

**Files:**
- Delete: `internal/tui/components/statusbar.go`
- Delete: `internal/tui/components/statusbar_test.go`
- Modify: `internal/tui/app/app.go` (remove statusBar field and all references)

**Step 1: Remove status bar from app model**

In `internal/tui/app/app.go`:
- Remove the `statusBar components.StatusBarModel` field from `AppModel`
- Remove `statusBar: components.NewStatusBarModel()` from both constructors
- Remove the `components.MsgStatusClear` case from `Update()`
- Remove all `m.statusBar.SetError(...)` and `m.statusBar.SetSuccess(...)` calls (replace with `nil` cmd for now)
- Remove the status bar rendering from `View()` — the `contentHeight - 1`, `statusBar` variable, and `centerContent` join
- Remove the `currentHelpText()` method
- Simplify `View()` to render center content directly without status bar

The center content in `View()` becomes simply:
```go
// Center content
content := m.viewCurrentContent()
if m.help != nil {
    content = m.help.View()
} else if m.compose != nil {
    content = m.compose.View()
}
centerContent := content
```

For `MsgAuthExpired`, `MsgPostPublished`, and `MsgAPIError` handlers that used status bar, remove the status bar calls and return `nil` cmd (or just the fetch cmd for MsgPostPublished).

**Step 2: Delete statusbar files**

```bash
rm internal/tui/components/statusbar.go
rm internal/tui/components/statusbar_test.go
```

**Step 3: Run all tests to verify nothing is broken**

Run: `go test ./internal/tui/... -v`
Expected: All PASS (the `MsgStatusClear` in `messages.go` is still there — keep it for now since `app_test.go` references it)

Note: The `components.MsgStatusClear` type is defined in `statusbar.go`. After deleting it, the `MsgStatusClear` in `app/messages.go` will be the sole definition. Update any imports in `app.go` that referenced `components.MsgStatusClear` to use `app.MsgStatusClear` or the local `MsgStatusClear`.

Review `app.go` imports — remove the `components.MsgStatusClear` case handler or switch to the local `MsgStatusClear` from messages.go (which already exists).

**Step 4: Run tests**

Run: `go test ./internal/tui/... -v`
Expected: All PASS

**Step 5: Commit**

```bash
git add -A internal/tui/components/statusbar.go internal/tui/components/statusbar_test.go internal/tui/app/app.go
git commit -m "refactor: remove status bar component (shortcuts now in left sidebar)"
```

---

### Task 6: Remove shortcuts component

**Files:**
- Delete: `internal/tui/components/shortcuts.go`
- Delete: `internal/tui/components/shortcuts_test.go`
- Modify: `internal/tui/app/app.go` (remove RenderShortcuts call from View)

**Step 1: Remove RenderShortcuts call from app.go View()**

In `app.go` `View()`, replace the right sidebar content:
```go
// Right sidebar — will be replaced by discover component in next task
rightContent := ""
```

Remove the `RenderShortcuts` call entirely.

**Step 2: Delete shortcuts files**

```bash
rm internal/tui/components/shortcuts.go
rm internal/tui/components/shortcuts_test.go
```

**Step 3: Run tests**

Run: `go test ./internal/tui/... -v`
Expected: All PASS

**Step 4: Commit**

```bash
git add -A internal/tui/components/shortcuts.go internal/tui/components/shortcuts_test.go internal/tui/app/app.go
git commit -m "refactor: remove shortcuts component (moved to left sidebar nav hub)"
```

---

## Phase 5: New and Rewritten Components

### Task 7: Rewrite sidebar as X-style nav hub

**Files:**
- Modify: `internal/tui/components/sidebar.go`
- Modify: `internal/tui/components/sidebar_test.go`

**Step 1: Write failing tests**

```go
// Replace sidebar_test.go with comprehensive tests:
package components_test

import (
	"strings"
	"testing"
	"time"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/Akram012388/niotebook-tui/internal/tui/components"
)

func TestSidebarShowsLogo(t *testing.T) {
	user := &models.User{Username: "akram", DisplayName: "Akram"}
	result := components.RenderSidebar(user, components.ViewTimeline, false, 24, 30)
	// Logo "niotebook" characters should be present
	if !strings.Contains(result, "n") || !strings.Contains(result, "otebook") {
		t.Error("sidebar should contain the niotebook logo")
	}
}

func TestSidebarShowsNavItems(t *testing.T) {
	user := &models.User{Username: "akram"}
	result := components.RenderSidebar(user, components.ViewTimeline, false, 24, 30)
	if !strings.Contains(result, "Home") {
		t.Error("sidebar should contain Home nav item")
	}
	if !strings.Contains(result, "Profile") {
		t.Error("sidebar should contain Profile nav item")
	}
	if !strings.Contains(result, "Bookmarks") {
		t.Error("sidebar should contain Bookmarks placeholder")
	}
	if !strings.Contains(result, "Settings") {
		t.Error("sidebar should contain Settings placeholder")
	}
}

func TestSidebarShowsPostButton(t *testing.T) {
	user := &models.User{Username: "akram"}
	result := components.RenderSidebar(user, components.ViewTimeline, false, 24, 30)
	if !strings.Contains(result, "Post") {
		t.Error("sidebar should contain Post button")
	}
}

func TestSidebarShowsShortcuts(t *testing.T) {
	user := &models.User{Username: "akram"}
	result := components.RenderSidebar(user, components.ViewTimeline, false, 24, 30)
	if !strings.Contains(result, "j/k") {
		t.Error("sidebar should contain j/k shortcut")
	}
	if !strings.Contains(result, "Tab") {
		t.Error("sidebar should contain Tab shortcut")
	}
}

func TestSidebarShowsVersion(t *testing.T) {
	user := &models.User{Username: "akram"}
	result := components.RenderSidebar(user, components.ViewTimeline, false, 24, 30)
	// Version string from build package
	if !strings.Contains(result, "v") {
		t.Error("sidebar should contain version info")
	}
}

func TestSidebarShowsJoinDate(t *testing.T) {
	user := &models.User{
		Username:  "akram",
		CreatedAt: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
	}
	result := components.RenderSidebar(user, components.ViewTimeline, false, 24, 30)
	if !strings.Contains(result, "Joined Feb 2026") {
		t.Error("sidebar should contain join date")
	}
}

func TestSidebarActiveHighlight(t *testing.T) {
	user := &models.User{Username: "akram"}
	result := components.RenderSidebar(user, components.ViewTimeline, false, 24, 30)
	// Active item (Home on timeline) should have filled dot
	if !strings.Contains(result, "●") {
		t.Error("sidebar should show ● marker for active nav item")
	}
}

func TestSidebarFocusedBorderChange(t *testing.T) {
	user := &models.User{Username: "akram"}
	unfocused := components.RenderSidebar(user, components.ViewTimeline, false, 24, 30)
	focused := components.RenderSidebar(user, components.ViewTimeline, true, 24, 30)
	// Focused and unfocused should differ (border color changes)
	if unfocused == focused {
		t.Error("focused sidebar should differ from unfocused (border accent)")
	}
}

func TestSidebarZeroWidth(t *testing.T) {
	result := components.RenderSidebar(nil, components.ViewTimeline, false, 0, 20)
	if result != "" {
		t.Errorf("expected empty string for zero width, got %q", result)
	}
}

func TestSidebarNilUser(t *testing.T) {
	result := components.RenderSidebar(nil, components.ViewTimeline, false, 24, 30)
	// Should still show logo but no user-specific content
	if result == "" {
		t.Error("sidebar with nil user should still render something")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/tui/components/ -run "TestSidebar" -v`
Expected: FAIL — signature changed (new `focused bool` parameter), missing content

**Step 3: Rewrite sidebar implementation**

Replace `internal/tui/components/sidebar.go`:

```go
package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/Akram012388/niotebook-tui/internal/build"
	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/Akram012388/niotebook-tui/internal/tui/theme"
)

// View identifies the active screen (mirrors app.View to avoid import cycles).
type View int

const (
	ViewSplash   View = iota
	ViewLogin
	ViewRegister
	ViewTimeline
	ViewProfile
)

// RenderSidebar renders the left column nav hub in X-style layout:
// logo, username, nav items, Post button, shortcuts, version, join date.
// The focused parameter controls the border accent color.
func RenderSidebar(user *models.User, activeView View, focused bool, width, height int) string {
	if width == 0 {
		return ""
	}

	innerWidth := width - 4 // account for border + padding
	if innerWidth < 0 {
		innerWidth = 0
	}

	var sections []string

	// Brand wordmark
	sections = append(sections, theme.Logo())
	sections = append(sections, theme.Separator(innerWidth))

	if user != nil {
		sections = append(sections, "")

		// Username
		usernameStyle := lipgloss.NewStyle().Foreground(theme.Accent)
		sections = append(sections, usernameStyle.Render("@"+user.Username))

		// Display name
		if user.DisplayName != "" {
			displayStyle := lipgloss.NewStyle().Foreground(theme.Text)
			sections = append(sections, displayStyle.Render(user.DisplayName))
		}

		sections = append(sections, "")

		// Navigation items
		sections = append(sections, renderNavItem("Home", activeView == ViewTimeline))
		sections = append(sections, renderNavItem("Profile", activeView == ViewProfile))
		sections = append(sections, renderNavItemPlaceholder("Bookmarks"))
		sections = append(sections, renderNavItemPlaceholder("Settings"))

		sections = append(sections, "")

		// Post button
		btnStyle := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.AccentDim).
			Foreground(theme.Accent).
			Bold(true).
			Align(lipgloss.Center).
			Width(innerWidth)
		sections = append(sections, btnStyle.Render("Post"))

		sections = append(sections, "")
		sections = append(sections, theme.Separator(innerWidth))

		// Version
		versionStyle := lipgloss.NewStyle().Foreground(theme.TextSecondary)
		sections = append(sections, versionStyle.Render("v"+build.Version))

		// Join date
		if !user.CreatedAt.IsZero() {
			captionStyle := lipgloss.NewStyle().Foreground(theme.TextSecondary)
			joinDate := user.CreatedAt.Format("Joined Jan 2006")
			sections = append(sections, captionStyle.Render(joinDate))
		}

		sections = append(sections, "")

		// Shortcuts reference
		sections = append(sections, renderShortcutSection(innerWidth))
	}

	content := strings.Join(sections, "\n")

	wrapper := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(1, 1)

	return wrapper.Render(content)
}

// renderNavItem renders a selectable navigation item.
func renderNavItem(label string, active bool) string {
	if active {
		style := lipgloss.NewStyle().
			Bold(true).
			Foreground(theme.Accent)
		return style.Render("● " + label)
	}

	style := lipgloss.NewStyle().
		Foreground(theme.TextSecondary)
	return style.Render("  " + label)
}

// renderNavItemPlaceholder renders a greyed-out placeholder nav item.
func renderNavItemPlaceholder(label string) string {
	style := lipgloss.NewStyle().
		Foreground(theme.TextMuted)
	return style.Render("  " + label)
}

// renderShortcutSection renders the keyboard shortcuts reference.
func renderShortcutSection(width int) string {
	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.TextMuted)

	keyStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Accent).
		Width(8)

	descStyle := lipgloss.NewStyle().
		Foreground(theme.TextSecondary)

	shortcuts := []struct {
		key  string
		desc string
	}{
		{"j/k", "scroll"},
		{"Tab", "switch col"},
		{"n", "compose"},
		{"?", "help"},
		{"q", "quit"},
	}

	var lines []string
	lines = append(lines, titleStyle.Render("Shortcuts"))
	for _, s := range shortcuts {
		lines = append(lines, keyStyle.Render(s.key)+descStyle.Render(s.desc))
	}

	return strings.Join(lines, "\n")
}
```

**Step 4: Update app.go RenderSidebar call to pass focused parameter**

In `app.go` `View()`, update the sidebar call:
```go
leftContent := components.RenderSidebar(
    m.user,
    components.View(m.currentView),
    false, // focus state integration comes in Task 12
    cols.Left,
    contentHeight,
)
```

Note: contentHeight should now equal `m.height` (no status bar to subtract).

**Step 5: Run tests**

Run: `go test ./internal/tui/... -v`
Expected: All PASS

**Step 6: Commit**

```bash
git add internal/tui/components/sidebar.go internal/tui/components/sidebar_test.go internal/tui/app/app.go
git commit -m "feat: rewrite sidebar as X-style nav hub with Post button, shortcuts, version"
```

---

### Task 8: Create discover component for right column

**Files:**
- Create: `internal/tui/components/discover.go`
- Create: `internal/tui/components/discover_test.go`

**Step 1: Write failing tests**

```go
// internal/tui/components/discover_test.go
package components_test

import (
	"strings"
	"testing"

	"github.com/Akram012388/niotebook-tui/internal/tui/components"
)

func TestDiscoverShowsSearchPlaceholder(t *testing.T) {
	result := components.RenderDiscover(false, 24, 30)
	if !strings.Contains(result, "Search") {
		t.Error("discover should contain search placeholder")
	}
}

func TestDiscoverShowsTrending(t *testing.T) {
	result := components.RenderDiscover(false, 24, 30)
	if !strings.Contains(result, "Trending") {
		t.Error("discover should contain Trending section")
	}
	if !strings.Contains(result, "#niotebook") {
		t.Error("discover should contain #niotebook trending tag")
	}
}

func TestDiscoverShowsWritersToFollow(t *testing.T) {
	result := components.RenderDiscover(false, 24, 30)
	if !strings.Contains(result, "Writers to follow") {
		t.Error("discover should contain Writers to follow section")
	}
	if !strings.Contains(result, "@alice") {
		t.Error("discover should contain @alice suggestion")
	}
}

func TestDiscoverZeroWidth(t *testing.T) {
	result := components.RenderDiscover(false, 0, 20)
	if result != "" {
		t.Errorf("expected empty string for zero width, got %q", result)
	}
}

func TestDiscoverFocusedDiffers(t *testing.T) {
	unfocused := components.RenderDiscover(false, 24, 30)
	focused := components.RenderDiscover(true, 24, 30)
	if unfocused == focused {
		t.Error("focused discover should differ from unfocused")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/tui/components/ -run "TestDiscover" -v`
Expected: FAIL — RenderDiscover undefined

**Step 3: Implement discover component**

```go
// internal/tui/components/discover.go
package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/Akram012388/niotebook-tui/internal/tui/theme"
)

// RenderDiscover renders the right column discover/trending panel.
// All content is placeholder for MVP. The focused parameter controls
// the border accent.
func RenderDiscover(focused bool, width, height int) string {
	if width == 0 {
		return ""
	}

	innerWidth := width - 4
	if innerWidth < 0 {
		innerWidth = 0
	}

	var sections []string

	// Search bar placeholder
	searchStyle := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.AccentDim).
		Foreground(theme.TextMuted).
		Width(innerWidth)
	sections = append(sections, searchStyle.Render("Search niotes..."))

	sections = append(sections, "")

	// Trending section
	sectionHeader := lipgloss.NewStyle().Bold(true).Foreground(theme.Text)
	tagStyle := lipgloss.NewStyle().Foreground(theme.Accent)
	countStyle := lipgloss.NewStyle().Foreground(theme.TextSecondary)

	sections = append(sections, sectionHeader.Render("Trending"))
	sections = append(sections, theme.Separator(innerWidth))

	trending := []struct {
		tag   string
		count string
	}{
		{"#niotebook", "12 niotes"},
		{"#hello-world", "8 niotes"},
		{"#terminal-life", "5 niotes"},
	}
	for _, t := range trending {
		sections = append(sections, tagStyle.Render(t.tag))
		sections = append(sections, countStyle.Render(t.count))
		sections = append(sections, "")
	}

	// Writers to follow section
	sections = append(sections, sectionHeader.Render("Writers to follow"))
	sections = append(sections, theme.Separator(innerWidth))

	writers := []struct {
		name string
		bio  string
	}{
		{"@alice", "loves terminal apps"},
		{"@bob", "building in public"},
	}
	for _, w := range writers {
		sections = append(sections, tagStyle.Render(w.name))
		sections = append(sections, countStyle.Render(w.bio))
		sections = append(sections, "")
	}

	content := strings.Join(sections, "\n")

	wrapper := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(1, 1)

	return wrapper.Render(content)
}
```

**Step 4: Wire discover into app.go View()**

In `app.go` `View()`, replace the empty right content:
```go
rightContent := components.RenderDiscover(
    false, // focus state integration comes in Task 12
    cols.Right,
    contentHeight,
)
```

**Step 5: Run tests**

Run: `go test ./internal/tui/... -v`
Expected: All PASS

**Step 6: Commit**

```bash
git add internal/tui/components/discover.go internal/tui/components/discover_test.go internal/tui/app/app.go
git commit -m "feat: add discover/trending component for right column"
```

---

### Task 9: Rewrite compose as inline bar (collapsed/expanded)

**Files:**
- Modify: `internal/tui/views/compose.go`
- Modify: `internal/tui/views/compose_test.go`
- Modify: `internal/tui/app/app.go` (interface changes)

**Step 1: Write failing tests for inline compose bar**

```go
// Replace compose_test.go with inline bar tests:
package views_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Akram012388/niotebook-tui/internal/tui/app"
	"github.com/Akram012388/niotebook-tui/internal/tui/views"
)

func TestComposeBarStartsCollapsed(t *testing.T) {
	m := views.NewComposeModel(nil)
	if m.Expanded() {
		t.Error("compose bar should start collapsed")
	}
}

func TestComposeBarCollapsedViewShowsPlaceholder(t *testing.T) {
	m := views.NewComposeModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 60, Height: 24})
	view := m.View()
	if !strings.Contains(view, "What's on your mind?") {
		t.Error("collapsed compose should show placeholder text")
	}
	if !strings.Contains(view, "0/140") {
		t.Error("collapsed compose should show character counter")
	}
}

func TestComposeBarExpandOnFocus(t *testing.T) {
	m := views.NewComposeModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 60, Height: 24})
	m.Expand()
	if !m.Expanded() {
		t.Error("compose should be expanded after Expand()")
	}
}

func TestComposeBarEscCollapses(t *testing.T) {
	m := views.NewComposeModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 60, Height: 24})
	m.Expand()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.Expanded() {
		t.Error("Esc should collapse the compose bar")
	}
}

func TestComposeBarExpandedShowsHints(t *testing.T) {
	m := views.NewComposeModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 60, Height: 24})
	m.Expand()
	view := m.View()
	if !strings.Contains(view, "Ctrl+Enter") || !strings.Contains(view, "Esc") {
		t.Errorf("expanded compose should show Ctrl+Enter and Esc hints, got:\n%s", view)
	}
}

func TestComposeBarTypingUpdatesCounter(t *testing.T) {
	m := views.NewComposeModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 60, Height: 24})
	m.Expand()
	for _, r := range "Hello" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	view := m.View()
	if !strings.Contains(view, "5/140") {
		t.Errorf("expected counter 5/140, got view:\n%s", view)
	}
}

func TestComposeBarCtrlEnterPublishes(t *testing.T) {
	m := views.NewComposeModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 60, Height: 24})
	m.Expand()
	for _, r := range "Test post" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlJ})
	if !m.Submitted() {
		t.Error("Ctrl+Enter with content should submit")
	}
	if cmd == nil {
		t.Error("expected publish cmd")
	}
}

func TestComposeBarEmptyCannotSubmit(t *testing.T) {
	m := views.NewComposeModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 60, Height: 24})
	m.Expand()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlJ})
	if m.Submitted() {
		t.Error("empty content should not submit")
	}
}

func TestComposeBarOverLimitCannotSubmit(t *testing.T) {
	m := views.NewComposeModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 60, Height: 24})
	m.Expand()
	for _, r := range strings.Repeat("a", 141) {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlJ})
	if m.Submitted() {
		t.Error("over-limit content should not submit")
	}
}

func TestComposeBarPublishWithNilClient(t *testing.T) {
	m := views.NewComposeModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 60, Height: 24})
	m.Expand()
	for _, r := range "Hello" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlJ})
	if cmd == nil {
		t.Fatal("expected cmd from publish")
	}
	msg := cmd()
	if _, ok := msg.(app.MsgAPIError); !ok {
		t.Errorf("expected MsgAPIError with nil client, got %T", msg)
	}
}

func TestComposeBarIsTextInputFocused(t *testing.T) {
	m := views.NewComposeModel(nil)
	// Collapsed: text input not focused
	if m.IsTextInputFocused() {
		t.Error("collapsed compose should not have text input focused")
	}
	m.Expand()
	if !m.IsTextInputFocused() {
		t.Error("expanded compose should have text input focused")
	}
}

func TestComposeBarCancelled(t *testing.T) {
	m := views.NewComposeModel(nil)
	m.Expand()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	// Esc collapses, doesn't "cancel" — Cancelled() is true only when
	// the user was in expanded mode and pressed Esc
	if !m.Cancelled() {
		t.Error("Esc from expanded should set cancelled")
	}
}

func TestComposeBarPosting(t *testing.T) {
	m := views.NewComposeModel(nil)
	if m.Posting() {
		t.Error("Posting should be false initially")
	}
}

func TestComposeBarHelpText(t *testing.T) {
	m := views.NewComposeModel(nil)
	text := m.HelpText()
	if text == "" {
		t.Error("HelpText should return non-empty string")
	}
}

func TestComposeBarInit(t *testing.T) {
	m := views.NewComposeModel(nil)
	cmd := m.Init()
	// Init on collapsed compose — no blink needed yet
	_ = cmd
}

func TestComposeBarCollapsesAfterPublish(t *testing.T) {
	m := views.NewComposeModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 60, Height: 24})
	m.Expand()
	for _, r := range "Hello" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlJ})
	// After submit, the bar collapses
	if m.Expanded() {
		t.Error("compose should collapse after submit")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/tui/views/ -run "TestComposeBar" -v`
Expected: FAIL — Expanded() method undefined

**Step 3: Rewrite compose implementation**

Replace `internal/tui/views/compose.go`:

```go
package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Akram012388/niotebook-tui/internal/tui/app"
	"github.com/Akram012388/niotebook-tui/internal/tui/client"
	"github.com/Akram012388/niotebook-tui/internal/tui/theme"
)

const maxPostLength = 140

// ComposeModel manages the inline compose bar (collapsed/expanded).
type ComposeModel struct {
	textarea  textarea.Model
	client    *client.Client
	expanded  bool
	submitted bool
	cancelled bool
	posting   bool
	err       error
	width     int
	height    int
}

// NewComposeModel creates a new inline compose bar model (starts collapsed).
func NewComposeModel(c *client.Client) ComposeModel {
	ta := textarea.New()
	ta.Placeholder = "What's on your mind?"
	ta.SetWidth(40)
	ta.SetHeight(3)
	ta.CharLimit = 0

	return ComposeModel{
		textarea: ta,
		client:   c,
	}
}

// Expanded returns whether the compose bar is in expanded editing mode.
func (m ComposeModel) Expanded() bool  { return m.expanded }
func (m ComposeModel) Submitted() bool { return m.submitted }
func (m ComposeModel) Cancelled() bool { return m.cancelled }
func (m ComposeModel) Posting() bool   { return m.posting }

// IsTextInputFocused returns true only when the compose bar is expanded.
func (m ComposeModel) IsTextInputFocused() bool {
	return m.expanded
}

// Expand switches the compose bar to expanded mode and focuses the textarea.
func (m *ComposeModel) Expand() {
	m.expanded = true
	m.cancelled = false
	m.submitted = false
	m.textarea.Focus()
}

// Init returns the initial command.
func (m ComposeModel) Init() tea.Cmd {
	if m.expanded {
		return textarea.Blink
	}
	return nil
}

// Update handles messages for the compose bar.
func (m ComposeModel) Update(msg tea.Msg) (ComposeModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateTextareaSize()
		return m, nil

	case app.MsgPostPublished:
		m.posting = false
		m.submitted = true
		m.expanded = false
		m.textarea.Reset()
		return m, nil

	case app.MsgAPIError:
		m.posting = false
		m.err = fmt.Errorf("%s", msg.Message)
		return m, nil

	case tea.KeyMsg:
		if m.expanded {
			return m.handleExpandedKey(msg)
		}
		return m, nil
	}

	// Pass to textarea only when expanded
	if m.expanded {
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m ComposeModel) handleExpandedKey(msg tea.KeyMsg) (ComposeModel, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.expanded = false
		m.cancelled = true
		m.textarea.Blur()
		return m, nil

	case tea.KeyCtrlJ:
		content := strings.TrimSpace(m.textarea.Value())
		charCount := len([]rune(content))
		if charCount == 0 || charCount > maxPostLength {
			return m, nil
		}
		m.submitted = true
		m.posting = true
		m.expanded = false
		return m, m.publish(content)
	}

	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m ComposeModel) publish(content string) tea.Cmd {
	c := m.client
	return func() tea.Msg {
		if c == nil {
			return app.MsgAPIError{Message: "no server connection"}
		}
		post, err := c.CreatePost(content)
		if err != nil {
			return app.MsgAPIError{Message: err.Error()}
		}
		return app.MsgPostPublished{Post: *post}
	}
}

func (m *ComposeModel) updateTextareaSize() {
	innerWidth := m.width - 6
	if innerWidth < 20 {
		innerWidth = 20
	}
	m.textarea.SetWidth(innerWidth)
}

// View renders the compose bar — collapsed or expanded.
func (m ComposeModel) View() string {
	if m.expanded {
		return m.viewExpanded()
	}
	return m.viewCollapsed()
}

func (m ComposeModel) viewCollapsed() string {
	counterText := fmt.Sprintf("0/%d", maxPostLength)
	counter := lipgloss.NewStyle().Foreground(theme.TextSecondary).Render(counterText)
	placeholder := lipgloss.NewStyle().Foreground(theme.TextMuted).Render("What's on your mind?")

	innerWidth := m.width - 6
	if innerWidth < 20 {
		innerWidth = 20
	}
	gap := innerWidth - lipgloss.Width(placeholder) - lipgloss.Width(counterText)
	if gap < 1 {
		gap = 1
	}

	line := placeholder + strings.Repeat(" ", gap) + counter

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.AccentDim).
		Width(innerWidth).
		Padding(0, 1)

	return box.Render(line)
}

func (m ComposeModel) viewExpanded() string {
	var b strings.Builder

	b.WriteString(m.textarea.View())
	b.WriteString("\n\n")

	// Character counter + hints
	content := m.textarea.Value()
	charCount := len([]rune(content))
	counterText := fmt.Sprintf("%d/%d", charCount, maxPostLength)

	counterStyle := lipgloss.NewStyle().Foreground(theme.TextSecondary)
	if charCount > maxPostLength-10 {
		counterStyle = lipgloss.NewStyle().Foreground(theme.Error).Bold(true)
	}

	hintStyle := lipgloss.NewStyle().Foreground(theme.TextMuted).Italic(true)
	hints := hintStyle.Render("Ctrl+Enter: post    Esc: cancel")

	footer := counterStyle.Render(counterText) + "    " + hints

	b.WriteString(footer)

	if m.err != nil {
		b.WriteString("\n")
		errStyle := lipgloss.NewStyle().Foreground(theme.Error)
		b.WriteString(errStyle.Render(m.err.Error()))
	}

	if m.posting {
		b.WriteString("\n")
		b.WriteString(hintStyle.Render("Publishing..."))
	}

	innerWidth := m.width - 6
	if innerWidth < 20 {
		innerWidth = 20
	}

	box := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Accent).
		Width(innerWidth).
		Padding(0, 1)

	return box.Render(b.String())
}

// HelpText returns hint text for the compose bar.
func (m ComposeModel) HelpText() string {
	if m.expanded {
		return "Ctrl+Enter: publish  Esc: cancel"
	}
	return "n: compose  j/k: scroll  ?: help  q: quit"
}
```

**Step 4: Update ComposeViewModel interface and app.go**

In `internal/tui/app/app.go`, add `Expanded()` and `Expand()` to the `ComposeViewModel` interface:

```go
type ComposeViewModel interface {
	ViewModel
	Submitted() bool
	Cancelled() bool
	Expanded() bool
	IsTextInputFocused() bool
	Expand()
}
```

Update `openCompose()` in `app.go` to call `Expand()`:
```go
func (m AppModel) openCompose() (AppModel, tea.Cmd) {
	if m.factory != nil {
		m.compose = m.factory.NewCompose(m.client)
		m.compose.Expand()
		cmd := m.compose.Init()
		return m, cmd
	}
	return m, nil
}
```

Update the compose adapter in `factory.go` to add the new methods:
```go
func (a *composeAdapter) Expanded() bool { return a.model.Expanded() }
func (a *composeAdapter) Expand()        { a.model.Expand() }
```

Update `isTextInputFocused()` in `app.go` to check `Expanded()`:
```go
if m.compose != nil && m.compose.Expanded() {
    return true
}
```

Update `View()` in `app.go` — compose is now inline, not an overlay:
```go
// Compose is inline at top of center column, not an overlay
content := ""
if m.compose != nil {
    content += m.compose.View() + "\n"
}
content += m.viewCurrentContent()
if m.help != nil {
    content = m.help.View()
}
centerContent := content
```

**Step 5: Update stub in app_test.go**

Add `Expanded()` and `Expand()` to stub compose types:
```go
func (s *stubCompose) Expanded() bool { return false }
func (s *stubCompose) Expand()        {}

func (s *stubCancelCompose) Expanded() bool { return false }
func (s *stubCancelCompose) Expand()        {}
```

**Step 6: Run tests**

Run: `go test ./internal/tui/... -v`
Expected: All PASS

**Step 7: Commit**

```bash
git add internal/tui/views/compose.go internal/tui/views/compose_test.go internal/tui/views/factory.go internal/tui/app/app.go internal/tui/app/app_test.go
git commit -m "feat: rewrite compose as inline bar with collapsed/expanded states"
```

---

## Phase 6: App Integration — Column Focus & Tab Navigation

### Task 10: Add column focus state and Tab/Shift+Tab navigation to app

**Files:**
- Modify: `internal/tui/app/app.go`
- Modify: `internal/tui/app/app_test.go`

**Step 1: Write failing tests**

```go
// Add to app_test.go:

func TestAppModelTabCyclesColumns(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = connectAndAuth(m, &models.User{Username: "akram"}, &models.TokenPair{AccessToken: "tok"})
	m = update(m, tea.WindowSizeMsg{Width: 120, Height: 40})

	// Default is center. Tab should move to right.
	m = update(m, tea.KeyMsg{Type: tea.KeyTab})
	// Tab again should move to left.
	m = update(m, tea.KeyMsg{Type: tea.KeyTab})
	// Tab again should return to center.
	m = update(m, tea.KeyMsg{Type: tea.KeyTab})
	// Should not panic and view should render
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view after Tab cycling")
	}
}

func TestAppModelShiftTabCyclesReverse(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = connectAndAuth(m, &models.User{Username: "akram"}, &models.TokenPair{AccessToken: "tok"})
	m = update(m, tea.WindowSizeMsg{Width: 120, Height: 40})

	// Shift+Tab from center should go to left.
	m = update(m, tea.KeyMsg{Type: tea.KeyShiftTab})
	// Should not panic
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view after Shift+Tab")
	}
}

func TestAppModelComposeDisablesTabNavigation(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = connectAndAuth(m, &models.User{Username: "akram"}, &models.TokenPair{AccessToken: "tok"})
	m = update(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	// Open compose
	m = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	// Tab should be routed to compose, not column navigation
	m = update(m, tea.KeyMsg{Type: tea.KeyTab})
	// Should not panic
	if !m.IsComposeOpen() {
		t.Error("compose should still be open — Tab should not close it")
	}
}

func TestAppModelEscReturnsToCenter(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = connectAndAuth(m, &models.User{Username: "akram"}, &models.TokenPair{AccessToken: "tok"})
	m = update(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	// Tab to right column
	m = update(m, tea.KeyMsg{Type: tea.KeyTab})
	// Esc should return to center
	m = update(m, tea.KeyMsg{Type: tea.KeyEsc})
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view after Esc")
	}
}

func TestAppModelNFromLeftColumnFocusesCenter(t *testing.T) {
	m := app.NewAppModelWithFactory(nil, nil, &stubFactory{}, "")
	m = connectAndAuth(m, &models.User{Username: "akram"}, &models.TokenPair{AccessToken: "tok"})
	m = update(m, tea.WindowSizeMsg{Width: 120, Height: 40})
	// Tab twice to get to left column (center -> right -> left)
	m = update(m, tea.KeyMsg{Type: tea.KeyTab})
	m = update(m, tea.KeyMsg{Type: tea.KeyTab})
	// Press n — should open compose (which is in center)
	m = update(m, tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'n'}})
	if !m.IsComposeOpen() {
		t.Error("n from left column should open compose")
	}
}
```

**Step 2: Run tests to verify they fail**

Run: `go test ./internal/tui/app/ -run "TestAppModelTab|TestAppModelShiftTab|TestAppModelComposeDisablesTab|TestAppModelEscReturnsToCenter|TestAppModelNFromLeft" -v`
Expected: Some may pass (Tab is already forwarded to compose), others may fail

**Step 3: Implement column focus in app.go**

Add to `AppModel` struct:
```go
// Column focus
focus layout.FocusState
```

Initialize in constructors:
```go
focus: layout.NewFocusState(),
```

Add Tab/Shift+Tab handling in Update(), after the compose/help overlay routing but before global shortcuts:

```go
// Tab/Shift+Tab: column navigation (disabled when compose expanded or text input focused)
if !m.isTextInputFocused() {
    switch msg.Type {
    case tea.KeyTab:
        m.focus.Next()
        return m, nil
    case tea.KeyShiftTab:
        m.focus.Prev()
        return m, nil
    }
}
```

Add Esc handling for non-center columns (return to center):
```go
// In the global shortcuts section, before the 'n' shortcut:
case msg.Type == tea.KeyEsc:
    if m.focus.Active() != layout.FocusCenter {
        m.focus.Reset()
        return m, nil
    }
```

Update `View()` to pass focus state to sidebar and discover:
```go
leftContent := components.RenderSidebar(
    m.user,
    components.View(m.currentView),
    m.focus.Active() == layout.FocusLeft,
    cols.Left,
    m.height,
)

rightContent := components.RenderDiscover(
    m.focus.Active() == layout.FocusRight,
    cols.Right,
    m.height,
)
```

Import the layout package (already imported).

**Step 4: Run all tests**

Run: `go test ./internal/tui/... -v -race`
Expected: All PASS

**Step 5: Commit**

```bash
git add internal/tui/app/app.go internal/tui/app/app_test.go
git commit -m "feat: add column focus with Tab/Shift+Tab navigation"
```

---

### Task 11: Update factory and view adapters for new compose interface

**Files:**
- Modify: `internal/tui/views/factory.go`

**Step 1: Verify factory compiles with new interface**

The `composeAdapter` needs `Expanded()` and `Expand()` methods (may already be added in Task 9).

Verify by running:
```bash
go build ./internal/tui/...
```

If it fails, add the missing adapter methods.

**Step 2: Run all tests**

Run: `go test ./internal/tui/... -v -race`
Expected: All PASS

**Step 3: Commit if changes were needed**

```bash
git add internal/tui/views/factory.go
git commit -m "fix: update factory adapters for new compose interface"
```

---

### Task 12: Final integration — compose inline in timeline view

**Files:**
- Modify: `internal/tui/app/app.go`

The compose model is now always present (not nil/overlay pattern). It renders at the top of the center column in its collapsed state. When the user presses `n`, it expands in place.

**Step 1: Refactor compose from overlay to inline**

In `AppModel` struct, change compose from overlay to persistent field:
```go
// Inline compose bar (always present when authenticated)
compose ComposeViewModel
```

The compose should be created along with timeline in `MsgServerConnected` handler and `handleAuthSuccess`:
```go
m.compose = m.factory.NewCompose(m.client)
```

Change `openCompose()` to just expand the existing compose:
```go
func (m AppModel) openCompose() (AppModel, tea.Cmd) {
    if m.compose != nil {
        m.compose.Expand()
        cmd := m.compose.Init()
        return m, cmd
    }
    return m, nil
}
```

Change `IsComposeOpen()` to check expanded state:
```go
func (m AppModel) IsComposeOpen() bool {
    return m.compose != nil && m.compose.Expanded()
}
```

Update `View()` to always render compose bar at top of center column:
```go
// Center column: compose bar + content
var centerParts []string

// Inline compose bar (always shown when authenticated)
if m.compose != nil {
    centerParts = append(centerParts, m.compose.View())
}

// Main content
if m.help != nil {
    centerParts = append(centerParts, m.help.View())
} else {
    centerParts = append(centerParts, m.viewCurrentContent())
}

centerContent := strings.Join(centerParts, "\n")
```

Update `MsgPostPublished` handler — compose no longer needs to be set to nil:
```go
case MsgPostPublished:
    // Compose auto-collapses via its own Update handler
    var cmds []tea.Cmd
    if m.compose != nil {
        updated, cmd := m.compose.Update(msg)
        if cv, ok := updated.(ComposeViewModel); ok {
            m.compose = cv
        }
        cmds = append(cmds, cmd)
    }
    if m.timeline != nil {
        cmds = append(cmds, m.timeline.FetchLatest())
    }
    return m, tea.Batch(cmds...)
```

Update `updateCompose` — Esc collapses but doesn't nil out:
```go
func (m AppModel) updateCompose(msg tea.Msg) (AppModel, tea.Cmd) {
    var updated ViewModel
    var cmd tea.Cmd
    updated, cmd = m.compose.Update(msg)
    if cv, ok := updated.(ComposeViewModel); ok {
        m.compose = cv
    }
    // No need to check Cancelled — compose auto-collapses
    return m, cmd
}
```

**Step 2: Run all tests**

Run: `go test ./internal/tui/... -v -race`
Expected: All PASS (some app_test.go tests may need adjustments)

**Step 3: Fix any failing tests**

Common fixes needed in app_test.go:
- `TestAppModelPostPublished` — compose is no longer nil after publish, it's collapsed
- `TestAppModelComposeCancelClosesOverlay` — compose is collapsed, not nil
- Tests checking `IsComposeOpen()` need to account for Expanded() check

**Step 4: Commit**

```bash
git add internal/tui/app/app.go internal/tui/app/app_test.go
git commit -m "feat: integrate inline compose bar with always-present rendering"
```

---

### Task 13: Propagate window size to compose bar correctly

**Files:**
- Modify: `internal/tui/app/app.go`

**Step 1: Update propagateWindowSize**

Ensure compose receives center column width (not full width):
```go
// In propagateWindowSize, compose section:
if m.compose != nil {
    var updated ViewModel
    var cmd tea.Cmd
    updated, cmd = m.compose.Update(centerMsg) // center column width
    if cv, ok := updated.(ComposeViewModel); ok {
        m.compose = cv
    }
    cmds = append(cmds, cmd)
}
```

This should already be the case from v1. Verify.

**Step 2: Run tests**

Run: `go test ./internal/tui/... -v -race`
Expected: All PASS

**Step 3: Commit if changes were needed**

```bash
git add internal/tui/app/app.go
git commit -m "fix: ensure compose bar receives center column width"
```

---

### Task 14: Clean up — remove HelpText from views, verify full test suite

**Files:**
- Multiple files (cleanup)

**Step 1: Remove HelpText() from view models that only returned status bar text**

`HelpText()` was used by the status bar. Since status bar is removed, `HelpText()` on views is unused. However, it's part of the `ViewModel` interface, so keep the interface but return `""` from views that no longer need it.

Or keep the interface for potential future use. Leave as-is if tests pass.

**Step 2: Run full test suite with race detector**

Run: `go test ./... -race -v`
Expected: All PASS

**Step 3: Run linter**

Run: `make lint`
Expected: Clean

**Step 4: Build both binaries**

Run: `make build`
Expected: Build succeeds

**Step 5: Commit any cleanup**

```bash
git add -A
git commit -m "chore: cleanup after TUI layout v2 implementation"
```

---

## Summary

| Task | Description | Files |
|------|-------------|-------|
| 1 | Update tagline to "the social terminal" | logo.go, logo_test.go |
| 2 | Bold logo + LogoSplash() + TaglineSplash() | logo.go, logo_test.go |
| 3 | Custom block spinner + splash 2.5s | splash.go, splash_test.go |
| 4 | Column focus state in layout | columns.go, columns_test.go |
| 5 | Remove status bar component | statusbar.go (delete), app.go |
| 6 | Remove shortcuts component | shortcuts.go (delete), app.go |
| 7 | Rewrite sidebar as X-style nav hub | sidebar.go, sidebar_test.go |
| 8 | Create discover component | discover.go (new), discover_test.go (new) |
| 9 | Rewrite compose as inline bar | compose.go, compose_test.go, factory.go, app.go |
| 10 | Column focus + Tab/Shift+Tab in app | app.go, app_test.go |
| 11 | Update factory adapters | factory.go |
| 12 | Inline compose integration in app | app.go, app_test.go |
| 13 | Window size propagation to compose | app.go |
| 14 | Full cleanup and verification | Multiple |
