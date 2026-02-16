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

// SidebarState holds interactive state for the left column when focused.
type SidebarState struct {
	NavCursor int // which nav item is highlighted (0=Home, 1=Profile, 2=Bookmarks, 3=Settings)
}

// NavItemCount is the number of navigable items.
const NavItemCount = 4

type navItem struct {
	label       string
	placeholder bool
}

var navItems = []navItem{
	{label: "Home", placeholder: false},
	{label: "Profile", placeholder: false},
	{label: "Bookmarks", placeholder: true},
	{label: "Settings", placeholder: true},
}

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
// and shortcuts pushed to the bottom.
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

	// === TOP: ASCII Logo ===
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

		statsStyle := lipgloss.NewStyle().Foreground(theme.TextMuted)
		topSections = append(topSections, statsStyle.Render("0 niotes · 0 following"))

		// === BOTTOM: Shortcuts (pushed to bottom) ===
		for _, s := range shortcuts {
			bottomSections = append(bottomSections, renderShortcut(s.key, s.desc))
		}
	}

	topContent := strings.Join(topSections, "\n")
	bottomContent := strings.Join(bottomSections, "\n")

	topLines := strings.Count(topContent, "\n") + 1
	bottomLines := 0
	if bottomContent != "" {
		bottomLines = strings.Count(bottomContent, "\n") + 1
	}
	padding := 2 // wrapper padding (top + bottom)

	gap := height - topLines - bottomLines - padding
	if gap < 1 {
		gap = 1
	}

	var content string
	if bottomContent != "" {
		content = topContent + strings.Repeat("\n", gap) + bottomContent
	} else {
		content = topContent
	}

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

func renderShortcut(key, desc string) string {
	keyStyle := lipgloss.NewStyle().Foreground(theme.Accent).Bold(true)
	descStyle := lipgloss.NewStyle().Foreground(theme.TextMuted)
	// Pad the key visually to align descriptions.
	// lipgloss.Width measures visual width (excluding ANSI codes).
	renderedKey := keyStyle.Render(key)
	keyWidth := lipgloss.Width(renderedKey)
	padWidth := 5 // fixed column width for keys
	pad := ""
	if keyWidth < padWidth {
		pad = strings.Repeat(" ", padWidth-keyWidth)
	}
	return renderedKey + pad + " " + descStyle.Render(desc)
}
