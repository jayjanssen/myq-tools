package viewer

import (
	"fmt"

	"github.com/jayjanssen/myq-tools/lib/blip"
)

type SubtractCol struct {
	colNum  `yaml:",inline"`
	Bigger  SourceKey `yaml:"bigger"`
	Smaller SourceKey `yaml:"smaller"`
}

// A list of source keys that this column requires
func (c SubtractCol) GetRequiredMetrics() []SourceKey {
	return []SourceKey{c.Bigger, c.Smaller}
}

// Data for this view based on the metrics
func (c SubtractCol) GetData(cache *blip.MetricCache) []string {
	var str string
	raw, err := c.getSubtract(cache)
	if err != nil {
		str = FitString(`-`, c.Length)
	} else {
		num := c.fitNumber(raw, c.Precision)
		str = FitString(num, c.Length) // adds padding if needed
	}
	return []string{str}
}

// Calculates the subtraction for the given MetricCache, returns an error if there's a data problem.
func (c SubtractCol) getSubtract(cache *blip.MetricCache) (float64, error) {
	// Get bigger value - must exist
	biggerMetric, ok := cache.GetMetric(c.Bigger.Domain, c.Bigger.Metric)
	if !ok {
		return 0, fmt.Errorf("metric not found: %s/%s", c.Bigger.Domain, c.Bigger.Metric)
	}

	// Get smaller value - must exist
	smallerMetric, ok := cache.GetMetric(c.Smaller.Domain, c.Smaller.Metric)
	if !ok {
		return 0, fmt.Errorf("metric not found: %s/%s", c.Smaller.Domain, c.Smaller.Metric)
	}

	// Return the calculated subtraction
	return (biggerMetric.Value - smallerMetric.Value), nil
}
