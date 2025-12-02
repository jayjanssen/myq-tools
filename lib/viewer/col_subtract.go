package viewer

import (
	myblip "github.com/jayjanssen/myq-tools/lib/blip"
)

type SubtractCol struct {
	colNum  `yaml:",inline"`
	Bigger  SourceKey `yaml:"bigger"`
	Smaller SourceKey `yaml:"smaller"`
}

// Data for this view based on the metrics
func (c SubtractCol) GetData(cache *myblip.MetricCache) []string {
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
func (c SubtractCol) getSubtract(cache *myblip.MetricCache) (float64, error) {
	// Get values
	bigger := cache.GetMetricValue(c.Bigger.Domain, c.Bigger.Metric)
	smaller := cache.GetMetricValue(c.Smaller.Domain, c.Smaller.Metric)

	// Return the calculated subtraction
	return (bigger - smaller), nil
}
