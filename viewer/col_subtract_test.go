package viewer

import (
	"reflect"
	"testing"

	"github.com/jayjanssen/myq-tools2/loader"
	"gopkg.in/yaml.v3"
)

func getTestSubtractCol() SubtractCol {
	bk := loader.SourceKey{SourceName: "status", Key: "wsrep_last_committed"}
	sk := loader.SourceKey{SourceName: "status", Key: "wsrep_local_cached_downto"}
	col := SubtractCol{}
	col.Name = "ist"
	col.Description = "Gcached transactions"
	col.Type = "Subtract"
	col.Bigger = bk
	col.Smaller = sk
	col.Units = NUMBER
	col.Length = 5
	col.Precision = 0

	return col
}

func TestSubtractCol(t *testing.T) {
	col := getTestSubtractCol()
	if col.Name != "ist" {
		t.Errorf("Unexpected col name (ist): %s", col.Name)
	}
}

func TestSubtractColImplementsStateViewer(t *testing.T) {
	var _ StateViewer = getTestSubtractCol()
}

func TestSubtractColParse(t *testing.T) {
	yaml_str := `---
- name: ist
  description: Gcached transactions
  type: Subtract
  bigger: status/wsrep_last_committed
  smaller: status/wsrep_local_cached_downto
  units: Number
  length: 5
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

	if col.GetShortHelp() != `ist: Gcached transactions` {
		t.Errorf("bad description: '%s'", cols[0].GetShortHelp())
	}

	rc := getTestSubtractCol()
	if !reflect.DeepEqual(rc, col) {
		t.Error(`cols not matching`)
		t.Logf("rc: %+v", rc)
		t.Logf("rc: %+v", col)
	}
}

// Create a state reader to test with
func getTestSubtractState(bigger, smaller string) loader.StateReader {
	sp := loader.NewState()
	cursamp := loader.NewSample()
	sp.GetCurrentWriter().SetSample(`status`, cursamp)

	cursamp.Data[`wsrep_last_committed`] = bigger
	cursamp.Data[`wsrep_local_cached_downto`] = smaller

	return sp
}

func TestSubtractColGetData(t *testing.T) {
	col := getTestSubtractCol()

	state := getTestSubtractState(`2`, `1`)
	outputs := col.GetData(state)
	if len(outputs) != 1 {
		t.Errorf(`unexpected amount of output strings %d`, len(outputs))
	}
	if outputs[0] != `    1` {
		t.Errorf(`unexpected GetData(): '%s'`, outputs[0])
	}

	// Negative
	state = getTestSubtractState(`1`, `2`)
	outputs = col.GetData(state)
	if len(outputs) != 1 {
		t.Errorf(`unexpected amount of output strings %d`, len(outputs))
	}
	if outputs[0] != `#####` {
		t.Errorf(`unexpected GetData(): '%s'`, outputs[0])
	}

	// Missing key
	col.Smaller.Key = `notfound`
	outputs = col.GetData(state)
	if len(outputs) != 1 {
		t.Errorf(`unexpected amount of output strings %d`, len(outputs))
	}
	if outputs[0] != `    -` {
		t.Errorf(`unexpected GetData(): '%s'`, outputs[0])
	}
	col.Bigger.Key = `notfound`
	outputs = col.GetData(state)
	if len(outputs) != 1 {
		t.Errorf(`unexpected amount of output strings %d`, len(outputs))
	}
	if outputs[0] != `    -` {
		t.Errorf(`unexpected GetData(): '%s'`, outputs[0])
	}

}
