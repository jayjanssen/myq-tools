package viewer

import (
	"fmt"
	"sort"

	"github.com/jayjanssen/myq-tools/lib/loader"
)

type SortedExpandedCountsCol struct {
	colNum       `yaml:",inline"`
	Keys         []loader.SourceKey `yaml:"keys"`
	expandedKeys []loader.SourceKey
}

func (secc SortedExpandedCountsCol) GetData(sr loader.StateReader) (output []string) {
	// Calculate expanded Keys once, because it's expensive
	if len(secc.expandedKeys) == 0 {
		secc.expandedKeys = sr.GetCurrent().ExpandSourceKeys(secc.Keys)
	}

	if len(secc.expandedKeys) == 0 {
		return []string{}
	}

	// Go through all the expandedKeys and compute their diffs
	var total_diff float64
	var all_diffs []float64
	diff_variables := map[float64][]string{}
	for _, sk := range secc.expandedKeys {
		curr := sr.GetCurrent().GetF(sk)
		// prev will be 0.0 if there is an error fetching it
		var prev float64
		if prevssp := sr.GetPrevious(); prevssp != nil {
			prev = prevssp.GetF(sk)
		}

		diff := CalculateDiff(curr, prev)
		// Skip those with no activity
		if diff <= 0 {
			continue
		}
		total_diff += diff

		// Create the [] slice for a rate we haven't seen yet
		if _, ok := diff_variables[diff]; !ok {
			diff_variables[diff] = make([]string, 0)
			all_diffs = append(all_diffs, diff) // record the diff the first time
		}

		// Push the variable name onto the rate slice
		diff_variables[diff] = append(diff_variables[diff], sk.Key)
	}

	// output the total diff
	numStr := FitString(secc.fitNumber(total_diff, 0), secc.Length)
	line := fmt.Sprintf("%s %v", numStr, "total")
	output = append(output, line)

	// Sort all the rates so we can iterate through them from big to small
	sort.Sort(sort.Reverse(sort.Float64Slice(all_diffs)))

	for _, diff := range all_diffs {
		numStr := FitString(secc.fitNumber(diff, 0), secc.Length)
		line := fmt.Sprintf("%s %v", numStr, diff_variables[diff])
		output = append(output, line)
	}
	return
}
