package viewer

import (
	"github.com/jayjanssen/myq-tools/lib/blip"
)

type DiffCol struct {
	colNum `yaml:",inline"`
	Key    SourceKey `yaml:"key"`
}

// A list of source keys that this column requires
func (c DiffCol) GetRequiredMetrics() []SourceKey {
	return []SourceKey{c.Key}
}

// Data for this view based on the metrics
func (c DiffCol) GetData(cache *blip.MetricCache) []string {
	var str string
	raw, err := c.getDiff(cache)
	if err != nil {
		str = FitString(`-`, c.Length)
	} else {
		num := c.fitNumber(raw, c.Precision)
		str = FitString(num, c.Length) // adds padding if needed
	}
	return []string{str}
}

// Calculates the diff for the given MetricCache, returns an error if there's a data problem.
func (c DiffCol) getDiff(cache *blip.MetricCache) (float64, error) {
	// Get current value
	cur := cache.GetMetricValue(c.Key.Domain, c.Key.Metric)

	// Get previous value
	prev := cache.GetPrevMetricValue(c.Key.Domain, c.Key.Metric)

	// Return the calculated diff
	return calculateDiff(cur, prev), nil
}
