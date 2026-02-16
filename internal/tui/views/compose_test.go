package views_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Akram012388/niotebook-tui/internal/tui/app"
	"github.com/Akram012388/niotebook-tui/internal/tui/views"
)

func TestComposeBarStartsCollapsed(t *testing.T) {
	m := views.NewComposeModel(nil)
	if m.Expanded() {
		t.Error("compose bar should start collapsed")
	}
}

func TestComposeBarCollapsedShowsPlaceholder(t *testing.T) {
	m := views.NewComposeModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 60, Height: 24})
	view := m.View()
	if !strings.Contains(view, "What's on your mind?") {
		t.Error("collapsed compose should show placeholder text")
	}
	if !strings.Contains(view, "0/140") {
		t.Error("collapsed compose should show 0/140 counter")
	}
}

func TestComposeBarExpand(t *testing.T) {
	m := views.NewComposeModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 60, Height: 24})
	m.Expand()
	if !m.Expanded() {
		t.Error("compose should be expanded after Expand()")
	}
}

func TestComposeBarEscCollapses(t *testing.T) {
	m := views.NewComposeModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 60, Height: 24})
	m.Expand()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if m.Expanded() {
		t.Error("Esc should collapse the compose bar")
	}
	if !m.Cancelled() {
		t.Error("Esc from expanded should set cancelled")
	}
}

func TestComposeBarExpandedShowsHints(t *testing.T) {
	m := views.NewComposeModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 60, Height: 24})
	m.Expand()
	view := m.View()
	if !strings.Contains(view, "Ctrl+Enter") {
		t.Error("expanded compose should show Ctrl+Enter hint")
	}
	if !strings.Contains(view, "Esc") {
		t.Error("expanded compose should show Esc hint")
	}
}

func TestComposeBarTypingUpdatesCounter(t *testing.T) {
	m := views.NewComposeModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 60, Height: 24})
	m.Expand()
	for _, r := range "Hello" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	view := m.View()
	if !strings.Contains(view, "5/140") {
		t.Errorf("expected counter 5/140, got view:\n%s", view)
	}
}

func TestComposeBarCtrlEnterPublishes(t *testing.T) {
	m := views.NewComposeModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 60, Height: 24})
	m.Expand()
	for _, r := range "Test post" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlJ})
	if !m.Submitted() {
		t.Error("Ctrl+Enter with content should submit")
	}
	if cmd == nil {
		t.Error("expected publish cmd")
	}
}

func TestComposeBarEmptyCannotSubmit(t *testing.T) {
	m := views.NewComposeModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 60, Height: 24})
	m.Expand()
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlJ})
	if m.Submitted() {
		t.Error("empty content should not submit")
	}
}

func TestComposeBarOverLimitCannotSubmit(t *testing.T) {
	m := views.NewComposeModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 60, Height: 24})
	m.Expand()
	for _, r := range strings.Repeat("a", 141) {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyCtrlJ})
	if m.Submitted() {
		t.Error("over-limit content should not submit")
	}
}

func TestComposeBarPublishWithNilClient(t *testing.T) {
	m := views.NewComposeModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 60, Height: 24})
	m.Expand()
	for _, r := range "Hello" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyCtrlJ})
	if cmd == nil {
		t.Fatal("expected cmd from publish")
	}
	msg := cmd()
	if _, ok := msg.(app.MsgAPIError); !ok {
		t.Errorf("expected MsgAPIError with nil client, got %T", msg)
	}
}

func TestComposeBarIsTextInputFocused(t *testing.T) {
	m := views.NewComposeModel(nil)
	if m.IsTextInputFocused() {
		t.Error("collapsed compose should not have text input focused")
	}
	m.Expand()
	if !m.IsTextInputFocused() {
		t.Error("expanded compose should have text input focused")
	}
}

func TestComposeBarPosting(t *testing.T) {
	m := views.NewComposeModel(nil)
	if m.Posting() {
		t.Error("Posting should be false initially")
	}
}

func TestComposeBarHelpText(t *testing.T) {
	m := views.NewComposeModel(nil)
	text := m.HelpText()
	if text == "" {
		t.Error("HelpText should return non-empty string")
	}
}

func TestComposeBarInit(t *testing.T) {
	m := views.NewComposeModel(nil)
	cmd := m.Init()
	_ = cmd // collapsed, may return nil
}
