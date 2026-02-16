package views_test

import (
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/Akram012388/niotebook-tui/internal/tui/app"
	"github.com/Akram012388/niotebook-tui/internal/tui/views"
)

func TestTimelineViewRendersPosts(t *testing.T) {
	posts := []models.Post{
		{
			ID:        "1",
			Author:    &models.User{Username: "akram"},
			Content:   "Hello, Niotebook!",
			CreatedAt: time.Now().Add(-5 * time.Minute),
		},
	}
	m := views.NewTimelineModel(nil)
	m.SetPosts(posts)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := m.View()
	if !strings.Contains(view, "@akram") {
		t.Error("view missing username")
	}
	if !strings.Contains(view, "Hello, Niotebook!") {
		t.Error("view missing post content")
	}
}

func TestTimelineViewScrolling(t *testing.T) {
	posts := make([]models.Post, 10)
	for i := range posts {
		posts[i] = models.Post{
			ID:      fmt.Sprintf("%d", i),
			Author:  &models.User{Username: "user"},
			Content: fmt.Sprintf("Post %d", i),
		}
	}
	m := views.NewTimelineModel(nil)
	m.SetPosts(posts)

	// Initial cursor at 0
	if m.CursorIndex() != 0 {
		t.Errorf("initial cursor = %d, want 0", m.CursorIndex())
	}

	// Press j to move down
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	if m.CursorIndex() != 1 {
		t.Errorf("after j cursor = %d, want 1", m.CursorIndex())
	}
}

func TestTimelineViewEmptyState(t *testing.T) {
	m := views.NewTimelineModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	view := m.View()
	if !strings.Contains(view, "No posts yet") {
		t.Error("expected empty state message")
	}
}

func TestTimelineModelAPIError(t *testing.T) {
	m := views.NewTimelineModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	// Sending an error message should not panic
	_, cmd := m.Update(app.MsgAPIError{Message: "server error"})
	_ = cmd // Timeline may or may not produce a command from this
}

func TestTimelineSelectedPost(t *testing.T) {
	m := views.NewTimelineModel(nil)
	// Empty timeline should return nil
	if m.SelectedPost() != nil {
		t.Error("empty timeline should return nil SelectedPost")
	}

	// With posts, should return the current post
	posts := []models.Post{
		{ID: "1", Content: "First"},
		{ID: "2", Content: "Second"},
	}
	m.SetPosts(posts)
	selected := m.SelectedPost()
	if selected == nil {
		t.Fatal("expected non-nil SelectedPost")
	}
	if selected.ID != "1" {
		t.Errorf("SelectedPost ID = %q, want %q", selected.ID, "1")
	}
}

func TestTimelineHelpText(t *testing.T) {
	m := views.NewTimelineModel(nil)
	text := m.HelpText()
	if text == "" {
		t.Error("HelpText should return non-empty string")
	}
}

func TestTimelineKeyNavigation(t *testing.T) {
	posts := make([]models.Post, 5)
	for i := range posts {
		posts[i] = models.Post{
			ID:      fmt.Sprintf("%d", i),
			Author:  &models.User{Username: "user"},
			Content: fmt.Sprintf("Post %d", i),
		}
	}
	m := views.NewTimelineModel(nil)
	m.SetPosts(posts)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	// k at top should stay at 0
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'k'}})
	if m.CursorIndex() != 0 {
		t.Errorf("cursor = %d after k at top, want 0", m.CursorIndex())
	}

	// g jumps to top
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'j'}})
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}})
	if m.CursorIndex() != 0 {
		t.Errorf("cursor = %d after g, want 0", m.CursorIndex())
	}

	// G jumps to bottom
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'G'}})
	if m.CursorIndex() != 4 {
		t.Errorf("cursor = %d after G, want 4", m.CursorIndex())
	}

	// Down arrow navigation
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'g'}}) // back to top
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	if m.CursorIndex() != 1 {
		t.Errorf("cursor = %d after Down, want 1", m.CursorIndex())
	}

	// Up arrow navigation
	m, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	if m.CursorIndex() != 0 {
		t.Errorf("cursor = %d after Up, want 0", m.CursorIndex())
	}
}

func TestTimelineLoaded(t *testing.T) {
	m := views.NewTimelineModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	cursor := "next-page"
	m, _ = m.Update(app.MsgTimelineLoaded{
		Posts:      []models.Post{{ID: "1", Content: "Loaded"}},
		NextCursor: cursor,
		HasMore:    true,
	})

	if m.CursorIndex() != 0 {
		t.Errorf("cursor = %d after load, want 0", m.CursorIndex())
	}
	view := m.View()
	if !strings.Contains(view, "Loaded") {
		t.Error("view should contain loaded post")
	}
}

func TestTimelineRefreshed(t *testing.T) {
	m := views.NewTimelineModel(nil)
	m, _ = m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	m, _ = m.Update(app.MsgTimelineRefreshed{
		Posts:      []models.Post{{ID: "1", Content: "Refreshed"}},
		NextCursor: "",
		HasMore:    false,
	})

	view := m.View()
	if !strings.Contains(view, "Refreshed") {
		t.Error("view should contain refreshed post")
	}
}
