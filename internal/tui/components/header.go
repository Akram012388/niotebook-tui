package components

import (
	"github.com/charmbracelet/lipgloss"

	"github.com/Akram012388/niotebook-tui/internal/tui/theme"
)

var (
	headerUsernameStyle = lipgloss.NewStyle().
				Foreground(theme.Accent)
	viewNameStyle = lipgloss.NewStyle().
			Foreground(theme.TextMuted)
)

// RenderHeader renders the app header bar with left-aligned app name + username
// and right-aligned view name, spanning the given width.
func RenderHeader(appName, username, viewName string, width int) string {
	left := theme.LogoCompact() + "  " + headerUsernameStyle.Render("@"+username)
	right := viewNameStyle.Render(viewName)

	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(right)
	gap := width - leftWidth - rightWidth
	if gap < 1 {
		gap = 1
	}

	row := left + lipgloss.NewStyle().Width(gap).Render("") + right
	return row
}
