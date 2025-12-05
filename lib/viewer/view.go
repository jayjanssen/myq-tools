package viewer

import (
	"fmt"

	myblip "github.com/jayjanssen/myq-tools/lib/blip"
)

// A view is made up of Groups of Cols
type View struct {
	// Views *are* a GroupCol in that they can have a Cols
	GroupCol `yaml:",inline"`

	// Usually a view would have Groups OR Cols, but not both.  If both, print groups first, then individual cols
	Groups []GroupCol `yaml:"groups"`
}

// How to print out the time with our output
var timeCol SampleTimeCol = NewSampleTimeCol()

// Get help for this view
func (v View) GetDetailedHelp() (output []string) {
	// Gather the svs
	var svs ViewerList
	for _, group := range v.Groups {
		svs = append(svs, group)
	}
	svs = append(svs, v.Cols...)

	// Gather and indent the lines
	output = append(output, v.GetShortHelp())
	for _, sv := range svs {
		for _, line := range sv.GetDetailedHelp() {
			output = append(output, fmt.Sprintf("   %s", line))
		}
	}
	return
}

// A list of domains that this view requires
func (v View) GetDomains() []string {
	domains := make(map[string]bool)

	// Collect domains from groups
	for _, group := range v.Groups {
		for _, d := range group.GetDomains() {
			domains[d] = true
		}
	}

	// Collect domains from cols
	for _, col := range v.Cols {
		for _, d := range col.GetDomains() {
			domains[d] = true
		}
	}

	// Convert to slice
	result := make([]string, 0, len(domains))
	for d := range domains {
		result = append(result, d)
	}
	return result
}

// GetMetricsByDomain returns a map of domain to list of metric names required by this view
func (v View) GetMetricsByDomain() map[string][]string {
	metricsByDomain := make(map[string]map[string]bool)

	// Helper function to add metrics from a source key
	addMetric := func(key SourceKey) {
		// Skip empty keys
		if key.Domain == "" || key.Metric == "" {
			return
		}

		// Check if this is a pattern (contains * or ^)
		// Patterns need special handling - for now we'll need to use "all" mode
		// TODO: handle patterns by expanding them at runtime
		if key.Metric == "" {
			return
		}

		if metricsByDomain[key.Domain] == nil {
			metricsByDomain[key.Domain] = make(map[string]bool)
		}
		metricsByDomain[key.Domain][key.Metric] = true
	}

	// Collect from groups
	for _, group := range v.Groups {
		for _, metricKey := range group.GetRequiredMetrics() {
			addMetric(metricKey)
		}
	}

	// Collect from cols
	for _, col := range v.Cols {
		for _, metricKey := range col.GetRequiredMetrics() {
			addMetric(metricKey)
		}
	}

	// Convert to map of domain -> []string
	result := make(map[string][]string)
	for domain, metricsMap := range metricsByDomain {
		metrics := make([]string, 0, len(metricsMap))
		for metric := range metricsMap {
			metrics = append(metrics, metric)
		}
		result[domain] = metrics
	}

	return result
}

// Header for this view
func (v View) GetHeader(cache *myblip.MetricCache) []string {
	// Collect all the Viewers for this view
	var svs ViewerList
	svs = append(svs, timeCol)
	for _, group := range v.Groups {
		svs = append(svs, group)
	}
	svs = append(svs, v.Cols...)

	// Get the header output of all those svs
	colOuts := pushColOutputDown(svs, func(sv Viewer) []string {
		return sv.GetHeader(cache)
	})

	// Get the length of this view based on the length of the first colOut
	if v.Length == 0 && len(colOuts) > 0 {
		v.Length = len(colOuts[0])
	}

	return colOuts
}

// Data for this view based on the metrics
func (v View) GetData(cache *myblip.MetricCache) (result []string) {
	// Collect all the Viewers for this view
	var svs ViewerList
	svs = append(svs, timeCol)
	for _, group := range v.Groups {
		svs = append(svs, group)
	}
	svs = append(svs, v.Cols...)

	// Get the data output of all those svs
	return pushColOutputUp(svs, func(sv Viewer) []string {
		return sv.GetData(cache)
	})
}
