package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/Akram012388/niotebook-tui/internal/tui/theme"
)

// trendingTag represents a hardcoded trending hashtag.
type trendingTag struct {
	tag   string
	count string
}

// suggestedWriter represents a hardcoded writer suggestion.
type suggestedWriter struct {
	handle string
	bio    string
}

// Hardcoded placeholder data for the MVP.
var (
	trendingTags = []trendingTag{
		{tag: "#niotebook", count: "12 niotes"},
		{tag: "#hello-world", count: "8 niotes"},
		{tag: "#terminal-life", count: "5 niotes"},
	}

	suggestedWriters = []suggestedWriter{
		{handle: "@alice", bio: "loves terminal apps"},
		{handle: "@bob", bio: "building in public"},
	}
)

// RenderDiscover renders the right-column discover/trending panel with a
// placeholder search bar, trending tags, and suggested writers. When width is
// 0 an empty string is returned. The focused parameter is accepted for future
// border highlighting but is unused in this MVP.
func RenderDiscover(_ bool, width, height int) string {
	if width == 0 {
		return ""
	}

	innerWidth := width - 2 // account for padding
	if innerWidth < 0 {
		innerWidth = 0
	}

	var sections []string

	// Search bar placeholder
	sections = append(sections, renderSearchBar(innerWidth))

	// Blank line
	sections = append(sections, "")

	// Trending section
	sections = append(sections, renderTrendingSection(innerWidth))

	// Blank line
	sections = append(sections, "")

	// Writers to follow section
	sections = append(sections, renderWritersSection(innerWidth))

	content := strings.Join(sections, "\n")

	wrapper := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(1, 1)

	return wrapper.Render(content)
}

// renderSearchBar renders a rounded-border placeholder search input.
func renderSearchBar(width int) string {
	placeholder := theme.Hint.Render("Search niotes...")

	bar := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.AccentDim).
		Width(width - 2). // account for border
		Padding(0, 1).
		Render(placeholder)

	return bar
}

// renderTrendingSection renders the "Trending" header followed by hardcoded
// trending tags.
func renderTrendingSection(width int) string {
	var lines []string

	// Section header
	lines = append(lines, theme.Label.Render("Trending"))
	lines = append(lines, theme.Separator(width))

	tagStyle := lipgloss.NewStyle().Foreground(theme.Accent)

	for _, tag := range trendingTags {
		lines = append(lines, tagStyle.Render(tag.tag))
		lines = append(lines, theme.Caption.Render(tag.count))
		lines = append(lines, "") // spacing between entries
	}

	// Remove trailing blank line
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return strings.Join(lines, "\n")
}

// renderWritersSection renders the "Writers to follow" header followed by
// hardcoded writer suggestions.
func renderWritersSection(width int) string {
	var lines []string

	// Section header
	lines = append(lines, theme.Label.Render("Writers to follow"))
	lines = append(lines, theme.Separator(width))

	handleStyle := lipgloss.NewStyle().Foreground(theme.Accent)

	for _, w := range suggestedWriters {
		lines = append(lines, handleStyle.Render(w.handle))
		lines = append(lines, theme.Caption.Render(w.bio))
		lines = append(lines, "") // spacing between entries
	}

	// Remove trailing blank line
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}

	return strings.Join(lines, "\n")
}
