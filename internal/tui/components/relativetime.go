package components

import (
	"fmt"
	"time"
)

// RelativeTimeFrom returns a human-readable relative time string
// comparing t to now, following the display rules:
//
//	< 1 minute  → "now"
//	1-59 min    → "Xm"
//	1-23 hours  → "Xh"
//	1-6 days    → "Xd"
//	7-29 days   → "Xw"
//	30+ days    → "Jan 15"
//	diff year   → "Jan 15, 2025"
func RelativeTimeFrom(t, now time.Time) string {
	d := now.Sub(t)

	switch {
	case d < time.Minute:
		return "now"
	case d < time.Hour:
		return fmt.Sprintf("%dm", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("%dh", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("%dd", int(d.Hours()/24))
	case d < 30*24*time.Hour:
		return fmt.Sprintf("%dw", int(d.Hours()/(24*7)))
	default:
		if t.Year() != now.Year() {
			return t.Format("Jan 2, 2006")
		}
		return t.Format("Jan 2")
	}
}
