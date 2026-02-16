package views_test

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Akram012388/niotebook-tui/internal/tui/app"
	"github.com/Akram012388/niotebook-tui/internal/tui/views"
)

func TestSplashScreenRenders(t *testing.T) {
	m := views.NewSplashModel("http://localhost:8080")
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view from splash screen")
	}
}

func TestSplashScreenInit(t *testing.T) {
	m := views.NewSplashModel("http://localhost:8080")
	cmd := m.Init()
	if cmd == nil {
		t.Error("expected non-nil command from Init")
	}
}

func TestSplashScreenNotDoneInitially(t *testing.T) {
	m := views.NewSplashModel("http://localhost:8080")
	if m.Done() {
		t.Error("splash screen should not be done initially")
	}
	if m.Failed() {
		t.Error("splash screen should not have failed initially")
	}
}

func TestSplashScreenConnected(t *testing.T) {
	m := views.NewSplashModel("http://localhost:8080")
	sm, _ := m.Update(app.MsgServerConnected{})
	if !sm.Done() {
		t.Error("expected Done() to be true after MsgServerConnected")
	}
	if sm.Failed() {
		t.Error("expected Failed() to be false after MsgServerConnected")
	}
}

func TestSplashScreenFailed(t *testing.T) {
	m := views.NewSplashModel("http://localhost:8080")
	sm, _ := m.Update(app.MsgServerFailed{Err: "connection refused"})
	if !sm.Failed() {
		t.Error("expected Failed() to be true after MsgServerFailed")
	}
	if sm.Done() {
		t.Error("expected Done() to be false after MsgServerFailed")
	}
	if sm.ErrorMessage() != "connection refused" {
		t.Errorf("ErrorMessage() = %q, want %q", sm.ErrorMessage(), "connection refused")
	}
}

func TestSplashScreenResize(t *testing.T) {
	m := views.NewSplashModel("http://localhost:8080")
	sm, _ := m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
	// Should not panic, and state should be unchanged
	if sm.Done() {
		t.Error("resize should not set done")
	}
	if sm.Failed() {
		t.Error("resize should not set failed")
	}
}
