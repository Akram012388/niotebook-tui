package components_test

import (
	"strings"
	"testing"
	"time"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/Akram012388/niotebook-tui/internal/tui/components"
)

func TestRenderPostCardWithAuthor(t *testing.T) {
	post := models.Post{
		Content:   "Hello, world!",
		Author:    &models.User{Username: "akram"},
		CreatedAt: time.Now().Add(-5 * time.Minute),
	}
	result := components.RenderPostCard(post, 80, false, time.Now())
	if !strings.Contains(result, "@akram") {
		t.Error("expected @akram in output")
	}
	if !strings.Contains(result, "Hello, world!") {
		t.Error("expected post content in output")
	}
}

func TestRenderPostCardSelected(t *testing.T) {
	post := models.Post{
		Content: "Test post",
		Author:  &models.User{Username: "test"},
	}
	result := components.RenderPostCard(post, 80, true, time.Now())
	if !strings.Contains(result, "Test post") {
		t.Error("expected post content in selected card")
	}
}

func TestRenderPostCardNilAuthor(t *testing.T) {
	post := models.Post{Content: "Orphan post", Author: nil}
	result := components.RenderPostCard(post, 80, false, time.Now())
	if !strings.Contains(result, "@unknown") {
		t.Error("expected @unknown for nil author")
	}
}

func TestRenderPostCardNarrowWidth(t *testing.T) {
	post := models.Post{
		Content: "This is a longer post that should wrap at narrow widths",
		Author:  &models.User{Username: "user"},
	}
	result := components.RenderPostCard(post, 20, false, time.Now())
	if result == "" {
		t.Error("expected non-empty output for narrow width")
	}
}
