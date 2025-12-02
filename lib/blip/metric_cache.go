package blip

import (
	"fmt"

	"github.com/cashapp/blip"
)

// MetricCache stores current and previous metrics for easy lookup
type MetricCache struct {
	current   *blip.Metrics
	previous  *blip.Metrics
	index     map[string]map[string]*blip.MetricValue // domain -> name -> value
	prevIndex map[string]map[string]*blip.MetricValue // domain -> name -> value
}

// NewMetricCache creates a new metric cache
func NewMetricCache() *MetricCache {
	return &MetricCache{
		index:     make(map[string]map[string]*blip.MetricValue),
		prevIndex: make(map[string]map[string]*blip.MetricValue),
	}
}

// Update updates the cache with new metrics
func (mc *MetricCache) Update(metrics *blip.Metrics) {
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
func (mc *MetricCache) SecondsDiff() float64 {
	if mc.current == nil || mc.previous == nil {
		return 0
	}
	return mc.current.End.Sub(mc.previous.End).Seconds()
}

// GetTimeString returns a timestamp string for display
func (mc *MetricCache) GetTimeString() string {
	if mc.current == nil {
		return ""
	}
	return mc.current.Begin.Format("15:04:05")
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
	// Simple implementation: just check if pattern (without ^) is a prefix
	// Could be enhanced with proper regex
	if len(pattern) > 0 && pattern[0] == '^' {
		pattern = pattern[1:]
	}
	if len(pattern) > 0 && pattern[len(pattern)-1] == '*' {
		pattern = pattern[:len(pattern)-1]
		return len(name) >= len(pattern) && name[:len(pattern)] == pattern
	}
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
