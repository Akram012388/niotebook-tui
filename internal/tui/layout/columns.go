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

// Layout breakpoint and fixed-width constants.
const (
	// ThreeColumnMin is the minimum terminal width for three-column layout.
	ThreeColumnMin = 100
	// TwoColumnMin is the minimum terminal width for two-column layout.
	TwoColumnMin = 80
	// LeftWidth is the fixed width of the left sidebar.
	LeftWidth = 20
	// RightWidth is the fixed width of the right sidebar.
	RightWidth = 18
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

// ComputeColumns returns the column widths for the given terminal width. The
// left and right sidebars have fixed widths; the center column receives all
// remaining space. Divider characters (one per adjacent column boundary) are
// subtracted from the center width.
func ComputeColumns(width int) Columns {
	mode := ModeForWidth(width)
	switch mode {
	case ThreeColumn:
		// Two dividers: left|center|right
		center := width - LeftWidth - RightWidth - 2
		if center < 1 {
			center = 1
		}
		return Columns{Left: LeftWidth, Center: center, Right: RightWidth, Mode: ThreeColumn}
	case TwoColumn:
		// One divider: left|center
		center := width - LeftWidth - 1
		if center < 1 {
			center = 1
		}
		return Columns{Left: LeftWidth, Center: center, Right: 0, Mode: TwoColumn}
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

// RenderColumns renders a three-column layout for the given terminal
// dimensions. Each content string is constrained to its column width and the
// specified height. Vertical dividers in the theme's Border color separate
// adjacent columns.
func RenderColumns(width, height int, leftContent, centerContent, rightContent string) string {
	cols := ComputeColumns(width)

	colStyle := func(w int) lipgloss.Style {
		return lipgloss.NewStyle().Width(w).Height(height)
	}

	switch cols.Mode {
	case ThreeColumn:
		left := colStyle(cols.Left).Render(leftContent)
		center := colStyle(cols.Center).Render(centerContent)
		right := colStyle(cols.Right).Render(rightContent)
		div := verticalDivider(height)
		return lipgloss.JoinHorizontal(lipgloss.Top, left, div, center, div, right)

	case TwoColumn:
		left := colStyle(cols.Left).Render(leftContent)
		center := colStyle(cols.Center).Render(centerContent)
		div := verticalDivider(height)
		return lipgloss.JoinHorizontal(lipgloss.Top, left, div, center)

	default:
		return colStyle(cols.Center).Render(centerContent)
	}
}
