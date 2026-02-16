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
