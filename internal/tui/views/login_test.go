package views_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/Akram012388/niotebook-tui/internal/tui/app"
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

func TestLoginSubmitEmptyEmail(t *testing.T) {
	m := views.NewLoginModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("expected nil cmd when email is empty")
	}
	view := m.View()
	if !strings.Contains(view, "email is required") {
		t.Error("expected email validation error in view")
	}
}

func TestLoginSubmitEmptyPassword(t *testing.T) {
	m := views.NewLoginModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	for _, r := range "test@example.com" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("expected nil cmd when password is empty")
	}
	view := m.View()
	if !strings.Contains(view, "password is required") {
		t.Error("expected password validation error in view")
	}
}

func TestLoginSubmitWithNilClient(t *testing.T) {
	m := views.NewLoginModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	for _, r := range "test@example.com" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	for _, r := range "password123" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected cmd from submit")
	}
	msg := cmd()
	if _, ok := msg.(app.MsgAuthError); !ok {
		t.Errorf("expected MsgAuthError with nil client, got %T", msg)
	}
}

func TestLoginAuthErrorShowsMessage(t *testing.T) {
	m := views.NewLoginModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = m.Update(app.MsgAuthError{Message: "wrong password", Field: "password"})
	view := m.View()
	if !strings.Contains(view, "wrong password") {
		t.Error("expected error message in view")
	}
}

func TestLoginKeypressClearsError(t *testing.T) {
	m := views.NewLoginModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = m.Update(app.MsgAuthError{Message: "bad password", Field: "password"})
	view := m.View()
	if !strings.Contains(view, "bad password") {
		t.Fatal("error should be visible")
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	view = m.View()
	if strings.Contains(view, "bad password") {
		t.Error("error should be cleared after keypress")
	}
}

func TestLoginTabPastPasswordSwitchesToRegister(t *testing.T) {
	m := views.NewLoginModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.FocusIndex() != 1 {
		t.Fatalf("focus = %d, want 1", m.FocusIndex())
	}
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if cmd == nil {
		t.Fatal("expected cmd for switching to register")
	}
	msg := cmd()
	if _, ok := msg.(app.MsgSwitchToRegister); !ok {
		t.Errorf("expected MsgSwitchToRegister, got %T", msg)
	}
}

func TestLoginInit(t *testing.T) {
	m := views.NewLoginModel(nil)
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init should return a blink command")
	}
}
