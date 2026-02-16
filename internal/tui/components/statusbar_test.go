package components_test

import (
	"strings"
	"testing"

	"github.com/Akram012388/niotebook-tui/internal/tui/components"
)

func TestStatusBarSetError(t *testing.T) {
	sb := components.NewStatusBarModel()
	cmd := sb.SetError("something went wrong")
	if cmd == nil {
		t.Error("expected auto-clear command")
	}
	result := sb.View("help text", 80)
	if !strings.Contains(result, "something went wrong") {
		t.Error("expected error message in status bar")
	}
}

func TestStatusBarSetSuccess(t *testing.T) {
	sb := components.NewStatusBarModel()
	cmd := sb.SetSuccess("Post published!")
	if cmd == nil {
		t.Error("expected auto-clear command")
	}
	result := sb.View("help text", 80)
	if !strings.Contains(result, "Post published!") {
		t.Error("expected success message in status bar")
	}
}

func TestStatusBarClear(t *testing.T) {
	sb := components.NewStatusBarModel()
	sb.SetError("error")
	sb.Clear()
	result := sb.View("help text", 80)
	if strings.Contains(result, "error") {
		t.Error("expected error to be cleared")
	}
}

func TestStatusBarDefaultShowsHelpText(t *testing.T) {
	sb := components.NewStatusBarModel()
	result := sb.View("n: new post  ?: help  q: quit", 80)
	if !strings.Contains(result, "n: new post") {
		t.Error("expected help text in default status bar")
	}
}
