// Package theme provides a centralized, adaptive color palette and typography
// styles for the Niotebook TUI. All colors use lipgloss.AdaptiveColor so they
// render appropriately in both dark and light terminal backgrounds.
package theme

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// ---------------------------------------------------------------------------
// Color Palette — adaptive light/dark
// ---------------------------------------------------------------------------

var (
	// Accent is the primary brand color (terracotta). Used for the logo dot,
	// selected items, and active navigation highlights.
	Accent = lipgloss.AdaptiveColor{Dark: "#D97757", Light: "#C15F3C"}

	// AccentDim is a muted variant of Accent. Used for active panel borders
	// and secondary accent elements.
	AccentDim = lipgloss.AdaptiveColor{Dark: "#B85C3A", Light: "#A04E30"}

	// Text is the primary foreground color for body text.
	Text = lipgloss.AdaptiveColor{Dark: "#FAFAF9", Light: "#141413"}

	// TextSecondary is used for timestamps, metadata, and counts.
	TextSecondary = lipgloss.AdaptiveColor{Dark: "#A8A29E", Light: "#57534E"}

	// TextMuted is used for hints, placeholders, and disabled text.
	TextMuted = lipgloss.AdaptiveColor{Dark: "#57534E", Light: "#A8A29E"}

	// Border is used for panel dividers and separators.
	Border = lipgloss.AdaptiveColor{Dark: "#44403C", Light: "#D6D3D1"}

	// Surface is the main background color.
	Surface = lipgloss.AdaptiveColor{Dark: "#141413", Light: "#FAFAF9"}

	// SurfaceRaised is used for sidebar backgrounds, cards, and modals.
	SurfaceRaised = lipgloss.AdaptiveColor{Dark: "#1C1917", Light: "#F5F5F4"}

	// Error is the color for error messages.
	Error = lipgloss.AdaptiveColor{Dark: "#EF4444", Light: "#DC2626"}

	// Success is the color for success confirmations.
	Success = lipgloss.AdaptiveColor{Dark: "#22C55E", Light: "#16A34A"}

	// Warning is the color for warning messages.
	Warning = lipgloss.AdaptiveColor{Dark: "#FBBF24", Light: "#D97706"}
)

// ---------------------------------------------------------------------------
// Typography Styles
// ---------------------------------------------------------------------------

var (
	// Heading is bold text in the Accent color, used for view titles and
	// section headers.
	Heading = lipgloss.NewStyle().
		Bold(true).
		Foreground(Accent)

	// Label is bold text in the primary Text color, used for form labels and
	// navigation items.
	Label = lipgloss.NewStyle().
		Bold(true).
		Foreground(Text)

	// Body is plain text in the primary Text color, used for post content and
	// descriptions.
	Body = lipgloss.NewStyle().
		Foreground(Text)

	// Caption is text in the TextSecondary color, used for timestamps, counts,
	// and metadata.
	Caption = lipgloss.NewStyle().
		Foreground(TextSecondary)

	// Hint is italic text in the TextMuted color, used for keyboard shortcuts
	// and placeholders.
	Hint = lipgloss.NewStyle().
		Italic(true).
		Foreground(TextMuted)

	// Key is bold text in the Accent color, used for keyboard key references
	// in help views.
	Key = lipgloss.NewStyle().
		Bold(true).
		Foreground(Accent)
)

// ---------------------------------------------------------------------------
// Border Styles
// ---------------------------------------------------------------------------

var (
	// PanelBorder is a rounded border in the Border color, used for column
	// dividers.
	PanelBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Border)

	// ActiveBorder is a rounded border in the AccentDim color, used for
	// active or focused panels.
	ActiveBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(AccentDim)

	// ModalBorder is a rounded border in the Accent color with padding,
	// used for compose and help overlays.
	ModalBorder = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(Accent).
			Padding(1, 2)
)

// ---------------------------------------------------------------------------
// Separator
// ---------------------------------------------------------------------------

// Separator returns a horizontal dashed line of the given width, rendered in
// the Border color. It is typically used as a visual divider between sections.
func Separator(width int) string {
	if width <= 0 {
		return ""
	}
	line := strings.Repeat("─", width)
	return lipgloss.NewStyle().Foreground(Border).Render(line)
}
