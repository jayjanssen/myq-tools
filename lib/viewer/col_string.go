package viewer

import (
	"github.com/jayjanssen/myq-tools/lib/loader"
)

type StringCol struct {
	defaultCol `yaml:",inline"`
	Key        loader.SourceKey `yaml:"key"`
	Fromend    bool             `yaml:"fromend"`
}

// Data for this view based on the state
func (c StringCol) GetData(sr loader.StateReader) []string {
	// get cur, or else return an error
	currssp := sr.GetCurrent()

	str, err := currssp.GetString(c.Key)
	if err != nil {
		str = `-`
	}

	if len(str) > c.Length {
		// Truncate the string
		if !c.Fromend {
			// First Length chars
			str = str[0:c.Length]
		} else {
			// Last Length chars
			str = str[len(str)-c.Length:]
		}
	}

	return []string{FitString(str, c.Length)}
}
