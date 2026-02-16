package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/Akram012388/niotebook-tui/internal/tui/theme"
)

// shortcut represents a single key binding with its description.
type shortcut struct {
	key  string
	desc string
}

// shortcutSection represents a titled group of shortcuts.
type shortcutSection struct {
	title string
	items []shortcut
}

// RenderShortcuts renders context-sensitive keyboard shortcuts for the right
// sidebar. The shortcuts change based on the active view and whether the
// compose overlay is open. When width is 0 an empty string is returned.
func RenderShortcuts(activeView View, composeOpen bool, width, height int) string {
	if width == 0 {
		return ""
	}

	var sections []shortcutSection

	if composeOpen {
		sections = []shortcutSection{
			{
				title: "Compose",
				items: []shortcut{
					{key: "Ctrl+J", desc: "publish"},
					{key: "Esc", desc: "cancel"},
				},
			},
		}
	} else {
		switch activeView {
		case ViewProfile:
			sections = []shortcutSection{
				{
					title: "Navigation",
					items: []shortcut{
						{key: "j/k", desc: "scroll"},
						{key: "Esc", desc: "back"},
					},
				},
				{
					title: "Actions",
					items: []shortcut{
						{key: "e", desc: "edit"},
						{key: "?", desc: "help"},
						{key: "q", desc: "quit"},
					},
				},
			}
		default: // ViewTimeline and all others
			sections = []shortcutSection{
				{
					title: "Navigation",
					items: []shortcut{
						{key: "j/k", desc: "scroll"},
						{key: "g/G", desc: "top/bottom"},
						{key: "Enter", desc: "profile"},
					},
				},
				{
					title: "Actions",
					items: []shortcut{
						{key: "n", desc: "compose"},
						{key: "r", desc: "refresh"},
						{key: "?", desc: "help"},
						{key: "q", desc: "quit"},
					},
				},
			}
		}
	}

	titleStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.TextMuted)

	keyStyle := lipgloss.NewStyle().
		Bold(true).
		Foreground(theme.Accent).
		Width(8)

	descStyle := lipgloss.NewStyle().
		Foreground(theme.TextSecondary)

	var lines []string
	for i, section := range sections {
		if i > 0 {
			// Blank line between sections
			lines = append(lines, "")
		}
		lines = append(lines, titleStyle.Render(section.title))
		for _, item := range section.items {
			line := keyStyle.Render(item.key) + descStyle.Render(item.desc)
			lines = append(lines, line)
		}
	}

	content := strings.Join(lines, "\n")

	wrapper := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(1, 1)

	return wrapper.Render(content)
}
