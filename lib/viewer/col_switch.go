package viewer

import (
	"fmt"

	myblip "github.com/jayjanssen/myq-tools/lib/blip"
)

type SwitchCol struct {
	defaultCol `yaml:",inline"`
	Key        SourceKey         `yaml:"key"`
	Cases      map[string]string `yaml:"cases"`
}

// A list of source keys that this column requires
func (c SwitchCol) GetRequiredMetrics() []SourceKey {
	return []SourceKey{c.Key}
}

// Data for this view based on the metrics
func (c SwitchCol) GetData(cache *myblip.MetricCache) []string {
	// Try to get the metric value as a string
	var str string
	if metric, ok := cache.GetMetric(c.Key.Domain, c.Key.Metric); ok {
		str = fmt.Sprintf("%.0f", metric.Value)
	} else {
		str = `-`
	}

	if val, ok := c.Cases[str]; ok {
		str = val
	} else {
		// Truncate string if it's too long
		if len(str) > c.Length {
			str = str[0:c.Length]
		}
	}

	return []string{FitString(str, c.Length)}
}
