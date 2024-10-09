package viewer

import (
	"fmt"

	"github.com/jayjanssen/myq-tools/lib/loader"
)

type RateSumCol struct {
	colNum       `yaml:",inline"`
	Keys         []loader.SourceKey `yaml:"keys"`
	expandedKeys []loader.SourceKey
}

func (rsc RateSumCol) GetData(sr loader.StateReader) []string {
	var str string
	raw, err := rsc.getRate(sr)
	if err != nil {
		str = FitString(`-`, rsc.Length)
	} else {
		num := rsc.fitNumber(raw, rsc.Precision)
		str = FitString(num, rsc.Length) // adds padding if needed
	}
	return []string{str}
}

func (rsc RateSumCol) getRate(sr loader.StateReader) (float64, error) {
	// Calculate expanded Keys once, because it's expensive
	if len(rsc.expandedKeys) == 0 {
		rsc.expandedKeys = sr.GetCurrent().ExpandSourceKeys(rsc.Keys)
	}

	if len(rsc.expandedKeys) == 0 {
		return 0, fmt.Errorf(`no keys found: %s`, rsc.Name)
	}

	// get cur, or else return an error
	curSum := sr.GetCurrent().GetFloatSum(rsc.expandedKeys)

	// prev will be 0.0 if there is an error fetching it
	var prevSum float64
	if prevssp := sr.GetPrevious(); prevssp != nil {
		prevSum = prevssp.GetFloatSum(rsc.expandedKeys)
	}

	// Return the calculated rate
	return calculateRate(curSum, prevSum, sr.SecondsDiff()), nil
}
