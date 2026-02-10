package blip

import (
	"fmt"
	"os"
	"time"

	"github.com/cashapp/blip"
)

// DebugCache enables debug logging for MetricCache
var DebugCache bool

// MetricCache stores current and previous metrics for easy lookup
type MetricCache struct {
	current       *blip.Metrics
	previous      *blip.Metrics
	index         map[string]map[string]*blip.MetricValue // domain -> name -> value
	prevIndex     map[string]map[string]*blip.MetricValue // domain -> name -> value
	isLiveMode    bool                                    // true for live MySQL connection, false for file replay
	firstUptime   int64                                   // uptime of first sample (for calculating relative time in file mode)
	uptimeTracked bool                                    // whether we've captured the first uptime yet
}

// NewMetricCache creates a new metric cache
func NewMetricCache(isLiveMode bool) *MetricCache {
	return &MetricCache{
		index:      make(map[string]map[string]*blip.MetricValue),
		prevIndex:  make(map[string]map[string]*blip.MetricValue),
		isLiveMode: isLiveMode,
	}
}

// Update updates the cache with new metrics
func (mc *MetricCache) Update(metrics *blip.Metrics) {
	// DEBUG: Log cache update
	if DebugCache {
		var prevEnd, curEnd string
		if mc.current != nil {
			prevEnd = mc.current.End.Format("15:04:05.000")
		} else {
			prevEnd = "nil"
		}
		if metrics != nil {
			curEnd = metrics.End.Format("15:04:05.000")
		} else {
			curEnd = "nil"
		}
		interval := uint(0)
		if metrics != nil {
			interval = metrics.Interval
		}
		fmt.Fprintf(os.Stderr, "DEBUG [Cache] Update: prev.End=%s -> cur.End=%s interval=%d\n",
			prevEnd, curEnd, interval)
	}

	// Shift current to previous
	mc.previous = mc.current
	mc.prevIndex = mc.index

	// Set new current
	mc.current = metrics
	mc.index = make(map[string]map[string]*blip.MetricValue)

	// Build index for fast lookups
	if metrics != nil && metrics.Values != nil {
		for domain, values := range metrics.Values {
			mc.index[domain] = make(map[string]*blip.MetricValue)
			for i := range values {
				mv := &values[i]
				mc.index[domain][mv.Name] = mv
			}
		}
	}

	// Track the first uptime for file mode duration calculation
	if !mc.isLiveMode && mc.HasCurrent() && !mc.uptimeTracked {
		mc.firstUptime = mc.GetUptime()
		mc.uptimeTracked = true
	}
}

// GetMetric returns the current metric value for a given domain and name
func (mc *MetricCache) GetMetric(domain, name string) (blip.MetricValue, bool) {
	if domainMap, ok := mc.index[domain]; ok {
		if mv, ok := domainMap[name]; ok {
			return *mv, true
		}
	}
	return blip.MetricValue{}, false
}

// GetPrevMetric returns the previous metric value for a given domain and name
func (mc *MetricCache) GetPrevMetric(domain, name string) (blip.MetricValue, bool) {
	if domainMap, ok := mc.prevIndex[domain]; ok {
		if mv, ok := domainMap[name]; ok {
			return *mv, true
		}
	}
	return blip.MetricValue{}, false
}

// GetMetricValue returns just the float64 value (convenience method)
func (mc *MetricCache) GetMetricValue(domain, name string) float64 {
	if mv, ok := mc.GetMetric(domain, name); ok {
		return mv.Value
	}
	return 0
}

// GetPrevMetricValue returns just the previous float64 value (convenience method)
func (mc *MetricCache) GetPrevMetricValue(domain, name string) float64 {
	if mv, ok := mc.GetPrevMetric(domain, name); ok {
		return mv.Value
	}
	return 0
}

// SecondsDiff returns the time difference between current and previous samples
// For file replay mode, uses uptime differences (more accurate when samples are irregular)
// For live mode, uses timestamp differences
func (mc *MetricCache) SecondsDiff() float64 {
	if mc.current == nil || mc.previous == nil {
		return 0
	}

	if mc.isLiveMode {
		// Live mode: use timestamp differences
		diff := mc.current.End.Sub(mc.previous.End).Seconds()
		if diff <= 0 {
			return 0
		}
		return diff
	} else {
		// File mode: use uptime differences for accurate rate calculations
		// This handles cases where actual sample intervals differ from configured interval
		currentUptime := mc.GetUptime()
		prevUptime := int64(0)
		if mv, ok := mc.GetPrevMetric("status.global", "uptime"); ok {
			prevUptime = int64(mv.Value)
		}
		uptimeDiff := currentUptime - prevUptime
		if uptimeDiff <= 0 {
			// Fallback to timestamp difference if uptime is missing or invalid
			return mc.current.End.Sub(mc.previous.End).Seconds()
		}
		return float64(uptimeDiff)
	}
}

