package theme

import "github.com/charmbracelet/lipgloss"

// Logo returns the full Niotebook brand logo: "niotebook" where the letter 'i'
// is rendered in the terracotta Accent color and all other characters use the
// primary Text color.
func Logo() string {
	text := lipgloss.NewStyle().Foreground(Text)
	accent := lipgloss.NewStyle().Foreground(Accent)

	return text.Render("n") + accent.Render("i") + text.Render("otebook")
}

// LogoCompact returns a compact variant of the brand logo suitable for sidebars
// and tight spaces. Currently identical to Logo.
func LogoCompact() string {
	return Logo()
}

// Tagline returns the brand tagline "the social terminal" styled in the Hint
// typography (italic, muted text).
func Tagline() string {
	return Hint.Render("the social terminal")
}
