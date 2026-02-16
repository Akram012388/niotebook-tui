package views

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// View constants for help binding context.
const (
	HelpViewTimeline = "timeline"
	HelpViewProfile  = "profile"
	HelpViewCompose  = "compose"
)

// HelpEntry represents a single key binding help entry.
type HelpEntry struct {
	Key         string
	Description string
}

var helpBindings = map[string][]HelpEntry{
	HelpViewTimeline: {
		{"j/k", "Scroll up/down"},
		{"n", "New post"},
		{"r", "Refresh"},
		{"Enter", "View profile"},
		{"p", "Own profile"},
		{"g/G", "Top/bottom"},
		{"?", "Close help"},
		{"q", "Quit"},
	},
	HelpViewProfile: {
		{"j/k", "Scroll up/down"},
		{"e", "Edit bio (own profile)"},
		{"Esc", "Back to timeline"},
		{"?", "Close help"},
		{"q", "Quit"},
	},
	HelpViewCompose: {
		{"Ctrl+Enter", "Publish post"},
		{"Esc", "Cancel"},
		{"?", "Close help"},
	},
}

var (
	helpBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("5")).
			Padding(1, 2)

	helpTitleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("5"))

	helpKeyStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("6")).
			Bold(true).
			Width(12)

	helpDescStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("7"))
)

// HelpModel manages the help overlay state.
type HelpModel struct {
	viewName  string
	dismissed bool
	width     int
	height    int
}

// NewHelpModel creates a new help overlay for the given view.
func NewHelpModel(viewName string) HelpModel {
	return HelpModel{
		viewName: viewName,
	}
}

// Dismissed returns whether the help overlay has been closed.
func (m HelpModel) Dismissed() bool {
	return m.dismissed
}

// Init returns the initial command.
func (m HelpModel) Init() tea.Cmd {
	return nil
}

// Update handles messages for the help overlay.
func (m HelpModel) Update(msg tea.Msg) (HelpModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil

	case tea.KeyMsg:
		switch {
		case msg.Type == tea.KeyEsc:
			m.dismissed = true
		case msg.Type == tea.KeyRunes && len(msg.Runes) == 1 && (msg.Runes[0] == '?' || msg.Runes[0] == 'q'):
			m.dismissed = true
		}
	}

	return m, nil
}

// View renders the help overlay as a centered bordered box.
func (m HelpModel) View() string {
	var b strings.Builder

	b.WriteString(helpTitleStyle.Render("Key Bindings"))
	b.WriteString("\n\n")

	bindings, ok := helpBindings[m.viewName]
	if !ok {
		bindings = helpBindings[HelpViewTimeline]
	}

	for _, entry := range bindings {
		b.WriteString(helpKeyStyle.Render(entry.Key))
		b.WriteString(helpDescStyle.Render(entry.Description))
		b.WriteString("\n")
	}

	b.WriteString("\n")
	b.WriteString(hintStyle.Render("Press ? or Esc to close"))

	content := helpBoxStyle.Render(b.String())

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, content)
}

// HelpText returns the status bar help text for the help overlay.
func (m HelpModel) HelpText() string {
	return "?: close  Esc: close  q: close"
}
