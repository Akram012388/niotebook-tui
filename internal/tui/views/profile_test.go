package views_test

import (
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/Akram012388/niotebook-tui/internal/tui/app"
	"github.com/Akram012388/niotebook-tui/internal/tui/views"
)

func TestProfileRendersUsernameAndBio(t *testing.T) {
	m := views.NewProfileModel(nil, "", true)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// Simulate profile loaded
	m, _ = m.Update(app.MsgProfileLoaded{
		User: &models.User{
			Username:    "akram",
			DisplayName: "Akram",
			Bio:         "Building tools for developers.",
			CreatedAt:   time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		},
		Posts: nil,
	})

	view := m.View()
	if !strings.Contains(view, "@akram") {
		t.Error("view missing username")
	}
	if !strings.Contains(view, "Akram") {
		t.Error("view missing display name")
	}
	if !strings.Contains(view, "Building tools for developers.") {
		t.Error("view missing bio")
	}
	if !strings.Contains(view, "Joined Feb 2026") {
		t.Error("view missing joined date")
	}
}

func TestProfileRendersUserPosts(t *testing.T) {
	m := views.NewProfileModel(nil, "", true)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	posts := []models.Post{
		{
			ID:        "1",
			Author:    &models.User{Username: "akram"},
			Content:   "First post content",
			CreatedAt: time.Now().Add(-5 * time.Minute),
		},
		{
			ID:        "2",
			Author:    &models.User{Username: "akram"},
			Content:   "Second post content",
			CreatedAt: time.Now().Add(-1 * time.Hour),
		},
	}

	m, _ = m.Update(app.MsgProfileLoaded{
		User: &models.User{
			Username:  "akram",
			CreatedAt: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		},
		Posts: posts,
	})

	view := m.View()
	if !strings.Contains(view, "First post content") {
		t.Error("view missing first post")
	}
	if !strings.Contains(view, "Second post content") {
		t.Error("view missing second post")
	}
}

func TestProfileEKeyOnOwnProfileSetsEditing(t *testing.T) {
	m := views.NewProfileModel(nil, "", true)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	m, _ = m.Update(app.MsgProfileLoaded{
		User: &models.User{
			Username:  "akram",
			CreatedAt: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		},
		Posts: nil,
	})

	// Press 'e' on own profile
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if !m.Editing() {
		t.Error("expected editing to be true after pressing e on own profile")
	}
}

func TestProfileEKeyOnOtherProfileDoesNotEdit(t *testing.T) {
	m := views.NewProfileModel(nil, "other-user-id", false)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	m, _ = m.Update(app.MsgProfileLoaded{
		User: &models.User{
			Username:  "other",
			CreatedAt: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		},
		Posts: nil,
	})

	// Press 'e' on other user's profile
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if m.Editing() {
		t.Error("expected editing to be false after pressing e on other user's profile")
	}
}

func TestProfileHelpText(t *testing.T) {
	m := views.NewProfileModel(nil, "", true)
	text := m.HelpText()
	if text == "" {
		t.Error("HelpText should return non-empty string")
	}
}

func TestProfileIsOwn(t *testing.T) {
	m := views.NewProfileModel(nil, "", true)
	if !m.IsOwn() {
		t.Error("expected IsOwn to be true")
	}

	m2 := views.NewProfileModel(nil, "other-id", false)
	if m2.IsOwn() {
		t.Error("expected IsOwn to be false for other user")
	}
}

func TestProfileUser(t *testing.T) {
	m := views.NewProfileModel(nil, "", true)
	// Before loading, User should be nil
	if m.User() != nil {
		t.Error("expected nil User before loading")
	}

	m, _ = m.Update(app.MsgProfileLoaded{
		User: &models.User{
			Username:  "akram",
			CreatedAt: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		},
		Posts: nil,
	})

	if m.User() == nil {
		t.Error("expected non-nil User after loading")
	}
	if m.User().Username != "akram" {
		t.Errorf("User.Username = %q, want %q", m.User().Username, "akram")
	}
}

func TestProfileEscSetsDismissed(t *testing.T) {
	m := views.NewProfileModel(nil, "", true)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	m, _ = m.Update(app.MsgProfileLoaded{
		User: &models.User{
			Username:  "akram",
			CreatedAt: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
		},
		Posts: nil,
	})

	// Press Esc
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyEsc})
	if !m.Dismissed() {
		t.Error("expected dismissed to be true after pressing Esc")
	}
}
