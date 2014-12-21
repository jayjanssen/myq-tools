package metricdefs

import (
    "testing"
    "encoding/json"
)

func TestJSON(t *testing.T) {
	metrics := map[string]MetricDef{
		"connections" : MetricDef{"cons", Counter},
		"threads_running" : MetricDef{"trun", Gauge},
	}
	// MetricDef( 'Connections', counter ),

	res, _ := json.MarshalIndent(metrics, "", "  ")
	// res, _ := json.Marshal(metrics)
	t.Log(string(res))

	var newmetrics map[string]MetricDef
	json.Unmarshal(res, &newmetrics)

	nres, _ := json.MarshalIndent(newmetrics, "", "  ")
	t.Log(string(nres))

    if string(res) != string(nres) {
        t.Fail()
    }

	return
}

func TestMarshallUndef(t *testing.T) {
  test := MetricDef{"cons", 255}
  res, _ := json.Marshal( test )
  if res != nil {
    t.Log(string(res))
    t.Fail()
  }
}

func TestUnMarshallUndef(t *testing.T) {
  var test MetricDef
  json.Unmarshal( []byte(`{"Header":"cons","type":"Fooey"}`), &test )
  if test.Type != Undefined {
    t.Fail()
  }
}