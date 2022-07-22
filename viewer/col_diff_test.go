package viewer

import (
	"reflect"
	"testing"

	"github.com/jayjanssen/myq-tools2/loader"
	"gopkg.in/yaml.v3"
)

func getTestDiffCol() DiffCol {
	sk := loader.SourceKey{SourceName: "status", Key: "bytes_received"}
	rc := DiffCol{}
	rc.Name = "recv"
	rc.Description = "Total since last sample"
	rc.Type = "Diff"
	rc.Key = sk
	rc.Length = 6
	rc.Units = MEMORY
	rc.Precision = 0

	return rc
}

func TestDiffCol(t *testing.T) {
	rc := getTestDiffCol()
	if rc.Name != "recv" {
		t.Errorf("Unexpected col name (recv): %s", rc.Name)
	}
}

func TestDiffColImplementsStateViewer(t *testing.T) {
	var _ StateViewer = getTestDiffCol()
}

func TestDiffColParse(t *testing.T) {
	yaml_str := `---
- name: recv
  description: Total since last sample
  type: Diff
  key: status/bytes_received
  units: Memory
  length: 6
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

	if col.GetShortHelp() != `recv: Total since last sample` {
		t.Errorf("bad description: %s", cols[0].GetShortHelp())
	}

	rc := getTestDiffCol()
	if !reflect.DeepEqual(rc, col) {
		t.Error(`cols not matching`)
		t.Logf("rc: %+v", rc)
		t.Logf("rc: %+v", col)
	}
}

// Create a state reader to test with
func getTestDiffState(con_prev, con_cur string) loader.StateReader {
	sp := loader.NewState()
	prevss := loader.NewSampleSet()

	cursamp := loader.NewSample()
	cursamp.Data[`bytes_received`] = con_cur

	sp.GetCurrentWriter().SetSample(`status`, cursamp)

	prevsamp := loader.NewSample()
	prevss.SetSample(`status`, prevsamp)
	sp.SetPrevious(prevss)

	prevsamp.Data[`bytes_received`] = con_prev

	return sp
}

func TestDiffColgetDiff(t *testing.T) {
	col := getTestDiffCol()

	// Normal Diff
	state := getTestDiffState(`19095555804`, `19096078726`)
	Diff, err := col.getDiff(state)
	if err != nil {
		t.Error(err)
	}
	if Diff != 522922 {
		t.Fatalf(`unexpected Diff: %f`, Diff)
	}
	outputs := col.GetData(state)
	if len(outputs) != 1 {
		t.Errorf(`unexpected amount of output strings %d`, len(outputs))
	}
	if outputs[0] != `510.7K` {
		t.Errorf(`unexpected GetData(): '%s'`, outputs[0])
	}

	// Blank prev Diff
	state = getTestDiffState(``, `19096078726`)
	Diff, err = col.getDiff(state)
	if err != nil {
		t.Error(err)
	}
	if Diff != 19096078726 {
		t.Errorf(`unexpected Diff: %f`, Diff)
	}

	// Bad value
	state = getTestDiffState(``, `notanumber`)
	_, err = col.getDiff(state)
	if err == nil {
		t.Error(`expected error parsing notanumber`)
	}
	outputs = col.GetData(state)
	if len(outputs) != 1 {
		t.Errorf(`unexpected amount of output strings %d`, len(outputs))
	}
	if len(outputs[0]) != 6 {
		t.Errorf(`output should be 4: %d`, len(outputs[0]))
	}
	if outputs[0] != `     -` {
		t.Errorf(`unexpected GetData(): '%s'`, outputs[0])
	}

}
