package viewer

import (
	"fmt"

	myblip "github.com/jayjanssen/myq-tools/lib/blip"
)

// A GroupCol is a list of (related) cols
type GroupCol struct {
	defaultCol `yaml:",inline"`
	Cols       ViewerList `yaml:"cols"`
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
func (gc GroupCol) GetHeader(cache *myblip.MetricCache) (result []string) {
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

func (gc GroupCol) GetData(cache *myblip.MetricCache) []string {
	getColOut := func(sv Viewer) []string {
		return sv.GetData(cache)
	}
	return pushColOutputUp(gc.Cols, getColOut)
}