// GetTimeString returns a timestamp string for display
func (mc *MetricCache) GetTimeString() string {
	if mc.current == nil {
		return ""
	}

	if mc.isLiveMode {
		// Live mode: show current time as HH:MM:SS
		return mc.current.Begin.Format("15:04:05")
	} else {
		// File mode: show duration based on uptime difference from first sample
		currentUptime := mc.GetUptime()

		// Handle case where uptime is missing in current sample
		if currentUptime == 0 {
			// If we had uptime before (firstUptime > 0), this means uptime is now missing
			// Fallback to timestamp-based display to avoid negative duration
			if mc.firstUptime > 0 && mc.previous != nil {
				// Use timestamp difference from previous sample
				elapsedDuration := mc.current.End.Sub(mc.previous.End)
				return elapsedDuration.String()
			}
			// If uptime was never available (firstUptime == 0), return "0s"
			return "0s"
		}

		elapsedSeconds := currentUptime - mc.firstUptime

		// Handle case where elapsedSeconds is negative (shouldn't happen with valid uptime)
		if elapsedSeconds < 0 {
			// Fallback to timestamp-based display if we have a previous sample
			if mc.previous != nil {
				elapsedDuration := mc.current.End.Sub(mc.previous.End)
				return elapsedDuration.String()
			}
			return "0s"
		}

		duration := time.Duration(elapsedSeconds) * time.Second
		return duration.String()
	}
}

// HasCurrent returns true if there's current data
func (mc *MetricCache) HasCurrent() bool {
	return mc.current != nil
}

// HasPrevious returns true if there's previous data
func (mc *MetricCache) HasPrevious() bool {
	return mc.previous != nil
}

// GetUptime returns the uptime from the current metrics
func (mc *MetricCache) GetUptime() int64 {
	// Try to get uptime from status.global
	if mv, ok := mc.GetMetric("status.global", "uptime"); ok {
		return int64(mv.Value)
	}
	return 0
}

// FindMetrics returns all metrics matching a pattern in a domain
func (mc *MetricCache) FindMetrics(domain, pattern string) []blip.MetricValue {
	results := []blip.MetricValue{}
	if domainMap, ok := mc.index[domain]; ok {
		for name, mv := range domainMap {
			// Simple pattern matching - could be enhanced
			if matchPattern(name, pattern) {
				results = append(results, *mv)
			}
		}
	}
	return results
}

// matchPattern does simple glob-style pattern matching
func matchPattern(name, pattern string) bool {
	// Handle caret - treat as "starts with"
	hasCaret := len(pattern) > 0 && pattern[0] == '^'
	if hasCaret {
		pattern = pattern[1:]
	}

	// Handle wildcard suffix
	hasWildcard := len(pattern) > 0 && pattern[len(pattern)-1] == '*'
	if hasWildcard {
		pattern = pattern[:len(pattern)-1]
	}

	// If we have a prefix pattern (from caret or wildcard), check prefix match
	if hasCaret || hasWildcard {
		return len(name) >= len(pattern) && name[:len(pattern)] == pattern
	}

	// Exact match
	return name == pattern
}

// DomainExists checks if a domain has any metrics
func (mc *MetricCache) DomainExists(domain string) bool {
	_, ok := mc.index[domain]
	return ok
}

// GetAllDomains returns all domains in the current cache
func (mc *MetricCache) GetAllDomains() []string {
	domains := make([]string, 0, len(mc.index))
	for domain := range mc.index {
		domains = append(domains, domain)
	}
	return domains
}

// Debug prints cache contents (for debugging)
func (mc *MetricCache) Debug() string {
	if mc.current == nil {
		return "MetricCache: no current metrics"
	}
	return fmt.Sprintf("MetricCache: %d domains, time=%s, interval=%d",
		len(mc.index), mc.GetTimeString(), mc.current.Interval)
}
