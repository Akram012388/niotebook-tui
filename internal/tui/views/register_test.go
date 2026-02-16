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

func TestRegisterHelpText(t *testing.T) {
	m := views.NewRegisterModel(nil)
	text := m.HelpText()
	if text == "" {
		t.Error("HelpText should return non-empty string")
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

func TestRegisterSubmitEmptyUsername(t *testing.T) {
	m := views.NewRegisterModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("expected nil cmd when username is empty")
	}
	view := m.View()
	if !strings.Contains(view, "username is required") {
		t.Error("expected username validation error")
	}
}

func TestRegisterSubmitShortUsername(t *testing.T) {
	m := views.NewRegisterModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	for _, r := range "ab" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("expected nil cmd for short username")
	}
	view := m.View()
	if !strings.Contains(view, "at least 3 characters") {
		t.Error("expected short username error")
	}
}

func TestRegisterSubmitInvalidEmail(t *testing.T) {
	m := views.NewRegisterModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	for _, r := range "akram" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	for _, r := range "notanemail" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("expected nil cmd for invalid email")
	}
	view := m.View()
	if !strings.Contains(view, "invalid email") {
		t.Error("expected invalid email error")
	}
}

func TestRegisterSubmitShortPassword(t *testing.T) {
	m := views.NewRegisterModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	for _, r := range "akram" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	for _, r := range "test@example.com" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	for _, r := range "short" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd != nil {
		t.Error("expected nil cmd for short password")
	}
	view := m.View()
	if !strings.Contains(view, "at least 8 characters") {
		t.Error("expected short password error")
	}
}

func TestRegisterSubmitWithNilClient(t *testing.T) {
	m := views.NewRegisterModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	for _, r := range "akram" {
		m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
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

func TestRegisterTabPastPasswordSwitchesToLogin(t *testing.T) {
	m := views.NewRegisterModel(nil)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyTab})
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyTab})
	if cmd == nil {
		t.Fatal("expected cmd for switching to login")
	}
	msg := cmd()
	if _, ok := msg.(app.MsgSwitchToLogin); !ok {
		t.Errorf("expected MsgSwitchToLogin, got %T", msg)
	}
}

func TestRegisterInit(t *testing.T) {
	m := views.NewRegisterModel(nil)
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init should return a blink command")
	}
}

func TestRegisterKeypressClearsError(t *testing.T) {
	m := views.NewRegisterModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = m.Update(app.MsgAuthError{Message: "email taken", Field: "email"})
	view := m.View()
	if !strings.Contains(view, "email taken") {
		t.Fatal("error should be visible")
	}
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	view = m.View()
	if strings.Contains(view, "email taken") {
		t.Error("error should be cleared after keypress")
	}
}
