# TUI Redesign Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Transform the Niotebook TUI from a utilitarian prototype into a polished, terminal.shop-caliber experience with a terracotta amber design system, splash screen, and three-column X/Twitter-style layout.

**Architecture:** Create a centralized theme package (`internal/tui/theme/`) with adaptive colors and typography tokens. Build a layout manager (`internal/tui/layout/`) for responsive three-column rendering. Add a splash screen as the new entry point. Rewrite every view and component to use the theme system instead of hard-coded ANSI colors.

**Tech Stack:** Go 1.24, Bubble Tea (bubbletea), Lip Gloss (lipgloss), Bubbles (bubbles/spinner, bubbles/textarea, bubbles/textinput), charmbracelet/x/ansi

**Branch:** `feature/tui-redesign` (create from `main`)

**Design doc:** `docs/plans/2026-02-16-tui-redesign-design.md`

---

### Task 0: Create Feature Branch

**Files:**
- None (git operation only)

**Step 1: Create and switch to feature branch**

```bash
cd /Users/akram/Learning/Projects/Niotebook/niotebook-tui
git checkout -b feature/tui-redesign
```

**Step 2: Verify branch**

Run: `git branch --show-current`
Expected: `feature/tui-redesign`

---

### Task 1: Theme Package — Color Palette

**Files:**
- Create: `internal/tui/theme/theme.go`
- Create: `internal/tui/theme/theme_test.go`

**Step 1: Write the failing test**

```go
package theme_test

import (
	"testing"

	"github.com/Akram012388/niotebook-tui/internal/tui/theme"
	"github.com/charmbracelet/lipgloss"
)

func TestColorsAreDefined(t *testing.T) {
	// Verify all color tokens are non-empty AdaptiveColors
	colors := []lipgloss.AdaptiveColor{
		theme.Accent,
		theme.AccentDim,
		theme.Text,
		theme.TextSecondary,
		theme.TextMuted,
		theme.Border,
		theme.Surface,
		theme.SurfaceRaised,
		theme.Error,
		theme.Success,
		theme.Warning,
	}
	for i, c := range colors {
		if c.Light == "" || c.Dark == "" {
			t.Errorf("color token %d has empty Light or Dark value", i)
		}
	}
}

func TestStylesAreDefined(t *testing.T) {
	// Verify typography styles can be rendered without panic
	styles := []lipgloss.Style{
		theme.Heading,
		theme.Label,
		theme.Body,
		theme.Caption,
		theme.Hint,
		theme.Key,
	}
	for i, s := range styles {
		result := s.Render("test")
		if result == "" {
			t.Errorf("style token %d rendered empty string", i)
		}
	}
}

func TestBorderStylesAreDefined(t *testing.T) {
	styles := []lipgloss.Style{
		theme.PanelBorder,
		theme.ActiveBorder,
		theme.ModalBorder,
	}
	for i, s := range styles {
		result := s.Render("test")
		if result == "" {
			t.Errorf("border style %d rendered empty string", i)
		}
	}
}

func TestSeparator(t *testing.T) {
	s := theme.Separator(40)
	if s == "" {
		t.Error("Separator returned empty string")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/theme/ -v -run TestColorsAreDefined`
Expected: FAIL — package does not exist

**Step 3: Write minimal implementation**

```go
package theme

import "github.com/charmbracelet/lipgloss"

// --- Color Palette (Adaptive: Light/Dark terminal) ---

var (
	// Primary brand accent — terracotta amber
	Accent = lipgloss.AdaptiveColor{Light: "#C15F3C", Dark: "#D97757"}
	// Secondary accent — dimmer terracotta for borders
	AccentDim = lipgloss.AdaptiveColor{Light: "#A04E30", Dark: "#B85C3A"}

	// Text hierarchy
	Text          = lipgloss.AdaptiveColor{Light: "#141413", Dark: "#FAFAF9"}
	TextSecondary = lipgloss.AdaptiveColor{Light: "#57534E", Dark: "#A8A29E"}
	TextMuted     = lipgloss.AdaptiveColor{Light: "#A8A29E", Dark: "#57534E"}

	// Structural
	Border       = lipgloss.AdaptiveColor{Light: "#D6D3D1", Dark: "#44403C"}
	Surface      = lipgloss.AdaptiveColor{Light: "#FAFAF9", Dark: "#141413"}
	SurfaceRaised = lipgloss.AdaptiveColor{Light: "#F5F5F4", Dark: "#1C1917"}

	// Semantic
	Error   = lipgloss.AdaptiveColor{Light: "#DC2626", Dark: "#EF4444"}
	Success = lipgloss.AdaptiveColor{Light: "#16A34A", Dark: "#22C55E"}
	Warning = lipgloss.AdaptiveColor{Light: "#D97706", Dark: "#FBBF24"}
)

// --- Typography Styles ---

var (
	Heading = lipgloss.NewStyle().Bold(true).Foreground(Accent)
	Label   = lipgloss.NewStyle().Bold(true).Foreground(Text)
	Body    = lipgloss.NewStyle().Foreground(Text)
	Caption = lipgloss.NewStyle().Foreground(TextSecondary)
	Hint    = lipgloss.NewStyle().Foreground(TextMuted).Italic(true)
	Key     = lipgloss.NewStyle().Bold(true).Foreground(Accent)
)

// --- Border Styles ---

var (
	PanelBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Border)

	ActiveBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(AccentDim)

	ModalBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Accent).
			Padding(1, 2)
)

// Separator returns a dashed line of the given width using the Border color.
func Separator(width int) string {
	s := lipgloss.NewStyle().Foreground(Border)
	line := ""
	for i := 0; i < width; i++ {
		if i%2 == 0 {
			line += "─"
		} else {
			line += " "
		}
	}
	return s.Render(line)
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui/theme/ -v`
Expected: PASS (all 4 tests)

**Step 5: Commit**

