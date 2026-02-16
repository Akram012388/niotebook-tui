package components_test

import (
	"strings"
	"testing"
	"time"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/Akram012388/niotebook-tui/internal/tui/components"
)

func TestSidebarShowsLogo(t *testing.T) {
	user := &models.User{Username: "akram", DisplayName: "Akram"}
	result := components.RenderSidebar(user, components.ViewTimeline, false, 24, 30)
	if !strings.Contains(result, "otebook") {
		t.Error("sidebar should contain the niotebook logo")
	}
}

func TestSidebarShowsNavItems(t *testing.T) {
	user := &models.User{Username: "akram"}
	result := components.RenderSidebar(user, components.ViewTimeline, false, 24, 30)
	if !strings.Contains(result, "Home") {
		t.Error("sidebar should contain Home nav item")
	}
	if !strings.Contains(result, "Profile") {
		t.Error("sidebar should contain Profile nav item")
	}
	if !strings.Contains(result, "Bookmarks") {
		t.Error("sidebar should contain Bookmarks placeholder")
	}
	if !strings.Contains(result, "Settings") {
		t.Error("sidebar should contain Settings placeholder")
	}
}

func TestSidebarShowsPostButton(t *testing.T) {
	user := &models.User{Username: "akram"}
	result := components.RenderSidebar(user, components.ViewTimeline, false, 24, 30)
	if !strings.Contains(result, "Post") {
		t.Error("sidebar should contain Post button")
	}
}

func TestSidebarShowsShortcuts(t *testing.T) {
	user := &models.User{Username: "akram"}
	result := components.RenderSidebar(user, components.ViewTimeline, false, 24, 30)
	if !strings.Contains(result, "j/k") {
		t.Error("sidebar should contain j/k shortcut")
	}
	if !strings.Contains(result, "Tab") {
		t.Error("sidebar should contain Tab shortcut")
	}
}

func TestSidebarShowsJoinDate(t *testing.T) {
	user := &models.User{
		Username:  "akram",
		CreatedAt: time.Date(2026, 2, 1, 0, 0, 0, 0, time.UTC),
	}
	result := components.RenderSidebar(user, components.ViewTimeline, false, 24, 30)
	if !strings.Contains(result, "Joined Feb 2026") {
		t.Error("sidebar should contain join date")
	}
}

func TestSidebarActiveHighlight(t *testing.T) {
	user := &models.User{Username: "akram"}
	result := components.RenderSidebar(user, components.ViewTimeline, false, 24, 30)
	if !strings.Contains(result, "●") {
		t.Error("sidebar should show ● marker for active nav item")
	}
}

func TestSidebarZeroWidth(t *testing.T) {
	result := components.RenderSidebar(nil, components.ViewTimeline, false, 0, 20)
	if result != "" {
		t.Errorf("expected empty string for zero width, got len=%d", len(result))
	}
}

func TestSidebarNilUser(t *testing.T) {
	result := components.RenderSidebar(nil, components.ViewTimeline, false, 24, 30)
	if result == "" {
		t.Error("sidebar with nil user should still render logo")
	}
}
