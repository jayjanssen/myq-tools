package viewer

import (
	"fmt"

	"github.com/jayjanssen/myq-tools/lib/blip"
)

// A GroupCol is a list of (related) cols
type GroupCol struct {
	defaultCol `yaml:",inline"`
	Cols       ViewerList `yaml:"cols"`
}

// A list of source keys that this group requires (aggregates from all its columns)
func (gc GroupCol) GetRequiredMetrics() []SourceKey {
	var keys []SourceKey
	for _, col := range gc.Cols {
		keys = append(keys, col.GetRequiredMetrics()...)
	}
	return keys
}

// A map of domain to list of metric names (aggregates from all columns)
func (gc GroupCol) GetMetricsByDomain() map[string][]string {
	result := make(map[string]map[string]bool)

	for _, col := range gc.Cols {
		colMetrics := col.GetMetricsByDomain()
		for domain, metrics := range colMetrics {
			if result[domain] == nil {
				result[domain] = make(map[string]bool)
			}
			for _, metric := range metrics {
				result[domain][metric] = true
			}
		}
	}

	// Convert to map of domain -> []string
	finalResult := make(map[string][]string)
	for domain, metricsMap := range result {
		metrics := make([]string, 0, len(metricsMap))
		for metric := range metricsMap {
			metrics = append(metrics, metric)
		}
		finalResult[domain] = metrics
	}

	return finalResult
}

// Get help for this view
func (gc GroupCol) GetDetailedHelp() (output []string) {
	// Gather and indent the lines
	output = append(output, gc.GetShortHelp())
	for _, col := range gc.Cols {
		for _, line := range col.GetDetailedHelp() {
			output = append(output, fmt.Sprintf("   %s", line))
		}
	}
	return
}

// Header for this Group, the name of the Group is first, then the headers of each individual col
func (gc GroupCol) GetHeader(cache *blip.MetricCache) (result []string) {
	getColOut := func(sv Viewer) []string {
		return sv.GetHeader(cache)
	}
	colOuts := pushColOutputDown(gc.Cols, getColOut)

	// Determine the length of this Group by the first line of output from the Cols
	if gc.Length == 0 && len(colOuts) > 0 {
		gc.Length = len(colOuts[0])
	}
	result = append(result, fitStringLeft(gc.Name, gc.Length))
	result = append(result, colOuts...)
	return
}

func (gc GroupCol) GetData(cache *blip.MetricCache) []string {
	getColOut := func(sv Viewer) []string {
		return sv.GetData(cache)
	}
	return pushColOutputUp(gc.Cols, getColOut)
}
