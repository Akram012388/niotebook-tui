package views_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Akram012388/niotebook-tui/internal/tui/app"
	"github.com/Akram012388/niotebook-tui/internal/tui/views"
)

func TestComposeTypingUpdatesCounterr(t *testing.T) {
	m := views.NewComposeModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Type some text
	for _, r := range "Hello" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	view := m.View()
	if !strings.Contains(view, "5/140") {
		t.Errorf("expected counter to show 5/140, got view:\n%s", view)
	}
}

func TestComposeCtrlEnterPublishes(t *testing.T) {
	m := views.NewComposeModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Type some text
	for _, r := range "Test post" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	// Press Ctrl+E to publish
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlE})
	if !m.Submitted() {
		t.Error("expected submitted to be true after Ctrl+Enter with content")
	}
	if cmd == nil {
		t.Error("expected publish cmd to be returned")
	}
}

func TestComposeEscCancels(t *testing.T) {
	m := views.NewComposeModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Press Esc
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !m.Cancelled() {
		t.Error("expected cancelled to be true after Esc")
	}
}

func TestComposeOverLimitDisablesSubmit(t *testing.T) {
	m := views.NewComposeModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Type more than 140 characters
	long := strings.Repeat("a", 141)
	for _, r := range long {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	// Try Ctrl+Enter — should not submit
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlE})
	if m.Submitted() {
		t.Error("expected submitted to be false when over character limit")
	}

	view := m.View()
	if !strings.Contains(view, "141/140") {
		t.Errorf("expected counter to show 141/140 when over limit, got view:\n%s", view)
	}
}

func TestComposeEmptyDisablesSubmit(t *testing.T) {
	m := views.NewComposeModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Try Ctrl+Enter with empty content
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlE})
	if m.Submitted() {
		t.Error("expected submitted to be false when content is empty")
	}
}

func TestComposePublishReturnsPostMessage(t *testing.T) {
	m := views.NewComposeModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	for _, r := range "Hello world" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}

	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlE})
	if cmd == nil {
		t.Fatal("expected cmd from publish")
	}

	// Execute the cmd — with nil client it should return an API error
	msg := cmd()
	if _, ok := msg.(app.MsgAPIError); !ok {
		t.Errorf("expected MsgAPIError with nil client, got %T", msg)
	}
}

func TestComposeModelAPIError(t *testing.T) {
	m := views.NewComposeModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	_, cmd := m.Update(app.MsgAPIError{Message: "post too long"})
	_ = cmd
}
