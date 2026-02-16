package theme

import "github.com/charmbracelet/lipgloss"

// Logo returns the full Niotebook brand logo: "n·otebook" where the middle dot
// (·) is rendered in the terracotta Accent color and all other characters use
// the primary Text color.
func Logo() string {
	text := lipgloss.NewStyle().Foreground(Text)
	dot := lipgloss.NewStyle().Foreground(Accent)

	return text.Render("n") + dot.Render("·") + text.Render("otebook")
}

// LogoCompact returns a compact variant of the brand logo suitable for headers
// and tight spaces. Currently identical to Logo.
func LogoCompact() string {
	return Logo()
}

// Tagline returns the brand tagline "a social notebook" styled in the Hint
// typography (italic, muted text).
func Tagline() string {
	return Hint.Render("a social notebook")
}
