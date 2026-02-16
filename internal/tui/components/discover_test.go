package components_test

import (
	"strings"
	"testing"

	"github.com/Akram012388/niotebook-tui/internal/tui/components"
)

func TestDiscoverShowsSearchPlaceholder(t *testing.T) {
	result := components.RenderDiscover(false, nil, 24, 30)
	if !strings.Contains(result, "Search") {
		t.Error("discover should contain search placeholder")
	}
}

func TestDiscoverShowsTrending(t *testing.T) {
	result := components.RenderDiscover(false, nil, 24, 30)
	if !strings.Contains(result, "Trending") {
		t.Error("discover should contain Trending section")
	}
	if !strings.Contains(result, "#niotebook") {
		t.Error("discover should contain #niotebook trending tag")
	}
}

func TestDiscoverShowsWritersToFollow(t *testing.T) {
	result := components.RenderDiscover(false, nil, 24, 30)
	if !strings.Contains(result, "Writers to follow") {
		t.Error("discover should contain Writers to follow section")
	}
	if !strings.Contains(result, "@alice") {
		t.Error("discover should contain @alice suggestion")
	}
}

func TestDiscoverZeroWidth(t *testing.T) {
	result := components.RenderDiscover(false, nil, 0, 20)
	if result != "" {
		t.Errorf("expected empty for zero width, got len=%d", len(result))
	}
}
