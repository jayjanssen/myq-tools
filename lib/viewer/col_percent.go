package viewer

import (
	"github.com/jayjanssen/myq-tools/lib/blip"
)

type PercentCol struct {
	colNum      `yaml:",inline"`
	Numerator   SourceKey `yaml:"numerator"`
	Denominator SourceKey `yaml:"denominator"`
}

// A list of source keys that this column requires
func (c PercentCol) GetRequiredMetrics() []SourceKey {
	return []SourceKey{c.Numerator, c.Denominator}
}

// Data for this view based on the metrics
func (c PercentCol) GetData(cache *blip.MetricCache) []string {
	var str string
	raw, err := c.getPercent(cache)
	if err != nil {
		str = FitString(`-`, c.Length)
	} else {
		num := c.fitNumber(raw, c.Precision)
		str = FitString(num, c.Length) // adds padding if needed
	}
	return []string{str}
}

// Calculates the percentage for the given MetricCache, returns an error if there's a data problem.
func (c PercentCol) getPercent(cache *blip.MetricCache) (float64, error) {
	// Get numerator
	numerator := cache.GetMetricValue(c.Numerator.Domain, c.Numerator.Metric)

	// Get denominator
	denominator := cache.GetMetricValue(c.Denominator.Domain, c.Denominator.Metric)

	if denominator == 0 {
		return 0, nil
	}

	// Return the calculated percentage
	return (numerator / denominator) * 100, nil
}
