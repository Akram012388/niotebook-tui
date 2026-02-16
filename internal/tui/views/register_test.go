package views_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Akram012388/niotebook-tui/internal/tui/app"
	"github.com/Akram012388/niotebook-tui/internal/tui/views"
)

func TestRegisterModelRenders(t *testing.T) {
	m := views.NewRegisterModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	output := m.View()
	if !strings.Contains(output, "Register") {
		t.Error("expected Register title in output")
	}
	if !strings.Contains(output, "Username") {
		t.Error("expected Username label in output")
	}
	if !strings.Contains(output, "Email") {
		t.Error("expected Email label in output")
	}
	if !strings.Contains(output, "Password") {
		t.Error("expected Password label in output")
	}
}

func TestRegisterModelTabNavigation(t *testing.T) {
	m := views.NewRegisterModel(nil)
	if m.FocusIndex() != 0 {
		t.Errorf("initial focus = %d, want 0", m.FocusIndex())
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.FocusIndex() != 1 {
		t.Errorf("after tab focus = %d, want 1", m.FocusIndex())
	}

	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if m.FocusIndex() != 2 {
		t.Errorf("after 2 tabs focus = %d, want 2", m.FocusIndex())
	}
}

func TestRegisterModelShiftTabNavigation(t *testing.T) {
	m := views.NewRegisterModel(nil)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab}) // go to email
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyShiftTab})
	if m.FocusIndex() != 0 {
		t.Errorf("after shift-tab focus = %d, want 0", m.FocusIndex())
	}
}

func TestRegisterModelAuthError(t *testing.T) {
	m := views.NewRegisterModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = m.Update(app.MsgAuthError{Message: "email already taken", Field: "email"})
	output := m.View()
	if !strings.Contains(output, "email already taken") {
		t.Error("expected error message in output")
	}
}
