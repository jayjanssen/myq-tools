package viewer

import (
	"github.com/jayjanssen/myq-tools2/loader"
)

type SubtractCol struct {
	colNum  `yaml:",inline"`
	Bigger  loader.SourceKey `yaml:"bigger"`
	Smaller loader.SourceKey `yaml:"smaller"`
}

// Data for this view based on the state
func (c SubtractCol) GetData(sr loader.StateReader) []string {
	var str string
	raw, err := c.getSubtract(sr)
	if err != nil {
		str = FitString(`-`, c.Length)
	} else {
		num := c.fitNumber(raw, c.Precision)
		str = FitString(num, c.Length) // adds padding if needed
	}
	return []string{str}
}

// Calculates the rate for the given StateReader, returns an error if there's a data problem.
func (c SubtractCol) getSubtract(sr loader.StateReader) (float64, error) {
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

	// Return the calculated rate
	return (bigger - smaller), nil
}
