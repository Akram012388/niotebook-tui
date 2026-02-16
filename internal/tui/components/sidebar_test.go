package components_test

import (
	"strings"
	"testing"
	"time"

	"github.com/Akram012388/niotebook-tui/internal/models"
	"github.com/Akram012388/niotebook-tui/internal/tui/components"
)

func TestSidebarRendersUsername(t *testing.T) {
	user := &models.User{
		ID:        "123",
		Username:  "akram",
		CreatedAt: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
	}
	result := components.RenderSidebar(user, components.ViewTimeline, 30, 20)
	if !strings.Contains(result, "akram") {
		t.Error("expected username 'akram' in sidebar output")
	}
}

func TestSidebarRendersNavItems(t *testing.T) {
	user := &models.User{
		ID:        "123",
		Username:  "testuser",
		CreatedAt: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
	}
	result := components.RenderSidebar(user, components.ViewTimeline, 30, 20)
	if !strings.Contains(result, "Home") {
		t.Error("expected 'Home' nav item in sidebar output")
	}
	if !strings.Contains(result, "Profile") {
		t.Error("expected 'Profile' nav item in sidebar output")
	}
}

func TestSidebarActiveIndicator(t *testing.T) {
	user := &models.User{
		ID:        "123",
		Username:  "testuser",
		CreatedAt: time.Date(2025, 1, 15, 0, 0, 0, 0, time.UTC),
	}
	result := components.RenderSidebar(user, components.ViewTimeline, 30, 20)
	if !strings.Contains(result, "●") {
		t.Error("expected active indicator '●' when viewing timeline")
	}
}

func TestSidebarLoggedOut(t *testing.T) {
	result := components.RenderSidebar(nil, components.ViewLogin, 30, 20)
	if strings.Contains(result, "@") {
		t.Error("expected no '@' prefix when logged out (nil user)")
	}
}

func TestSidebarZeroWidth(t *testing.T) {
	user := &models.User{
		ID:       "123",
		Username: "testuser",
	}
	result := components.RenderSidebar(user, components.ViewTimeline, 0, 20)
	if result != "" {
		t.Errorf("expected empty string for zero width, got %q", result)
	}
}