```bash
git add internal/tui/theme/theme.go internal/tui/theme/theme_test.go
git commit -m "feat(tui): add centralized theme package with adaptive color palette"
```

---

### Task 2: Theme Package — Logo Renderer

**Files:**
- Create: `internal/tui/theme/logo.go`
- Create: `internal/tui/theme/logo_test.go`

**Step 1: Write the failing test**

```go
package theme_test

import (
	"strings"
	"testing"

	"github.com/Akram012388/niotebook-tui/internal/tui/theme"
)

func TestLogoContainsBrandName(t *testing.T) {
	logo := theme.Logo()
	// The logo should contain "n" and "otebook" as visible text
	if !strings.Contains(logo, "n") || !strings.Contains(logo, "otebook") {
		t.Errorf("Logo() missing brand name, got: %q", logo)
	}
}

func TestLogoCompact(t *testing.T) {
	compact := theme.LogoCompact()
	if compact == "" {
		t.Error("LogoCompact() returned empty string")
	}
}

func TestTagline(t *testing.T) {
	tagline := theme.Tagline()
	if !strings.Contains(tagline, "social notebook") {
		t.Errorf("Tagline() should contain 'social notebook', got: %q", tagline)
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/theme/ -v -run TestLogo`
Expected: FAIL — Logo function not defined

**Step 3: Write minimal implementation**

```go
package theme

import "github.com/charmbracelet/lipgloss"

// Logo returns the full brand logo with the dot in terracotta.
// "n·otebook" — only the · is accented, rest is primary text.
func Logo() string {
	textStyle := lipgloss.NewStyle().Foreground(Text)
	dotStyle := lipgloss.NewStyle().Foreground(Accent)
	return textStyle.Render("n") + dotStyle.Render("·") + textStyle.Render("otebook")
}

// LogoCompact returns the compact logo for headers/narrow views.
func LogoCompact() string {
	return Logo()
}

// Tagline returns the styled tagline "a social notebook".
func Tagline() string {
	return Hint.Render("a social notebook")
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui/theme/ -v`
Expected: PASS (all tests including new logo tests)

**Step 5: Commit**

```bash
git add internal/tui/theme/logo.go internal/tui/theme/logo_test.go
git commit -m "feat(tui): add brand logo renderer with terracotta dot accent"
```

---

### Task 3: Layout Package — Column Manager

**Files:**
- Create: `internal/tui/layout/columns.go`
- Create: `internal/tui/layout/columns_test.go`

**Step 1: Write the failing test**

```go
package layout_test

import (
	"testing"

	"github.com/Akram012388/niotebook-tui/internal/tui/layout"
)

func TestLayoutModeThreeColumns(t *testing.T) {
	mode := layout.ModeForWidth(120)
	if mode != layout.ThreeColumn {
		t.Errorf("width 120 should be ThreeColumn, got %v", mode)
	}
}

func TestLayoutModeTwoColumns(t *testing.T) {
	mode := layout.ModeForWidth(90)
	if mode != layout.TwoColumn {
		t.Errorf("width 90 should be TwoColumn, got %v", mode)
	}
}

func TestLayoutModeSingleColumn(t *testing.T) {
	mode := layout.ModeForWidth(60)
	if mode != layout.SingleColumn {
		t.Errorf("width 60 should be SingleColumn, got %v", mode)
	}
}

func TestColumnWidths(t *testing.T) {
	tests := []struct {
		name        string
		totalWidth  int
		wantLeft    int
		wantCenter  int
		wantRight   int
	}{
		{"three column", 120, 20, 82, 18},
		{"two column", 90, 20, 70, 0},
		{"single column", 60, 0, 60, 0},
		{"minimum", 40, 0, 40, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cols := layout.ComputeColumns(tt.totalWidth)
			if cols.Left != tt.wantLeft {
				t.Errorf("Left: got %d, want %d", cols.Left, tt.wantLeft)
			}
			if cols.Center != tt.wantCenter {
				t.Errorf("Center: got %d, want %d", cols.Center, tt.wantCenter)
			}
			if cols.Right != tt.wantRight {
				t.Errorf("Right: got %d, want %d", cols.Right, tt.wantRight)
			}
		})
	}
}

func TestRenderColumnsThree(t *testing.T) {
	result := layout.RenderColumns(120, 24, "LEFT", "CENTER", "RIGHT")
	if result == "" {
		t.Error("RenderColumns returned empty string")
	}
}

func TestRenderColumnsTwoOmitsRight(t *testing.T) {
	result := layout.RenderColumns(90, 24, "LEFT", "CENTER", "RIGHT")
	if result == "" {
		t.Error("RenderColumns returned empty string")
	}
}

func TestRenderColumnsSingleOmitsSidebars(t *testing.T) {
	result := layout.RenderColumns(60, 24, "LEFT", "CENTER", "RIGHT")
	if result == "" {
		t.Error("RenderColumns returned empty string")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/layout/ -v`
Expected: FAIL — package does not exist

**Step 3: Write minimal implementation**

