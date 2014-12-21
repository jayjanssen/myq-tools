package metrics

import (
  "../metricdefs"
)

type Metric struct {
  Value interface{}
  Def *metricdefs.MetricDef
}

// type Metric interface {
//   Get() interface{}
//   Set( interface{} )
// }

// Int Metrics
type MetricInt struct {
  Value int64
  Def *metricdefs.MetricDef
}

func NewMetricInt(val int64, def *metricdefs.MetricDef) MetricInt {
  return MetricInt{ val, def }
}

func (t *MetricInt) Get() int64 {
  return t.Value
}

func (t *MetricInt) Set(v int64) {
  t.Value = v
}

// type MetricFloat float64