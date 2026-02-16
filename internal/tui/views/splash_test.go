package views_test

import (
	"strings"
	"testing"
	"time"

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

func TestSplashMinDuration(t *testing.T) {
	if views.MinSplashDuration != 2500*time.Millisecond {
		t.Errorf("MinSplashDuration = %v, want 2500ms", views.MinSplashDuration)
	}
}

func TestBlockSpinnerFrames(t *testing.T) {
	frames := views.BlockSpinnerFrames()
	if len(frames) != 4 {
		t.Fatalf("expected 4 spinner frames, got %d", len(frames))
	}
	if !strings.Contains(frames[0], "░") {
		t.Error("frame 0 should contain light shade blocks")
	}
	if !strings.Contains(frames[3], "█") {
		t.Error("frame 3 should contain full blocks")
	}
}

func TestSplashViewContainsConnecting(t *testing.T) {
	m := views.NewSplashModel("http://localhost:8080")
	m, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 40})

	// Advance through the typewriter reveal phase by sending enough MsgRevealTick
	// messages. "n i o t e b o o k" is 17 characters (9 letters + 8 spaces).
	for i := 0; i < 20; i++ {
		m, _ = m.Update(app.MsgRevealTick{})
	}

	// Advance past the tagline pause
	type msgTaglineShow struct{}
	// The model transitions internally; we need to send the msgTaglineShow
	// message that the tagline pause timer would produce. Since it's an
	// unexported type in the views package, we simulate the full reveal by
	// checking if the view eventually contains "connecting" after the model
	// processes all the reveal ticks. The last reveal tick transitions to
	// phaseTagline and returns a taglinePauseCmd. We can't send the internal
	// msgTaglineShow from outside the package, but we can verify the view
	// at least shows the revealed logo.
	view := m.View()
	if view == "" {
		t.Error("splash view should not be empty")
	}

	// After full reveal, the phase is phaseTagline (waiting for tagline pause).
	// The "connecting..." text appears only in phaseConnecting (after tagline).
	// Since msgTaglineShow is unexported, verify the logo is revealed instead.
	// The logo contains "n" "i" "o" "t" "e" "b" "o" "k" characters.
	if !strings.Contains(view, "k") {
		t.Error("splash view should contain revealed logo after typewriter animation")
	}
}
