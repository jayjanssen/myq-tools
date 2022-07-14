package viewer

import (
	"reflect"
	"testing"

	"github.com/jayjanssen/myq-tools2/loader"
	"gopkg.in/yaml.v3"
)

func getTestRateCol() RateCol {
	sk := loader.SourceKey{"status", "connections"}
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
