package components

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"

	"github.com/Akram012388/niotebook-tui/internal/models"
)

var (
	usernameStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("6")).
			Bold(true)

	selectedUsernameStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("5")).
				Bold(true)

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Faint(true)

	separatorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	markerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("5")).
			Bold(true)
)

// RenderPostCard renders a single post card. If selected is true, the post
// is highlighted with an accent marker. width is the total terminal width.
func RenderPostCard(post models.Post, width int, selected bool, now time.Time) string {
	var b strings.Builder

	// Marker: "▸ " when selected, "  " when not
	marker := "  "
	if selected {
		marker = markerStyle.Render("▸") + " "
	}

	// Username
	username := "@"
	if post.Author != nil {
		username += post.Author.Username
	} else {
		username += "unknown"
	}
	var styledUsername string
	if selected {
		styledUsername = selectedUsernameStyle.Render(username)
	} else {
		styledUsername = usernameStyle.Render(username)
	}

	// Separator dot and relative time
	sep := dimStyle.Render(" · ")
	relTime := dimStyle.Render(RelativeTimeFrom(post.CreatedAt, now))

	// Header line
	b.WriteString(marker)
	b.WriteString(styledUsername)
	b.WriteString(sep)
	b.WriteString(relTime)
	b.WriteString("\n")

	// Content, word-wrapped with 2-char left padding
	contentWidth := width - 2
	if contentWidth < 10 {
		contentWidth = 10
	}
	wrapped := ansi.Wordwrap(post.Content, contentWidth, "")
	for _, line := range strings.Split(wrapped, "\n") {
		b.WriteString("  ")
		b.WriteString(line)
		b.WriteString("\n")
	}

	// Separator line
	if width < 1 {
		width = 1
	}
	b.WriteString(separatorStyle.Render(strings.Repeat("─", width)))

	return b.String()
}
