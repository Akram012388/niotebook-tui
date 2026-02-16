package components_test

import (
	"strings"
	"testing"

	"github.com/Akram012388/niotebook-tui/internal/tui/components"
)

func TestShortcutsTimeline(t *testing.T) {
	result := components.RenderShortcuts(components.ViewTimeline, false, 30, 20)
	if !strings.Contains(result, "j/k") {
		t.Error("expected 'j/k' shortcut in timeline shortcuts")
	}
	if !strings.Contains(result, "compose") {
		t.Error("expected 'compose' action in timeline shortcuts")
	}
}

func TestShortcutsProfile(t *testing.T) {
	result := components.RenderShortcuts(components.ViewProfile, false, 30, 20)
	if !strings.Contains(result, "scroll") {
		t.Error("expected 'scroll' description in profile shortcuts")
	}
}

func TestShortcutsCompose(t *testing.T) {
	result := components.RenderShortcuts(components.ViewTimeline, true, 30, 20)
	if !strings.Contains(result, "Ctrl+J") {
		t.Error("expected 'Ctrl+J' shortcut when compose is open")
	}
}

func TestShortcutsZeroWidth(t *testing.T) {
	result := components.RenderShortcuts(components.ViewTimeline, false, 0, 20)
	if result != "" {
		t.Errorf("expected empty string for zero width, got %q", result)
	}
}
