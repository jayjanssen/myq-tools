package viewer

import (
	"reflect"
	"testing"

	"github.com/jayjanssen/myq-tools/lib/loader"
	"gopkg.in/yaml.v3"
)

func getTestStringCol() StringCol {
	sk := loader.SourceKey{SourceName: "status", Key: "wsrep_cluster_status"}
	col := StringCol{}
	col.Name = "P"
	col.Description = "Primary (P) or Non-primary (N)"
	col.Type = "String"
	col.Key = sk
	col.Length = 1

	return col
}

func TestStringCol(t *testing.T) {
	col := getTestStringCol()
	if col.Name != "P" {
		t.Errorf("Unexpected col name (P): %s", col.Name)
	}
}

func TestStringColImplementsViewer(t *testing.T) {
	var _ Viewer = getTestStringCol()
}

func TestStringColParse(t *testing.T) {
	yaml_str := `---
- name: P
  description: 'Primary (P) or Non-primary (N)'
  key: status/wsrep_cluster_status
  type: String
  length: 1
`

	var cols ViewerList
	err := yaml.Unmarshal([]byte(yaml_str), &cols)

	if err != nil {
		t.Error(err)
	}

	if len(cols) != 1 {
		t.Errorf("not enough cols parsed: %d", len(cols))
	}

	col := cols[0]

	if col.GetShortHelp() != `P: Primary (P) or Non-primary (N)` {
		t.Errorf("bad description: '%s'", cols[0].GetShortHelp())
	}

	rc := getTestStringCol()
	if !reflect.DeepEqual(rc, col) {
		t.Error(`cols not matching`)
		t.Logf("rc: %+v", rc)
		t.Logf("rc: %+v", col)
	}
}

// Create a state reader to test with
func getTestStringState(con_cur string) loader.StateReader {
	sp := loader.NewState()
	cursamp := loader.NewSample()
	sp.GetCurrentWriter().SetSample(`status`, cursamp)

	cursamp.Data[`wsrep_cluster_status`] = con_cur

	return sp
}

func TestStringColGetData(t *testing.T) {
	col := getTestStringCol()

	state := getTestStringState(`Primary`)
	outputs := col.GetData(state)
	if len(outputs) != 1 {
		t.Errorf(`unexpected amount of output strings %d`, len(outputs))
	}
	if outputs[0] != `P` {
		t.Errorf(`unexpected GetData(): '%s'`, outputs[0])
	}

	// Missing key
	col.Key.Key = `notfound`
	outputs = col.GetData(state)
	if len(outputs) != 1 {
		t.Errorf(`unexpected amount of output strings %d`, len(outputs))
	}
	if outputs[0] != `-` {
		t.Errorf(`unexpected GetData(): '%s'`, outputs[0])
	}

}

func TestStringColGetDataFromEnd(t *testing.T) {
	col := getTestStringCol()
	col.Fromend = true
	col.Length = 3

	state := getTestStringState(`Primary`)
	outputs := col.GetData(state)
	if len(outputs) != 1 {
		t.Errorf(`unexpected amount of output strings %d`, len(outputs))
	}
	if outputs[0] != `ary` {
		t.Errorf(`unexpected GetData(): '%s'`, outputs[0])
	}
}
