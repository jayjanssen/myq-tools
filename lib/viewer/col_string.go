package viewer

import (
	"fmt"

	myblip "github.com/jayjanssen/myq-tools/lib/blip"
)

type StringCol struct {
	defaultCol `yaml:",inline"`
	Key        SourceKey `yaml:"key"`
	Fromend    bool      `yaml:"fromend"`
}

// Data for this view based on the metrics
func (c StringCol) GetData(cache *myblip.MetricCache) []string {
	// Try to get string representation of the metric
	var str string
	if metric, ok := cache.GetMetric(c.Key.Domain, c.Key.Metric); ok {
		str = fmt.Sprintf("%.0f", metric.Value) // Convert numeric to string
	} else {
		str = `-`
	}

	if len(str) > c.Length {
		// Truncate the string
		if !c.Fromend {
			// First Length chars
			str = str[0:c.Length]
		} else {
			// Last Length chars
			str = str[len(str)-c.Length:]
		}
	}

	return []string{FitString(str, c.Length)}
}
