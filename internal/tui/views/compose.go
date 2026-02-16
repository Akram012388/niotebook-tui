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

var (
	composeBoxStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(theme.Accent).
			Padding(1, 2)
	composeTitleStyle = lipgloss.NewStyle().
				Bold(true).
				Foreground(theme.Accent)
	composePromptStyle = lipgloss.NewStyle().
				Foreground(theme.Text).
				MarginBottom(1)
	counterNormalStyle = lipgloss.NewStyle().
				Foreground(theme.TextSecondary)
	counterWarningStyle = lipgloss.NewStyle().
				Foreground(theme.Error).
				Bold(true)
	composeHintStyle = lipgloss.NewStyle().
				Foreground(theme.TextMuted).
				Italic(true)
)

// ComposeModel manages the compose modal state.
type ComposeModel struct {
	textarea  textarea.Model
	client    *client.Client
	submitted bool
	cancelled bool
	posting   bool
	err       error
	width     int
	height    int
}

// NewComposeModel creates a new compose modal model.
func NewComposeModel(c *client.Client) ComposeModel {
	ta := textarea.New()
	ta.Placeholder = "What's on your mind?"
	ta.SetWidth(40)
	ta.SetHeight(4)
	ta.Focus()
	ta.CharLimit = 0 // No hard limit — we handle it ourselves for the counter

	return ComposeModel{
		textarea: ta,
		client:   c,
	}
}

// Submitted returns whether the post was published.
func (m ComposeModel) Submitted() bool {
	return m.submitted
}

// Cancelled returns whether the user cancelled.
func (m ComposeModel) Cancelled() bool {
	return m.cancelled
}

// Posting returns whether a publish request is in flight.
func (m ComposeModel) Posting() bool {
	return m.posting
}

// IsTextInputFocused returns true since the textarea is always focused in compose mode.
func (m ComposeModel) IsTextInputFocused() bool {
	return true
}

// Init returns the initial command for the compose modal.
func (m ComposeModel) Init() tea.Cmd {
	return textarea.Blink
}

// Update handles messages for the compose modal.
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
		return m, nil

	case app.MsgAPIError:
		m.posting = false
		m.err = fmt.Errorf("%s", msg.Message)
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)
	}

	// Pass to textarea
	var cmd tea.Cmd
	m.textarea, cmd = m.textarea.Update(msg)
	return m, cmd
}

func (m ComposeModel) handleKey(msg tea.KeyMsg) (ComposeModel, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.cancelled = true
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
	// Modal is roughly 50 chars wide or half the terminal, whichever is larger
	modalWidth := m.width / 2
	if modalWidth < 40 {
		modalWidth = 40
	}
	if modalWidth > 80 {
		modalWidth = 80
	}
	// Account for border and padding (2 border + 4 padding = 6)
	innerWidth := modalWidth - 6
	if innerWidth < 20 {
		innerWidth = 20
	}
	m.textarea.SetWidth(innerWidth)
}

// View renders the compose modal as a centered overlay.
func (m ComposeModel) View() string {
	var b strings.Builder

	b.WriteString(composeTitleStyle.Render("New Post"))
	b.WriteString("\n\n")

	b.WriteString(composePromptStyle.Render("What's on your mind?"))
	b.WriteString("\n")

	b.WriteString(m.textarea.View())
	b.WriteString("\n\n")

	// Character counter
	content := m.textarea.Value()
	charCount := len([]rune(content))
	counterText := fmt.Sprintf("%d/%d", charCount, maxPostLength)

	if charCount > maxPostLength-10 {
		b.WriteString(counterWarningStyle.Render(counterText))
	} else {
		b.WriteString(counterNormalStyle.Render(counterText))
	}
	b.WriteString("\n\n")

	// Help text
	if m.posting {
		b.WriteString(composeHintStyle.Render("Publishing..."))
	} else {
		b.WriteString(composeHintStyle.Render("Ctrl+Enter: post    Esc: cancel"))
	}

	if m.err != nil {
		b.WriteString("\n")
		b.WriteString(errMsgStyle.Render(m.err.Error()))
	}

	modalContent := composeBoxStyle.Render(b.String())

	return lipgloss.Place(m.width, m.height, lipgloss.Center, lipgloss.Center, modalContent)
}

// HelpText returns the status bar help text for the compose modal.
func (m ComposeModel) HelpText() string {
	return "Ctrl+Enter: publish  Esc: cancel"
}
