package viewer

import (
	"reflect"
	"testing"

	"github.com/jayjanssen/myq-tools2/loader"
	"gopkg.in/yaml.v3"
)

func getTestGaugeCol() GaugeCol {
	sk := loader.SourceKey{SourceName: "status", Key: "threads_connect"}
	rc := GaugeCol{}
	rc.Name = "conn"
	rc.Description = "Threads connected"
	rc.Type = "Gauge"
	rc.Key = sk
	rc.Length = 4
	rc.Units = NUMBER
	rc.Precision = 0

	return rc
}

func TestGaugeCol(t *testing.T) {
	rc := getTestRateCol()
	if rc.Name != "cons" {
		t.Errorf("Unexpected col name (test): %s", rc.Name)
	}
}

func TestGaugeColImplementsStateViewer(t *testing.T) {
	var _ StateViewer = getTestGaugeCol()
}

func TestGaugeColParse(t *testing.T) {
	yaml_str := `---
- name: conn
  description: Threads connected
  key: status/threads_connect
  type: Gauge
  units: Number
  length: 4
  precision: 0
`

	var cols StateViewerList
	err := yaml.Unmarshal([]byte(yaml_str), &cols)

	if err != nil {
		t.Error(err)
	}

	if len(cols) != 1 {
		t.Errorf("not enough cols parsed: %d", len(cols))
	}

	col := cols[0]

	if col.GetShortHelp() != `conn: Threads connected` {
		t.Errorf("bad description: %s", cols[0].GetShortHelp())
	}

	rc := getTestGaugeCol()
	if !reflect.DeepEqual(rc, col) {
		t.Error(`cols not matching`)
		t.Logf("rc: %+v", rc)
		t.Logf("rc: %+v", col)
	}
}

// Create a state reader to test with
func getTestGaugeState(con_cur string) loader.StateReader {
	sp := loader.NewState()
	cursamp := loader.NewSample()
	sp.GetCurrentWriter().SetSample(`status`, cursamp)

	cursamp.Data[`threads_connect`] = con_cur

	return sp
}

func TestGaugeColGetData(t *testing.T) {
	col := getTestGaugeCol()

	// Normal gauge
	state := getTestGaugeState(`10`)
	outputs := col.GetData(state)
	if len(outputs) != 1 {
		t.Errorf(`unexpected amount of output strings %d`, len(outputs))
	}
	if outputs[0] != `  10` {
		t.Errorf(`unexpected GetData(): '%s'`, outputs[0])
	}

	// Normal gauge string
	state = getTestGaugeState(`S`)
	outputs = col.GetData(state)
	if len(outputs) != 1 {
		t.Errorf(`unexpected amount of output strings %d`, len(outputs))
	}
	if outputs[0] != `   S` {
		t.Errorf(`unexpected GetData(): '%s'`, outputs[0])
	}

	// Missing key
	col.Key.Key = `notfound`
	outputs = col.GetData(state)
	if len(outputs) != 1 {
		t.Errorf(`unexpected amount of output strings %d`, len(outputs))
	}
	if outputs[0] != `   -` {
		t.Errorf(`unexpected GetData(): '%s'`, outputs[0])
	}

}
