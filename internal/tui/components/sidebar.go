package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/Akram012388/niotebook-tui/internal/tui/theme"
)

// View identifies the active screen. These constants mirror the values in the
// app package so that components can receive a view identifier without creating
// an import cycle (app already imports components).
type View int

const (
	ViewSplash   View = iota // 0
	ViewLogin                // 1
	ViewRegister             // 2
	ViewTimeline             // 3
	ViewProfile              // 4
)

// RenderSidebar renders the left sidebar with profile info and navigation.
// When user is nil (logged out), only the logo is shown. When width is 0 an
// empty string is returned.
func RenderSidebar(user *models.User, activeView View, width, height int) string {
	if width == 0 {
		return ""
	}

	var sections []string

	// Brand logo
	sections = append(sections, theme.LogoCompact())

	if user != nil {
		// Blank line after logo
		sections = append(sections, "")

		// Username with @ prefix
		usernameStyle := lipgloss.NewStyle().Foreground(theme.Accent)
		sections = append(sections, usernameStyle.Render("@"+user.Username))

		// Display name (if not empty)
		if user.DisplayName != "" {
			displayStyle := lipgloss.NewStyle().Foreground(theme.Text)
			sections = append(sections, displayStyle.Render(user.DisplayName))
		}

		// Separator
		innerWidth := width - 2 // account for padding
		if innerWidth < 0 {
			innerWidth = 0
		}
		sections = append(sections, theme.Separator(innerWidth))

		// Blank line
		sections = append(sections, "")

		// Navigation items
		sections = append(sections, renderNavItem("Home", activeView == ViewTimeline))
		sections = append(sections, renderNavItem("Profile", activeView == ViewProfile))

		// Blank line
		sections = append(sections, "")

		// Separator
		sections = append(sections, theme.Separator(innerWidth))

		// Join date
		if !user.CreatedAt.IsZero() {
			captionStyle := lipgloss.NewStyle().Foreground(theme.TextSecondary)
			joinDate := user.CreatedAt.Format("Joined Jan 2006")
			sections = append(sections, captionStyle.Render(joinDate))
		}
	}

	content := strings.Join(sections, "\n")

	wrapper := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(1, 1)

	return wrapper.Render(content)
}

// renderNavItem renders a single navigation item with an active indicator.
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