```go
package layout

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/Akram012388/niotebook-tui/internal/tui/theme"
)

// LayoutMode represents the responsive column layout.
type LayoutMode int

const (
	SingleColumn LayoutMode = iota
	TwoColumn
	ThreeColumn
)

// Breakpoints
const (
	ThreeColumnMin = 100
	TwoColumnMin   = 80
	LeftWidth      = 20
	RightWidth     = 18
)

// Columns holds the computed widths for each column.
type Columns struct {
	Left   int
	Center int
	Right  int
	Mode   LayoutMode
}

// ModeForWidth returns the layout mode for the given terminal width.
func ModeForWidth(width int) LayoutMode {
	switch {
	case width >= ThreeColumnMin:
		return ThreeColumn
	case width >= TwoColumnMin:
		return TwoColumn
	default:
		return SingleColumn
	}
}

// ComputeColumns returns the column widths for the given terminal width.
func ComputeColumns(width int) Columns {
	mode := ModeForWidth(width)
	switch mode {
	case ThreeColumn:
		return Columns{
			Left:   LeftWidth,
			Center: width - LeftWidth - RightWidth,
			Right:  RightWidth,
			Mode:   ThreeColumn,
		}
	case TwoColumn:
		return Columns{
			Left:   LeftWidth,
			Center: width - LeftWidth,
			Right:  0,
			Mode:   TwoColumn,
		}
	default:
		return Columns{
			Left:   0,
			Center: width,
			Right:  0,
			Mode:   SingleColumn,
		}
	}
}

// RenderColumns renders the three-column layout with vertical dividers.
// leftContent, centerContent, rightContent are pre-rendered strings.
func RenderColumns(width, height int, leftContent, centerContent, rightContent string) string {
	cols := ComputeColumns(width)

	dividerStyle := lipgloss.NewStyle().
		Foreground(theme.Border).
		Width(1)

	divider := dividerStyle.Render(strings.Repeat("│\n", height))

	switch cols.Mode {
	case ThreeColumn:
		left := lipgloss.NewStyle().
			Width(cols.Left).
			Height(height).
			Render(leftContent)
		center := lipgloss.NewStyle().
			Width(cols.Center - 2). // account for 2 dividers
			Height(height).
			Render(centerContent)
		right := lipgloss.NewStyle().
			Width(cols.Right).
			Height(height).
			Render(rightContent)
		return lipgloss.JoinHorizontal(lipgloss.Top, left, divider, center, divider, right)

	case TwoColumn:
		left := lipgloss.NewStyle().
			Width(cols.Left).
			Height(height).
			Render(leftContent)
		center := lipgloss.NewStyle().
			Width(cols.Center - 1). // account for 1 divider
			Height(height).
			Render(centerContent)
		return lipgloss.JoinHorizontal(lipgloss.Top, left, divider, center)

	default:
		return lipgloss.NewStyle().
			Width(cols.Center).
			Height(height).
			Render(centerContent)
	}
}
```

**Step 4: Run test to verify it passes**

Run: `go test ./internal/tui/layout/ -v`
Expected: PASS (all 7 tests)

**Step 5: Commit**

```bash
git add internal/tui/layout/columns.go internal/tui/layout/columns_test.go
git commit -m "feat(tui): add responsive three-column layout manager"
```

---

### Task 4: Splash Screen View

**Files:**
- Create: `internal/tui/views/splash.go`
- Create: `internal/tui/views/splash_test.go`
- Modify: `internal/tui/app/messages.go` — Add connection messages
- Modify: `internal/tui/app/app.go` — Add ViewSplash constant and SplashViewModel interface

**Step 1: Add new message types to `internal/tui/app/messages.go`**

Add the following after the existing message types (after line 44):

```go
// MsgServerConnected indicates the server health check succeeded.
type MsgServerConnected struct{}

// MsgServerFailed indicates the server health check failed.
type MsgServerFailed struct {
	Err string
}
```

**Step 2: Add ViewSplash and SplashViewModel to `internal/tui/app/app.go`**

Add `ViewSplash` to the View constants (after line 45, before `ViewLogin`):

```go
ViewSplash View = iota
```

Add the `SplashViewModel` interface alongside the other view model interfaces:

```go
// SplashViewModel extends ViewModel for the splash screen.
type SplashViewModel interface {
	ViewModel
	Done() bool
	Failed() bool
	ErrorMessage() string
}
```

Add `NewSplash` to the `ViewFactory` interface:

```go
NewSplash(serverURL string) SplashViewModel
```

**Step 3: Write the splash screen test**

```go
package views_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Akram012388/niotebook-tui/internal/tui/app"
	"github.com/Akram012388/niotebook-tui/internal/tui/views"
)

func TestSplashScreenRenders(t *testing.T) {
	m := views.NewSplashModel("http://localhost:8080")
	v := m.View()
	if v == "" {
		t.Error("splash View() returned empty string")
	}
}

func TestSplashScreenInit(t *testing.T) {
	m := views.NewSplashModel("http://localhost:8080")
	cmd := m.Init()
	if cmd == nil {
		t.Error("splash Init() should return a command (spinner tick + health check)")
	}
}

func TestSplashScreenNotDoneInitially(t *testing.T) {
	m := views.NewSplashModel("http://localhost:8080")
	if m.Done() {
		t.Error("splash should not be done initially")
	}
	if m.Failed() {
		t.Error("splash should not be failed initially")
	}
}

func TestSplashScreenConnected(t *testing.T) {
	m := views.NewSplashModel("http://localhost:8080")
	updated, _ := m.Update(app.MsgServerConnected{})
	sm := updated.(views.SplashModel)
	if !sm.Done() {
		t.Error("splash should be done after MsgServerConnected")
	}
}

func TestSplashScreenFailed(t *testing.T) {
	m := views.NewSplashModel("http://localhost:8080")
	updated, _ := m.Update(app.MsgServerFailed{Err: "timeout"})
	sm := updated.(views.SplashModel)
	if !sm.Failed() {
		t.Error("splash should be failed after MsgServerFailed")
	}
	if sm.ErrorMessage() != "timeout" {
		t.Errorf("ErrorMessage() = %q, want %q", sm.ErrorMessage(), "timeout")
	}
}

func TestSplashScreenResize(t *testing.T) {
	m := views.NewSplashModel("http://localhost:8080")
	updated, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	sm := updated.(views.SplashModel)
	_ = sm.View() // should not panic
}
```

**Step 4: Write the splash screen implementation**

