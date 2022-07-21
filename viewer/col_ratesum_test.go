package viewer

import (
	"reflect"
	"testing"

	"github.com/jayjanssen/myq-tools2/loader"
	"gopkg.in/yaml.v3"
)

func getTestRateSumCol() RateSumCol {
	sk := loader.SourceKey{SourceName: "status", Key: "com_set.*"}
	rc := RateSumCol{}
	rc.Name = "set"
	rc.Description = "SET commands per second"
	rc.Type = "RateSum"
	rc.Keys = []loader.SourceKey{sk}
	rc.Length = 5
	rc.Units = NUMBER
	rc.Precision = 0

	return rc
}

func TestRateSumCol(t *testing.T) {
	rc := getTestRateSumCol()
	if rc.Name != "set" {
		t.Errorf("Unexpected col name (set): %s", rc.Name)
	}
}

func TestRateSumColImplementsStateViewer(t *testing.T) {
	var _ StateViewer = getTestRateSumCol()
}

func TestRateSumColParse(t *testing.T) {
	yaml_str := `---
- name: set
  description: SET commands per second
  keys: 
    - status/com_set.*
  type: RateSum
  units: Number
  length: 5
  precision: 0
`

	var cols StateViewerList
	err := yaml.Unmarshal([]byte(yaml_str), &cols)

	if err != nil {
		t.Fatal(err)
	}

	if len(cols) != 1 {
		t.Fatalf("not enough cols parsed: %d", len(cols))
	}

	col := cols[0]

	if col.GetShortHelp() != `set: SET commands per second` {
		t.Errorf("bad description: %s", cols[0].GetShortHelp())
	}

	rc := getTestRateSumCol()
	if !reflect.DeepEqual(rc, col) {
		t.Error(`cols not matching`)
		t.Logf("rc: %+v", rc)
		t.Logf("col: %+v", col)
	}
}

// Create a state reader to test with
func getTestRateSumState(con_prev, con_cur string) loader.StateReader {
	sp := loader.NewState()
	prevss := loader.NewSampleSet()

	cursamp := loader.NewSample()
	cursamp.Data[`com_set_option`] = con_cur
	cursamp.Data[`com_set_password`] = con_cur
	cursamp.Data[`com_set_resource_group`] = con_cur

	sp.GetCurrentWriter().SetSample(`status`, cursamp)

	prevsamp := loader.NewSample()
	prevss.SetSample(`status`, prevsamp)
	sp.SetPrevious(prevss)

	prevsamp.Data[`com_set_option`] = con_prev
	prevsamp.Data[`com_set_password`] = con_prev
	prevsamp.Data[`com_set_resource_group`] = con_prev

	return sp
}

func TestRateSumColgetRate(t *testing.T) {
	col := getTestRateSumCol()

	// Normal rate
	state := getTestRateSumState(`10`, `15`)
	rate, err := col.getRate(state)
	if err != nil {
		t.Error(err)
	}
	if rate != 15 {
		t.Fatalf(`unexpected rate: %f`, rate)
	}
	outputs := col.GetData(state)
	if len(outputs) != 1 {
		t.Errorf(`unexpected amount of output strings %d`, len(outputs))
	}
	if outputs[0] != `   15` {
		t.Errorf(`unexpected GetData(): '%s'`, outputs[0])
	}

	// Blank prev rate
	state = getTestRateSumState(``, `15`)
	rate, err = col.getRate(state)
	if err != nil {
		t.Error(err)
	}
	if rate != 45 {
		t.Errorf(`unexpected rate: %f`, rate)
	}

	// Bad value
	state = getTestRateSumState(``, `notanumber`)
	_, err = col.getRate(state)
	if err != nil {
		t.Fatalf(`unexpected error parsing notanumber: %s`, err)
	}
	outputs = col.GetData(state)
	if len(outputs) != 1 {
		t.Errorf(`unexpected amount of output strings %d`, len(outputs))
	}
	if len(outputs[0]) != 5 {
		t.Errorf(`output should be 5: %d`, len(outputs[0]))
	}
	if outputs[0] != `    0` {
		t.Errorf(`unexpected GetData(): '%s'`, outputs[0])
	}

}

func TestRateSumColgetRateNoMatches(t *testing.T) {
	// State with no matching keys
	sp := loader.NewState()
	cursamp := loader.NewSample()
	cursamp.Data[`com_something`] = `10`
	cursamp.Data[`connections`] = `12`
	sp.GetCurrentWriter().SetSample(`status`, cursamp)

	col := getTestRateSumCol()
	_, err := col.getRate(sp)
	if err == nil {
		t.Fatalf(`expected error parsing no matches`)
	}
	outputs := col.GetData(sp)
	if len(outputs) != 1 {
		t.Errorf(`unexpected amount of output strings %d`, len(outputs))
	}
	if len(outputs[0]) != 5 {
		t.Errorf(`output should be 5: %d`, len(outputs[0]))
	}
	if outputs[0] != `    -` {
		t.Errorf(`unexpected GetData(): '%s'`, outputs[0])
	}

}
