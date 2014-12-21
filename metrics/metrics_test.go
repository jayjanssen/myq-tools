package metrics

import (
    "testing"
    "../metricdefs"
)

func TestMetric(t *testing.T) {
  def := metricdefs.MetricDef{"cons", metricdefs.Counter}
  test := NewMetricInt( 64, &def )

  t.Log(test.Get())
  if test.Get() != 64 {
    t.Fail()
  }
  
  test.Set(128)
  if test.Get() != 128 {
    t.Fail()
  }  
}