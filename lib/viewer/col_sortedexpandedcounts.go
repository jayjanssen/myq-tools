package viewer

import (
	"fmt"
	"sort"
	"strings"

	"github.com/jayjanssen/myq-tools/lib/blip"
)

type SortedExpandedCountsCol struct {
	colNum `yaml:",inline"`
	Keys   []SourceKey `yaml:"keys"`
}

// A list of source keys that this column requires
func (secc SortedExpandedCountsCol) GetRequiredMetrics() []SourceKey {
	return secc.Keys
}

func (secc SortedExpandedCountsCol) GetData(cache *blip.MetricCache) (output []string) {
	// For each key, find matching metrics using pattern
	// Track metrics with their domains for proper prev metric lookup
	type metricWithDomain struct {
		metric blip.MetricValue
		domain string
	}
	var allMetrics []metricWithDomain

	for _, key := range secc.Keys {
		// Check if this is a pattern (ends with *)
		pattern := key.Metric
		if strings.HasSuffix(pattern, "*") || strings.HasPrefix(pattern, "^") {
			// Find all metrics matching the pattern
			metrics := cache.FindMetrics(key.Domain, pattern)
			for _, metric := range metrics {
				allMetrics = append(allMetrics, metricWithDomain{metric: metric, domain: key.Domain})
			}
		} else {
			// Single metric
			if metric, ok := cache.GetMetric(key.Domain, key.Metric); ok {
				allMetrics = append(allMetrics, metricWithDomain{metric: metric, domain: key.Domain})
			}
		}
	}

	if len(allMetrics) == 0 {
		return []string{}
	}

	// Calculate diffs for each metric
	var totalDiff float64
	var allDiffs []float64
	diffVariables := map[float64][]string{}

	for _, mwd := range allMetrics {
		metric := mwd.metric
		curr := metric.Value
		var prev float64
		if prevMetric, ok := cache.GetPrevMetric(mwd.domain, metric.Name); ok {
			prev = prevMetric.Value
		}

		diff := calculateDiff(curr, prev)
		// Skip those with no activity
		if diff <= 0 {
			continue
		}
		totalDiff += diff

		// Create the [] slice for a diff we haven't seen yet
		if _, ok := diffVariables[diff]; !ok {
			diffVariables[diff] = make([]string, 0)
			allDiffs = append(allDiffs, diff)
		}

		// Push the variable name onto the diff slice
		diffVariables[diff] = append(diffVariables[diff], metric.Name)
	}

	// Output the total diff
	numStr := FitString(secc.fitNumber(totalDiff, 0), secc.Length)
	line := fmt.Sprintf("%s %v", numStr, "total")
	output = append(output, line)

	// Sort all the diffs so we can iterate through them from big to small
	sort.Sort(sort.Reverse(sort.Float64Slice(allDiffs)))

	for _, diff := range allDiffs {
		// Sort metric names alphabetically for consistent output when counts are equal
		metricNames := diffVariables[diff]
		sort.Strings(metricNames)
		numStr := FitString(secc.fitNumber(diff, 0), secc.Length)
		line := fmt.Sprintf("%s %v", numStr, metricNames)
		output = append(output, line)
	}
	return
}
