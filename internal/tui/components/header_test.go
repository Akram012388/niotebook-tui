package components_test

import (
	"strings"
	"testing"

	"github.com/Akram012388/niotebook-tui/internal/tui/components"
)

func TestRenderHeaderShowsAppName(t *testing.T) {
	result := components.RenderHeader("niotebook", "akram", "Timeline", 80)
	if !strings.Contains(result, "niotebook") {
		t.Error("expected app name in header")
	}
}

func TestRenderHeaderShowsUsername(t *testing.T) {
	result := components.RenderHeader("niotebook", "akram", "Timeline", 80)
	if !strings.Contains(result, "@akram") {
		t.Error("expected @akram in header")
	}
}

func TestRenderHeaderEmptyUsername(t *testing.T) {
	result := components.RenderHeader("niotebook", "", "Timeline", 80)
	if !strings.Contains(result, "niotebook") {
		t.Error("expected app name in header even with empty username")
	}
}

func TestRenderHeaderNarrowWidth(t *testing.T) {
	result := components.RenderHeader("niotebook", "akram", "Timeline", 10)
	if result == "" {
		t.Error("expected non-empty output for narrow width")
	}
}
