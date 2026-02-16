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
	// The logo renders "niotebook" with ANSI styling on the 'i'. Check raw
	// characters are present (they will be interspersed with escape codes).
	if !strings.Contains(logo, "n") {
		t.Error("Logo() does not contain 'n'")
	}
	if !strings.Contains(logo, "i") {
		t.Error("Logo() does not contain 'i'")
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

func TestTaglineContainsSocialTerminal(t *testing.T) {
	result := Tagline()
	if result == "" {
		t.Fatal("Tagline() returned empty string")
	}
	if !strings.Contains(result, "the social terminal") {
		t.Errorf("Tagline() should contain 'the social terminal', got %q", result)
	}
}

func TestLogoSplashReturnsSpacedLetters(t *testing.T) {
	result := LogoSplash()
	if result == "" {
		t.Fatal("LogoSplash() returned empty string")
	}
	// Should contain individual characters (with ANSI codes between them)
	if !strings.Contains(result, "o") {
		t.Error("LogoSplash() should contain letter 'o'")
	}
}

func TestTaglineSplashReturnsSpacedText(t *testing.T) {
	result := TaglineSplash()
	if result == "" {
		t.Fatal("TaglineSplash() returned empty string")
	}
}
