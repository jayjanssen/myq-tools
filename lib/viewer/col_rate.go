package viewer

import (
	myblip "github.com/jayjanssen/myq-tools/lib/blip"
)

type RateCol struct {
	colNum `yaml:",inline"`
	Key    SourceKey `yaml:"key"`
}

// Data for this view based on the metrics
func (c RateCol) GetData(cache *myblip.MetricCache) []string {
	var str string
	raw, err := c.getRate(cache)
	if err != nil {
		str = FitString(`-`, c.Length)
	} else {
		num := c.fitNumber(raw, c.Precision)
		str = FitString(num, c.Length) // adds padding if needed
	}
	return []string{str}
}

// Calculates the rate for the given MetricCache, returns an error if there's a data problem.
func (c RateCol) getRate(cache *myblip.MetricCache) (float64, error) {
	// Get current value
	cur, ok := cache.GetMetric(c.Key.Domain, c.Key.Metric)
	if !ok {
		return 0, nil // No data available
	}

	// Get previous value (0 if not available)
	var prev float64
	if prevMetric, ok := cache.GetPrevMetric(c.Key.Domain, c.Key.Metric); ok {
		prev = prevMetric.Value
	}

	// Return the calculated rate
	return calculateRate(cur.Value, prev, cache.SecondsDiff()), nil
}