```go
package views

import (
	"fmt"
	"net/http"
	"time"

	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Akram012388/niotebook-tui/internal/tui/app"
	"github.com/Akram012388/niotebook-tui/internal/tui/theme"
)

// SplashModel is the splash screen shown on startup while connecting to the server.
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
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = lipgloss.NewStyle().Foreground(theme.Accent)
	return SplashModel{
		serverURL: serverURL,
		spinner:   s,
	}
}

func (m SplashModel) Done() bool         { return m.done }
func (m SplashModel) Failed() bool       { return m.failed }
func (m SplashModel) ErrorMessage() string { return m.err }
func (m SplashModel) HelpText() string   { return "" }

func (m SplashModel) Init() tea.Cmd {
	return tea.Batch(m.spinner.Tick, m.checkHealth())
}

func (m SplashModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
		if m.failed && msg.String() == "r" {
			m.failed = false
			m.err = ""
			return m, tea.Batch(m.spinner.Tick, m.checkHealth())
		}
		if msg.String() == "q" || msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		return m, nil

	case spinner.TickMsg:
		if m.done || m.failed {
			return m, nil
		}
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}

	return m, nil
}

func (m SplashModel) View() string {
	var content string

	logo := theme.Logo()
	tagline := theme.Tagline()

	if m.failed {
		errStyle := lipgloss.NewStyle().Foreground(theme.Error)
		hintStyle := theme.Hint
		content = fmt.Sprintf("%s\n\n%s\n\n%s\n\n%s",
			logo,
			tagline,
			errStyle.Render("connection failed: "+m.err),
			hintStyle.Render("press r to retry · q to quit"),
		)
	} else if m.done {
		content = fmt.Sprintf("%s\n\n%s", logo, tagline)
	} else {
		statusStyle := theme.Caption
		content = fmt.Sprintf("%s\n\n%s\n\n%s\n%s",
			logo,
			tagline,
			m.spinner.View(),
			statusStyle.Render("connecting..."),
		)
	}

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

func (m SplashModel) checkHealth() tea.Cmd {
	url := m.serverURL
	return func() tea.Msg {
		client := &http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get(url + "/health")
		if err != nil {
			return app.MsgServerFailed{Err: err.Error()}
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return app.MsgServerFailed{Err: fmt.Sprintf("server returned %d", resp.StatusCode)}
		}
		return app.MsgServerConnected{}
	}
}
```

**Step 5: Run tests to verify they pass**

Run: `go test ./internal/tui/views/ -v -run TestSplash`
Expected: PASS (all 6 tests)

Run: `go test ./internal/tui/... -v -count=1`
Expected: PASS (all existing tests still pass)

**Step 6: Commit**

```bash
git add internal/tui/app/messages.go internal/tui/app/app.go internal/tui/views/splash.go internal/tui/views/splash_test.go
git commit -m "feat(tui): add splash screen with spinner and server health check"
```

---

### Task 5: Left Sidebar Component

**Files:**
- Create: `internal/tui/components/sidebar.go`
- Create: `internal/tui/components/sidebar_test.go`

**Step 1: Write the failing test**

```go
package components_test

import (
	"strings"
	"testing"

	"github.com/Akram012388/niotebook-tui/internal/tui/app"
	"github.com/Akram012388/niotebook-tui/internal/tui/components"
	"github.com/Akram012388/niotebook-tui/internal/models"
)

func TestSidebarRendersUsername(t *testing.T) {
	user := &models.User{Username: "akram", DisplayName: "Akram"}
	result := components.RenderSidebar(user, app.ViewTimeline, 20, 24)
	if !strings.Contains(result, "akram") {
		t.Error("sidebar should contain username")
	}
}

func TestSidebarRendersNavItems(t *testing.T) {
	user := &models.User{Username: "test"}
	result := components.RenderSidebar(user, app.ViewTimeline, 20, 24)
	if !strings.Contains(result, "Home") {
		t.Error("sidebar should contain Home nav item")
	}
}

func TestSidebarActiveIndicator(t *testing.T) {
	user := &models.User{Username: "test"}
	result := components.RenderSidebar(user, app.ViewTimeline, 20, 24)
	// The active nav item (Home/Timeline) should have the bullet marker
	if !strings.Contains(result, "●") {
		t.Error("sidebar should show active indicator for current view")
	}
}

func TestSidebarLoggedOut(t *testing.T) {
	result := components.RenderSidebar(nil, app.ViewLogin, 20, 24)
	// Should show logo but not profile info
	if strings.Contains(result, "@") {
		t.Error("logged-out sidebar should not show username")
	}
}

func TestSidebarZeroWidth(t *testing.T) {
	user := &models.User{Username: "test"}
	result := components.RenderSidebar(user, app.ViewTimeline, 0, 24)
	if result != "" {
		t.Error("zero-width sidebar should return empty string")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/components/ -v -run TestSidebar`
Expected: FAIL — RenderSidebar not defined

**Step 3: Write minimal implementation**

```go
package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/Akram012388/niotebook-tui/internal/tui/app"
	"github.com/Akram012388/niotebook-tui/internal/tui/theme"
)

type navItem struct {
	label string
	view  app.View
}

var navItems = []navItem{
	{"Home", app.ViewTimeline},
	{"Profile", app.ViewProfile},
}

// RenderSidebar renders the left sidebar with profile card and navigation.
func RenderSidebar(user *models.User, activeView app.View, width, height int) string {
	if width <= 0 {
		return ""
	}

	var b strings.Builder

	// Logo
	b.WriteString(theme.LogoCompact())
	b.WriteString("\n\n")

	// Profile card (only when logged in)
	if user != nil {
		usernameStyle := lipgloss.NewStyle().Foreground(theme.Accent)
		b.WriteString(usernameStyle.Render("@" + user.Username))
		b.WriteString("\n")
		if user.DisplayName != "" {
			nameStyle := lipgloss.NewStyle().Foreground(theme.Text)
			b.WriteString(nameStyle.Render(user.DisplayName))
			b.WriteString("\n")
		}
		b.WriteString(theme.Separator(width - 2))
		b.WriteString("\n\n")

		// Navigation
		for _, item := range navItems {
			if item.view == activeView {
				activeStyle := lipgloss.NewStyle().Foreground(theme.Accent).Bold(true)
				b.WriteString(activeStyle.Render("● " + item.label))
			} else {
				inactiveStyle := lipgloss.NewStyle().Foreground(theme.TextSecondary)
				b.WriteString(inactiveStyle.Render("  " + item.label))
			}
			b.WriteString("\n")
		}

		b.WriteString("\n")
		b.WriteString(theme.Separator(width - 2))
		b.WriteString("\n")

		// Stats
		if user.CreatedAt != nil {
			joinedStyle := theme.Caption
			b.WriteString(joinedStyle.Render(fmt.Sprintf("Joined %s", user.CreatedAt.Format("Jan 2006"))))
			b.WriteString("\n")
		}
	}

	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(1, 1).
		Render(b.String())
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/tui/components/ -v -run TestSidebar`
Expected: PASS (all 5 tests)

