package viewer

import (
	"reflect"
	"testing"

	"github.com/jayjanssen/myq-tools2/loader"
	"gopkg.in/yaml.v3"
)

func getTestRateCol() RateCol {
	sk := loader.SourceKey{SourceName: "status", Key: "connections"}
	rc := RateCol{}
	rc.Name = "cons"
	rc.Description = "Connections per second"
	rc.Type = "Rate"
	rc.Key = sk
	rc.Length = 4
	rc.Units = NUMBER
	rc.Precision = 0

	return rc
}

func TestRateCol(t *testing.T) {
	rc := getTestRateCol()
	if rc.Name != "cons" {
		t.Errorf("Unexpected col name (test): %s", rc.Name)
	}
}

func TestRateColImplementsStateViewer(t *testing.T) {
	var _ StateViewer = getTestRateCol()
}

func TestRateColParse(t *testing.T) {
	yaml_str := `---
- name: cons
  description: Connections per second
  key: status/connections
  type: Rate
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

	if col.GetShortHelp() != `cons: Connections per second` {
		t.Errorf("bad description: %s", cols[0].GetShortHelp())
	}

	rc := getTestRateCol()
	if !reflect.DeepEqual(rc, col) {
		t.Error(`cols not matching`)
		t.Logf("rc: %+v", rc)
		t.Logf("rc: %+v", col)
	}
}

// Create a state reader to test with
func getTestRateState(con_prev, con_cur string) loader.StateReader {
	sp := loader.NewState()
	curss := loader.NewSampleSet()
	prevss := loader.NewSampleSet()

	cursamp := loader.NewSample()
	curss.SetSample(`status`, cursamp)

	prevsamp := loader.NewSample()
	prevss.SetSample(`status`, prevsamp)

	sp.SetCurrent(curss)
	sp.SetPrevious(prevss)

	cursamp.Data[`connections`] = con_cur
	prevsamp.Data[`connections`] = con_prev

	return sp
}

func TestRateColgetRate(t *testing.T) {
	col := getTestRateCol()

	// Normal rate
	state := getTestRateState(`10`, `15`)
	rate, err := col.getRate(state)
	if err != nil {
		t.Error(err)
	}
	if rate != 5 {
		t.Errorf(`unexpected rate: %f`, rate)
	}
	outputs := col.GetData(state)
	if len(outputs) != 1 {
		t.Errorf(`unexpected amount of output strings %d`, len(outputs))
	}
	if outputs[0] != `   5` {
		t.Errorf(`unexpected GetData(): '%s'`, outputs[0])
	}

	// Blank prev rate
	state = getTestRateState(``, `15`)
	rate, err = col.getRate(state)
	if err != nil {
		t.Error(err)
	}
	if rate != 15 {
		t.Errorf(`unexpected rate: %f`, rate)
	}

	// Bad value
	state = getTestRateState(``, `notanumber`)
	_, err = col.getRate(state)
	if err == nil {
		t.Error(`expected error parsing notanumber`)
	}
	outputs = col.GetData(state)
	if len(outputs) != 1 {
		t.Errorf(`unexpected amount of output strings %d`, len(outputs))
	}
	if len(outputs[0]) != 4 {
		t.Errorf(`output should be 4: %d`, len(outputs[0]))
	}
	if outputs[0] != `   -` {
		t.Errorf(`unexpected GetData(): '%s'`, outputs[0])
	}

}

func TestRateColBadSourceKey(t *testing.T) {
	// the key value is incorrect, it should be <source>/<key>
	yaml_str := `---
- name: acls
  description: Aborted clients (existing connections)
  source: status
  key: aborted_clients
  type: Rate
  units: Number
  length: 4
  precision: 0
`

	var cols StateViewerList
	err := yaml.Unmarshal([]byte(yaml_str), &cols)

	if err == nil {
		t.Error(`expected error parsing bad sourcekey`)
	}
}
