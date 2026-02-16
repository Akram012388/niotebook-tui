package views_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/Akram012388/niotebook-tui/internal/tui/views"
)

func TestLoginViewRender(t *testing.T) {
	m := views.NewLoginModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := m.View()
	if view == "" {
		t.Error("expected non-empty view")
	}
	// Should contain login form elements
	if !containsAny(view, "Email", "Password", "Login") {
		t.Error("view missing form elements")
	}
}

func TestLoginViewTabSwitchesField(t *testing.T) {
	m := views.NewLoginModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Initial focus on email
	if m.FocusIndex() != 0 {
		t.Errorf("initial focus = %d, want 0 (email)", m.FocusIndex())
	}

	// Tab moves to password
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.FocusIndex() != 1 {
		t.Errorf("after tab focus = %d, want 1 (password)", m.FocusIndex())
	}
}

func TestLoginHelpText(t *testing.T) {
	m := views.NewLoginModel(nil)
	text := m.HelpText()
	if text == "" {
		t.Error("HelpText should return non-empty string")
	}
}

func TestLoginShiftTabMovesFocus(t *testing.T) {
	m := views.NewLoginModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Move to password
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.FocusIndex() != 1 {
		t.Errorf("focus = %d, want 1", m.FocusIndex())
	}

	// Shift+Tab back to email
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if m.FocusIndex() != 0 {
		t.Errorf("focus = %d after shift-tab, want 0", m.FocusIndex())
	}
}

func containsAny(s string, substrs ...string) bool {
	for _, sub := range substrs {
		if strings.Contains(s, sub) {
			return true
		}
	}
	return false
}
