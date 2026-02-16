package theme

import (
	"testing"

	"github.com/charmbracelet/lipgloss"
)

func TestColorTokensHaveNonEmptyValues(t *testing.T) {
	colors := []struct {
		name  string
		color lipgloss.AdaptiveColor
	}{
		{"Accent", Accent},
		{"AccentDim", AccentDim},
		{"Text", Text},
		{"TextSecondary", TextSecondary},
		{"TextMuted", TextMuted},
		{"Border", Border},
		{"Surface", Surface},
		{"SurfaceRaised", SurfaceRaised},
		{"Error", Error},
		{"Success", Success},
		{"Warning", Warning},
	}

	for _, tc := range colors {
		t.Run(tc.name, func(t *testing.T) {
			if tc.color.Dark == "" {
				t.Errorf("%s: Dark value is empty", tc.name)
			}
			if tc.color.Light == "" {
				t.Errorf("%s: Light value is empty", tc.name)
			}
		})
	}
}

func TestTypographyStylesRenderNonEmpty(t *testing.T) {
	styles := []struct {
		name  string
		style lipgloss.Style
	}{
		{"Heading", Heading},
		{"Label", Label},
		{"Body", Body},
		{"Caption", Caption},
		{"Hint", Hint},
		{"Key", Key},
	}

	for _, tc := range styles {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.style.Render("test")
			if result == "" {
				t.Errorf("%s.Render(\"test\") returned empty string", tc.name)
			}
		})
	}
}

func TestBorderStylesRenderNonEmpty(t *testing.T) {
	styles := []struct {
		name  string
		style lipgloss.Style
	}{
		{"PanelBorder", PanelBorder},
		{"ActiveBorder", ActiveBorder},
		{"ModalBorder", ModalBorder},
	}

	for _, tc := range styles {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.style.Render("test")
			if result == "" {
				t.Errorf("%s.Render(\"test\") returned empty string", tc.name)
			}
		})
	}
}

func TestSeparatorReturnsNonEmpty(t *testing.T) {
	result := Separator(40)
	if result == "" {
		t.Error("Separator(40) returned empty string")
	}
}

func TestSeparatorZeroWidth(t *testing.T) {
	result := Separator(0)
	if result != "" {
		t.Errorf("Separator(0) should return empty string, got %q", result)
	}
}

func TestSeparatorNegativeWidth(t *testing.T) {
	result := Separator(-5)
	if result != "" {
		t.Errorf("Separator(-5) should return empty string, got %q", result)
	}
}
