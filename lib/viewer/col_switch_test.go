package viewer

import (
	"reflect"
	"testing"
	"time"

	"github.com/cashapp/blip"
	myqblip "github.com/jayjanssen/myq-tools/lib/blip"
	"gopkg.in/yaml.v3"
)

func getTestSwitchCol() SwitchCol {
	sk, _ := ParseSourceKey("status/wsrep_local_state_comment")
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

func TestSwitchColImplementsViewer(t *testing.T) {
	var _ Viewer = getTestSwitchCol()
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

	var cols ViewerList
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
		t.Logf("col: %+v", col)
	}
}

// Create a metric cache to test with
func getTestSwitchCache(value string) *myqblip.MetricCache {
	cache := myqblip.NewMetricCache(false)

	if value == "" {
		// Missing metric - don't add it
		return cache
	}

	metricValue := blip.MetricValue{
		Name:  "wsrep_local_state_comment",
		Value: 0,
		Type:  blip.GAUGE,
		Meta: map[string]string{
			"string_value": value,
		},
	}

	cache.Update(&blip.Metrics{
		Begin: time.Now(),
		End:   time.Now(),
		Values: map[string][]blip.MetricValue{
			"status.global": {metricValue},
		},
	})

	return cache
}

func TestSwitchColGetData(t *testing.T) {
	col := getTestSwitchCol()

	cache := getTestSwitchCache(`Joining`)
	outputs := col.GetData(cache)
	if len(outputs) != 1 {
		t.Errorf(`unexpected amount of output strings %d`, len(outputs))
	}
	if outputs[0] != `Jing` {
		t.Errorf(`unexpected GetData(): '%s'`, outputs[0])
	}

	cache = getTestSwitchCache(`Joining: requested State Transfer`)
	outputs = col.GetData(cache)
	if len(outputs) != 1 {
		t.Errorf(`unexpected amount of output strings %d`, len(outputs))
	}
	if outputs[0] != `J:Rq` {
		t.Errorf(`unexpected GetData(): '%s'`, outputs[0])
	}

	cache = getTestSwitchCache(`Something not in the switch`)
	outputs = col.GetData(cache)
	if len(outputs) != 1 {
		t.Errorf(`unexpected amount of output strings %d`, len(outputs))
	}
	if outputs[0] != `Some` {
		t.Errorf(`unexpected GetData(): '%s'`, outputs[0])
	}

	// Missing key
	col.Key.Metric = `notfound`
	cache = getTestSwitchCache(`Joining`)
	outputs = col.GetData(cache)
	if len(outputs) != 1 {
		t.Errorf(`unexpected amount of output strings %d`, len(outputs))
	}
	if outputs[0] != `   -` {
		t.Errorf(`unexpected GetData(): '%s'`, outputs[0])
	}

}
