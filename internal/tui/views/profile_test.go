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

func TestProfileKeyNavigationJK(t *testing.T) {
	m := views.NewProfileModel(nil, "", true)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	user := &models.User{
		ID: "user-1", Username: "testuser", DisplayName: "Test",
		Bio: "bio", CreatedAt: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
	}
	posts := []models.Post{
		{ID: "1", Content: "Post 1", Author: &models.User{Username: "testuser"}, CreatedAt: time.Now()},
		{ID: "2", Content: "Post 2", Author: &models.User{Username: "testuser"}, CreatedAt: time.Now()},
		{ID: "3", Content: "Post 3", Author: &models.User{Username: "testuser"}, CreatedAt: time.Now()},
	}
	m, _ = m.Update(app.MsgProfileLoaded{User: user, Posts: posts})

	// Navigate down with j
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	// Navigate up with k
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	// Go to bottom with G
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	// Go to top with g
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})

	view := m.View()
	if view == "" {
		t.Error("expected non-empty profile view")
	}
}

func TestProfileArrowKeyNavigation(t *testing.T) {
	m := views.NewProfileModel(nil, "", true)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	user := &models.User{
		ID: "user-1", Username: "testuser",
		CreatedAt: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
	}
	posts := []models.Post{
		{ID: "1", Content: "Post 1", Author: &models.User{Username: "testuser"}, CreatedAt: time.Now()},
		{ID: "2", Content: "Post 2", Author: &models.User{Username: "testuser"}, CreatedAt: time.Now()},
	}
	m, _ = m.Update(app.MsgProfileLoaded{User: user, Posts: posts})

	// Down arrow
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	// Up arrow
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})

	if m.Dismissed() {
		t.Error("navigation should not dismiss profile")
	}
}

func TestProfileHelpTextOther(t *testing.T) {
	m := views.NewProfileModel(nil, "other-user", false)
	text := m.HelpText()
	if text == "" {
		t.Error("expected non-empty help text for other profile")
	}
	if strings.Contains(text, "edit") {
		t.Error("other user's help text should not mention edit")
	}
}

func TestProfileInitReturnsCmd(t *testing.T) {
	m := views.NewProfileModel(nil, "", true)
	cmd := m.Init()
	if cmd == nil {
		t.Error("Init should return a fetch command")
	}
	// With nil client, the cmd should return an API error
	msg := cmd()
	if _, ok := msg.(app.MsgAPIError); !ok {
		t.Errorf("expected MsgAPIError with nil client, got %T", msg)
	}
}

func TestProfileUpdatedClearsEditing(t *testing.T) {
	m := views.NewProfileModel(nil, "", true)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m, _ = m.Update(app.MsgProfileLoaded{
		User:  &models.User{Username: "akram", CreatedAt: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)},
		Posts: nil,
	})

	// Start editing
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'e'}})
	if !m.Editing() {
		t.Fatal("expected editing after e key")
	}

	// Simulate profile update
	m, _ = m.Update(app.MsgProfileUpdated{
		User: &models.User{Username: "akram", Bio: "Updated bio", CreatedAt: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC)},
	})
	if m.Editing() {
		t.Error("expected editing to be false after profile update")
	}
}

func TestProfileLoadingState(t *testing.T) {
	m := views.NewProfileModel(nil, "", true)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := m.View()
	if !strings.Contains(view, "Loading") {
		t.Error("expected loading state")
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
