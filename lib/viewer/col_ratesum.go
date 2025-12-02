package viewer

import (
	"fmt"

	myblip "github.com/jayjanssen/myq-tools/lib/blip"
)

type RateSumCol struct {
	colNum       `yaml:",inline"`
	Keys         []SourceKey `yaml:"keys"`
	expandedKeys []SourceKey
}

func (rsc RateSumCol) GetData(cache *myblip.MetricCache) []string {
	var str string
	raw, err := rsc.getRate(cache)
	if err != nil {
		str = FitString(`-`, rsc.Length)
	} else {
		num := rsc.fitNumber(raw, rsc.Precision)
		str = FitString(num, rsc.Length) // adds padding if needed
	}
	return []string{str}
}

func (rsc RateSumCol) getRate(cache *myblip.MetricCache) (float64, error) {
	// Calculate expanded Keys once if they contain patterns
	// For now, just use the keys as-is (pattern expansion can be added later)
	if len(rsc.expandedKeys) == 0 {
		rsc.expandedKeys = rsc.Keys
	}

	if len(rsc.expandedKeys) == 0 {
		return 0, fmt.Errorf(`no keys found: %s`, rsc.Name)
	}

	// Sum current values
	var curSum float64
	for _, key := range rsc.expandedKeys {
		curSum += cache.GetMetricValue(key.Domain, key.Metric)
	}

	// Sum previous values
	var prevSum float64
	for _, key := range rsc.expandedKeys {
		prevSum += cache.GetPrevMetricValue(key.Domain, key.Metric)
	}

	// Return the calculated rate
	return calculateRate(curSum, prevSum, cache.SecondsDiff()), nil
}
