package viewer

import (
	myblip "github.com/jayjanssen/myq-tools/lib/blip"
)

type GaugeCol struct {
	colNum `yaml:",inline"`
	Key    SourceKey `yaml:"key"`
}

// Data for this view based on the metrics
func (c GaugeCol) GetData(cache *myblip.MetricCache) []string {
	var str string

	// Try getting the metric value
	if metric, ok := cache.GetMetric(c.Key.Domain, c.Key.Metric); ok {
		str = c.fitNumber(metric.Value, c.Precision)
	} else {
		str = `-`
	}

	return []string{FitString(str, c.Length)}
}
