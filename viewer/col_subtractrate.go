package viewer

import (
	"github.com/jayjanssen/myq-tools2/loader"
)

type SubtractRateCol struct {
	colNum  `yaml:",inline"`
	Bigger  loader.SourceKey `yaml:"bigger"`
	Smaller loader.SourceKey `yaml:"smaller"`
}

// Data for this view based on the state
func (c SubtractRateCol) GetData(sr loader.StateReader) []string {
	var str string
	raw, err := c.getSubtractDiff(sr)
	if err != nil {
		str = FitString(`-`, c.Length)
	} else {
		num := c.fitNumber(raw, c.Precision)
		str = FitString(num, c.Length) // adds padding if needed
	}
	return []string{str}
}

// Calculates the rate for the given StateReader, returns an error if there's a data problem.
func (c SubtractRateCol) getSubtractDiff(sr loader.StateReader) (float64, error) {
	// get cur, or else return an error
	currssp := sr.GetCurrent()
	bigger, err := currssp.GetFloat(c.Bigger)
	if err != nil {
		return 0, err
	}
	smaller, err := currssp.GetFloat(c.Smaller)
	if err != nil {
		return 0, err
	}

	currDiff := bigger - smaller

	// prev will be 0.0 if there is an error fetching it
	var prevDiff float64
	if prevssp := sr.GetPrevious(); prevssp != nil {
		prevBigger := prevssp.GetF(c.Bigger)
		prevSmaller := prevssp.GetF(c.Smaller)

		prevDiff = prevBigger - prevSmaller
	}

	// Return the calculated rate
	return CalculateRate(currDiff, prevDiff, sr.SecondsDiff()), nil
}
