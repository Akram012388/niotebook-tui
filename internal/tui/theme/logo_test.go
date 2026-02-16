package theme

import (
	"strings"
	"testing"
)

func TestLogoContainsExpectedParts(t *testing.T) {
	logo := Logo()
	if logo == "" {
		t.Fatal("Logo() returned empty string")
	}
	if !strings.Contains(logo, "n") {
		t.Error("Logo() does not contain 'n'")
	}
	if !strings.Contains(logo, "otebook") {
		t.Error("Logo() does not contain 'otebook'")
	}
}

func TestLogoCompactReturnsNonEmpty(t *testing.T) {
	result := LogoCompact()
	if result == "" {
		t.Error("LogoCompact() returned empty string")
	}
}

func TestTaglineContainsSocialNotebook(t *testing.T) {
	result := Tagline()
	if result == "" {
		t.Fatal("Tagline() returned empty string")
	}
	if !strings.Contains(result, "social notebook") {
		t.Errorf("Tagline() should contain 'social notebook', got %q", result)
	}
}
