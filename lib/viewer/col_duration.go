package viewer

import (
	"fmt"
	"math"
	"time"

	myblip "github.com/jayjanssen/myq-tools/lib/blip"
)

// DurationCol displays a time duration in human-readable format
// The metric value is expected to be in seconds (float64)
type DurationCol struct {
	defaultCol `yaml:",inline"`
	Key        SourceKey `yaml:"key"`
}

// A list of source keys that this column requires
func (c DurationCol) GetRequiredMetrics() []SourceKey {
	return []SourceKey{c.Key}
}

// Data for this view based on the metrics
func (c DurationCol) GetData(cache *myblip.MetricCache) []string {
	var str string

	// Try getting the metric value
	if metric, ok := cache.GetMetric(c.Key.Domain, c.Key.Metric); ok {
		str = c.formatDuration(metric.Value)
	} else {
		str = `-`
	}

	return []string{FitString(str, c.Length)}
}

// formatDuration converts seconds (float64) to a compact human-readable duration string
// that fits within the column length
func (c DurationCol) formatDuration(seconds float64) string {
	if seconds == 0 {
		return "0s"
	}

	// Handle negative values
	if seconds < 0 {
		return "-" + c.formatDuration(-seconds)
	}

	// Convert to time.Duration (nanoseconds)
	duration := time.Duration(seconds * float64(time.Second))

	// For very small durations (< 1 second), use milliseconds or microseconds
	if duration < time.Second {
		if duration < time.Microsecond {
			return fmt.Sprintf("%.0fns", float64(duration))
		} else if duration < time.Millisecond {
			return fmt.Sprintf("%.0fµs", float64(duration)/float64(time.Microsecond))
		} else {
			return fmt.Sprintf("%.0fms", float64(duration)/float64(time.Millisecond))
		}
	}

	// For durations >= 1 second, use a compact format
	return c.compactDuration(duration)
}

// compactDuration creates a compact duration string that fits in small columns
// Examples: "5s", "2m30s", "1h15m", "2d5h", "4w2d"
// Implements intelligent rounding: when units are dropped due to space constraints, rounds up if needed
func (c DurationCol) compactDuration(d time.Duration) string {
	if d == 0 {
		return "0s"
	}

	totalSeconds := int64(d / time.Second)
	minutes := totalSeconds / 60
	hours := minutes / 60
	days := hours / 24
	weeks := days / 7

	// Calculate remainders for each unit
	seconds := totalSeconds % 60
	minutes = minutes % 60
	hours = hours % 24
	days = days % 7

	// All time components in order from largest to smallest
	parts := []struct {
		value int64
		unit  string
	}{
		{weeks, "w"},
		{days, "d"},
		{hours, "h"},
		{minutes, "m"},
		{seconds, "s"},
	}

	// Find the first non-zero part (most significant unit)
	firstNonZero := -1
	for i, part := range parts {
		if part.value > 0 {
			firstNonZero = i
			break
		}
	}

	if firstNonZero == -1 {
		return "0s"
	}

	// Try to build a string with up to 3 units, then 2, then 1
	// For each unit count, first try without rounding, then with rounding if needed
	for maxUnits := 3; maxUnits >= 1; maxUnits-- {
		// Try without rounding first
		result, needsRounding := c.buildDurationString(parts, firstNonZero, maxUnits, false)

		// If there are units being dropped, also try with rounding
		if needsRounding {
			resultRounded, _ := c.buildDurationString(parts, firstNonZero, maxUnits, true)
			// Prefer the rounded version if it fits (even if non-rounded also fits)
			if len(resultRounded) <= c.Length {
				return resultRounded
			}
		}

		// Fall back to non-rounded if rounding didn't work or wasn't needed
		if len(result) <= c.Length {
			return result
		}
	}

	// If nothing fits, use the fallback single-unit formatter
	return c.fitDurationValue(parts[firstNonZero].value, parts[firstNonZero].unit)
}

// buildDurationString builds a duration string with up to maxUnits
// Returns the string and a flag indicating if there are remaining units that would require rounding
func (c DurationCol) buildDurationString(parts []struct {
	value int64
	unit  string
}, firstNonZero int, maxUnits int, applyRounding bool) (string, bool) {
	// Collect ALL non-zero parts first
	var allNonZeroParts []struct {
		value int64
		unit  string
		index int
	}

	for i := firstNonZero; i < len(parts); i++ {
		if parts[i].value > 0 {
			allNonZeroParts = append(allNonZeroParts, struct {
				value int64
				unit  string
				index int
			}{parts[i].value, parts[i].unit, i})
		}
	}

	if len(allNonZeroParts) == 0 {
		return "0s", false
	}

	// Select up to maxUnits from the non-zero parts
	selectedCount := maxUnits
	if selectedCount > len(allNonZeroParts) {
		selectedCount = len(allNonZeroParts)
	}
	selectedParts := allNonZeroParts[:selectedCount]

	// Check if we're dropping units (there are more non-zero parts than we're showing)
	hasRemainingUnits := len(allNonZeroParts) > selectedCount

	// Determine if we should round up when dropping units
	// Round up if the first dropped unit is >= half of its maximum value
	shouldRound := false
	if hasRemainingUnits && applyRounding {
		firstDropped := allNonZeroParts[selectedCount]
		// Determine the threshold for rounding based on the unit
		threshold := int64(30) // default for most units (half of 60)
		if firstDropped.unit == "d" {
			threshold = 12 // half of 24 hours
		} else if firstDropped.unit == "w" {
			threshold = 4 // half of 7 days (rounded up)
		}

		if firstDropped.value >= threshold {
			shouldRound = true
		}
	}

	// Build the result string
	result := ""
	for i, part := range selectedParts {
		value := part.value
		// Round up the last part if we should round
		if i == len(selectedParts)-1 && shouldRound {
			value++
		}
		result += fmt.Sprintf("%d%s", value, part.unit)
	}

	return result, hasRemainingUnits
}

// fitDurationValue formats a single duration value to fit in the column
// If it doesn't fit, it uses scientific notation or abbreviates
func (c DurationCol) fitDurationValue(value int64, unit string) string {
	str := fmt.Sprintf("%d%s", value, unit)
	if len(str) <= c.Length {
		return str
	}

	// Try with k/m/g suffix for large values
	if value >= 1000000 {
		v := float64(value) / 1000000.0
		str = fmt.Sprintf("%.0fm%s", v, unit)
	} else if value >= 1000 {
		v := float64(value) / 1000.0
		str = fmt.Sprintf("%.0fk%s", v, unit)
	}

	if len(str) <= c.Length {
		return str
	}

	// Last resort: scientific notation
	exp := int(math.Log10(float64(value)))
	mantissa := float64(value) / math.Pow10(exp)
	return fmt.Sprintf("%.0fe%d%s", mantissa, exp, unit)
}
