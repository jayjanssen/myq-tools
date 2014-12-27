package myqlib

import (
	"errors"
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

type MySQLMetricDef struct {
	Header string     `json:header`
	Type   MetricType `json:"type"`
}

func DefaultMetricDefs() map[string]MySQLMetricDef {
	return map[string]MySQLMetricDef{
		"connections":     MySQLMetricDef{"cons", Counter},
		"threads_running": MySQLMetricDef{"trun", Gauge},
	}
}
