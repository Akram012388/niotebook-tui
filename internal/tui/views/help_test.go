package views_test

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Akram012388/niotebook-tui/internal/tui/views"
)

func TestHelpModelNewAndDismissed(t *testing.T) {
	m := views.NewHelpModel(views.HelpViewTimeline)
	if m.Dismissed() {
		t.Error("new help model should not be dismissed")
	}
}

func TestHelpModelInit(t *testing.T) {
	m := views.NewHelpModel(views.HelpViewTimeline)
	cmd := m.Init()
	if cmd != nil {
		t.Error("Init should return nil")
	}
}

func TestHelpModelDismissOnEsc(t *testing.T) {
	m := views.NewHelpModel(views.HelpViewTimeline)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !m.Dismissed() {
		t.Error("expected dismissed after Esc")
	}
}

func TestHelpModelDismissOnQuestionMark(t *testing.T) {
	m := views.NewHelpModel(views.HelpViewTimeline)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'?'}})
	if !m.Dismissed() {
		t.Error("expected dismissed after ?")
	}
}

func TestHelpModelDismissOnQ(t *testing.T) {
	m := views.NewHelpModel(views.HelpViewTimeline)
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'q'}})
	if !m.Dismissed() {
		t.Error("expected dismissed after q")
	}
}

func TestHelpModelWindowSize(t *testing.T) {
	m := views.NewHelpModel(views.HelpViewTimeline)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 100, Height: 50})
	// Should not panic and should not be dismissed
	if m.Dismissed() {
		t.Error("window size should not dismiss help")
	}
}

func TestHelpModelViewContainsBindings(t *testing.T) {
	m := views.NewHelpModel(views.HelpViewTimeline)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := m.View()
	if !strings.Contains(view, "Key Bindings") {
		t.Error("view missing title")
	}
	if !strings.Contains(view, "j/k") {
		t.Error("view missing j/k binding")
	}
	if !strings.Contains(view, "Quit") {
		t.Error("view missing Quit description")
	}
}

func TestHelpModelViewProfileBindings(t *testing.T) {
	m := views.NewHelpModel(views.HelpViewProfile)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := m.View()
	if !strings.Contains(view, "Edit bio") {
		t.Error("profile help missing edit bio binding")
	}
}

func TestHelpModelViewComposeBindings(t *testing.T) {
	m := views.NewHelpModel(views.HelpViewCompose)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := m.View()
	if !strings.Contains(view, "Publish") {
		t.Error("compose help missing publish binding")
	}
}

func TestHelpModelViewUnknownFallback(t *testing.T) {
	m := views.NewHelpModel("unknown-view")
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := m.View()
	// Should fallback to timeline bindings
	if !strings.Contains(view, "j/k") {
		t.Error("unknown view should fallback to timeline bindings")
	}
}

func TestHelpModelHelpText(t *testing.T) {
	m := views.NewHelpModel(views.HelpViewTimeline)
	text := m.HelpText()
	if text == "" {
		t.Error("HelpText should return non-empty string")
	}
	if !strings.Contains(text, "close") {
		t.Error("HelpText should mention close")
	}
}