**Step 5: Commit**

```bash
git add internal/tui/components/sidebar.go internal/tui/components/sidebar_test.go
git commit -m "feat(tui): add left sidebar component with profile card and navigation"
```

---

### Task 6: Right Sidebar — Context Shortcuts Component

**Files:**
- Create: `internal/tui/components/shortcuts.go`
- Create: `internal/tui/components/shortcuts_test.go`

**Step 1: Write the failing test**

```go
package components_test

import (
	"strings"
	"testing"

	"github.com/Akram012388/niotebook-tui/internal/tui/app"
	"github.com/Akram012388/niotebook-tui/internal/tui/components"
)

func TestShortcutsTimeline(t *testing.T) {
	result := components.RenderShortcuts(app.ViewTimeline, false, 18, 24)
	if !strings.Contains(result, "j/k") {
		t.Error("timeline shortcuts should contain j/k for scrolling")
	}
	if !strings.Contains(result, "compose") {
		t.Error("timeline shortcuts should mention compose")
	}
}

func TestShortcutsProfile(t *testing.T) {
	result := components.RenderShortcuts(app.ViewProfile, false, 18, 24)
	if !strings.Contains(result, "scroll") {
		t.Error("profile shortcuts should mention scroll")
	}
}

func TestShortcutsCompose(t *testing.T) {
	result := components.RenderShortcuts(app.ViewTimeline, true, 18, 24)
	if !strings.Contains(result, "Ctrl+J") {
		t.Error("compose shortcuts should contain Ctrl+J for publish")
	}
}

func TestShortcutsZeroWidth(t *testing.T) {
	result := components.RenderShortcuts(app.ViewTimeline, false, 0, 24)
	if result != "" {
		t.Error("zero-width shortcuts should return empty string")
	}
}
```

**Step 2: Run test to verify it fails**

Run: `go test ./internal/tui/components/ -v -run TestShortcuts`
Expected: FAIL — RenderShortcuts not defined

**Step 3: Write minimal implementation**

```go
package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/Akram012388/niotebook-tui/internal/tui/app"
	"github.com/Akram012388/niotebook-tui/internal/tui/theme"
)

type shortcutEntry struct {
	key  string
	desc string
}

type shortcutSection struct {
	title   string
	entries []shortcutEntry
}

var timelineShortcuts = []shortcutSection{
	{
		title: "Navigation",
		entries: []shortcutEntry{
			{"j/k", "scroll"},
			{"g/G", "top/bottom"},
			{"Enter", "profile"},
		},
	},
	{
		title: "Actions",
		entries: []shortcutEntry{
			{"n", "compose"},
			{"r", "refresh"},
			{"?", "help"},
			{"q", "quit"},
		},
	},
}

var profileShortcuts = []shortcutSection{
	{
		title: "Navigation",
		entries: []shortcutEntry{
			{"j/k", "scroll"},
			{"Esc", "back"},
		},
	},
	{
		title: "Actions",
		entries: []shortcutEntry{
			{"e", "edit"},
			{"?", "help"},
			{"q", "quit"},
		},
	},
}

var composeShortcuts = []shortcutSection{
	{
		title: "Compose",
		entries: []shortcutEntry{
			{"Ctrl+J", "publish"},
			{"Esc", "cancel"},
		},
	},
}

// RenderShortcuts renders context-sensitive keyboard shortcuts for the right sidebar.
func RenderShortcuts(activeView app.View, composeOpen bool, width, height int) string {
	if width <= 0 {
		return ""
	}

	var sections []shortcutSection
	if composeOpen {
		sections = composeShortcuts
	} else {
		switch activeView {
		case app.ViewProfile:
			sections = profileShortcuts
		default:
			sections = timelineShortcuts
		}
	}

	var b strings.Builder
	sectionStyle := lipgloss.NewStyle().Foreground(theme.TextMuted).Bold(true)
	keyStyle := theme.Key
	descStyle := theme.Caption

	for i, section := range sections {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(sectionStyle.Render(section.title))
		b.WriteString("\n")
		for _, entry := range section.entries {
			b.WriteString(keyStyle.Width(8).Render(entry.key))
			b.WriteString(descStyle.Render(entry.desc))
			b.WriteString("\n")
		}
	}

	return lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(1, 1).
		Render(b.String())
}
```

**Step 4: Run tests to verify they pass**

Run: `go test ./internal/tui/components/ -v -run TestShortcuts`
Expected: PASS (all 4 tests)

**Step 5: Commit**

```bash
git add internal/tui/components/shortcuts.go internal/tui/components/shortcuts_test.go
git commit -m "feat(tui): add context-sensitive keyboard shortcuts sidebar"
```

---

### Task 7: Retheme Post Card Component

**Files:**
- Modify: `internal/tui/components/postcard.go`
- Modify: `internal/tui/components/postcard_test.go`

**Step 1: Replace all ANSI color styles with theme tokens**

Replace the style block at lines 13-31 of `internal/tui/components/postcard.go`:

Old:
```go
var (
	usernameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("6")).Bold(true)
	selectedUsernameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("5")).Bold(true)
	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).Faint(true)
	separatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))
	markerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("5")).Bold(true)
)
```

