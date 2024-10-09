package viewer

import (
	"github.com/jayjanssen/myq-tools/lib/loader"
)

type SwitchCol struct {
	defaultCol `yaml:",inline"`
	Key        loader.SourceKey  `yaml:"key"`
	Cases      map[string]string `yaml:"cases"`
}

// Data for this view based on the state
func (c SwitchCol) GetData(sr loader.StateReader) []string {
	// get cur, or else return an error
	currssp := sr.GetCurrent()

	str, err := currssp.GetString(c.Key)
	if err != nil {
		str = `-`
	}

	if val, ok := c.Cases[str]; ok {
		str = val
	} else {
		// Truncate string if it's too long
		if len(str) > c.Length {
			str = str[0:c.Length]
		}
	}

	return []string{FitString(str, c.Length)}
}
