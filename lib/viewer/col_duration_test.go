package viewer

import (
	"testing"
	"time"
)

func TestDurationColFormatting(t *testing.T) {
	tests := []struct {
		name     string
		seconds  float64
		length   int
		expected string
	}{
		{"zero", 0, 6, "0s"},
		{"1 second", 1, 6, "1s"},
		{"30 seconds", 30, 6, "30s"},
		{"1 minute", 60, 6, "1m"},
		{"90 seconds", 90, 6, "1m30s"},
		{"1 hour", 3600, 6, "1h"},
		{"1.5 hours", 5400, 6, "1h30m"},
		{"1 day", 86400, 6, "1d"},
		{"1 day 5 hours", 104400, 6, "1d5h"},
		{"1 week", 604800, 6, "1w"},
		{"1 week 2 days", 777600, 6, "1w2d"},
		{"500 milliseconds", 0.5, 6, "500ms"},
		{"100 microseconds", 0.0001, 6, "100µs"},
		{"small column 1h30m", 5400, 5, "1h30m"},
		{"large value in small column", 8640000, 5, "14w2d"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			col := DurationCol{
				defaultCol: defaultCol{
					Length: tt.length,
				},
			}
			result := col.formatDuration(tt.seconds)
			if result != tt.expected {
				t.Errorf("formatDuration(%v) with length %d = %q, want %q",
					tt.seconds, tt.length, result, tt.expected)
			}
		})
	}
}

func TestCompactDuration(t *testing.T) {
	tests := []struct {
		duration time.Duration
		length   int
		want     string
	}{
		{5 * time.Second, 6, "5s"},
		{61 * time.Second, 3, "1m"},
		{90 * time.Second, 3, "2m"},                                   // round up to 2m
		{2*time.Hour + 30*time.Minute + 1*time.Second, 10, "2h30m1s"}, // full size
		{2*time.Hour + 30*time.Minute + 1*time.Second, 6, "2h30m"},    // not enough room for seconds
		{2*time.Hour + 30*time.Minute + 1*time.Second, 3, "3h"},       // not enough room for seconds
		{25 * time.Hour, 6, "1d1h"},
		{8 * 24 * time.Hour, 6, "1w1d"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			col := DurationCol{
				defaultCol: defaultCol{
					Length: tt.length,
				},
			}
			result := col.compactDuration(tt.duration)
			if result != tt.want {
				t.Errorf("compactDuration(%v) with length %d = %q, want %q",
					tt.duration, tt.length, result, tt.want)
			}
		})
	}
}