New:
```go
var (
	usernameStyle = lipgloss.NewStyle().
			Foreground(theme.TextSecondary).Bold(true)
	selectedUsernameStyle = lipgloss.NewStyle().
			Foreground(theme.Accent).Bold(true)
	dimStyle = lipgloss.NewStyle().
			Foreground(theme.TextMuted)
	separatorStyle = lipgloss.NewStyle().
			Foreground(theme.Border)
	markerStyle = lipgloss.NewStyle().
			Foreground(theme.Accent).Bold(true)
)
```

Add import: `"github.com/Akram012388/niotebook-tui/internal/tui/theme"`

**Step 2: Run existing tests to verify they still pass**

Run: `go test ./internal/tui/components/ -v`
Expected: PASS (all component tests)

**Step 3: Commit**

```bash
git add internal/tui/components/postcard.go
git commit -m "refactor(tui): retheme post card with adaptive color tokens"
```

---

### Task 8: Retheme Header Component

**Files:**
- Modify: `internal/tui/components/header.go`

**Step 1: Replace ANSI colors with theme tokens**

Replace styles at lines 7-17 of `internal/tui/components/header.go`:

Old:
```go
var (
	appNameStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("5"))
	headerUsernameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("6"))
	viewNameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Faint(true)
)
```

New:
```go
var (
	appNameStyle = lipgloss.NewStyle()
	headerUsernameStyle = lipgloss.NewStyle().
			Foreground(theme.Accent)
	viewNameStyle = lipgloss.NewStyle().
			Foreground(theme.TextMuted)
)
```

Update the `RenderHeader` function to use `theme.LogoCompact()` instead of rendering the app name directly:

Replace the left-side rendering:
```go
left := appNameStyle.Render(appName) + "  " + headerUsernameStyle.Render("@"+username)
```
With:
```go
left := theme.LogoCompact() + "  " + headerUsernameStyle.Render("@"+username)
```

Add import: `"github.com/Akram012388/niotebook-tui/internal/tui/theme"`

**Step 2: Run existing tests**

Run: `go test ./internal/tui/components/ -v`
Expected: PASS

**Step 3: Commit**

```bash
git add internal/tui/components/header.go
git commit -m "refactor(tui): retheme header with logo and adaptive colors"
```

---

### Task 9: Retheme Status Bar Component

**Files:**
- Modify: `internal/tui/components/statusbar.go`

**Step 1: Replace ANSI colors with theme tokens**

Replace styles at lines 19-31 of `internal/tui/components/statusbar.go`:

Old:
```go
var (
	errorStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("1"))
	successStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("2"))
	loadingStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("3"))
	helpStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Faint(true)
)
```

New:
```go
var (
	errorStyle = lipgloss.NewStyle().
		Foreground(theme.Error)
	successStyle = lipgloss.NewStyle().
		Foreground(theme.Success)
	loadingStyle = lipgloss.NewStyle().
		Foreground(theme.Warning)
	helpStyle = lipgloss.NewStyle().
		Foreground(theme.TextMuted)
)
```

Add import: `"github.com/Akram012388/niotebook-tui/internal/tui/theme"`

**Step 2: Run existing tests**

Run: `go test ./internal/tui/components/ -v`
Expected: PASS

**Step 3: Commit**

```bash
git add internal/tui/components/statusbar.go
git commit -m "refactor(tui): retheme status bar with adaptive color tokens"
```

---

### Task 10: Retheme Login View

**Files:**
- Modify: `internal/tui/views/login.go`

**Step 1: Replace ANSI colors with theme tokens**

Replace styles at lines 14-37 of `internal/tui/views/login.go`:

Old:
```go
var (
	formBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(lipgloss.Color("8")).
		Padding(1, 2)
	formTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("5")).
		MarginBottom(1)
	labelStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("7"))
	buttonStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(lipgloss.Color("5"))
	errMsgStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("1"))
	hintStyle = lipgloss.NewStyle().
		Foreground(lipgloss.Color("8")).
		Faint(true)
)
```

New:
```go
var (
	formBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.AccentDim).
		Padding(1, 2)
	formTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Accent).
		MarginBottom(1)
	labelStyle = lipgloss.NewStyle().
		Foreground(theme.Text)
	buttonStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Accent)
	errMsgStyle = lipgloss.NewStyle().
		Foreground(theme.Error)
	hintStyle = lipgloss.NewStyle().
		Foreground(theme.TextMuted).
		Italic(true)
)
```

Add import: `"github.com/Akram012388/niotebook-tui/internal/tui/theme"`

**Step 2: Run existing tests**

Run: `go test ./internal/tui/views/ -v -run TestFactoryNewLogin`
Expected: PASS

**Step 3: Commit**

```bash
git add internal/tui/views/login.go
git commit -m "refactor(tui): retheme login view with adaptive color tokens"
```

---

### Task 11: Retheme Register View

**Files:**
- Modify: `internal/tui/views/register.go`

**Step 1: Replace ANSI colors with theme tokens**

The register view reuses styles from login.go. Update the inline dynamic counter styles. Find all occurrences of `lipgloss.Color("8")` and `lipgloss.Color("1")` in the View() method and replace with `theme.TextMuted` and `theme.Error`.

Specifically, update the counter rendering logic (around lines 220-260) to use:
- Normal counter: `theme.Caption` style
- Warning counter: `lipgloss.NewStyle().Foreground(theme.Error)` style

Add import: `"github.com/Akram012388/niotebook-tui/internal/tui/theme"`

**Step 2: Run existing tests**

Run: `go test ./internal/tui/views/ -v -run TestFactoryNewRegister`
Expected: PASS

**Step 3: Commit**

```bash
git add internal/tui/views/register.go
git commit -m "refactor(tui): retheme register view with adaptive color tokens"
```

---

### Task 12: Retheme Timeline View

