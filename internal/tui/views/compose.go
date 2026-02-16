package views

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textarea"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/Akram012388/niotebook-tui/internal/tui/app"
	"github.com/Akram012388/niotebook-tui/internal/tui/client"
	"github.com/Akram012388/niotebook-tui/internal/tui/theme"
)

const maxPostLength = 140

// Collapsed bar styles: AccentDim border, muted placeholder, secondary counter.
var (
	collapsedBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(theme.AccentDim).
				Padding(0, 1)

	collapsedPlaceholderStyle = lipgloss.NewStyle().
					Foreground(theme.TextMuted)

	counterNormalStyle = lipgloss.NewStyle().
				Foreground(theme.TextSecondary)

	counterWarningStyle = lipgloss.NewStyle().
				Foreground(theme.Error).
				Bold(true)
)

// Expanded bar styles: Accent border, hints inside the box.
var (
	expandedBoxStyle = lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(theme.Accent).
				Padding(0, 1)

	composeHintStyle = lipgloss.NewStyle().
				Foreground(theme.TextMuted).
				Italic(true)
)

// ComposeModel manages the inline compose bar state.
type ComposeModel struct {
	textarea  textarea.Model
	client    *client.Client
	expanded  bool
	submitted bool
	cancelled bool
	posting   bool
	err       error
	width     int
	height    int
}

// NewComposeModel creates a new inline compose bar model. It starts collapsed.
func NewComposeModel(c *client.Client) ComposeModel {
	ta := textarea.New()
	ta.Placeholder = "What's on your mind?"
	ta.SetWidth(40)
	ta.SetHeight(3)
	ta.CharLimit = 0 // No hard limit; we handle it ourselves for the counter
	// Start unfocused — textarea is only focused when expanded.
	ta.Blur()

	return ComposeModel{
		textarea: ta,
		client:   c,
	}
}

// Submitted returns whether the post was published.
func (m ComposeModel) Submitted() bool {
	return m.submitted
}

// Cancelled returns whether the user cancelled from expanded state.
func (m ComposeModel) Cancelled() bool {
	return m.cancelled
}

// Expanded returns whether the compose bar is in expanded mode.
func (m ComposeModel) Expanded() bool {
	return m.expanded
}

// Posting returns whether a publish request is in flight.
func (m ComposeModel) Posting() bool {
	return m.posting
}

// IsTextInputFocused returns true only when the compose bar is expanded.
func (m ComposeModel) IsTextInputFocused() bool {
	return m.expanded
}

// Expand switches the compose bar to expanded mode and focuses the textarea.
func (m *ComposeModel) Expand() {
	m.expanded = true
	m.cancelled = false
	m.submitted = false
	m.textarea.Focus()
}

// Init returns the initial command. Collapsed state returns nil.
func (m ComposeModel) Init() tea.Cmd {
	if m.expanded {
		return textarea.Blink
	}
	return nil
}

// Update handles messages for the inline compose bar.
func (m ComposeModel) Update(msg tea.Msg) (ComposeModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.updateTextareaSize()
		return m, nil

	case app.MsgPostPublished:
		m.posting = false
		m.submitted = true
		m.expanded = false
		m.textarea.Reset()
		m.textarea.Blur()
		return m, nil

	case app.MsgAPIError:
		m.posting = false
		m.err = fmt.Errorf("%s", msg.Message)
		return m, nil

	case tea.KeyMsg:
		if !m.expanded {
			// Collapsed: keys are not consumed
			return m, nil
		}
		return m.handleKey(msg)
	}

	// Pass to textarea only when expanded
	if m.expanded {
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd
	}
	return m, nil
}

func (m ComposeModel) handleKey(msg tea.KeyMsg) (ComposeModel, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.expanded = false
		m.cancelled = true
		m.textarea.Reset()
		m.textarea.Blur()
		return m, nil

	case tea.KeyCtrlJ: // Ctrl+Enter (terminal sends Ctrl+J / LF for Ctrl+Enter)
		content := strings.TrimSpace(m.textarea.Value())
		charCount := len([]rune(content))
		if charCount == 0 || charCount > maxPostLength {
			return m, nil
		}
		m.submitted = true
		m.posting = true
		return m, m.publish(content)
	}

	// Pass to textarea
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m ComposeModel) publish(content string) tea.Cmd {
	c := m.client
	return func() tea.Msg {
		if c == nil {
			return app.MsgAPIError{Message: "no server connection"}
		}
		post, err := c.CreatePost(content)
		if err != nil {
			return app.MsgAPIError{Message: err.Error()}
		}
		return app.MsgPostPublished{Post: *post}
	}
}

func (m *ComposeModel) updateTextareaSize() {
	// The expanded box style has: border (2) + padding (2) = 4 chars horizontal
	// The textarea itself needs to fit inside that, so subtract 4 from m.width
	boxWidth := m.width - 4
	if boxWidth < 20 {
		boxWidth = 20
	}
	// Textarea width should match boxWidth exactly for clean hard-wrap
	m.textarea.SetWidth(boxWidth)
}

// View renders the compose bar — either collapsed (single line) or expanded (textarea).
func (m ComposeModel) View() string {
	if m.expanded {
		return m.viewExpanded()
	}
	return m.viewCollapsed()
}

// viewCollapsed renders a single-line bar with placeholder and counter.
func (m ComposeModel) viewCollapsed() string {
	placeholder := collapsedPlaceholderStyle.Render("What's on your mind?")
	counter := counterNormalStyle.Render("0/140")

	// Calculate spacing between placeholder and counter
	boxWidth := m.width - 4 // border (2) + padding (2)
	if boxWidth < 20 {
		boxWidth = 20
	}

	placeholderLen := lipgloss.Width(placeholder)
	counterLen := lipgloss.Width(counter)
	gap := boxWidth - placeholderLen - counterLen
	if gap < 1 {
		gap = 1
	}

	line := placeholder + strings.Repeat(" ", gap) + counter

	return collapsedBoxStyle.Width(boxWidth).Render(line)
}

// viewExpanded renders the full textarea with counter and hints.
func (m ComposeModel) viewExpanded() string {
	boxWidth := m.width - 4 // border (2) + padding (2)
	if boxWidth < 20 {
		boxWidth = 20
	}

	var b strings.Builder

	b.WriteString(m.textarea.View())
	b.WriteString("\n")

	// Character counter
	content := m.textarea.Value()
	charCount := len([]rune(content))
	counterText := fmt.Sprintf("%d/%d", charCount, maxPostLength)

	var counterRendered string
	if charCount > maxPostLength-10 {
		counterRendered = counterWarningStyle.Render(counterText)
	} else {
		counterRendered = counterNormalStyle.Render(counterText)
	}

	// Hints
	var hints string
	if m.posting {
		hints = composeHintStyle.Render("Publishing...")
	} else {
		hints = composeHintStyle.Render("Ctrl+Enter: post    Esc: cancel")
	}

	// Counter on left, hints on right
	hintsLen := lipgloss.Width(hints)
	counterRenderedLen := lipgloss.Width(counterRendered)
	gap := boxWidth - counterRenderedLen - hintsLen
	if gap < 1 {
		gap = 1
	}

	b.WriteString(counterRendered + strings.Repeat(" ", gap) + hints)

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(lipgloss.NewStyle().Foreground(theme.Error).Render(m.err.Error()))
	}

	return expandedBoxStyle.Width(boxWidth).Render(b.String())
}

// HelpText returns the status bar help text for the compose bar.
func (m ComposeModel) HelpText() string {
	if m.expanded {
		return "Ctrl+Enter: publish  Esc: cancel"
	}
	return "n: compose"
}
