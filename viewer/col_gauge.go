package viewer

import (
	"github.com/jayjanssen/myq-tools2/loader"
)

type GaugeCol struct {
	colNum    `yaml:",inline"`
	oneKeyCol `yaml:",inline"`
}

// Data for this view based on the state
func (c GaugeCol) GetData(sr loader.StateReader) []string {
	// get cur, or else return an error
	currssp := sr.GetCurrent()

	var str string

	// Try parsing a float first, then a string, else report `-`
	if val, err := currssp.GetFloat(c.Key); err == nil {
		str = c.fitNumber(val, c.Precision)
	} else if val, err := currssp.GetString(c.Key); err == nil {
		str = val
	} else {
		str = `-`
	}

	return []string{FitString(str, c.Length)}
}
