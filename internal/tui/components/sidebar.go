package components

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/Akram012388/niotebook-tui/internal/build"
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

// shortcut is a key-description pair for the shortcuts reference section.
type shortcut struct {
	key  string
	desc string
}

// shortcuts lists the keyboard shortcuts shown at the bottom of the sidebar.
var shortcuts = []shortcut{
	{"j/k", "scroll"},
	{"Tab", "switch col"},
	{"n", "compose"},
	{"?", "help"},
	{"q", "quit"},
}

// RenderSidebar renders the left sidebar with profile info, navigation, post
// button, version, join date, and keyboard shortcuts. When user is nil (logged
// out), only the logo is shown. When width is 0 an empty string is returned.
// The focused parameter is accepted for future use (border color changes).
func RenderSidebar(user *models.User, activeView View, focused bool, width, height int) string {
	if width == 0 {
		return ""
	}

	innerWidth := width - 2 // account for padding
	if innerWidth < 0 {
		innerWidth = 0
	}

	var sections []string

	// Brand logo
	sections = append(sections, theme.LogoCompact())

	// Separator below logo
	sections = append(sections, theme.Separator(innerWidth))

	if user != nil {
		// Blank line
		sections = append(sections, "")

		// Username with @ prefix
		usernameStyle := lipgloss.NewStyle().Foreground(theme.Accent)
		sections = append(sections, usernameStyle.Render("@"+user.Username))

		// Display name (if not empty)
		if user.DisplayName != "" {
			displayStyle := lipgloss.NewStyle().Foreground(theme.Text)
			sections = append(sections, displayStyle.Render(user.DisplayName))
		}

		// Blank line
		sections = append(sections, "")

		// Navigation items
		sections = append(sections, renderNavItem("Home", activeView == ViewTimeline))
		sections = append(sections, renderNavItem("Profile", activeView == ViewProfile))
		sections = append(sections, renderNavItemPlaceholder("Bookmarks"))
		sections = append(sections, renderNavItemPlaceholder("Settings"))

		// Blank line
		sections = append(sections, "")

		// Post button
		postBtn := lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.AccentDim).
			Foreground(theme.Accent).
			Bold(true).
			Align(lipgloss.Center).
			Width(innerWidth)
		sections = append(sections, postBtn.Render("Post"))

		// Blank line
		sections = append(sections, "")

		// Separator
		sections = append(sections, theme.Separator(innerWidth))

		// Version
		versionStyle := lipgloss.NewStyle().Foreground(theme.TextMuted)
		sections = append(sections, versionStyle.Render("v"+build.Version))

		// Join date
		if !user.CreatedAt.IsZero() {
			captionStyle := lipgloss.NewStyle().Foreground(theme.TextSecondary)
			joinDate := user.CreatedAt.Format("Joined Jan 2006")
			sections = append(sections, captionStyle.Render(joinDate))
		}

		// Blank line
		sections = append(sections, "")

		// Shortcuts header
		shortcutsHeader := lipgloss.NewStyle().Foreground(theme.TextSecondary).Bold(true)
		sections = append(sections, shortcutsHeader.Render("Shortcuts"))

		// Shortcut key-desc pairs
		for _, s := range shortcuts {
			sections = append(sections, renderShortcut(s.key, s.desc))
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

// renderNavItemPlaceholder renders a greyed-out nav item in TextMuted to
// indicate a feature that is not yet available.
func renderNavItemPlaceholder(label string) string {
	style := lipgloss.NewStyle().
		Foreground(theme.TextMuted)
	return style.Render("  " + label)
}

// renderShortcut renders a single shortcut line with key in Accent and
// description in TextMuted.
func renderShortcut(key, desc string) string {
	keyStyle := lipgloss.NewStyle().Foreground(theme.Accent).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(theme.TextMuted)
	return fmt.Sprintf("%-6s %s", keyStyle.Render(key), descStyle.Render(desc))
}
