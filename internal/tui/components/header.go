package components

import (
	"github.com/charmbracelet/lipgloss"
)

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

// RenderHeader renders the app header bar with left-aligned app name + username
// and right-aligned view name, spanning the given width.
func RenderHeader(appName, username, viewName string, width int) string {
	left := appNameStyle.Render(appName) + "  " + headerUsernameStyle.Render("@"+username)
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
