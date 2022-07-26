package viewer

import (
	"github.com/jayjanssen/myq-tools2/loader"
)

type PercentCol struct {
	colNum      `yaml:",inline"`
	Numerator   loader.SourceKey `yaml:"numerator"`
	Denominator loader.SourceKey `yaml:"denominator"`
}

func (c PercentCol) GetSources() []loader.SourceName {
	return []loader.SourceName{
		c.Numerator.SourceName,
		c.Denominator.SourceName,
	}
}

// Data for this view based on the state
func (c PercentCol) GetData(sr loader.StateReader) []string {
	var str string
	raw, err := c.getPercent(sr)
	if err != nil {
		str = FitString(`-`, c.Length)
	} else {
		num := c.fitNumber(raw, c.Precision)
		str = FitString(num, c.Length) // adds padding if needed
	}
	return []string{str}
}

// Calculates the rate for the given StateReader, returns an error if there's a data problem.
func (c PercentCol) getPercent(sr loader.StateReader) (float64, error) {
	// get cur, or else return an error
	currssp := sr.GetCurrent()
	numerator, err := currssp.GetFloat(c.Numerator)
	if err != nil {
		return 0, err
	}
	denominator, err := currssp.GetFloat(c.Denominator)
	if err != nil {
		return 0, err
	}

	// Return the calculated rate
	return (numerator / denominator) * 100, nil
}
