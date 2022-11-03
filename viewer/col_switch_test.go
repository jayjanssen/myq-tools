package viewer

import (
	"reflect"
	"testing"

	"github.com/jayjanssen/myq-tools2/loader"
	"gopkg.in/yaml.v3"
)

func getTestSwitchCol() SwitchCol {
	sk := loader.SourceKey{SourceName: "status", Key: "wsrep_local_state_comment"}
	col := SwitchCol{}
	col.Name = "state"
	col.Description = "State of this node"
	col.Type = "Switch"
	col.Key = sk
	col.Length = 4
	col.Cases = map[string]string{
		`Joining`:                                `Jing`,
		`Joining: preparing for State Transfer`:  `J:Pr`,
		`Joining: requested State Transfer`:      `J:Rq`,
		`Joining: receiving State Transfer`:      `J:Rc`,
		`Joining: State Transfer request failed`: `J:RF`,
		`Joining: State Transfer failed`:         `J:F`,
		`Joined`:                                 `Jned`,
	}

	return col
}

func TestSwitchCol(t *testing.T) {
	col := getTestSwitchCol()
	if col.Name != "state" {
		t.Errorf("Unexpected col name (state): %s", col.Name)
	}
}

func TestSwitchColImplementsStateViewer(t *testing.T) {
	var _ StateViewer = getTestSwitchCol()
}

func TestSwitchColParse(t *testing.T) {
	yaml_str := `---
- name: state
  description: State of this node
  type: Switch
  key: status/wsrep_local_state_comment
  length: 4
  cases:
    Joining: Jing
    'Joining: preparing for State Transfer': 'J:Pr'
    'Joining: requested State Transfer': 'J:Rq'
    'Joining: receiving State Transfer': 'J:Rc'
    'Joining: State Transfer request failed': 'J:RF'
    'Joining: State Transfer failed': 'J:F'
    Joined: Jned
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

	if col.GetShortHelp() != `state: State of this node` {
		t.Errorf("bad description: '%s'", cols[0].GetShortHelp())
	}

	rc := getTestSwitchCol()
	if !reflect.DeepEqual(rc, col) {
		t.Error(`cols not matching`)
		t.Logf("rc: %+v", rc)
		t.Logf("rc: %+v", col)
	}
}

// Create a state reader to test with
func getTestSwitchState(con_cur string) loader.StateReader {
	sp := loader.NewState()
	cursamp := loader.NewSample()
	sp.GetCurrentWriter().SetSample(`status`, cursamp)

	cursamp.Data[`wsrep_local_state_comment`] = con_cur

	return sp
}

func TestSwitchColGetData(t *testing.T) {
	col := getTestSwitchCol()

	state := getTestSwitchState(`Joining`)
	outputs := col.GetData(state)
	if len(outputs) != 1 {
		t.Errorf(`unexpected amount of output strings %d`, len(outputs))
	}
	if outputs[0] != `Jing` {
		t.Errorf(`unexpected GetData(): '%s'`, outputs[0])
	}

	state = getTestSwitchState(`Joining: requested State Transfer`)
	outputs = col.GetData(state)
	if len(outputs) != 1 {
		t.Errorf(`unexpected amount of output strings %d`, len(outputs))
	}
	if outputs[0] != `J:Rq` {
		t.Errorf(`unexpected GetData(): '%s'`, outputs[0])
	}

	state = getTestSwitchState(`Something not in the switch`)
	outputs = col.GetData(state)
	if len(outputs) != 1 {
		t.Errorf(`unexpected amount of output strings %d`, len(outputs))
	}
	if outputs[0] != `Some` {
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
