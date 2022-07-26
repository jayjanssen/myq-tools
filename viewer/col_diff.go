package viewer

import (
	"github.com/jayjanssen/myq-tools2/loader"
)

type DiffCol struct {
	colNum    `yaml:",inline"`
	oneKeyCol `yaml:",inline"`
}

// Data for this view based on the state
func (c DiffCol) GetData(sr loader.StateReader) []string {
	var str string
	raw, err := c.getDiff(sr)
	if err != nil {
		str = FitString(`-`, c.Length)
	} else {
		num := c.fitNumber(raw, c.Precision)
		str = FitString(num, c.Length) // adds padding if needed
	}
	return []string{str}
}

// Calculates the rate for the given StateReader, returns an error if there's a data problem.
func (c DiffCol) getDiff(sr loader.StateReader) (float64, error) {
	// get cur, or else return an error
	currssp := sr.GetCurrent()
	cur, err := currssp.GetFloat(c.Key)
	if err != nil {
		return 0, err
	}

	// prev will be 0.0 if there is an error fetching it
	var prev float64
	if prevssp := sr.GetPrevious(); prevssp != nil {
		prev = prevssp.GetF(c.Key)
	}

	// Return the calculated rate
	return CalculateDiff(cur, prev), nil
}
