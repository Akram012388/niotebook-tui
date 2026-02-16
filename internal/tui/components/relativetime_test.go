package components_test

import (
	"testing"
	"time"

	"github.com/Akram012388/niotebook-tui/internal/tui/components"
)

func TestRelativeTime(t *testing.T) {
	now := time.Date(2026, 2, 16, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		t    time.Time
		want string
	}{
		{"just now", now.Add(-30 * time.Second), "now"},
		{"1 minute", now.Add(-1 * time.Minute), "1m"},
		{"5 minutes", now.Add(-5 * time.Minute), "5m"},
		{"59 minutes", now.Add(-59 * time.Minute), "59m"},
		{"1 hour", now.Add(-1 * time.Hour), "1h"},
		{"23 hours", now.Add(-23 * time.Hour), "23h"},
		{"1 day", now.Add(-24 * time.Hour), "1d"},
		{"6 days", now.Add(-6 * 24 * time.Hour), "6d"},
		{"1 week", now.Add(-7 * 24 * time.Hour), "1w"},
		{"3 weeks", now.Add(-21 * 24 * time.Hour), "3w"},
		{"31 days same year", now.Add(-31 * 24 * time.Hour), "Jan 16"},
		{"different year", time.Date(2025, 6, 15, 0, 0, 0, 0, time.UTC), "Jun 15, 2025"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := components.RelativeTimeFrom(tt.t, now)
			if got != tt.want {
				t.Errorf("RelativeTimeFrom() = %q, want %q", got, tt.want)
			}
		})
	}
}
