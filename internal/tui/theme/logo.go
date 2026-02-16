package theme

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

// Logo returns the Niotebook brand wordmark: "niotebook" in bold,
// with the letter 'i' in the terracotta Accent color.
func Logo() string {
	text := lipgloss.NewStyle().Bold(true).Foreground(Text)
	accent := lipgloss.NewStyle().Bold(true).Foreground(Accent)
	return text.Render("n") + accent.Render("i") + text.Render("otebook")
}

// LogoCompact returns a compact variant of the brand logo.
func LogoCompact() string {
	return Logo()
}

// LogoSplash returns the splash screen variant with letter-spacing:
// "n i o t e b o o k" — bold, 'i' in Accent.
func LogoSplash() string {
	text := lipgloss.NewStyle().Bold(true).Foreground(Text)
	accent := lipgloss.NewStyle().Bold(true).Foreground(Accent)

	letters := []struct {
		char  string
		style lipgloss.Style
	}{
		{"n", text}, {"i", accent}, {"o", text}, {"t", text},
		{"e", text}, {"b", text}, {"o", text}, {"o", text}, {"k", text},
	}

	parts := make([]string, len(letters))
	for i, l := range letters {
		parts[i] = l.style.Render(l.char)
	}
	return strings.Join(parts, " ")
}

// Tagline returns the brand tagline in Hint style.
func Tagline() string {
	return Hint.Render("the social terminal")
}

// TaglineSplash returns the splash screen variant of the tagline with
// letter-spacing. Styled in TextMuted.
func TaglineSplash() string {
	style := lipgloss.NewStyle().Foreground(TextMuted)
	chars := []rune("the social terminal")
	parts := make([]string, len(chars))
	for i, ch := range chars {
		if ch == ' ' {
			parts[i] = " "
		} else {
			parts[i] = style.Render(string(ch))
		}
	}
	return strings.Join(parts, " ")
}

// LogoASCII returns a multi-line ASCII art rendering of "niotebook".
// The dot on the letter 'i' is rendered in Accent color.
// Designed to fit within a ~20 character wide sidebar column.
func LogoASCII(_ int) string {
	bold := lipgloss.NewStyle().Bold(true).Foreground(Text)
	dot := lipgloss.NewStyle().Bold(true).Foreground(Accent)

	// 3-line compact art: accent dot above, bold wordmark, separator
	line1 := bold.Render("  ") + dot.Render("•")
	line2 := bold.Render("n") + bold.Render("i") + bold.Render("otebook")
	line3 := lipgloss.NewStyle().Foreground(Accent).Render("━━━━━━━━━")

	return line1 + "\n" + line2 + "\n" + line3
}

// LogoSplashASCII returns the splash screen ASCII art variant.
func LogoSplashASCII() string {
	return LogoASCII(40)
}