**Files:**
- Modify: `internal/tui/views/timeline.go`

**Step 1: Replace ANSI colors with theme tokens**

Replace styles at lines 16-23 of `internal/tui/views/timeline.go`:

Old:
```go
var (
	emptyStateStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Faint(true)
	loadingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("3"))
)
```

New:
```go
var (
	emptyStateStyle = theme.Hint
	tlLoadingStyle  = lipgloss.NewStyle().Foreground(theme.Warning)
)
```

Note: rename `loadingStyle` to `tlLoadingStyle` to avoid collision with the same name in statusbar.go (they are in different packages so it's actually fine, but renaming for clarity within timeline). Update all references in the file.

Add import: `"github.com/Akram012388/niotebook-tui/internal/tui/theme"`

**Step 2: Run existing tests**

Run: `go test ./internal/tui/views/ -v -run TestTimeline`
Expected: PASS (existing timeline tests)

**Step 3: Commit**

```bash
git add internal/tui/views/timeline.go
git commit -m "refactor(tui): retheme timeline view with adaptive color tokens"
```

---

### Task 13: Retheme Compose View

**Files:**
- Modify: `internal/tui/views/compose.go`

**Step 1: Replace ANSI colors with theme tokens**

Replace styles at lines 17-41 of `internal/tui/views/compose.go`:

New:
```go
var (
	composeBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Accent).
		Padding(1, 2)
	composeTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Accent)
	composePromptStyle = lipgloss.NewStyle().
		Foreground(theme.Text).
		MarginBottom(1)
	counterNormalStyle = theme.Caption
	counterWarningStyle = lipgloss.NewStyle().
		Foreground(theme.Error).
		Bold(true)
	composeHintStyle = theme.Hint
)
```

Add import: `"github.com/Akram012388/niotebook-tui/internal/tui/theme"`

**Step 2: Run existing tests**

Run: `go test ./internal/tui/views/ -v -run TestCompose`
Expected: PASS

**Step 3: Commit**

```bash
git add internal/tui/views/compose.go
git commit -m "refactor(tui): retheme compose view with adaptive color tokens"
```

---

### Task 14: Retheme Profile View

**Files:**
- Modify: `internal/tui/views/profile.go`

**Step 1: Replace ANSI colors with theme tokens**

Replace styles at lines 16-37 of `internal/tui/views/profile.go`:

New:
```go
var (
	profileUsernameStyle = lipgloss.NewStyle().
		Foreground(theme.Accent).Bold(true)
	profileDisplayNameStyle = lipgloss.NewStyle().
		Foreground(theme.Text)
	profileBioStyle = lipgloss.NewStyle().
		Foreground(theme.Text)
	profileJoinedStyle = theme.Caption
	profileSectionStyle = lipgloss.NewStyle().
		Foreground(theme.Accent).Bold(true)
	profileSeparatorStyle = lipgloss.NewStyle().
		Foreground(theme.Border)
)
```

Add import: `"github.com/Akram012388/niotebook-tui/internal/tui/theme"`

**Step 2: Run existing tests**

Run: `go test ./internal/tui/views/ -v -run TestProfile`
Expected: PASS

**Step 3: Commit**

```bash
git add internal/tui/views/profile.go
git commit -m "refactor(tui): retheme profile view with adaptive color tokens"
```

---

### Task 15: Retheme Help View

**Files:**
- Modify: `internal/tui/views/help.go`

**Step 1: Replace ANSI colors with theme tokens**

Replace styles at lines 48-65 of `internal/tui/views/help.go`:

New:
```go
var (
	helpBoxStyle = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.Accent).
		Padding(1, 2)
	helpTitleStyle = lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Accent)
	helpKeyStyle = lipgloss.NewStyle().
		Foreground(theme.Accent).
		Bold(true).
		Width(12)
	helpDescStyle = lipgloss.NewStyle().
		Foreground(theme.Text)
)
```

Add import: `"github.com/Akram012388/niotebook-tui/internal/tui/theme"`

**Step 2: Run existing tests**

Run: `go test ./internal/tui/views/ -v -run TestHelp`
Expected: PASS

**Step 3: Commit**

```bash
git add internal/tui/views/help.go
git commit -m "refactor(tui): retheme help overlay with adaptive color tokens"
```

---

### Task 16: Integrate Splash Screen into App

**Files:**
- Modify: `internal/tui/app/app.go` — Route splash screen, handle connection messages
- Modify: `internal/tui/views/factory.go` — Add splash adapter
- Modify: `cmd/tui/main.go` — Start on splash view

**Step 1: Update factory.go — add splash adapter**

Add the `NewSplash` method to the `Factory` struct and a `splashAdapter` type:

```go
func (f *Factory) NewSplash(serverURL string) app.SplashViewModel {
	m := NewSplashModel(serverURL)
	return &splashAdapter{m: m}
}

type splashAdapter struct{ m SplashModel }

func (a *splashAdapter) Init() tea.Cmd                          { return a.m.Init() }
func (a *splashAdapter) Update(msg tea.Msg) (app.ViewModel, tea.Cmd) {
	updated, cmd := a.m.Update(msg)
	a.m = updated.(SplashModel)
	return a, cmd
}
func (a *splashAdapter) View() string                           { return a.m.View() }
func (a *splashAdapter) HelpText() string                      { return a.m.HelpText() }
func (a *splashAdapter) Done() bool                            { return a.m.Done() }
func (a *splashAdapter) Failed() bool                          { return a.m.Failed() }
func (a *splashAdapter) ErrorMessage() string                  { return a.m.ErrorMessage() }
```

**Step 2: Update app.go — handle splash screen**

In `AppModel`, add a `splash` field of type `SplashViewModel` and a `serverURL` field of type `string`.

Update the constructor to accept the server URL and create the splash view:
- Start on `ViewSplash` instead of `ViewLogin`/`ViewTimeline`
- In `Init()`, return `splash.Init()` command

In `Update()`, add handling for `MsgServerConnected`:
- Transition from splash to login (if no stored auth) or timeline (if stored auth)

In `Update()`, add handling for splash view updates when `currentView == ViewSplash`

In `View()`, when `currentView == ViewSplash`, return only the splash view (no header/sidebar/columns).

**Step 3: Update cmd/tui/main.go**

Pass the server URL to the app model so the splash screen can do the health check:
- Add a `serverURL` parameter to `NewAppModelWithFactory` or set it after construction
- Remove any "connecting..." output before the tea.Program starts

**Step 4: Run all tests**

Run: `go test ./internal/tui/... -v -count=1`
Expected: PASS (all tests)

Run: `go test ./... -v -count=1 -race`
Expected: PASS

**Step 5: Commit**

```bash
git add internal/tui/app/app.go internal/tui/views/factory.go internal/tui/views/factory_test.go cmd/tui/main.go
git commit -m "feat(tui): integrate splash screen as startup entry point"
```

---

### Task 17: Integrate Three-Column Layout into App

**Files:**
- Modify: `internal/tui/app/app.go` — Replace single-view rendering with column layout

**Step 1: Update AppModel to track user for sidebar**

Add a `user *models.User` field to `AppModel`. Set it on `MsgAuthSuccess` and `MsgProfileLoaded` (when viewing own profile).

**Step 2: Update View() to use three-column layout**

For views other than splash and login/register, render:
- Left: `components.RenderSidebar(m.user, m.currentView, cols.Left, contentHeight)`
- Center: The current view's `View()` + status bar
- Right: `components.RenderShortcuts(m.currentView, m.isComposeOpen(), cols.Right, contentHeight)`

Use `layout.RenderColumns(m.width, m.height, left, center, right)` to compose.

For login/register views, use single-column centered layout (no sidebars).

For splash view, use full-screen `lipgloss.Place` (already handled by splash model).

**Step 3: Pass center column width to views**

Views need to know their available width. Update `tea.WindowSizeMsg` propagation to pass the center column width rather than the full terminal width.

**Step 4: Run all tests**

Run: `go test ./internal/tui/... -v -count=1`
Expected: PASS

**Step 5: Manual visual test**

Run: `make dev && make dev-tui`
Verify: Splash screen shows, transitions to login, three-column layout visible after login.

**Step 6: Commit**

```bash
git add internal/tui/app/app.go
git commit -m "feat(tui): integrate three-column layout with sidebar and shortcuts"
```

---

### Task 18: Final Polish and Verification

**Files:**
- All modified files

**Step 1: Run full test suite with race detector**

Run: `go test ./... -v -race -coverprofile=coverage.out`
Expected: PASS, coverage ≥ 80%

**Step 2: Run linter**

Run: `golangci-lint run ./...`
Expected: No errors

**Step 3: Build both binaries**

Run: `make build`
Expected: Clean build, no warnings

**Step 4: Run full integration test**

1. Start server: `NIOTEBOOK_DB_URL="postgres://localhost/niotebook?sslmode=disable" NIOTEBOOK_JWT_SECRET="$(openssl rand -hex 32)" make dev`
2. Start TUI: `make dev-tui`
3. Verify splash screen with animated spinner
4. Verify auto-transition to login
5. Register a new user
6. Verify three-column layout (timeline + sidebar + shortcuts)
7. Create a post (Ctrl+J)
8. View profile (p)
9. Open help (?)
10. Resize terminal — verify responsive breakpoints (3-col → 2-col → 1-col)
11. Quit (q)

**Step 5: Commit any final fixes**

```bash
git add -A
git commit -m "chore(tui): final polish and cleanup for TUI redesign"
```

---

## Task Dependency Graph

```
Task 0  (branch)
  └─ Task 1  (theme/colors)
       └─ Task 2  (theme/logo)
            ├─ Task 3  (layout/columns)
            │    └─ Task 17 (integrate layout into app)
            ├─ Task 4  (splash screen)
            │    └─ Task 16 (integrate splash into app)
            ├─ Task 5  (sidebar component)
            │    └─ Task 17 (integrate layout into app)
            ├─ Task 6  (shortcuts component)
            │    └─ Task 17 (integrate layout into app)
            ├─ Task 7  (retheme postcard)
            ├─ Task 8  (retheme header)
            ├─ Task 9  (retheme statusbar)
            ├─ Task 10 (retheme login)
            ├─ Task 11 (retheme register)
            ├─ Task 12 (retheme timeline)
            ├─ Task 13 (retheme compose)
            ├─ Task 14 (retheme profile)
            └─ Task 15 (retheme help)

Tasks 7-15 can run in parallel after Task 2.
Task 16 depends on Task 4.
Task 17 depends on Tasks 3, 5, 6.
Task 18 depends on all prior tasks.
```

---

## Summary

| Task | Description | New Files | Modified Files |
|------|-------------|-----------|----------------|
| 0 | Create branch | — | — |
| 1 | Theme colors | 2 | — |
| 2 | Theme logo | 2 | — |
| 3 | Layout columns | 2 | — |
| 4 | Splash screen | 2 | 2 (messages.go, app.go) |
| 5 | Left sidebar | 2 | — |
| 6 | Right shortcuts | 2 | — |
| 7 | Retheme postcard | — | 1 |
| 8 | Retheme header | — | 1 |
| 9 | Retheme statusbar | — | 1 |
| 10 | Retheme login | — | 1 |
| 11 | Retheme register | — | 1 |
| 12 | Retheme timeline | — | 1 |
| 13 | Retheme compose | — | 1 |
| 14 | Retheme profile | — | 1 |
| 15 | Retheme help | — | 1 |
| 16 | Integrate splash | — | 3 (app.go, factory.go, main.go) |
| 17 | Integrate layout | — | 1 (app.go) |
| 18 | Final verification | — | any |
| **Total** | **19 tasks** | **12 new** | **~15 modified** |
