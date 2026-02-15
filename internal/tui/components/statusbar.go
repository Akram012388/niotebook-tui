package components

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type statusKind int

const (
	statusNone statusKind = iota
	statusError
	statusSuccess
	statusLoading
)

var (
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2"))

	loadingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("3"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8")).
			Faint(true)
)

// MsgStatusClear is sent when the status bar auto-clear timer fires.
type MsgStatusClear struct{}

// StatusBarModel manages the status bar state.
type StatusBarModel struct {
	message string
	kind    statusKind
}

// NewStatusBarModel returns an empty status bar.
func NewStatusBarModel() StatusBarModel {
	return StatusBarModel{}
}

// SetError sets an error message and returns a tea.Cmd to auto-clear after 5s.
func (m *StatusBarModel) SetError(msg string) tea.Cmd {
	m.message = msg
	m.kind = statusError
	return clearAfter(5 * time.Second)
}

// SetSuccess sets a success message and returns a tea.Cmd to auto-clear after 5s.
func (m *StatusBarModel) SetSuccess(msg string) tea.Cmd {
	m.message = msg
	m.kind = statusSuccess
	return clearAfter(5 * time.Second)
}

// SetLoading sets a loading message (no auto-clear).
func (m *StatusBarModel) SetLoading(msg string) {
	m.message = msg
	m.kind = statusLoading
}

// Clear resets the status bar.
func (m *StatusBarModel) Clear() {
	m.message = ""
	m.kind = statusNone
}

// View renders the status bar. The helpText appears on the left when no status
// is active. The status message appears on the right side.
func (m StatusBarModel) View(helpText string, width int) string {
	if m.kind == statusNone || m.message == "" {
		return helpStyle.Width(width).Render(helpText)
	}

	var styled string
	switch m.kind {
	case statusError:
		styled = errorStyle.Render(m.message)
	case statusSuccess:
		styled = successStyle.Render(m.message)
	case statusLoading:
		styled = loadingStyle.Render(m.message)
	default:
		styled = m.message
	}

	left := helpStyle.Render(helpText)
	leftWidth := lipgloss.Width(left)
	rightWidth := lipgloss.Width(styled)
	gap := width - leftWidth - rightWidth
	if gap < 1 {
		gap = 1
	}

	return left + lipgloss.NewStyle().Width(gap).Render("") + styled
}

func clearAfter(d time.Duration) tea.Cmd {
	return tea.Tick(d, func(_ time.Time) tea.Msg {
		return MsgStatusClear{}
	})
}
