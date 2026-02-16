package components

import (
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/Akram012388/niotebook-tui/internal/tui/theme"
)

type trendingTag struct {
	tag   string
	count string
}

type suggestedWriter struct {
	handle string
	bio    string
}

var trendingTags = []trendingTag{
	{tag: "#niotebook", count: "12 niotes"},
	{tag: "#hello-world", count: "8 niotes"},
	{tag: "#terminal-life", count: "5 niotes"},
	{tag: "#claude-code", count: "42 niotes"},
	{tag: "#codex", count: "28 niotes"},
	{tag: "#opencode", count: "19 niotes"},
	{tag: "#skills", count: "15 niotes"},
	{tag: "#openclaw", count: "7 niotes"},
	{tag: "#mcp-servers", count: "33 niotes"},
	{tag: "#agentic-coding", count: "21 niotes"},
	{tag: "#terminal-love", count: "14 niotes"},
}

var suggestedWriters = []suggestedWriter{
	{handle: "@alice", bio: "loves terminal apps"},
	{handle: "@bob", bio: "building in public"},
	{handle: "@dev_sarah", bio: "Go enthusiast"},
	{handle: "@terminal_fan", bio: "CLI everything"},
	{handle: "@rust_rover", bio: "systems thinker"},
	{handle: "@gopher_grace", bio: "open source lover"},
	{handle: "@vim_master", bio: "keyboard warrior"},
	{handle: "@cloud_nina", bio: "infra nerd"},
}

// DiscoverSection identifies which section is active in the right column.
type DiscoverSection int

const (
	SectionTrending DiscoverSection = 0
	SectionWriters  DiscoverSection = 1
)

// DiscoverState holds interactive state for the right column.
type DiscoverState struct {
	ActiveSection  DiscoverSection
	TrendingCursor int
	WritersCursor  int
	TrendingScroll int
	WritersScroll  int
}

// TrendingCount returns the number of trending tags.
func TrendingCount() int { return len(trendingTags) }

// WritersCount returns the number of suggested writers.
func WritersCount() int { return len(suggestedWriters) }

// RenderDiscover renders the right-column with search bar, trending, and
// writers to follow. The two sections split the available height equally.
// When focused, the active section and cursor item are highlighted.
func RenderDiscover(focused bool, state *DiscoverState, width, height int) string {
	if width == 0 {
		return ""
	}

	innerWidth := width - 2
	if innerWidth < 0 {
		innerWidth = 0
	}

	var sections []string

	// Search bar
	sections = append(sections, renderSearchBar(innerWidth))
	sections = append(sections, "")

	// Calculate available height for the two sections.
	// Search bar takes ~3 lines (border top + content + border bottom) + 1 blank.
	searchLines := 4
	availableHeight := height - searchLines - 2 // padding
	if availableHeight < 4 {
		availableHeight = 4
	}
	halfHeight := availableHeight / 2

	// Trending section
	trendingActive := focused && state != nil && state.ActiveSection == SectionTrending
	sections = append(sections, renderTrendingSection(innerWidth, halfHeight, trendingActive, state))

	sections = append(sections, "")

	// Writers section
	writersActive := focused && state != nil && state.ActiveSection == SectionWriters
	sections = append(sections, renderWritersSection(innerWidth, halfHeight, writersActive, state))

	content := strings.Join(sections, "\n")

	wrapper := lipgloss.NewStyle().
		Width(width).
		Height(height).
		Padding(1, 1)

	return wrapper.Render(content)
}

func renderSearchBar(width int) string {
	placeholder := theme.Hint.Render("Search niotes...")
	bar := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(theme.AccentDim).
		Width(width - 2).
		Padding(0, 1).
		Render(placeholder)
	return bar
}

func renderTrendingSection(width, maxHeight int, sectionActive bool, state *DiscoverState) string {
	var lines []string

	headerStyle := theme.Label
	if sectionActive {
		headerStyle = lipgloss.NewStyle().Bold(true).Foreground(theme.Accent)
	}
	lines = append(lines, headerStyle.Render("Trending"))
	lines = append(lines, theme.Separator(width))

	scrollOffset := 0
	cursor := -1
	if state != nil {
		scrollOffset = state.TrendingScroll
		if sectionActive {
			cursor = state.TrendingCursor
		}
	}

	// Each tag takes 2 lines: tag + count
	headerLines := 2
	itemHeight := 2
	visibleItems := (maxHeight - headerLines) / itemHeight
	if visibleItems < 1 {
		visibleItems = 1
	}

	if scrollOffset > 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(theme.TextMuted).Render("  ▲"))
	}

	tagStyle := lipgloss.NewStyle().Foreground(theme.Accent)
	selectedTagStyle := lipgloss.NewStyle().Foreground(theme.Accent).Bold(true).Reverse(true)

	end := scrollOffset + visibleItems
	if end > len(trendingTags) {
		end = len(trendingTags)
	}

	for i := scrollOffset; i < end; i++ {
		tag := trendingTags[i]
		if i == cursor {
			lines = append(lines, selectedTagStyle.Render(" "+tag.tag+" "))
		} else {
			lines = append(lines, tagStyle.Render(tag.tag))
		}
		lines = append(lines, theme.Caption.Render(tag.count))
	}

	if end < len(trendingTags) {
		lines = append(lines, lipgloss.NewStyle().Foreground(theme.TextMuted).Render("  ▼"))
	}

	return strings.Join(lines, "\n")
}

func renderWritersSection(width, maxHeight int, sectionActive bool, state *DiscoverState) string {
	var lines []string

	headerStyle := theme.Label
	if sectionActive {
		headerStyle = lipgloss.NewStyle().Bold(true).Foreground(theme.Accent)
	}
	lines = append(lines, headerStyle.Render("Writers to follow"))
	lines = append(lines, theme.Separator(width))

	scrollOffset := 0
	cursor := -1
	if state != nil {
		scrollOffset = state.WritersScroll
		if sectionActive {
			cursor = state.WritersCursor
		}
	}

	headerLines := 2
	itemHeight := 2
	visibleItems := (maxHeight - headerLines) / itemHeight
	if visibleItems < 1 {
		visibleItems = 1
	}

	if scrollOffset > 0 {
		lines = append(lines, lipgloss.NewStyle().Foreground(theme.TextMuted).Render("  ▲"))
	}

	handleStyle := lipgloss.NewStyle().Foreground(theme.Accent)
	selectedHandleStyle := lipgloss.NewStyle().Foreground(theme.Accent).Bold(true).Reverse(true)

	end := scrollOffset + visibleItems
	if end > len(suggestedWriters) {
		end = len(suggestedWriters)
	}

	for i := scrollOffset; i < end; i++ {
		w := suggestedWriters[i]
		if i == cursor {
			lines = append(lines, selectedHandleStyle.Render(" "+w.handle+" "))
		} else {
			lines = append(lines, handleStyle.Render(w.handle))
		}
		lines = append(lines, theme.Caption.Render(w.bio))
	}

	if end < len(suggestedWriters) {
		lines = append(lines, lipgloss.NewStyle().Foreground(theme.TextMuted).Render("  ▼"))
	}

	return strings.Join(lines, "\n")
}
