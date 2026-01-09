package viewer

import (
	"reflect"
	"testing"
	"time"

	"github.com/cashapp/blip"
	myqblip "github.com/jayjanssen/myq-tools/lib/blip"
	"gopkg.in/yaml.v3"
)

func getTestStringCol() StringCol {
	sk, _ := ParseSourceKey("status/wsrep_cluster_status")
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
		t.Logf("col: %+v", col)
	}
}

// Create a metric cache to test with
func getTestStringCache(value string) *myqblip.MetricCache {
	cache := myqblip.NewMetricCache(false)

	if value == "" {
		// Missing metric - don't add it
		return cache
	}

	metricValue := blip.MetricValue{
		Name:  "wsrep_cluster_status",
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

func TestStringColGetData(t *testing.T) {
	col := getTestStringCol()

	cache := getTestStringCache(`Primary`)
	outputs := col.GetData(cache)
	if len(outputs) != 1 {
		t.Errorf(`unexpected amount of output strings %d`, len(outputs))
	}
	if outputs[0] != `P` {
		t.Errorf(`unexpected GetData(): '%s'`, outputs[0])
	}

	// Missing key
	col.Key.Metric = `notfound`
	cache = getTestStringCache(`Primary`)
	outputs = col.GetData(cache)
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

	cache := getTestStringCache(`Primary`)
	outputs := col.GetData(cache)
	if len(outputs) != 1 {
		t.Errorf(`unexpected amount of output strings %d`, len(outputs))
	}
	if outputs[0] != `ary` {
		t.Errorf(`unexpected GetData(): '%s'`, outputs[0])
	}
}
