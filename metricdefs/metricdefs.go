package metricdefs

import (
	"errors"
)

type MetricType uint8

const (
	Undefined MetricType = iota
	Gauge
	Counter
	Misc
)

func (t MetricType) MarshalJSON() ([]byte, error) {
	switch t {
	case Undefined:
		return []byte(`"Undefined"`), nil
	case Gauge:
		return []byte(`"Gauge"`), nil
	case Counter:
		return []byte(`"Counter"`), nil
	case Misc:
		return []byte(`"Misc"`), nil
	default:
		return nil, errors.New("Unknown type")
	}
}

func (t *MetricType) UnmarshalJSON(b []byte) error {
	switch string(b) {
	case `"Gauge"`:
		*t = Gauge
	case `"Counter"`:
		*t = Counter
	case `"Misc"`:
		*t = Misc
	default:
		*t = Undefined
		return errors.New("Unknown MetricType")
	}

	return nil
}

type MetricDef struct {
	Header string     `json:header`
	Type   MetricType `json:"type"`
}

func New() map[string]MetricDef {
	return map[string]MetricDef{
		"connections":     MetricDef{"cons", Counter},
		"threads_running": MetricDef{"trun", Gauge},
	}
}
