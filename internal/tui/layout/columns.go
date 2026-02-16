// Package layout provides a responsive three-column layout manager for the
// Niotebook TUI. It adapts between single, two, and three column modes based
// on terminal width, matching the X/Twitter-style sidebar-content-sidebar
// pattern.
package layout

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/Akram012388/niotebook-tui/internal/tui/theme"
)

// LayoutMode describes how many columns the layout uses.
type LayoutMode int

const (
	// SingleColumn renders only the center content area.
	SingleColumn LayoutMode = iota
	// TwoColumn renders a left sidebar and center content area.
	TwoColumn
	// ThreeColumn renders left sidebar, center content, and right sidebar.
	ThreeColumn
)

// Layout breakpoint and proportion constants.
const (
	// ThreeColumnMin is the minimum terminal width for three-column layout.
	ThreeColumnMin = 100
	// TwoColumnMin is the minimum terminal width for two-column layout.
	TwoColumnMin = 80
	// SidebarPercent is the percentage of terminal width each sidebar occupies
	// in three-column mode (20/60/20 split).
	SidebarPercent = 20
)

// Columns holds the computed widths for each column and the active layout
// mode.
type Columns struct {
	Left   int
	Center int
	Right  int
	Mode   LayoutMode
}

// ModeForWidth returns the appropriate layout mode for the given terminal
// width.
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

// ComputeColumns returns the column widths for the given terminal width using
// a 20/60/20 percentage split. Both sidebars are equal width. Divider
// characters (one per adjacent column boundary) are subtracted from the center.
func ComputeColumns(width int) Columns {
	mode := ModeForWidth(width)
	switch mode {
	case ThreeColumn:
		sidebar := width * SidebarPercent / 100
		// Two dividers: left|center|right
		center := width - 2*sidebar - 2
		if center < 1 {
			center = 1
		}
		return Columns{Left: sidebar, Center: center, Right: sidebar, Mode: ThreeColumn}
	case TwoColumn:
		sidebar := width * SidebarPercent / 100
		// One divider: left|center
		center := width - sidebar - 1
		if center < 1 {
			center = 1
		}
		return Columns{Left: sidebar, Center: center, Right: 0, Mode: TwoColumn}
	default:
		return Columns{Left: 0, Center: width, Right: 0, Mode: SingleColumn}
	}
}

// dividerStyle returns a lipgloss style for the vertical divider character,
// colored with the theme's Border token.
var dividerStyle = lipgloss.NewStyle().Foreground(theme.Border)

// verticalDivider returns a single-character-wide column of "│" characters
// repeated for the given height, rendered in the Border color.
func verticalDivider(height int) string {
	if height <= 0 {
		return ""
	}
	lines := make([]string, height)
	ch := dividerStyle.Render("│")
	for i := range lines {
		lines[i] = ch
	}
	return strings.Join(lines, "\n")
}

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

// RenderColumns renders a three-column layout for the given terminal
// dimensions. Each column gets a header bar: thick amber for the focused
// column, thin dim for inactive. Vertical dividers separate columns.
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

// Prev moves focus to the previous column (Right→Center→Left→Right).
func (f *FocusState) Prev() {
	f.active = (f.active + 2) % 3
}

// Reset returns focus to the center column.
func (f *FocusState) Reset() {
	f.active = FocusCenter
}
